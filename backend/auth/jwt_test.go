package auth

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestAudienceAllowed(t *testing.T) {
	cases := []struct {
		name    string
		allowed []string
		aud     jwt.ClaimStrings
		want    bool
	}{
		{"single match", []string{"a", "b"}, jwt.ClaimStrings{"b"}, true},
		{"multi token audience intersects", []string{"a", "b"}, jwt.ClaimStrings{"x", "a"}, true},
		{"no match", []string{"a", "b"}, jwt.ClaimStrings{"c"}, false},
		{"empty token audience", []string{"a", "b"}, jwt.ClaimStrings{}, false},
		{"frontend audience accepted", []string{"salon-mobile", "fejd-frontend"}, jwt.ClaimStrings{"fejd-frontend"}, true},
		{"mobile audience accepted", []string{"salon-mobile", "fejd-frontend"}, jwt.ClaimStrings{"salon-mobile"}, true},
		{"backend-only audience rejected", []string{"salon-mobile", "fejd-frontend"}, jwt.ClaimStrings{"fejd-backend"}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, audienceAllowed(tc.aud, tc.allowed))
		})
	}
}

func TestIsRealmAdmin(t *testing.T) {
	claims := func(realmRoles []string, resourceAccess map[string]any) *Claims {
		return &Claims{
			RealmAccess:    map[string][]string{"roles": realmRoles},
			ResourceAccess: resourceAccess,
		}
	}

	t.Run("nil claims", func(t *testing.T) {
		assert.False(t, IsRealmAdmin(nil))
	})

	t.Run("realm admin role", func(t *testing.T) {
		assert.True(t, IsRealmAdmin(claims([]string{"admin"}, nil)))
	})

	t.Run("realm-management realm-admin client role", func(t *testing.T) {
		ra := map[string]any{
			"realm-management": map[string]any{"roles": []any{"realm-admin"}},
		}
		assert.True(t, IsRealmAdmin(claims([]string{"default-roles-fejd"}, ra)))
	})

	t.Run("ordinary owner", func(t *testing.T) {
		assert.False(t, IsRealmAdmin(claims([]string{"Owner"}, nil)))
	})

	t.Run("no admin signal", func(t *testing.T) {
		ra := map[string]any{
			"account": map[string]any{"roles": []any{"manage-account"}},
		}
		assert.False(t, IsRealmAdmin(claims([]string{"Customer"}, ra)))
	})
}
