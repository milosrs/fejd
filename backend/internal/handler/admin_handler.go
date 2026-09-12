package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"fejd-backend/internal/authutil"
	"fejd-backend/internal/db"
	"fejd-backend/internal/dto"
	"fejd-backend/internal/models"
	"fejd-backend/internal/service"
	"fejd-backend/internal/store"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
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
	pageStore           *store.PageStore
	sectionStore        *store.SectionStore
	businessHoursStore  *store.BusinessHoursStore
	workingHoursService *service.WorkingHoursService
	appointmentStore    *store.AppointmentStore
	slotService         *service.SlotService
	imageService        *service.ImageService
	employeeService     *service.EmployeeService
	pool                *pgxpool.Pool
}

func NewAdminHandler(
	businessStore *store.BusinessStore,
	buStore *store.BusinessUserStore,
	serviceStore *store.ServiceStore,
	pageStore *store.PageStore,
	sectionStore *store.SectionStore,
	businessHoursStore *store.BusinessHoursStore,
	workingHoursService *service.WorkingHoursService,
	appointmentStore *store.AppointmentStore,
	slotService *service.SlotService,
	imageService *service.ImageService,
	employeeService *service.EmployeeService,
	pool *pgxpool.Pool,
) *AdminHandler {
	return &AdminHandler{
		businessStore:       businessStore,
		buStore:             buStore,
		serviceStore:        serviceStore,
		pageStore:           pageStore,
		sectionStore:        sectionStore,
		businessHoursStore:  businessHoursStore,
		workingHoursService: workingHoursService,
		appointmentStore:    appointmentStore,
		slotService:         slotService,
		imageService:        imageService,
		employeeService:     employeeService,
		pool:                pool,
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
		WorkingHours: dto.WorkingHoursFromModels(hours),
		Overrides:    dto.WorkingHoursOverridesFromModels(overrides),
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

	hours := make([]models.WorkingHours, len(body.WorkingHours))
	for i, wh := range body.WorkingHours {
		hours[i] = models.WorkingHours{
			DayOfWeek: wh.DayOfWeek,
			StartTime: wh.StartTime,
			EndTime:   wh.EndTime,
		}
	}

	if err := h.workingHoursService.SetWeeklyHours(r.Context(), businessID, targetUserID, hours); err != nil {
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
// @Param        override body WorkingHoursOverrideInput true "Override details"
// @Success      201 {object} dto.WorkingHoursOverride
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

	var body WorkingHoursOverrideInput
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, errInvalidRequestBody.Error())
		return
	}

	override := models.WorkingHoursOverride{
		OverrideDate: body.OverrideDate,
		StartTime:    body.StartTime,
		EndTime:      body.EndTime,
		IsOff:        body.IsOff,
		Reason:       body.Reason,
	}

	if err := h.workingHoursService.AddOverride(r.Context(), businessID, targetUserID, &override); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, dto.WorkingHoursOverrideFromModel(override))
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
// @Param        service body ServiceInput true "Service details"
// @Success      201 {object} dto.Service
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

	var body ServiceInput
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, errInvalidRequestBody.Error())
		return
	}

	svc := &models.Service{
		BusinessID:      businessID,
		Name:            body.Name,
		DurationMinutes: body.DurationMinutes,
		Price:           body.Price,
		Active:          body.Active,
		Description:     body.Description,
		PictureID:       body.PictureID,
	}

	if err := h.serviceStore.Create(r.Context(), svc); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, dto.ServiceFromModel(*svc))
}

// UpdateService godoc
// @Summary      Update a service
// @Description  Updates an existing service for a business.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        serviceID path string true "Service UUID"
// @Param        service body ServiceInput true "Updated service details"
// @Success      200 {object} dto.Service
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

	var body ServiceInput
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, errInvalidRequestBody.Error())
		return
	}

	svc := &models.Service{
		ID:              serviceID,
		BusinessID:      businessID,
		Name:            body.Name,
		DurationMinutes: body.DurationMinutes,
		Price:           body.Price,
		Active:          body.Active,
		Description:     body.Description,
		PictureID:       body.PictureID,
	}

	if err := h.serviceStore.Update(r.Context(), svc); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, dto.ServiceFromModel(*svc))
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

	svc, err := h.serviceStore.GetByID(r.Context(), serviceID)
	if err != nil || svc.BusinessID != businessID {
		writeError(w, http.StatusNotFound, "service not found")
		return
	}

	if err := h.serviceStore.Delete(r.Context(), serviceID, businessID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			writeError(w, http.StatusConflict, "service has appointments and cannot be deleted")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := h.imageService.UnlinkAndMaybeDeleteAll(r.Context(), "service", serviceID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, MessageResponse{Message: "service deleted"})
}

