package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fejd-backend/internal/store"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPushHandler_RegisterAndUnregister(t *testing.T) {
	pool := setupHandlerTestDB(t)
	h := NewPushHandler(store.NewPushTokenStore(pool))

	registerReq := withUser(
		httptest.NewRequest(http.MethodPost, "/api/me/push-token", bytes.NewBufferString(`{"token":"fcm-token-1","platform":"android"}`)),
		"user-1", "approved",
	)
	rr := httptest.NewRecorder()
	h.Register(rr, registerReq)
	require.Equal(t, http.StatusOK, rr.Code)

	tokens, err := store.NewPushTokenStore(pool).ListByUserIDs(registerReq.Context(), []string{"user-1"})
	require.NoError(t, err)
	require.Len(t, tokens, 1)
	assert.Equal(t, "fcm-token-1", tokens[0].Token)
	assert.Equal(t, "android", tokens[0].Platform)

	unregisterReq := withUser(
		httptest.NewRequest(http.MethodDelete, "/api/me/push-token", bytes.NewBufferString(`{"token":"fcm-token-1"}`)),
		"user-1", "approved",
	)
	rr = httptest.NewRecorder()
	h.Unregister(rr, unregisterReq)
	require.Equal(t, http.StatusOK, rr.Code)

	tokens, err = store.NewPushTokenStore(pool).ListByUserIDs(unregisterReq.Context(), []string{"user-1"})
	require.NoError(t, err)
	assert.Empty(t, tokens)
}

func TestPushHandler_RegisterRejectsEmptyToken(t *testing.T) {
	pool := setupHandlerTestDB(t)
	h := NewPushHandler(store.NewPushTokenStore(pool))

	req := withUser(
		httptest.NewRequest(http.MethodPost, "/api/me/push-token", bytes.NewBufferString(`{"token":"   "}`)),
		"user-1", "approved",
	)
	rr := httptest.NewRecorder()
	h.Register(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestPushHandler_RequiresAuth(t *testing.T) {
	pool := setupHandlerTestDB(t)
	h := NewPushHandler(store.NewPushTokenStore(pool))

	req := httptest.NewRequest(http.MethodPost, "/api/me/push-token", bytes.NewBufferString(`{"token":"t"}`))
	rr := httptest.NewRecorder()
	h.Register(rr, req)
	require.Equal(t, http.StatusUnauthorized, rr.Code)

	var body ErrorResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &body))
	assert.NotEmpty(t, body.Error)
}
