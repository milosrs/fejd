package handler

import (
	"encoding/json"
	"errors"
	"fejd-backend/internal/models"
	"fejd-backend/internal/service"
	"fejd-backend/internal/store"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

var (
	errInvalidBusinessID  = errors.New("invalid business ID")
	errInvalidRequestBody = errors.New("invalid request body")
	errInvalidServiceID   = errors.New("invalid service ID")
)

type AdminHandler struct {
	businessStore       *store.BusinessStore
	buStore             *store.BusinessUserStore
	serviceStore        *store.ServiceStore
	workingHoursService *service.WorkingHoursService
	appointmentStore    *store.AppointmentStore
}

func NewAdminHandler(
	businessStore *store.BusinessStore,
	buStore *store.BusinessUserStore,
	serviceStore *store.ServiceStore,
	workingHoursService *service.WorkingHoursService,
	appointmentStore *store.AppointmentStore,
) *AdminHandler {
	return &AdminHandler{
		businessStore:       businessStore,
		buStore:             buStore,
		serviceStore:        serviceStore,
		workingHoursService: workingHoursService,
		appointmentStore:    appointmentStore,
	}
}

// GetWorkingHours godoc
// @Summary      Get employee working hours
// @Description  Returns weekly working hours and date overrides for an employee.
// @Tags         admin
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        userID path string true "User ID (Keycloak sub)"
// @Success      200 {object} WorkingHoursResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/employees/{userID}/working-hours [get]
func (h *AdminHandler) GetWorkingHours(w http.ResponseWriter, r *http.Request) {
	businessID, err := uuid.Parse(chi.URLParam(r, "businessID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errInvalidBusinessID.Error())
		return
	}

	targetUserID := chi.URLParam(r, "userID")

	hours, err := h.workingHoursService.GetWeeklyHours(r.Context(), businessID, targetUserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	from := time.Now().AddDate(0, -3, 0)
	to := time.Now().AddDate(0, 3, 0)
	overrides, _ := h.workingHoursService.GetOverrides(r.Context(), businessID, targetUserID, from, to)

	writeJSON(w, http.StatusOK, WorkingHoursResponse{
		WorkingHours: hours,
		Overrides:    overrides,
	})
}

// SetWorkingHours godoc
// @Summary      Set employee working hours
// @Description  Replaces the weekly working hours for an employee.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        userID path string true "User ID (Keycloak sub)"
// @Param        body body SetWorkingHoursRequest true "Weekly working hours"
// @Success      200 {object} MessageResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/employees/{userID}/working-hours [put]
func (h *AdminHandler) SetWorkingHours(w http.ResponseWriter, r *http.Request) {
	businessID, err := uuid.Parse(chi.URLParam(r, "businessID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errInvalidBusinessID.Error())
		return
	}

	targetUserID := chi.URLParam(r, "userID")

	var body SetWorkingHoursRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, errInvalidRequestBody.Error())
		return
	}

	if err := h.workingHoursService.SetWeeklyHours(r.Context(), businessID, targetUserID, body.WorkingHours); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, MessageResponse{Message: "working hours updated"})
}

// AddOverride godoc
// @Summary      Add a date override
// @Description  Adds a working hours override (e.g. holiday) for an employee.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        userID path string true "User ID (Keycloak sub)"
// @Param        override body models.WorkingHoursOverride true "Override details"
// @Success      201 {object} models.WorkingHoursOverride
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/employees/{userID}/overrides [post]
func (h *AdminHandler) AddOverride(w http.ResponseWriter, r *http.Request) {
	businessID, err := uuid.Parse(chi.URLParam(r, "businessID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errInvalidBusinessID.Error())
		return
	}

	targetUserID := chi.URLParam(r, "userID")

	var override models.WorkingHoursOverride
	if err := json.NewDecoder(r.Body).Decode(&override); err != nil {
		writeError(w, http.StatusBadRequest, errInvalidRequestBody.Error())
		return
	}

	if err := h.workingHoursService.AddOverride(r.Context(), businessID, targetUserID, &override); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, override)
}

// DeleteOverride godoc
// @Summary      Delete a date override
// @Description  Removes a working hours override for an employee.
// @Tags         admin
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        userID path string true "User ID (Keycloak sub)"
// @Param        overrideID path string true "Override UUID"
// @Success      200 {object} MessageResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/employees/{userID}/overrides/{overrideID} [delete]
func (h *AdminHandler) DeleteOverride(w http.ResponseWriter, r *http.Request) {
	businessID, err := uuid.Parse(chi.URLParam(r, "businessID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errInvalidBusinessID.Error())
		return
	}

	overrideID, err := uuid.Parse(chi.URLParam(r, "overrideID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid override ID")
		return
	}

	if err := h.workingHoursService.DeleteOverride(r.Context(), businessID, overrideID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, MessageResponse{Message: "override deleted"})
}

// CreateService godoc
// @Summary      Create a service
// @Description  Adds a new service to a business.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        service body models.Service true "Service details"
// @Success      201 {object} models.Service
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/services [post]
func (h *AdminHandler) CreateService(w http.ResponseWriter, r *http.Request) {
	businessID, err := uuid.Parse(chi.URLParam(r, "businessID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errInvalidBusinessID.Error())
		return
	}

	var svc models.Service
	if err := json.NewDecoder(r.Body).Decode(&svc); err != nil {
		writeError(w, http.StatusBadRequest, errInvalidRequestBody.Error())
		return
	}

	svc.BusinessID = businessID
	if err := h.serviceStore.Create(r.Context(), &svc); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, svc)
}

// UpdateService godoc
// @Summary      Update a service
// @Description  Updates an existing service for a business.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        serviceID path string true "Service UUID"
// @Param        service body models.Service true "Updated service details"
// @Success      200 {object} models.Service
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/services/{serviceID} [put]
func (h *AdminHandler) UpdateService(w http.ResponseWriter, r *http.Request) {
	businessID, err := uuid.Parse(chi.URLParam(r, "businessID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errInvalidBusinessID.Error())
		return
	}

	serviceID, err := uuid.Parse(chi.URLParam(r, "serviceID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errInvalidServiceID.Error())
		return
	}

	var svc models.Service
	if err := json.NewDecoder(r.Body).Decode(&svc); err != nil {
		writeError(w, http.StatusBadRequest, errInvalidRequestBody.Error())
		return
	}

	svc.ID = serviceID
	svc.BusinessID = businessID
	if err := h.serviceStore.Update(r.Context(), &svc); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, svc)
}

// DeleteService godoc
// @Summary      Delete a service
// @Description  Removes a service from a business.
// @Tags         admin
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        serviceID path string true "Service UUID"
// @Success      200 {object} MessageResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/services/{serviceID} [delete]
func (h *AdminHandler) DeleteService(w http.ResponseWriter, r *http.Request) {
	businessID, err := uuid.Parse(chi.URLParam(r, "businessID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errInvalidBusinessID.Error())
		return
	}

	serviceID, err := uuid.Parse(chi.URLParam(r, "serviceID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errInvalidServiceID.Error())
		return
	}

	if err := h.serviceStore.Delete(r.Context(), serviceID, businessID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, MessageResponse{Message: "service deleted"})
}

// GetEmployees godoc
// @Summary      List all business users
// @Description  Returns all business users for a business (admin only).
// @Tags         admin
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Success      200 {array} models.BusinessUser
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/employees [get]
func (h *AdminHandler) GetEmployees(w http.ResponseWriter, r *http.Request) {
	businessID, err := uuid.Parse(chi.URLParam(r, "businessID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errInvalidBusinessID.Error())
		return
	}

	users, err := h.buStore.ListByBusiness(r.Context(), businessID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, users)
}