// CreateEmployee godoc
// @Summary      Invite a new employee
// @Description  Creates a Keycloak user (invite email), a business_user row, and assigns services.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        body body CreateEmployeeRequest true "Employee"
// @Success      201 {object} dto.BusinessUser
// @Failure      400 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/employees [post]
func (h *AdminHandler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	businessID, err := uuid.Parse(chi.URLParam(r, "businessID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errInvalidBusinessID.Error())
		return
	}

	var body CreateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, errInvalidRequestBody.Error())
		return
	}

	bu, err := h.employeeService.Invite(r.Context(), businessID, body.Name, body.Email, body.ServiceIDs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, dto.BusinessUserFromModel(*bu))
}

// GetEmployees godoc
// @Summary      List all business users
// @Description  Returns all business users for a business (admin only).
// @Tags         admin
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Success      200 {array} dto.BusinessUser
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

	dtos := dto.BusinessUsersFromModels(users)
	if h.imageService != nil {
		for i := range dtos {
			dtos[i].Avatar = h.imageService.BusinessUserAvatarURL(r.Context(), users[i].ID)
		}
	}

	writeJSON(w, http.StatusOK, dtos)
}

// RemoveEmployee godoc
// @Summary      Remove an employee
// @Description  Soft-deletes an employee: future reservations are reassigned or cancelled.
// @Tags         admin
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        userID path string true "User ID (Keycloak sub)"
// @Success      200 {object} RemoveEmployeeResponse
// @Failure      400 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/employees/{userID} [delete]
func (h *AdminHandler) RemoveEmployee(w http.ResponseWriter, r *http.Request) {
	businessID, err := uuid.Parse(chi.URLParam(r, "businessID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errInvalidBusinessID.Error())
		return
	}

	targetUserID := chi.URLParam(r, "userID")

	bu, err := h.buStore.GetByBusinessAndUser(r.Context(), businessID, targetUserID)
	if err != nil {
		writeError(w, http.StatusNotFound, "employee not found")
		return
	}

	result, err := h.slotService.RemoveEmployee(r.Context(), businessID, bu.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, RemoveEmployeeResponse{
		Reassigned: result.Reassigned,
		Cancelled:  result.Cancelled,
		Message:    "employee removed",
	})
}

// SetEmployeeServices godoc
// @Summary      Set employee services
// @Description  Replaces the set of services an employee offers.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        userID path string true "User ID (Keycloak sub)"
// @Param        body body SetEmployeeServicesRequest true "Service IDs"
// @Success      200 {object} MessageResponse
// @Failure      400 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/employees/{userID}/services [put]
func (h *AdminHandler) SetEmployeeServices(w http.ResponseWriter, r *http.Request) {
	businessID, err := uuid.Parse(chi.URLParam(r, "businessID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errInvalidBusinessID.Error())
		return
	}

	targetUserID := chi.URLParam(r, "userID")

	var body SetEmployeeServicesRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, errInvalidRequestBody.Error())
		return
	}

	if err := h.slotService.SetEmployeeServices(r.Context(), businessID, targetUserID, body.ServiceIDs); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, MessageResponse{Message: "employee services updated"})
}

// GetServiceEmployees godoc
// @Summary      List employees mapped to a service
// @Description  Returns the business users (owner and employees) currently mapped to a service.
// @Tags         admin
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        serviceID path string true "Service UUID"
// @Success      200 {array} dto.BusinessUser
// @Failure      400 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/services/{serviceID}/employees [get]
func (h *AdminHandler) GetServiceEmployees(w http.ResponseWriter, r *http.Request) {
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

	users, err := h.slotService.ListServiceEmployees(r.Context(), businessID, serviceID)
	if err != nil {
		writeError(w, http.StatusNotFound, "service not found")
		return
	}
	if users == nil {
		users = []models.BusinessUser{}
	}

	writeJSON(w, http.StatusOK, dto.BusinessUsersFromModels(users))
}

// SetServiceEmployees godoc
// @Summary      Set service employees
// @Description  Replaces the set of employees mapped to a service.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        serviceID path string true "Service UUID"
// @Param        body body SetServiceEmployeesRequest true "Business user IDs"
// @Success      200 {object} MessageResponse
// @Failure      400 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/services/{serviceID}/employees [put]
func (h *AdminHandler) SetServiceEmployees(w http.ResponseWriter, r *http.Request) {
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

	var body SetServiceEmployeesRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, errInvalidRequestBody.Error())
		return
	}

	if err := h.slotService.SetServiceEmployees(r.Context(), businessID, serviceID, body.BusinessUserIDs); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, MessageResponse{Message: "service employees updated"})
}

