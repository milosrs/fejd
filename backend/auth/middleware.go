package auth

import (
	"context"
	"fmt"
	"net/http"
)

type Middleware struct {
	jwksClient *JWKSClient
}

func NewMiddleware(config KeycloakConfig) (*Middleware, error) {
	jwksClient, err := NewJWKSClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize JWKS client: %w", err)
	}
	return &Middleware{jwksClient: jwksClient}, nil
}

func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString, err := ExtractBearerToken(r)
		if err != nil {
			http.Error(w, "unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		claims, err := m.jwksClient.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "unauthorized: invalid token", http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, ContextKeyClaims, claims)
		ctx = context.WithValue(ctx, ContextKeyUserID, claims.Subject)

		var roles []string
		if realmRoles, exists := claims.RealmAccess["roles"]; exists {
			roles = append(roles, realmRoles...)
		}
		ctx = context.WithValue(ctx, ContextKeyRoles, roles)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaimsFromRequest(r)
			if claims == nil {
				http.Error(w, "unauthorized: missing claims", http.StatusUnauthorized)
				return
			}

			if realmRoles, exists := claims.RealmAccess["roles"]; exists {
				for _, rl := range realmRoles {
					if rl == role {
						next.ServeHTTP(w, r)
						return
					}
				}
			}

			for _, res := range claims.ResourceAccess {
				resMap, ok := res.(map[string]any)
				if !ok {
					continue
				}
				resourceRoles, exists := resMap["roles"].([]any)
				if !exists {
					continue
				}
				for _, rl := range resourceRoles {
					if roleStr, ok := rl.(string); ok && roleStr == role {
						next.ServeHTTP(w, r)
						return
					}
				}
			}

			http.Error(w, "forbidden: missing required role", http.StatusForbidden)
		})
	}
}

func (m *Middleware) RequireAnyRole(roles []string) func(http.Handler) http.Handler {
	required := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		required[role] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaimsFromRequest(r)
			if claims == nil {
				http.Error(w, "unauthorized: missing claims", http.StatusUnauthorized)
				return
			}

			if realmRoles, exists := claims.RealmAccess["roles"]; exists {
				for _, role := range realmRoles {
					if _, ok := required[role]; ok {
						next.ServeHTTP(w, r)
						return
					}
				}
			}

			for _, res := range claims.ResourceAccess {
				resMap, ok := res.(map[string]any)
				if !ok {
					continue
				}
				resourceRoles, exists := resMap["roles"].([]any)
				if !exists {
					continue
				}
				for _, rl := range resourceRoles {
					if roleStr, ok := rl.(string); ok {
						if _, ok := required[roleStr]; ok {
							next.ServeHTTP(w, r)
							return
						}
					}
				}
			}

			http.Error(w, "forbidden: insufficient permissions", http.StatusForbidden)
		})
	}
}

func GetClaimsFromRequest(r *http.Request) *Claims {
	if claims, ok := r.Context().Value(ContextKeyClaims).(*Claims); ok {
		return claims
	}
	return nil
}

func GetUserIDFromRequest(r *http.Request) string {
	if userID, ok := r.Context().Value(ContextKeyUserID).(string); ok {
		return userID
	}
	return ""
}

func GetRolesFromRequest(r *http.Request) []string {
	if roles, ok := r.Context().Value(ContextKeyRoles).([]string); ok {
		return roles
	}
	return nil
}

func Chain(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}
