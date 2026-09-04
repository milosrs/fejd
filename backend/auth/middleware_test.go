package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
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

			if rr.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d", tc.wantStatus, rr.Code)
			}
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

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}
