package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequireApproved(t *testing.T) {
	cases := []struct {
		name       string
		status     string
		wantStatus int
	}{
		{"approved passes", "approved", http.StatusOK},
		{"pending rejected", "pending", http.StatusForbidden},
		{"rejected rejected", "rejected", http.StatusForbidden},
		{"empty rejected", "", http.StatusForbidden},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := &Middleware{}
			handler := m.RequireApproved(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			ctx := context.WithValue(context.Background(), ContextKeyClaims, &Claims{ApprovalStatus: tc.status})
			req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.wantStatus, rr.Code)
		})
	}
}

func TestRequireApprovedMissingClaims(t *testing.T) {
	m := &Middleware{}
	handler := m.RequireApproved(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