// AddUnavailability godoc
// @Summary      Mark an employee unavailable
// @Description  Blocks a time range (e.g. vacation) for an employee.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        userID path string true "User ID (Keycloak sub)"
// @Param        body body CreateUnavailabilityRequest true "Unavailability range"
// @Success      201 {object} dto.EmployeeUnavailability
// @Failure      400 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/employees/{userID}/unavailability [post]
func (h *AdminHandler) AddUnavailability(w http.ResponseWriter, r *http.Request) {
	businessID, err := uuid.Parse(chi.URLParam(r, "businessID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errInvalidBusinessID.Error())
		return
	}

	targetUserID := chi.URLParam(r, "userID")

	var body CreateUnavailabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, errInvalidRequestBody.Error())
		return
	}

	startTime, err := time.Parse(time.RFC3339, body.StartTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid start_time format, use RFC3339")
		return
	}

	endTime, err := time.Parse(time.RFC3339, body.EndTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid end_time format, use RFC3339")
		return
	}

	u := &models.EmployeeUnavailability{
		StartTime: startTime,
		EndTime:   endTime,
		Reason:    body.Reason,
	}

	if err := h.slotService.AddEmployeeUnavailability(r.Context(), businessID, targetUserID, u); err != nil {
		writeUnavailabilityError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, dto.EmployeeUnavailabilityFromModel(*u))
}

// DeleteUnavailability godoc
// @Summary      Remove an unavailability block
// @Description  Deletes a previously created unavailability range.
// @Tags         admin
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        userID path string true "User ID (Keycloak sub)"
// @Param        unavailabilityID path string true "Unavailability UUID"
// @Success      200 {object} MessageResponse
// @Failure      400 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/employees/{userID}/unavailability/{unavailabilityID} [delete]
func (h *AdminHandler) DeleteUnavailability(w http.ResponseWriter, r *http.Request) {
	businessID, err := uuid.Parse(chi.URLParam(r, "businessID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errInvalidBusinessID.Error())
		return
	}

	targetUserID := chi.URLParam(r, "userID")

	unavailabilityID, err := uuid.Parse(chi.URLParam(r, "unavailabilityID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid unavailability ID")
		return
	}

	if err := h.slotService.DeleteEmployeeUnavailability(r.Context(), businessID, targetUserID, unavailabilityID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, MessageResponse{Message: "unavailability deleted"})
}

// ListMyUnavailability godoc
// @Summary      List my blocked time slots
// @Description  Returns the blocked time ranges the authenticated member (owner or employee) reserved for themselves.
// @Tags         admin
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Success      200 {array} dto.EmployeeUnavailability
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/me/unavailability [get]
func (h *AdminHandler) ListMyUnavailability(w http.ResponseWriter, r *http.Request) {
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

	unavail, err := h.slotService.ListEmployeeUnavailability(r.Context(), businessID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, dto.EmployeeUnavailabilitysFromModels(unavail))
}

// AddMyUnavailability godoc
// @Summary      Reserve one of my time slots
// @Description  Blocks a time range for the authenticated member so it is not offered to customers.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        body body CreateUnavailabilityRequest true "Blocked range"
// @Success      201 {object} dto.EmployeeUnavailability
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/me/unavailability [post]
func (h *AdminHandler) AddMyUnavailability(w http.ResponseWriter, r *http.Request) {
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

	var body CreateUnavailabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, errInvalidRequestBody.Error())
		return
	}

	startTime, err := time.Parse(time.RFC3339, body.StartTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid start_time format, use RFC3339")
		return
	}

	endTime, err := time.Parse(time.RFC3339, body.EndTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid end_time format, use RFC3339")
		return
	}

	u := &models.EmployeeUnavailability{
		StartTime: startTime,
		EndTime:   endTime,
		Reason:    body.Reason,
		Status:    models.UnavailabilityStatusPending,
	}

	// An owner reserving their own time is auto-accepted; an employee's
	// reservation waits for the owner to acknowledge it.
	if isAdmin, err := h.buStore.IsAdmin(r.Context(), businessID, userID); err == nil && isAdmin {
		u.Status = models.UnavailabilityStatusConfirmed
	}

	if err := h.slotService.AddEmployeeUnavailability(r.Context(), businessID, userID, u); err != nil {
		writeUnavailabilityError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, dto.EmployeeUnavailabilityFromModel(*u))
}

