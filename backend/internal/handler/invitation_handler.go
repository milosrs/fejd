package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"fejd-backend/internal/authutil"
	"fejd-backend/internal/dto"
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

// GetInvitation godoc
// @Summary      Resolve an invite link
// @Description  Returns the target salon for an invite token, so the landing page can greet the invitee before they register. Never returns the token itself.
// @Tags         invitations
// @Produce      json
// @Param        token path string true "Invitation token"
// @Success      200 {object} PublicInvitationResponse
// @Failure      404 {object} ErrorResponse
// @Failure      410 {object} ErrorResponse
// @Router       /api/invitations/{token} [get]
func (h *InvitationHandler) GetInvitation(w http.ResponseWriter, r *http.Request) {
	inv, err := h.invitations.GetInvitation(r.Context(), chi.URLParam(r, "token"))
	if err != nil {
		if status, ok := invitationErrorStatus(err); ok {
			writeError(w, status, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, PublicInvitationResponse{
		SalonName: inv.SalonName,
		SalonSlug: inv.SalonSlug,
		Role:      inv.Role,
		ExpiresAt: inv.ExpiresAt,
	})
}

// AcceptInvitation godoc
// @Summary      Accept an invite link
// @Description  Links the authenticated user to the invited salon as an employee and grants the Employee role. Idempotent; works for new and existing users.
// @Tags         invitations
// @Produce      json
// @Param        token path string true "Invitation token"
// @Success      200 {object} dto.Business
// @Failure      401 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      409 {object} ErrorResponse
// @Failure      410 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/invitations/{token}/accept [post]
func (h *InvitationHandler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	userID, err := authutil.GetUserID(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	business, err := h.invitations.AcceptInvitation(r.Context(), chi.URLParam(r, "token"), userID)
	if err != nil {
		if status, ok := invitationErrorStatus(err); ok {
			writeError(w, status, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, dto.BusinessFromModel(*business))
}

// invitationErrorStatus maps invitation service sentinel errors to HTTP codes.
func invitationErrorStatus(err error) (int, bool) {
	switch {
	case errors.Is(err, service.ErrInvitationNotFound):
		return http.StatusNotFound, true
	case errors.Is(err, service.ErrInvitationExpired), errors.Is(err, service.ErrInvitationUsed):
		return http.StatusGone, true
	case errors.Is(err, service.ErrCannotInviteOwner):
		return http.StatusConflict, true
	default:
		return 0, false
	}
}
