package auth

import (
	"context"
	"fmt"
	"log"
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
	return m.authenticate(next, true)
}

// AuthenticateUnverified authenticates like Authenticate but skips the
// email-verification check. It is used by registration-finalization routes
// (claim-role) that must run immediately after self-registration, when the
// user's email is necessarily still unverified.
func (m *Middleware) AuthenticateUnverified(next http.Handler) http.Handler {
	return m.authenticate(next, false)
}

func (m *Middleware) authenticate(next http.Handler, requireVerified bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString, err := ExtractBearerToken(r)
		if err != nil {
			log.Printf("[auth] missing/invalid bearer token: %v (path=%s)", err, r.URL.Path)
			http.Error(w, "unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		claims, err := m.jwksClient.ValidateToken(tokenString)
		if err != nil {
			log.Printf("[auth] token validation failed: %v (path=%s)", err, r.URL.Path)
			http.Error(w, "unauthorized: invalid token", http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, ContextKeyClaims, claims)
		ctx = context.WithValue(ctx, ContextKeyUserID, claims.Subject)
		ctx = context.WithValue(ctx, ContextKeyApprovalStatus, claims.ApprovalStatus)

		var roles []string
		if realmRoles, exists := claims.RealmAccess["roles"]; exists {
			roles = append(roles, realmRoles...)
		}
		ctx = context.WithValue(ctx, ContextKeyRoles, roles)

		log.Printf("[auth] authenticated user=%q path=%s roles=%v", claims.Subject, r.URL.Path, roles)

		// Unverified emails may only read; every write (POST/PUT/DELETE/PATCH)
		// is blocked until the user confirms their address.
		if requireVerified && !requireVerifiedEmail(w, r) {
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// requireVerifiedEmail enforces read-only access for unverified emails and
// writes the 403 response when blocked. It returns true when the request may
// proceed. Callers must run Authenticate first so claims are in the context.
func requireVerifiedEmail(w http.ResponseWriter, r *http.Request) bool {
	claims := GetClaimsFromRequest(r)
	if claims == nil || claims.EmailVerified || isReadOnly(r.Method) {
		return true
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte(`{"error":"email not verified"}`))
	return false
}

// isReadOnly reports whether an HTTP method is a safe, read-only method that
// an unverified user is still allowed to perform.
func isReadOnly(method string) bool {
	return method == http.MethodGet || method == http.MethodHead
}

// OptionalAuthenticate validates the bearer token when present but never
// rejects the request: a valid token populates the usual context keys, an
// invalid token sets ContextKeyAuthInvalid, and a missing token leaves the
// request anonymous. The handler decides whether auth matters.
func (m *Middleware) OptionalAuthenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString, err := ExtractBearerToken(r)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		claims, err := m.jwksClient.ValidateToken(tokenString)
		if err != nil {
			ctx := context.WithValue(r.Context(), ContextKeyAuthInvalid, true)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, ContextKeyClaims, claims)
		ctx = context.WithValue(ctx, ContextKeyUserID, claims.Subject)
		ctx = context.WithValue(ctx, ContextKeyApprovalStatus, claims.ApprovalStatus)

		var roles []string
		if realmRoles, exists := claims.RealmAccess["roles"]; exists {
			roles = append(roles, realmRoles...)
		}
		ctx = context.WithValue(ctx, ContextKeyRoles, roles)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func IsAuthInvalidFromRequest(r *http.Request) bool {
	if invalid, ok := r.Context().Value(ContextKeyAuthInvalid).(bool); ok {
		return invalid
	}
	return false
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

// RequireApproved blocks requests whose user has not been approved
// (approval_status != "approved"). Callers must run Authenticate first so
// claims are present in the context.
func (m *Middleware) RequireApproved(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := GetClaimsFromRequest(r)
		if claims == nil {
			http.Error(w, "unauthorized: missing claims", http.StatusUnauthorized)
			return
		}
		// Realm administrators are super-admins and are not subject to the
		// owner-onboarding approval flag.
		if claims.ApprovalStatus != "approved" && !IsRealmAdmin(claims) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error":"account pending approval"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireRealmAdmin blocks requests whose user is not a Keycloak realm
// administrator (realm-level "admin" role or the realm-management
// "realm-admin" client role). Callers must run Authenticate first.
func (m *Middleware) RequireRealmAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !IsRealmAdmin(GetClaimsFromRequest(r)) {
			http.Error(w, "forbidden: realm admin required", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
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

func GetApprovalStatusFromRequest(r *http.Request) string {
	if status, ok := r.Context().Value(ContextKeyApprovalStatus).(string); ok {
		return status
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
