package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"fejd-backend/internal/authutil"
	"fejd-backend/internal/models"
	"fejd-backend/internal/store"
)

// PushHandler manages a user's push-notification device tokens.
type PushHandler struct {
	tokenStore *store.PushTokenStore
}

func NewPushHandler(tokenStore *store.PushTokenStore) *PushHandler {
	return &PushHandler{tokenStore: tokenStore}
}

// Register godoc
// @Summary      Register a push token
// @Description  Associates the current device's FCM token with the authenticated user so they can receive push notifications.
// @Tags         me
// @Accept       json
// @Produce      json
// @Param        body body RegisterPushTokenRequest true "Device token"
// @Success      200 {object} MessageResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/me/push-token [post]
func (h *PushHandler) Register(w http.ResponseWriter, r *http.Request) {
	userID, err := authutil.GetUserID(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var body RegisterPushTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, errInvalidRequestBody.Error())
		return
	}

	token := strings.TrimSpace(body.Token)
	if token == "" {
		writeError(w, http.StatusBadRequest, "token is required")
		return
	}

	platform := strings.TrimSpace(body.Platform)
	if platform == "" {
		platform = "unknown"
	}

	if err := h.tokenStore.Upsert(r.Context(), &models.PushToken{
		UserID:   userID,
		Token:    token,
		Platform: platform,
	}); err != nil {
		writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, MessageResponse{Message: "push token registered"})
}

// Unregister godoc
// @Summary      Remove a push token
// @Description  Removes the current device's FCM token so it no longer receives push notifications.
// @Tags         me
// @Accept       json
// @Produce      json
// @Param        body body UnregisterPushTokenRequest true "Device token"
// @Success      200 {object} MessageResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/me/push-token [delete]
func (h *PushHandler) Unregister(w http.ResponseWriter, r *http.Request) {
	userID, err := authutil.GetUserID(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var body UnregisterPushTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, errInvalidRequestBody.Error())
		return
	}

	token := strings.TrimSpace(body.Token)
	if token == "" {
		writeError(w, http.StatusBadRequest, "token is required")
		return
	}

	if err := h.tokenStore.Delete(r.Context(), userID, token); err != nil {
		writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, MessageResponse{Message: "push token removed"})
}