// DeleteMyUnavailability godoc
// @Summary      Remove one of my blocked time slots
// @Description  Deletes a blocked time range the authenticated member reserved for themselves.
// @Tags         admin
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        unavailabilityID path string true "Unavailability UUID"
// @Success      200 {object} MessageResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/me/unavailability/{unavailabilityID} [delete]
func (h *AdminHandler) DeleteMyUnavailability(w http.ResponseWriter, r *http.Request) {
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

	unavailabilityID, err := uuid.Parse(chi.URLParam(r, "unavailabilityID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid unavailability ID")
		return
	}

	if err := h.slotService.DeleteOwnEmployeeUnavailability(r.Context(), businessID, userID, unavailabilityID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, MessageResponse{Message: "unavailability deleted"})
}

// ListMyReservations godoc
// @Summary      List my reservations for a date
// @Description  Returns the authenticated member's reservations (appointments) for a given date, with status and service name.
// @Tags         admin
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        date query string true "Date (YYYY-MM-DD)"
// @Success      200 {array} dto.Appointment
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/me/appointments [get]
func (h *AdminHandler) ListMyReservations(w http.ResponseWriter, r *http.Request) {
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

	date, err := time.Parse(time.DateOnly, r.URL.Query().Get("date"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid date format, use YYYY-MM-DD")
		return
	}

	appointments, err := h.slotService.ListOwnAppointments(r.Context(), businessID, userID, date)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if appointments == nil {
		appointments = []models.Appointment{}
	}

	serviceNames, err := h.serviceNameMap(r.Context(), businessID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	noShowAfterHours := 0
	if b, err := h.businessStore.GetByID(r.Context(), businessID); err == nil {
		noShowAfterHours = b.NoShowAfterHours
	}

	writeJSON(w, http.StatusOK, dto.StaffAppointmentsFromModels(appointments, serviceNames, noShowAfterHours))
}

// CancelMyReservation godoc
// @Summary      Cancel one of my reservations
// @Description  Cancels the authenticated member's own reservation. A cancellation reason is required.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        appointmentID path string true "Appointment UUID"
// @Param        body body CancelAppointmentRequest true "Cancellation reason"
// @Success      200 {object} MessageResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/me/appointments/{appointmentID} [delete]
func (h *AdminHandler) CancelMyReservation(w http.ResponseWriter, r *http.Request) {
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

	appointmentID, err := uuid.Parse(chi.URLParam(r, "appointmentID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid appointment ID")
		return
	}

	var body CancelAppointmentRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, errInvalidRequestBody.Error())
		return
	}

	reason := strings.TrimSpace(body.CancellationReason)
	if reason == "" {
		writeError(w, http.StatusBadRequest, "cancellation_reason is required")
		return
	}

	if err := h.slotService.CancelOwnAppointment(r.Context(), businessID, userID, appointmentID, reason); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, MessageResponse{Message: "reservation cancelled"})
}

