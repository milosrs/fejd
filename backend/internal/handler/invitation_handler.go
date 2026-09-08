package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"fejd-backend/internal/authutil"
	"fejd-backend/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type InvitationHandler struct {
	invitations *service.InvitationService
	defaultTTL  time.Duration
}

func NewInvitationHandler(invitations *service.InvitationService, defaultTTL time.Duration) *InvitationHandler {
	return &InvitationHandler{invitations: invitations, defaultTTL: defaultTTL}
}

// CreateInvitation godoc
// @Summary      Create an invite link / QR code
// @Description  Generates a single-use invitation link that, when opened, links the recipient to this salon as an employee. Accessible to owners and employees.
// @Tags         invitations
// @Accept       json
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        body body CreateInvitationRequest false "Invitation options"
// @Success      201 {object} InvitationResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/invitations [post]
func (h *InvitationHandler) CreateInvitation(w http.ResponseWriter, r *http.Request) {
	businessID, err := uuid.Parse(chi.URLParam(r, "businessID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errInvalidBusinessID.Error())
		return
	}

	userID, err := authutil.GetUserID(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var body CreateInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, errInvalidRequestBody.Error())
		return
	}

	ttl := h.defaultTTL
	if body.ExpiresInHours > 0 {
		ttl = time.Duration(body.ExpiresInHours) * time.Hour
	}

	out, err := h.invitations.CreateInvitation(r.Context(), businessID, userID, ttl)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, InvitationResponse{
		ID:        out.ID,
		URL:       out.URL,
		Token:     out.Token,
		ExpiresAt: out.ExpiresAt,
	})
}
