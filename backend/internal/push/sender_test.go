package push

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testServiceAccountJSON(t *testing.T) []byte {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	der, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)

	pemKey := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})

	sa := map[string]any{
		"type":         "service_account",
		"project_id":   "fejd-c8cab",
		"client_email": "firebase-adminsdk@fejd-c8cab.iam.gserviceaccount.com",
		"private_key":  string(pemKey),
		"token_uri":    "https://oauth2.googleapis.com/token",
	}
	out, err := json.Marshal(sa)
	require.NoError(t, err)
	return out
}

func TestNewSenderParsesServiceAccount(t *testing.T) {
	s, err := NewSender(testServiceAccountJSON(t))
	require.NoError(t, err)

	assert.Equal(t, "fejd-c8cab", s.projectID)
	assert.Equal(t, "firebase-adminsdk@fejd-c8cab.iam.gserviceaccount.com", s.clientEmail)
	assert.Equal(t, "https://oauth2.googleapis.com/token", s.tokenURI)
}

func TestNewSenderDefaultsTokenURI(t *testing.T) {
	var sa map[string]any
	require.NoError(t, json.Unmarshal(testServiceAccountJSON(t), &sa))
	delete(sa, "token_uri")
	raw, err := json.Marshal(sa)
	require.NoError(t, err)

	s, err := NewSender(raw)
	require.NoError(t, err)
	assert.Equal(t, oauthURL, s.tokenURI)
}

func TestNewSenderRejectsMissingProject(t *testing.T) {
	var sa map[string]any
	require.NoError(t, json.Unmarshal(testServiceAccountJSON(t), &sa))
	delete(sa, "project_id")
	raw, err := json.Marshal(sa)
	require.NoError(t, err)

	_, err = NewSender(raw)
	require.Error(t, err)
}

func TestNewSenderRejectsInvalidJSON(t *testing.T) {
	_, err := NewSender([]byte("{not json"))
	require.Error(t, err)
}

func TestMessagePayload(t *testing.T) {
	payload := messagePayload("token-1", Message{Title: "Hi", Body: "There", Data: map[string]string{"k": "v"}})

	assert.Equal(t, "token-1", payload["token"])
	notification := payload["notification"].(map[string]any)
	assert.Equal(t, "Hi", notification["title"])
	assert.Equal(t, "There", notification["body"])
	assert.Equal(t, map[string]string{"k": "v"}, payload["data"])
}

func TestMessagePayloadOmitsEmptyData(t *testing.T) {
	payload := messagePayload("token-1", Message{Title: "Hi", Body: "There"})
	assert.NotContains(t, payload, "data")
}

func TestBuildJWTAssertion(t *testing.T) {
	s, err := NewSender(testServiceAccountJSON(t))
	require.NoError(t, err)

	assertion, err := s.buildJWTAssertion(time.Now())
	require.NoError(t, err)

	parts := strings.Split(assertion, ".")
	require.Len(t, parts, 3)

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	require.NoError(t, err)

	var claims map[string]any
	require.NoError(t, json.Unmarshal(payload, &claims))
	assert.Equal(t, s.clientEmail, claims["iss"])
	assert.Equal(t, oauthScope, claims["scope"])
	assert.Equal(t, s.tokenURI, claims["aud"])
}
