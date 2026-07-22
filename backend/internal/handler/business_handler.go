package handler

import (
	"encoding/json"
	"fejd-backend/internal/models"
	"fejd-backend/internal/service"
	"fejd-backend/internal/store"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type BusinessHandler struct {
	businessStore *store.BusinessStore
	buStore       *store.BusinessUserStore
	serviceStore  *store.ServiceStore
	slotService   *service.SlotService
}

func NewBusinessHandler(
	businessStore *store.BusinessStore,
	buStore *store.BusinessUserStore,
	serviceStore *store.ServiceStore,
	slotService *service.SlotService,
) *BusinessHandler {
	return &BusinessHandler{
		businessStore: businessStore,
		buStore:       buStore,
		serviceStore:  serviceStore,
		slotService:   slotService,
	}
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

	writeJSON(w, http.StatusOK, BusinessResponse{
		Business:  *b,
		Services:  services,
		Employees: employees,
	})
}

// GetServices godoc
// @Summary      List business services
// @Description  Returns all active services for a business.
// @Tags         public
// @Produce      json
// @Param        slug path string true "Business slug"
// @Success      200 {array} models.Service
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

	writeJSON(w, http.StatusOK, services)
}

// GetEmployees godoc
// @Summary      List business employees
// @Description  Returns all employees for a business.
// @Tags         public
// @Produce      json
// @Param        slug path string true "Business slug"
// @Success      200 {array} models.BusinessUser
// @Failure      404 {object} ErrorResponse
// @Router       /api/business/{slug}/employees [get]
func (h *BusinessHandler) GetEmployees(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	b, err := h.businessStore.GetBySlug(r.Context(), slug)
	if err != nil {
		writeError(w, http.StatusNotFound, "business not found")
		return
	}

	employees, err := h.buStore.ListEmployeesByBusiness(r.Context(), b.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get employees")
		return
	}

	writeJSON(w, http.StatusOK, employees)
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
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if slots == nil {
		slots = []models.TimeSlot{}
	}

	writeJSON(w, http.StatusOK, SlotsResponse{
		Slots: slots,
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
