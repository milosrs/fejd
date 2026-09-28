package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"fejd-backend/internal/dto"
	"fejd-backend/internal/models"
	"fejd-backend/internal/service"
	"fejd-backend/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type BusinessHandler struct {
	businessStore        *store.BusinessStore
	buStore              *store.BusinessUserStore
	userStore            *store.UserStore
	serviceStore         *store.ServiceStore
	pageStore            *store.PageStore
	sectionStore         *store.SectionStore
	imageLinkStore       *store.ImageLinkStore
	employeeServiceStore *store.EmployeeServiceStore
	businessClosureStore *store.BusinessClosureStore
	slotService          *service.SlotService
}

func NewBusinessHandler(
	businessStore *store.BusinessStore,
	buStore *store.BusinessUserStore,
	userStore *store.UserStore,
	serviceStore *store.ServiceStore,
	pageStore *store.PageStore,
	sectionStore *store.SectionStore,
	imageLinkStore *store.ImageLinkStore,
	employeeServiceStore *store.EmployeeServiceStore,
	businessClosureStore *store.BusinessClosureStore,
	slotService *service.SlotService,
) *BusinessHandler {
	return &BusinessHandler{
		businessStore:        businessStore,
		buStore:              buStore,
		userStore:            userStore,
		serviceStore:         serviceStore,
		pageStore:            pageStore,
		sectionStore:         sectionStore,
		imageLinkStore:       imageLinkStore,
		employeeServiceStore: employeeServiceStore,
		businessClosureStore: businessClosureStore,
		slotService:          slotService,
	}
}

// ListBusinesses godoc
// @Summary      List salons
// @Description  Returns all salons for the public directory.
// @Tags         public
// @Produce      json
// @Success      200 {array} dto.DirectoryBusiness
// @Failure      500 {object} ErrorResponse
// @Router       /api/businesses [get]
func (h *BusinessHandler) ListBusinesses(w http.ResponseWriter, r *http.Request) {
	businesses, err := h.businessStore.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list businesses")
		return
	}
	if businesses == nil {
		businesses = []models.Business{}
	}

	writeJSON(w, http.StatusOK, dto.DirectoryBusinessesFromModels(businesses))
}

// GetBusiness godoc
// @Summary      Get business details
// @Description  Returns a business with its services and employees by slug.
// @Tags         public
// @Produce      json
// @Param        slug path string true "Business slug"
// @Success      200 {object} BusinessResponse
// @Failure      404 {object} ErrorResponse
// @Router       /api/business/{slug} [get]
func (h *BusinessHandler) GetBusiness(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	b, err := h.businessStore.GetBySlug(r.Context(), slug)
	if err != nil {
		writeError(w, http.StatusNotFound, "business not found")
		return
	}

	services, _ := h.serviceStore.ListByBusiness(r.Context(), b.ID)
	employees, _ := h.buStore.ListEmployeesByBusiness(r.Context(), b.ID)
	images := h.businessImages(r.Context(), b.ID)

	writeJSON(w, http.StatusOK, BusinessResponse{
		Business:  dto.BusinessFromModel(*b),
		Services:  dto.ServicesFromModels(services),
		Employees: dto.BusinessUsersFromModels(employees),
		Images:    images,
	})
}

func (h *BusinessHandler) businessImages(ctx context.Context, businessID uuid.UUID) BusinessImages {
	links, err := h.imageLinkStore.ListByEntity(ctx, "business", businessID)
	if err != nil {
		return BusinessImages{}
	}

	var images BusinessImages
	for _, l := range links {
		url := "/api/images/" + l.ImageID.String()
		switch l.Purpose {
		case "hero":
			images.Hero = url
		case "logo":
			images.Logo = url
		case "background":
			images.Background = url
		}
	}
	return images
}

// GetSections godoc
// @Summary      List landing page sections
// @Description  Returns the salon's landing page sections ordered by position.
// @Tags         public
// @Produce      json
// @Param        slug path string true "Business slug"
// @Success      200 {array} dto.Section
// @Failure      404 {object} ErrorResponse
// @Router       /api/business/{slug}/sections [get]
func (h *BusinessHandler) GetSections(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	b, err := h.businessStore.GetBySlug(r.Context(), slug)
	if err != nil {
		writeError(w, http.StatusNotFound, "business not found")
		return
	}

	page, err := h.pageStore.GetByBusinessAndName(r.Context(), b.ID, store.LandingPageName)
	if err != nil {
		writeJSON(w, http.StatusOK, dto.SectionsFromModels(nil))
		return
	}

	sections, err := h.sectionStore.ListByPage(r.Context(), page.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get sections")
		return
	}
	if sections == nil {
		sections = []models.Section{}
	}

	writeJSON(w, http.StatusOK, dto.SectionsFromModels(sections))
}

