package handler

import (
	"encoding/json"
	"errors"
	"fejd-backend/internal/db"
	"fejd-backend/internal/dto"
	"fejd-backend/internal/models"
	"fejd-backend/internal/service"
	"fejd-backend/internal/store"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
	workingHoursService *service.WorkingHoursService
	appointmentStore    *store.AppointmentStore
	slotService         *service.SlotService
	imageService        *service.ImageService
	pool                *pgxpool.Pool
}

func NewAdminHandler(
	businessStore *store.BusinessStore,
	buStore *store.BusinessUserStore,
	serviceStore *store.ServiceStore,
	pageStore *store.PageStore,
	sectionStore *store.SectionStore,
	workingHoursService *service.WorkingHoursService,
	appointmentStore *store.AppointmentStore,
	slotService *service.SlotService,
	imageService *service.ImageService,
	pool *pgxpool.Pool,
) *AdminHandler {
	return &AdminHandler{
		businessStore:       businessStore,
		buStore:             buStore,
		serviceStore:        serviceStore,
		pageStore:           pageStore,
		sectionStore:        sectionStore,
		workingHoursService: workingHoursService,
		appointmentStore:    appointmentStore,
		slotService:         slotService,
		imageService:        imageService,
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
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := h.imageService.UnlinkAndMaybeDeleteAll(r.Context(), "service", serviceID); err != nil {
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

	writeJSON(w, http.StatusOK, dto.BusinessUsersFromModels(users))
}

// RemoveEmployee godoc
// @Summary      Remove an employee
// @Description  Soft-deletes an employee: future reservations are reassigned or cancelled.
// @Tags         admin
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        userID path string true "User ID (Keycloak sub)"
// @Success      200 {object} MessageResponse
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

	if err := h.slotService.RemoveEmployee(r.Context(), businessID, bu.ID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, MessageResponse{Message: "employee removed"})
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
		writeError(w, http.StatusInternalServerError, err.Error())
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
