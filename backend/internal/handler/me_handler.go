package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"fejd-backend/auth"
	"fejd-backend/internal/authutil"
	"fejd-backend/internal/db"
	"fejd-backend/internal/dto"
	"fejd-backend/internal/models"
	"fejd-backend/internal/store"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const maxSlugLength = 100

type MeHandler struct {
	businessStore *store.BusinessStore
	buStore       *store.BusinessUserStore
	pool          *pgxpool.Pool
}

func NewMeHandler(businessStore *store.BusinessStore, buStore *store.BusinessUserStore, pool *pgxpool.Pool) *MeHandler {
	return &MeHandler{businessStore: businessStore, buStore: buStore, pool: pool}
}

// GetMe godoc
// @Summary      Current user's onboarding state
// @Description  Returns the caller's approval status, whether they already own a salon, and the businesses they belong to.
// @Tags         me
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} dto.Me
// @Failure      401 {object} ErrorResponse
// @Router       /api/me [get]
func (h *MeHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, err := authutil.GetUserID(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	memberships, err := h.businessStore.ListMembershipsByUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list businesses")
		return
	}
	if memberships == nil {
		memberships = []models.BusinessMembership{}
	}

	hasSalon := false
	for _, m := range memberships {
		if m.Role == "admin" {
			hasSalon = true
			break
		}
	}

	writeJSON(w, http.StatusOK, dto.Me{
		ApprovalStatus: authutil.GetApprovalStatus(r),
		HasSalon:       hasSalon,
		Businesses:     dto.MeBusinessesFromModels(memberships),
	})
}

// CreateBusiness godoc
// @Summary      Create the caller's salon
// @Description  Creates a business and the caller's owner (admin) row. Requires an approved account with the Owner role.
// @Tags         me
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body BusinessCreateInput true "Business name"
// @Success      201 {object} dto.Business
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      409 {object} ErrorResponse
// @Router       /api/me/business [post]
func (h *MeHandler) CreateBusiness(w http.ResponseWriter, r *http.Request) {
	var input BusinessCreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	userID, err := authutil.GetUserID(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	if !authutil.HasRole(r, auth.RoleOwner) {
		writeError(w, http.StatusForbidden, "owner role required")
		return
	}

	hasSalon, err := h.buStore.HasAdminBusiness(r.Context(), h.pool, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check salon")
		return
	}
	if hasSalon {
		writeError(w, http.StatusConflict, "already has a business")
		return
	}

	slug, err := h.availableSlug(r, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate slug")
		return
	}

	business := &models.Business{Name: name, Slug: slug}
	err = db.WithTx(r.Context(), h.pool, func(tx pgx.Tx) error {
		if err := h.businessStore.Create(r.Context(), tx, business); err != nil {
			return err
		}
		owner := &models.BusinessUser{
			BusinessID: business.ID,
			UserID:     userID,
			Role:       "admin",
		}
		return h.buStore.Create(r.Context(), tx, owner)
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			writeError(w, http.StatusConflict, "already has a business")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create business")
		return
	}

	writeJSON(w, http.StatusCreated, dto.BusinessFromModel(*business))
}

func (h *MeHandler) availableSlug(r *http.Request, name string) (string, error) {
	base := slugify(name)
	slug := base
	for i := 2; ; i++ {
		exists, err := h.businessStore.SlugExists(r.Context(), h.pool, slug)
		if err != nil {
			return "", err
		}
		if !exists {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", base, i)
	}
}

func slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	lastDash := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDash = false
		} else if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if len(out) > maxSlugLength {
		out = out[:maxSlugLength]
	}
	if out == "" {
		return "salon"
	}
	return out
}