// AcceptAppointment godoc
// @Summary      Accept a customer appointment
// @Description  Acknowledges a pending appointment. Allowed for the appointment's provider or the business owner.
// @Tags         admin
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        appointmentID path string true "Appointment UUID"
// @Success      200 {object} MessageResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/appointments/{appointmentID}/accept [post]
func (h *AdminHandler) AcceptAppointment(w http.ResponseWriter, r *http.Request) {
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

	appointmentID, err := uuid.Parse(chi.URLParam(r, "appointmentID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid appointment ID")
		return
	}

	if err := h.slotService.AcceptAppointment(r.Context(), businessID, userID, appointmentID); err != nil {
		if errors.Is(err, service.ErrForbidden) {
			writeError(w, http.StatusForbidden, "not allowed to accept this appointment")
			return
		}
		if errors.Is(err, service.ErrAppointmentNotFound) {
			writeError(w, http.StatusNotFound, "appointment not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, MessageResponse{Message: "appointment accepted"})
}

// AcceptUnavailability godoc
// @Summary      Accept an employee's reserved time
// @Description  Acknowledges a pending blocked-time reservation. Only the business owner may accept it.
// @Tags         admin
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        unavailabilityID path string true "Unavailability UUID"
// @Success      200 {object} MessageResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/unavailability/{unavailabilityID}/accept [post]
func (h *AdminHandler) AcceptUnavailability(w http.ResponseWriter, r *http.Request) {
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

	unavailabilityID, err := uuid.Parse(chi.URLParam(r, "unavailabilityID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid unavailability ID")
		return
	}

	if err := h.slotService.AcceptUnavailability(r.Context(), businessID, userID, unavailabilityID); err != nil {
		if errors.Is(err, service.ErrForbidden) {
			writeError(w, http.StatusForbidden, "only the owner can accept time reservations")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, MessageResponse{Message: "reservation accepted"})
}

// RejectAppointment godoc
// @Summary      Reject a customer appointment
// @Description  Rejects (cancels) a pending appointment with a stated reason. Allowed for the appointment's provider or the business owner.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        appointmentID path string true "Appointment UUID"
// @Param        body body RejectRequest true "Rejection reason"
// @Success      200 {object} MessageResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/appointments/{appointmentID}/reject [post]
func (h *AdminHandler) RejectAppointment(w http.ResponseWriter, r *http.Request) {
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

	appointmentID, err := uuid.Parse(chi.URLParam(r, "appointmentID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid appointment ID")
		return
	}

	var body RejectRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, errInvalidRequestBody.Error())
		return
	}

	reason := strings.TrimSpace(body.Reason)
	if reason == "" {
		writeError(w, http.StatusBadRequest, "reason is required")
		return
	}

	if err := h.slotService.RejectAppointment(r.Context(), businessID, userID, appointmentID, reason); err != nil {
		if errors.Is(err, service.ErrForbidden) {
			writeError(w, http.StatusForbidden, "not allowed to reject this appointment")
			return
		}
		if errors.Is(err, service.ErrAppointmentNotFound) {
			writeError(w, http.StatusNotFound, "appointment not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, MessageResponse{Message: "appointment rejected"})
}

// RejectUnavailability godoc
// @Summary      Reject an employee's reserved time
// @Description  Rejects a pending blocked-time reservation with a stated reason. Only the business owner may reject it.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        unavailabilityID path string true "Unavailability UUID"
// @Param        body body RejectRequest true "Rejection reason"
// @Success      200 {object} MessageResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/unavailability/{unavailabilityID}/reject [post]
func (h *AdminHandler) RejectUnavailability(w http.ResponseWriter, r *http.Request) {
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

	unavailabilityID, err := uuid.Parse(chi.URLParam(r, "unavailabilityID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid unavailability ID")
		return
	}

	var body RejectRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, errInvalidRequestBody.Error())
		return
	}

	reason := strings.TrimSpace(body.Reason)
	if reason == "" {
		writeError(w, http.StatusBadRequest, "reason is required")
		return
	}

	if err := h.slotService.RejectUnavailability(r.Context(), businessID, userID, unavailabilityID, reason); err != nil {
		if errors.Is(err, service.ErrForbidden) {
			writeError(w, http.StatusForbidden, "only the owner can reject time reservations")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, MessageResponse{Message: "reservation rejected"})
}

// ListBusinessAppointments godoc
// @Summary      List all salon appointments
// @Description  Returns every appointment in the salon with service name, for the owner to review and acknowledge pending bookings.
// @Tags         admin
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Success      200 {array} dto.Appointment
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/appointments [get]
func (h *AdminHandler) ListBusinessAppointments(w http.ResponseWriter, r *http.Request) {
	businessID, err := uuid.Parse(chi.URLParam(r, "businessID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errInvalidBusinessID.Error())
		return
	}

	appointments, err := h.appointmentStore.ListByBusiness(r.Context(), businessID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if appointments == nil {
		appointments = []models.Appointment{}
	}

	serviceNames, err := h.serviceNameMap(r.Context(), businessID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	noShowAfterHours := 0
	if b, err := h.businessStore.GetByID(r.Context(), businessID); err == nil {
		noShowAfterHours = b.NoShowAfterHours
	}

	writeJSON(w, http.StatusOK, dto.StaffAppointmentsFromModels(appointments, serviceNames, noShowAfterHours))
}

// ListBusinessUnavailability godoc
// @Summary      List all reserved time
// @Description  Returns every blocked-time reservation in the salon, for the owner to review and acknowledge pending employee reservations.
// @Tags         admin
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Success      200 {array} dto.EmployeeUnavailability
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/unavailability [get]
func (h *AdminHandler) ListBusinessUnavailability(w http.ResponseWriter, r *http.Request) {
	businessID, err := uuid.Parse(chi.URLParam(r, "businessID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errInvalidBusinessID.Error())
		return
	}

	unavail, err := h.slotService.ListBusinessUnavailability(r.Context(), businessID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if unavail == nil {
		unavail = []models.EmployeeUnavailability{}
	}

	writeJSON(w, http.StatusOK, dto.EmployeeUnavailabilitysFromModels(unavail))
}

func (h *AdminHandler) serviceNameMap(ctx context.Context, businessID uuid.UUID) (map[uuid.UUID]string, error) {
	services, err := h.serviceStore.ListByBusiness(ctx, businessID)
	if err != nil {
		return nil, err
	}
	names := make(map[uuid.UUID]string, len(services))
	for _, svc := range services {
		names[svc.ID] = svc.Name
	}
	return names, nil
}

// GetSalonPolicy godoc
// @Summary      Get salon policy
// @Description  Returns the salon's cancellation, no-show, slot interval and working-hours policies.
// @Tags         admin
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Success      200 {object} SalonPolicyResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/policy [get]
func (h *AdminHandler) GetSalonPolicy(w http.ResponseWriter, r *http.Request) {
	businessID, err := uuid.Parse(chi.URLParam(r, "businessID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errInvalidBusinessID.Error())
		return
	}

	b, err := h.businessStore.GetByID(r.Context(), businessID)
	if err != nil {
		writeError(w, http.StatusNotFound, "business not found")
		return
	}

	hours, err := h.businessHoursStore.ListByBusiness(r.Context(), businessID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, SalonPolicyResponse{
		CancellationLeadHours: b.CancellationLeadHours,
		NoShowAfterHours:      b.NoShowAfterHours,
		SlotIntervalMinutes:   b.SlotIntervalMinutes,
		WorkingHours:          dto.BusinessHoursFromModels(hours),
	})
}

// UpdateSalonPolicy godoc
// @Summary      Update salon policy
// @Description  Sets the salon's cancellation, no-show, slot interval and working-hours policies.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        body body UpdateSalonPolicyRequest true "Policy"
// @Success      200 {object} SalonPolicyResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/policy [put]
func (h *AdminHandler) UpdateSalonPolicy(w http.ResponseWriter, r *http.Request) {
	businessID, err := uuid.Parse(chi.URLParam(r, "businessID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errInvalidBusinessID.Error())
		return
	}

	var body UpdateSalonPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, errInvalidRequestBody.Error())
		return
	}

	if body.CancellationLeadHours < 0 {
		writeError(w, http.StatusBadRequest, "cancellation_lead_hours must be zero or greater")
		return
	}
	if body.NoShowAfterHours < 0 {
		writeError(w, http.StatusBadRequest, "no_show_after_hours must be zero or greater")
		return
	}
	if body.SlotIntervalMinutes <= 0 {
		writeError(w, http.StatusBadRequest, "slot_interval_minutes must be greater than zero")
		return
	}

	hours := make([]models.BusinessHours, 0, len(body.WorkingHours))
	for _, wh := range body.WorkingHours {
		start, err := parseTimeOnly(wh.StartTime)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		end, err := parseTimeOnly(wh.EndTime)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if !end.After(start) {
			writeError(w, http.StatusBadRequest, "end_time must be after start_time")
			return
		}
		hours = append(hours, models.BusinessHours{
			BusinessID: businessID,
			DayOfWeek:  wh.DayOfWeek,
			StartTime:  start,
			EndTime:    end,
		})
	}

	if err := h.businessStore.UpdatePolicy(r.Context(), businessID, body.CancellationLeadHours, body.NoShowAfterHours, body.SlotIntervalMinutes); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := h.businessHoursStore.ReplaceByBusiness(r.Context(), businessID, hours); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, SalonPolicyResponse{
		CancellationLeadHours: body.CancellationLeadHours,
		NoShowAfterHours:      body.NoShowAfterHours,
		SlotIntervalMinutes:   body.SlotIntervalMinutes,
		WorkingHours:          dto.BusinessHoursFromModels(hours),
	})
}

// RenameBusiness godoc
// @Summary      Rename a salon
// @Description  Updates the salon's display name and regenerates its slug from the new name.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        body body RenameBusinessRequest true "New name"
// @Success      200 {object} dto.Business
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      409 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/name [put]
func (h *AdminHandler) RenameBusiness(w http.ResponseWriter, r *http.Request) {
	businessID, err := uuid.Parse(chi.URLParam(r, "businessID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errInvalidBusinessID.Error())
		return
	}

	var body RenameBusinessRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, errInvalidRequestBody.Error())
		return
	}

	name := strings.TrimSpace(body.Name)
	if name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	business, err := h.businessStore.GetByID(r.Context(), businessID)
	if err != nil {
		writeError(w, http.StatusNotFound, "business not found")
		return
	}

	newSlug := sanitizeSlug(slugify(name))
	if newSlug != business.Slug {
		base := newSlug
		for i := 2; ; i++ {
			taken, err := h.businessStore.SlugTakenByOther(r.Context(), h.pool, newSlug, businessID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to generate slug")
				return
			}
			if !taken {
				break
			}
			newSlug = fmt.Sprintf("%s-%d", base, i)
		}
	}

	if err := h.businessStore.Rename(r.Context(), h.pool, businessID, name, newSlug); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			writeError(w, http.StatusConflict, "a salon with this name already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to rename business")
		return
	}

	business.Name = name
	business.Slug = newSlug
	writeJSON(w, http.StatusOK, dto.BusinessFromModel(*business))
}

// parseTimeOnly parses an "HH:MM" or "HH:MM:SS" string into a time value.
func parseTimeOnly(s string) (time.Time, error) {
	for _, layout := range []string{time.TimeOnly, "15:04"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid time: %s", s)
}

// writeUnavailabilityError maps an unavailability write error to a friendly
// HTTP response (e.g. the exclusion constraint fires on overlapping ranges).
func writeUnavailabilityError(w http.ResponseWriter, err error) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23P01" {
		writeError(w, http.StatusConflict, "this time range overlaps an existing unavailability period")
		return
	}
	writeError(w, http.StatusInternalServerError, err.Error())
}

// MarkNoShow godoc
// @Summary      Mark one of my reservations as no-show
// @Description  Marks the authenticated member's own reservation as no-show, once the salon's configured grace period has passed.
// @Tags         admin
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        appointmentID path string true "Appointment UUID"
// @Success      200 {object} MessageResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      409 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/me/appointments/{appointmentID}/no-show [post]
func (h *AdminHandler) MarkNoShow(w http.ResponseWriter, r *http.Request) {
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

	appointmentID, err := uuid.Parse(chi.URLParam(r, "appointmentID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid appointment ID")
		return
	}

	if err := h.slotService.MarkNoShow(r.Context(), businessID, userID, appointmentID); err != nil {
		if errors.Is(err, service.ErrNoShowTooEarly) {
			writeError(w, http.StatusConflict, "no-show can only be marked after the configured grace period")
			return
		}
		if errors.Is(err, service.ErrAppointmentNotFound) {
			writeError(w, http.StatusNotFound, "appointment not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, MessageResponse{Message: "appointment marked as no-show"})
}

// ListMyServices godoc
// @Summary      List services I offer
// @Description  Returns the active services the authenticated member offers.
// @Tags         admin
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Success      200 {array} dto.Service
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/me/services [get]
func (h *AdminHandler) ListMyServices(w http.ResponseWriter, r *http.Request) {
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

	services, err := h.slotService.ListMyServices(r.Context(), businessID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, dto.ServicesFromModels(services))
}

// BookOwnAppointment godoc
// @Summary      Add an appointment to my calendar
// @Description  Books an appointment on the authenticated member's own calendar, optionally for an existing customer (walk-in if omitted).
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        body body CreateOwnAppointmentRequest true "Appointment details"
// @Success      201 {object} dto.Appointment
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      409 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/me/appointments [post]
func (h *AdminHandler) BookOwnAppointment(w http.ResponseWriter, r *http.Request) {
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

	var body CreateOwnAppointmentRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, errInvalidRequestBody.Error())
		return
	}

	if body.ServiceID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "invalid service_id")
		return
	}

	startTime, err := time.Parse(time.RFC3339, body.StartTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid start_time format, use RFC3339")
		return
	}

	appointment, err := h.slotService.BookOwnAppointment(r.Context(), businessID, userID, body.ServiceID, startTime, body.CustomerUserID)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, dto.AppointmentFromModel(*appointment))
}

// ListCustomers godoc
// @Summary      List existing customers
// @Description  Returns the distinct customers who have booked with the business, with their locally-cached display names.
// @Tags         admin
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Success      200 {array} dto.Customer
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/customers [get]
func (h *AdminHandler) ListCustomers(w http.ResponseWriter, r *http.Request) {
	businessID, err := uuid.Parse(chi.URLParam(r, "businessID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errInvalidBusinessID.Error())
		return
	}

	customers, err := h.appointmentStore.ListCustomers(r.Context(), businessID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if customers == nil {
		customers = []models.Customer{}
	}

	writeJSON(w, http.StatusOK, dto.CustomersFromModels(customers))
}

// CreateSection godoc
// @Summary      Add a landing page section
// @Description  Creates a new section on the salon's landing page, creating the page if needed.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        body body CreateSectionRequest true "Section"
// @Success      201 {object} dto.Section
// @Failure      400 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/sections [post]
func (h *AdminHandler) CreateSection(w http.ResponseWriter, r *http.Request) {
	businessID, err := uuid.Parse(chi.URLParam(r, "businessID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errInvalidBusinessID.Error())
		return
	}

	var body CreateSectionRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, errInvalidRequestBody.Error())
		return
	}
	if !models.IsValidSectionType(body.Type) {
		writeError(w, http.StatusBadRequest, "invalid section type")
		return
	}

	content := body.Content
	if len(content) == 0 {
		content = json.RawMessage(`{}`)
	}

	page, err := h.ensureLandingPage(r.Context(), businessID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load landing page")
		return
	}

	sections, err := h.sectionStore.ListByPage(r.Context(), page.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list sections")
		return
	}

	position := body.Position
	if position == 0 {
		position = len(sections)
	}

	section := &models.Section{
		PageID:   page.ID,
		Type:     body.Type,
		Content:  []byte(content),
		Position: position,
	}
	if err := h.sectionStore.Create(r.Context(), h.pool, section); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create section")
		return
	}

	writeJSON(w, http.StatusCreated, dto.SectionFromModel(*section))
}

// UpdateSection godoc
// @Summary      Update a landing page section's content
// @Description  Replaces the content JSON of a section on the salon's landing page.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        sectionID path string true "Section UUID"
// @Param        body body UpdateSectionRequest true "Section content"
// @Success      200 {object} dto.Section
// @Failure      400 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/sections/{sectionID} [put]
func (h *AdminHandler) UpdateSection(w http.ResponseWriter, r *http.Request) {
	businessID, err := uuid.Parse(chi.URLParam(r, "businessID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errInvalidBusinessID.Error())
		return
	}

	sectionID, err := uuid.Parse(chi.URLParam(r, "sectionID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid section ID")
		return
	}

	var body UpdateSectionRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, errInvalidRequestBody.Error())
		return
	}

	section, err := h.ownedSection(r.Context(), businessID, sectionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "section not found")
		return
	}

	if err := h.sectionStore.UpdateContent(r.Context(), h.pool, sectionID, []byte(body.Content)); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update section")
		return
	}
	section.Content = []byte(body.Content)

	writeJSON(w, http.StatusOK, dto.SectionFromModel(*section))
}

