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

func TestRequireVerifiedEmail(t *testing.T) {
	cases := []struct {
		name       string
		verified   bool
		method     string
		wantStatus int
	}{
		{"verified user may write", true, http.MethodPost, http.StatusOK},
		{"verified user may read", true, http.MethodGet, http.StatusOK},
		{"unverified user may read", false, http.MethodGet, http.StatusOK},
		{"unverified user may head", false, http.MethodHead, http.StatusOK},
		{"unverified user may not write", false, http.MethodPost, http.StatusForbidden},
		{"unverified user may not put", false, http.MethodPut, http.StatusForbidden},
		{"unverified user may not delete", false, http.MethodDelete, http.StatusForbidden},
		{"unverified user may not patch", false, http.MethodPatch, http.StatusForbidden},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), ContextKeyClaims, &Claims{EmailVerified: tc.verified})
			req := httptest.NewRequest(tc.method, "/", nil).WithContext(ctx)
			rr := httptest.NewRecorder()

			allowed := requireVerifiedEmail(rr, req)
			if tc.wantStatus == http.StatusOK {
				assert.True(t, allowed)
				assert.Equal(t, http.StatusOK, rr.Code)
			} else {
				assert.False(t, allowed)
				assert.Equal(t, http.StatusForbidden, rr.Code)
			}
		})
	}
}

func TestRequireVerifiedEmailMissingClaims(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rr := httptest.NewRecorder()

	assert.True(t, requireVerifiedEmail(rr, req))
}