// GetClosures godoc
// @Summary      List non-working days
// @Description  Returns the days the salon is closed, so the booking calendar can disable them.
// @Tags         public
// @Produce      json
// @Param        slug path string true "Business slug"
// @Success      200 {array} dto.BusinessClosure
// @Failure      404 {object} ErrorResponse
// @Router       /api/business/{slug}/closures [get]
func (h *BusinessHandler) GetClosures(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	b, err := h.businessStore.GetBySlug(r.Context(), slug)
	if err != nil {
		writeError(w, http.StatusNotFound, "business not found")
		return
	}

	if h.businessClosureStore == nil {
		writeJSON(w, http.StatusOK, []dto.BusinessClosure{})
		return
	}

	closures, err := h.businessClosureStore.ListByBusiness(r.Context(), b.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get closures")
		return
	}
	if closures == nil {
		closures = []models.BusinessClosure{}
	}

	writeJSON(w, http.StatusOK, dto.BusinessClosuresFromModels(closures))
}

// GetServices godoc
// @Summary      List business services
// @Description  Returns all active services for a business.
// @Tags         public
// @Produce      json
// @Param        slug path string true "Business slug"
// @Success      200 {array} dto.Service
// @Failure      404 {object} ErrorResponse
// @Router       /api/business/{slug}/services [get]
func (h *BusinessHandler) GetServices(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	b, err := h.businessStore.GetBySlug(r.Context(), slug)
	if err != nil {
		writeError(w, http.StatusNotFound, "business not found")
		return
	}

	services, err := h.serviceStore.ListByBusiness(r.Context(), b.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get services")
		return
	}

	writeJSON(w, http.StatusOK, dto.ServicesFromModels(services))
}

// GetEmployees godoc
// @Summary      List business employees
// @Description  Returns all employees for a business.
// @Tags         public
// @Produce      json
// @Param        slug path string true "Business slug"
// @Success      200 {array} dto.BusinessUser
// @Failure      404 {object} ErrorResponse
// @Router       /api/business/{slug}/employees [get]
func (h *BusinessHandler) GetEmployees(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	b, err := h.businessStore.GetBySlug(r.Context(), slug)
	if err != nil {
		writeError(w, http.StatusNotFound, "business not found")
		return
	}

	members, err := h.buStore.ListByBusiness(r.Context(), b.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get employees")
		return
	}

	providerIDs, err := h.employeeServiceStore.ListServiceProviderIDs(r.Context(), b.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get employees")
		return
	}
	offers := make(map[uuid.UUID]bool, len(providerIDs))
	for _, id := range providerIDs {
		offers[id] = true
	}

	staff := make([]models.BusinessUser, 0, len(members))
	for _, m := range members {
		if m.Active && offers[m.ID] {
			staff = append(staff, m)
		}
	}

	users := dto.BusinessUsersFromModels(staff)
	for i := range users {
		users[i].Avatar = h.avatarURL(r.Context(), staff[i].ID)
		users[i].DisplayName = h.displayName(r.Context(), staff[i])
	}

	writeJSON(w, http.StatusOK, users)
}

func (h *BusinessHandler) avatarURL(ctx context.Context, businessUserID uuid.UUID) string {
	links, err := h.imageLinkStore.ListByEntity(ctx, "business_user", businessUserID)
	if err == nil {
		for _, l := range links {
			if l.Purpose == "avatar" {
				return "/api/images/" + l.ImageID.String()
			}
		}
	}

	if h.userStore == nil {
		return ""
	}
	bu, err := h.buStore.GetByID(ctx, businessUserID)
	if err != nil {
		return ""
	}
	u, err := h.userStore.GetByID(ctx, bu.UserID)
	if err != nil || u.AvatarID == nil {
		return ""
	}
	return "/api/images/" + u.AvatarID.String()
}

