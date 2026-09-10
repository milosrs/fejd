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

// maxDNSLabelLength caps a slug so it can serve as a single DNS label (a
// subdomain) as well as a path segment. RFC 1035 limits labels to 63 octets.
const maxDNSLabelLength = 63

// reservedSubdomains are labels that must never become a salon slug, because
// they are (or will be) used for Fejd infrastructure: the app shell, the API
// edge, auth, etc. Keep in sync with the backfill migration
// 000015_reserved_slug_backfill.up.sql.
var reservedSubdomains = map[string]struct{}{
	"www": {}, "api": {}, "app": {}, "auth": {}, "keycloak": {},
	"admin": {}, "m": {}, "static": {}, "cdn": {}, "smtp": {},
	"mail": {}, "help": {}, "support": {}, "status": {}, "docs": {},
	"blog": {}, "staging": {}, "dev": {}, "test": {},
}

type MeHandler struct {
	businessStore      *store.BusinessStore
	buStore            *store.BusinessUserStore
	userStore          *store.UserStore
	businessHoursStore *store.BusinessHoursStore
	pool               *pgxpool.Pool
}

func NewMeHandler(businessStore *store.BusinessStore, buStore *store.BusinessUserStore, userStore *store.UserStore, businessHoursStore *store.BusinessHoursStore, pool *pgxpool.Pool) *MeHandler {
	return &MeHandler{businessStore: businessStore, buStore: buStore, userStore: userStore, businessHoursStore: businessHoursStore, pool: pool}
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

	avatar := ""
	if u, err := h.userStore.GetByID(r.Context(), userID); err == nil && u.AvatarID != nil {
		avatar = "/api/images/" + u.AvatarID.String()
	}

	writeJSON(w, http.StatusOK, dto.Me{
		ApprovalStatus: authutil.GetApprovalStatus(r),
		HasSalon:       hasSalon,
		Businesses:     dto.MeBusinessesFromModels(memberships),
		Avatar:         avatar,
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
		if err := h.buStore.Create(r.Context(), tx, owner); err != nil {
			return err
		}
		return h.businessHoursStore.SeedDefaults(r.Context(), tx, business.ID)
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
	base := sanitizeSlug(slugify(name))
	slug := base
	for i := 2; ; i++ {
		if !isValidDNSLabel(slug) || isReservedSubdomain(slug) {
			slug = fmt.Sprintf("%s-%d", base, i)
			continue
		}
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

// sanitizeSlug trims a slugified name to a valid DNS label and guarantees a
// non-empty result.
func sanitizeSlug(base string) string {
	if len(base) > maxDNSLabelLength {
		base = base[:maxDNSLabelLength]
	}
	base = strings.Trim(base, "-")
	if base == "" {
		return "salon"
	}
	return base
}

// isValidDNSLabel reports whether s is a valid single DNS label: at most 63
// characters, lowercase alphanumerics and hyphens, no leading/trailing hyphen.
func isValidDNSLabel(s string) bool {
	if s == "" || len(s) > maxDNSLabelLength {
		return false
	}
	if s[0] == '-' || s[len(s)-1] == '-' {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			continue
		}
		return false
	}
	return true
}

func isReservedSubdomain(s string) bool {
	_, ok := reservedSubdomains[s]
	return ok
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
