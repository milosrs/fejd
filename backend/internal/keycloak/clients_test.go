package keycloak

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"fejd-backend/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnsureClientRedirect(t *testing.T) {
	var mu sync.Mutex
	clientRep := map[string]any{
		"id":           "frontend-uuid",
		"clientId":     "fejd-frontend",
		"protocol":     "openid-connect",
		"redirectUris": []any{"https://fejd.fyi/*"},
		"webOrigins":   []any{"https://fejd.fyi"},
	}
	var putBodies []map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/realms/fejd/protocol/openid-connect/token" && r.Method == http.MethodPost:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"tok","expires_in":300}`))
		case r.URL.Path == "/admin/realms/fejd/clients" && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{"id":"frontend-uuid","clientId":"fejd-frontend"}]`))
		case r.URL.Path == "/admin/realms/fejd/clients/frontend-uuid" && r.Method == http.MethodGet:
			mu.Lock()
			defer mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(clientRep)
		case r.URL.Path == "/admin/realms/fejd/clients/frontend-uuid" && r.Method == http.MethodPut:
			mu.Lock()
			defer mu.Unlock()
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			putBodies = append(putBodies, body)
			clientRep = body
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client := NewClient(config.KeycloakConfig{
		AdminURL:          srv.URL,
		Realm:             "fejd",
		AdminClientID:     "fejd-admin",
		AdminClientSecret: "secret",
	})

	err := client.EnsureClientRedirect(context.Background(), "fejd-frontend", "https://my-salon.fejd.fyi/*", "https://my-salon.fejd.fyi")
	require.NoError(t, err)

	// Second call is idempotent and must not duplicate the entries.
	err = client.EnsureClientRedirect(context.Background(), "fejd-frontend", "https://my-salon.fejd.fyi/*", "https://my-salon.fejd.fyi")
	require.NoError(t, err)

	require.Len(t, putBodies, 2)
	final := putBodies[1]

	redirects := final["redirectUris"].([]any)
	origins := final["webOrigins"].([]any)

	assert.Len(t, redirects, 2)
	assert.Contains(t, redirects, "https://fejd.fyi/*")
	assert.Contains(t, redirects, "https://my-salon.fejd.fyi/*")

	assert.Len(t, origins, 2)
	assert.Contains(t, origins, "https://fejd.fyi")
	assert.Contains(t, origins, "https://my-salon.fejd.fyi")
}

func TestAnyContains(t *testing.T) {
	items := []any{"a", 1, "b"}
	assert.True(t, anyContains(items, "a"))
	assert.True(t, anyContains(items, "b"))
	assert.False(t, anyContains(items, "c"))
	assert.False(t, anyContains(nil, "a"))
}