// displayName returns a business user's display name, falling back to their
// account profile name when the salon row has none (e.g. the owner).
func (h *BusinessHandler) displayName(ctx context.Context, bu models.BusinessUser) string {
	if bu.DisplayName != "" {
		return bu.DisplayName
	}
	if h.userStore == nil {
		return ""
	}
	u, err := h.userStore.GetByID(ctx, bu.UserID)
	if err != nil {
		return ""
	}
	return u.DisplayName
}

// GetServiceEmployees godoc
// @Summary      List staff who offer a service
// @Description  Returns active business users (owner and employees) assigned to the given service.
// @Tags         public
// @Produce      json
// @Param        slug path string true "Business slug"
// @Param        serviceID path string true "Service UUID"
// @Success      200 {array} dto.BusinessUser
// @Failure      400 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Router       /api/business/{slug}/services/{serviceID}/employees [get]
func (h *BusinessHandler) GetServiceEmployees(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	b, err := h.businessStore.GetBySlug(r.Context(), slug)
	if err != nil {
		writeError(w, http.StatusNotFound, "business not found")
		return
	}

	serviceID, err := uuid.Parse(chi.URLParam(r, "serviceID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid service ID")
		return
	}

	svc, err := h.serviceStore.GetByID(r.Context(), serviceID)
	if err != nil || svc.BusinessID != b.ID {
		writeError(w, http.StatusNotFound, "service not found")
		return
	}

	links, err := h.employeeServiceStore.ListByService(r.Context(), serviceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list employees")
		return
	}

	members, err := h.buStore.ListByBusiness(r.Context(), b.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list employees")
		return
	}

	capable := make(map[uuid.UUID]bool, len(links))
	for _, l := range links {
		capable[l.BusinessUserID] = true
	}

	result := make([]models.BusinessUser, 0, len(members))
	for _, m := range members {
		if m.Active && capable[m.ID] {
			result = append(result, m)
		}
	}

	users := dto.BusinessUsersFromModels(result)
	for i := range users {
		users[i].Avatar = h.avatarURL(r.Context(), result[i].ID)
		users[i].DisplayName = h.displayName(r.Context(), result[i])
	}

	writeJSON(w, http.StatusOK, users)
}

// GetAvailableSlots godoc
// @Summary      Get available time slots
// @Description  Returns available time slots for a service and employee on a given date.
// @Tags         public
// @Produce      json
// @Param        slug path string true "Business slug"
// @Param        service_id query string true "Service UUID"
// @Param        employee_id query string true "Employee UUID"
// @Param        date query string true "Date (YYYY-MM-DD)"
// @Success      200 {object} SlotsResponse
// @Failure      400 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Router       /api/business/{slug}/slots [get]
func (h *BusinessHandler) GetAvailableSlots(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	b, err := h.businessStore.GetBySlug(r.Context(), slug)
	if err != nil {
		writeError(w, http.StatusNotFound, "business not found")
		return
	}

	serviceID, err := uuid.Parse(r.URL.Query().Get("service_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid service_id")
		return
	}

	employeeID, err := uuid.Parse(r.URL.Query().Get("employee_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid employee_id")
		return
	}

	dateStr := r.URL.Query().Get("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid date format, use YYYY-MM-DD")
		return
	}

	slots, err := h.slotService.GetAvailableSlots(r.Context(), b.ID, serviceID, employeeID, date)
	if err != nil {
		writeInternalError(w, err)
		return
	}

	if slots == nil {
		slots = []models.TimeSlot{}
	}

	writeJSON(w, http.StatusOK, SlotsResponse{
		Slots: dto.TimeSlotsFromModels(slots),
		Date:  dateStr,
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, ErrorResponse{Error: msg})
}

// writeInternalError logs the real error server-side and returns a generic 500
// response, so internal details (SQL errors, connection info) never reach the
// client.
func writeInternalError(w http.ResponseWriter, err error) {
	log.Printf("[handler] internal error: %v", err)
	writeError(w, http.StatusInternalServerError, "internal server error")
}

// writeBookingError maps a booking error to a 409 Conflict response. Business
// rule messages (e.g. "time slot is no longer available") are user-safe and
// returned as-is; any underlying database error is logged and replaced with a
// generic message so raw SQL never reaches the client.
func writeBookingError(w http.ResponseWriter, err error) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) || errors.Is(err, pgx.ErrNoRows) {
		log.Printf("[handler] booking error: %v", err)
		writeError(w, http.StatusConflict, "unable to complete booking")
		return
	}
	writeError(w, http.StatusConflict, err.Error())
}
