package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/time/rate"
)

type Claims struct {
	jwt.RegisteredClaims
	RealmAccess    map[string][]string `json:"realm_access"`
	ResourceAccess map[string]any      `json:"resource_access"`
	Email          string              `json:"email"`
	EmailVerified  bool                `json:"email_verified"`
}

type ContextKey string

const (
	ContextKeyClaims ContextKey = "claims"
	ContextKeyUserID ContextKey = "user_id"
	ContextKeyRoles  ContextKey = "roles"
)

type KeycloakConfig struct {
	RealmURL     string
	ClientID     string
	RequiredRole string
}

type JWKSClient struct {
	jwks   keyfunc.Keyfunc
	config KeycloakConfig
}

func NewJWKSClient(config KeycloakConfig) (*JWKSClient, error) {
	options := keyfunc.Override{
		RefreshInterval:   1 * time.Hour,
		HTTPTimeout:       10 * time.Second,
		RefreshUnknownKID: rate.NewLimiter(rate.Every(jwt.TimePrecision), 5),
		RateLimitWaitMax:  time.Second * 2,
	}

	jwks, err := keyfunc.NewDefaultOverrideCtx(context.Background(), []string{config.RealmURL + "/protocol/openid-connect/certs"}, options)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	return &JWKSClient{jwks: jwks, config: config}, nil
}

func (k *JWKSClient) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		k.jwks.Keyfunc,
		jwt.WithAudience(k.config.ClientID),
		jwt.WithIssuer(k.config.RealmURL),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

func (k *KeycloakConfig) RequireRole(claims *Claims) error {
	if k.RequiredRole == "" {
		return nil
	}

	if roles, exists := claims.RealmAccess["roles"]; exists {
		for _, role := range roles {
			if role == k.RequiredRole {
				return nil
			}
		}
	}

	for _, res := range claims.ResourceAccess {
		resMap, ok := res.(map[string]any)
		if !ok {
			continue
		}
		roles, exists := resMap["roles"].([]any)
		if !exists {
			continue
		}
		for _, role := range roles {
			if roleStr, ok := role.(string); ok && roleStr == k.RequiredRole {
				return nil
			}
		}
	}

	return fmt.Errorf("user missing required role: %s", k.RequiredRole)
}

func ExtractBearerToken(r *http.Request) (string, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	const prefix = "Bearer "
	if len(auth) < len(prefix) || auth[:len(prefix)] != prefix {
		return "", fmt.Errorf("invalid authorization format, expected Bearer token")
	}

	return auth[len(prefix):], nil
}
