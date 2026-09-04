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
