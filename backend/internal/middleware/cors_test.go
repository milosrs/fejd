package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testCORSMiddleware(allowedOrigins []string, suffix string) http.Handler {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return CORS(allowedOrigins, suffix)(next)
}

func TestCORSAllowsExactOrigin(t *testing.T) {
	h := testCORSMiddleware([]string{"http://localhost:5173"}, "")

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "http://localhost:5173", rr.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", rr.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCORSRejectsUnknownOrigin(t *testing.T) {
	h := testCORSMiddleware([]string{"http://localhost:5173"}, "")

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Origin", "https://evil.com")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Empty(t, rr.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSAllowsSubdomainBySuffix(t *testing.T) {
	h := testCORSMiddleware(nil, "fejd.com")

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Origin", "https://dragicevic.fejd.com")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	assert.Equal(t, "https://dragicevic.fejd.com", rr.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSRejectsLookalikeSuffix(t *testing.T) {
	h := testCORSMiddleware(nil, "fejd.com")

	for _, origin := range []string{
		"https://evilfejd.com",
		"https://fejd.com.evil.com",
	} {
		req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
		req.Header.Set("Origin", origin)
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		assert.Empty(t, rr.Header().Get("Access-Control-Allow-Origin"), "origin %q", origin)
	}
}

func TestCORSPreflight(t *testing.T) {
	h := testCORSMiddleware([]string{"http://localhost:5173"}, "")

	req := httptest.NewRequest(http.MethodOptions, "/api/me", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	require.Equal(t, http.StatusNoContent, rr.Code)
	assert.Equal(t, "http://localhost:5173", rr.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, rr.Header().Get("Access-Control-Allow-Methods"), "POST")
	assert.Contains(t, rr.Header().Get("Access-Control-Allow-Headers"), "Authorization")
}
