package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fejd-backend/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withLocale(r *http.Request, locale string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("locale", locale)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestI18nHandler_GetTranslations(t *testing.T) {
	pool := setupHandlerTestDB(t)
	h := NewI18nHandler(store.NewTranslationStore(pool))
	ctx := context.Background()

	_, err := pool.Exec(ctx, `
		INSERT INTO translations (key, locale, value) VALUES
		('greeting', 'de', 'Hallo'),
		('farewell', 'de', 'Tschüss')`)
	require.NoError(t, err)

	req := withLocale(httptest.NewRequest(http.MethodGet, "/api/i18n/de", nil), "de")
	rr := httptest.NewRecorder()

	h.GetTranslations(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var out map[string]string
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &out))
	assert.Equal(t, "Hallo", out["greeting"])
	assert.Equal(t, "Tschüss", out["farewell"])
}

func TestI18nHandler_GetTranslations_Seeded(t *testing.T) {
	pool := setupHandlerTestDB(t)
	h := NewI18nHandler(store.NewTranslationStore(pool))

	req := withLocale(httptest.NewRequest(http.MethodGet, "/api/i18n/en", nil), "en")
	rr := httptest.NewRecorder()

	h.GetTranslations(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var out map[string]string
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &out))
	assert.Equal(t, "This page isn't set up yet", out["landing.empty.title"])
	assert.Equal(t, "The salon hasn't published any content yet. Check back soon.", out["landing.empty.body"])
}
