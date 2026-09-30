// Package push delivers FCM push notifications. It authenticates against the
// Firebase Cloud Messaging HTTP v1 API using a Google service account (OAuth2
// JWT-bearer flow) and sends messages to individual device tokens.
package push

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	fcmSendURL  = "https://fcm.googleapis.com/v1/projects/%s/messages:send"
	oauthScope  = "https://www.googleapis.com/auth/firebase.messaging"
	oauthURL    = "https://oauth2.googleapis.com/token"
	tokenBuffer = time.Minute
)

// Message is a user-facing notification payload.
type Message struct {
	Title string
	Body  string
	// Data carries optional key/value pairs delivered with the notification
	// (e.g. appointment ID, type). Values are strings only.
	Data map[string]string
}

// serviceAccount is the subset of a Google service-account JSON the sender
// needs.
type serviceAccount struct {
	ProjectID   string `json:"project_id"`
	ClientEmail string `json:"client_email"`
	PrivateKey  string `json:"private_key"`
	TokenURI    string `json:"token_uri"`
}

// Sender sends FCM messages authenticated with a service account.
type Sender struct {
	projectID   string
	clientEmail string
	tokenURI    string
	privateKey  *rsa.PrivateKey
	http        *http.Client

	mu          sync.Mutex
	accessToken string
	expiresAt   time.Time
}

// NewSender parses a Google service-account JSON document and returns a sender
// for its project. The JSON is what the Firebase console emits from
// Project settings > Service accounts > Generate new private key.
func NewSender(serviceAccountJSON []byte) (*Sender, error) {
	var sa serviceAccount
	if err := json.Unmarshal(serviceAccountJSON, &sa); err != nil {
		return nil, fmt.Errorf("parse service account: %w", err)
	}
	if sa.ProjectID == "" {
		return nil, fmt.Errorf("service account is missing project_id")
	}
	if sa.ClientEmail == "" {
		return nil, fmt.Errorf("service account is missing client_email")
	}

	key, err := parsePrivateKey(sa.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("parse service account private key: %w", err)
	}

	tokenURI := strings.TrimSpace(sa.TokenURI)
	if tokenURI == "" {
		tokenURI = oauthURL
	}

	return &Sender{
		projectID:   sa.ProjectID,
		clientEmail: sa.ClientEmail,
		tokenURI:    tokenURI,
		privateKey:  key,
		http:        &http.Client{Timeout: 15 * time.Second},
	}, nil
}

// Send delivers a message to a single device token.
func (s *Sender) Send(ctx context.Context, token string, msg Message) error {
	if token == "" {
		return fmt.Errorf("push: empty token")
	}

	body, err := json.Marshal(map[string]any{
		"message": messagePayload(token, msg),
	})
	if err != nil {
		return fmt.Errorf("push: marshal message: %w", err)
	}

	return s.do(ctx, fmt.Sprintf(fcmSendURL, s.projectID), body)
}

// SendMany delivers a message to many device tokens. It attempts every token
// and aggregates failures so a single bad token does not prevent the rest.
func (s *Sender) SendMany(ctx context.Context, tokens []string, msg Message) error {
	var errs []error
	for _, token := range tokens {
		if err := s.Send(ctx, token, msg); err != nil {
			errs = append(errs, fmt.Errorf("token: %w", err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("push: %d of %d sends failed: %v", len(errs), len(tokens), errs[0])
	}
	return nil
}

func (s *Sender) do(ctx context.Context, url string, body []byte) error {
	token, err := s.accessTokenFor(ctx)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("push: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.http.Do(req)
	if err != nil {
		return fmt.Errorf("push: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		io.Copy(io.Discard, resp.Body)
		return nil
	}

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return fmt.Errorf("push: fcm returned %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
}

// accessTokenFor returns a cached access token, minting a new one when the
// cached token is missing or about to expire.
func (s *Sender) accessTokenFor(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.accessToken != "" && time.Now().Before(s.expiresAt) {
		return s.accessToken, nil
	}

	assertion, err := s.buildJWTAssertion(time.Now())
	if err != nil {
		return "", fmt.Errorf("push: build jwt: %w", err)
	}

	form := url.Values{}
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	form.Set("assertion", assertion)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.tokenURI, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("push: build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("push: token exchange: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("push: token exchange returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("push: decode token response: %w", err)
	}
	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("push: token exchange returned no access_token")
	}

	ttl := time.Duration(tokenResp.ExpiresIn) * time.Second
	if ttl <= tokenBuffer {
		ttl = time.Hour
	}
	s.accessToken = tokenResp.AccessToken
	s.expiresAt = time.Now().Add(ttl - tokenBuffer)

	return s.accessToken, nil
}

// buildJWTAssertion mints the signed JWT used in the OAuth2 JWT-bearer grant.
func (s *Sender) buildJWTAssertion(now time.Time) (string, error) {
	header, _ := json.Marshal(map[string]string{"alg": "RS256", "typ": "JWT"})
	claims, _ := json.Marshal(map[string]any{
		"iss":   s.clientEmail,
		"scope": oauthScope,
		"aud":   s.tokenURI,
		"iat":   now.Unix(),
		"exp":   now.Add(time.Hour).Unix(),
	})

	unsigned := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(claims)

	digest := sha256.Sum256([]byte(unsigned))
	signature, err := rsa.SignPKCS1v15(rand.Reader, s.privateKey, crypto.SHA256, digest[:])
	if err != nil {
		return "", err
	}

	return unsigned + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func messagePayload(token string, msg Message) map[string]any {
	m := map[string]any{
		"token": token,
		"notification": map[string]any{
			"title": msg.Title,
			"body":  msg.Body,
		},
	}
	if len(msg.Data) > 0 {
		m["data"] = msg.Data
	}
	return m
}

func parsePrivateKey(pemKey string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemKey))
	if block == nil {
		return nil, fmt.Errorf("no PEM block found")
	}

	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		if rsaKey, ok := key.(*rsa.PrivateKey); ok {
			return rsaKey, nil
		}
		return nil, fmt.Errorf("private key is not RSA")
	}

	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	return nil, fmt.Errorf("unsupported private key format")
}