// DeleteSection godoc
// @Summary      Remove a landing page section
// @Description  Deletes a section from the salon's landing page.
// @Tags         admin
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        sectionID path string true "Section UUID"
// @Success      200 {object} MessageResponse
// @Failure      400 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/sections/{sectionID} [delete]
func (h *AdminHandler) DeleteSection(w http.ResponseWriter, r *http.Request) {
	businessID, err := uuid.Parse(chi.URLParam(r, "businessID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errInvalidBusinessID.Error())
		return
	}

	sectionID, err := uuid.Parse(chi.URLParam(r, "sectionID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid section ID")
		return
	}

	if _, err := h.ownedSection(r.Context(), businessID, sectionID); err != nil {
		writeError(w, http.StatusNotFound, "section not found")
		return
	}

	if err := h.sectionStore.Delete(r.Context(), h.pool, sectionID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete section")
		return
	}

	writeJSON(w, http.StatusOK, MessageResponse{Message: "section deleted"})
}

// ReorderSections godoc
// @Summary      Reorder landing page sections
// @Description  Sets the position of every landing page section from the given ordered list of section IDs.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        body body ReorderSectionsRequest true "Ordered section IDs"
// @Success      200 {array} dto.Section
// @Failure      400 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/sections/reorder [put]
func (h *AdminHandler) ReorderSections(w http.ResponseWriter, r *http.Request) {
	businessID, err := uuid.Parse(chi.URLParam(r, "businessID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errInvalidBusinessID.Error())
		return
	}

	var body ReorderSectionsRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, errInvalidRequestBody.Error())
		return
	}

	page, err := h.pageStore.GetByBusinessAndName(r.Context(), businessID, store.LandingPageName)
	if err != nil {
		writeJSON(w, http.StatusOK, dto.SectionsFromModels(nil))
		return
	}

	owned, err := h.sectionStore.ListByPage(r.Context(), page.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list sections")
		return
	}
	ownedIDs := make(map[uuid.UUID]bool, len(owned))
	for _, s := range owned {
		ownedIDs[s.ID] = true
	}
	for _, id := range body.SectionIDs {
		if !ownedIDs[id] {
			writeError(w, http.StatusBadRequest, "section does not belong to this business")
			return
		}
	}

	err = db.WithTx(r.Context(), h.pool, func(tx pgx.Tx) error {
		for i, id := range body.SectionIDs {
			if err := h.sectionStore.UpdatePosition(r.Context(), tx, id, i); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to reorder sections")
		return
	}

	sections, err := h.sectionStore.ListByPage(r.Context(), page.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list sections")
		return
	}
	writeJSON(w, http.StatusOK, dto.SectionsFromModels(sections))
}
