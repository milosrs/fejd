package handler

import (
	"encoding/json"
	"fejd-backend/internal/authutil"
	"fejd-backend/internal/models"
	"fejd-backend/internal/service"
	"fejd-backend/internal/store"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type AppointmentHandler struct {
	appointments *store.AppointmentStore
	services     *store.ServiceStore
	business     *store.BusinessStore
	buStore      *store.BusinessUserStore
	slotService  *service.SlotService
}

func NewAppointmentHandler(
	appointments *store.AppointmentStore,
	services *store.ServiceStore,
	business *store.BusinessStore,
	buStore *store.BusinessUserStore,
	slotService *service.SlotService,
) *AppointmentHandler {
	return &AppointmentHandler{
		appointments: appointments,
		services:     services,
		business:     business,
		buStore:      buStore,
		slotService:  slotService,
	}
}

// Create godoc
// @Summary      Book an appointment
// @Description  Creates a new appointment for the authenticated customer.
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Param        appointment body CreateAppointmentRequest true "Appointment details"
// @Success      201 {object} models.Appointment
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      409 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/appointments [post]
func (h *AppointmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := authutil.GetUserID(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var body CreateAppointmentRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	businessID, err := uuid.Parse(body.BusinessID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid business_id")
		return
	}

	serviceID, err := uuid.Parse(body.ServiceID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid service_id")
		return
	}

	businessUserID, err := uuid.Parse(body.BusinessUserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid business_user_id")
		return
	}

	startTime, err := time.Parse(time.RFC3339, body.StartTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid start_time format, use RFC3339")
		return
	}

	svc, err := h.services.GetByID(r.Context(), serviceID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "service not found")
		return
	}

	appointment := &models.Appointment{
		BusinessID:     businessID,
		ServiceID:      serviceID,
		BusinessUserID: businessUserID,
		CustomerUserID: userID,
		StartTime:      startTime,
		EndTime:        startTime.Add(time.Duration(svc.DurationMinutes) * time.Minute),
		Status:         models.AppointmentStatusConfirmed,
	}

	if err := h.slotService.BookAppointment(r.Context(), appointment); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, appointment)
}

// ListMyAppointments godoc
// @Summary      List customer appointments
// @Description  Returns all appointments for the authenticated customer.
// @Tags         appointments
// @Produce      json
// @Success      200 {array} models.Appointment
// @Failure      401 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/my/appointments [get]
func (h *AppointmentHandler) ListMyAppointments(w http.ResponseWriter, r *http.Request) {
	userID, err := authutil.GetUserID(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	appointments, err := h.appointments.ListByCustomer(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, appointments)
}

// Cancel godoc
// @Summary      Cancel an appointment
// @Description  Cancels an appointment owned by the authenticated customer.
// @Tags         appointments
// @Produce      json
// @Param        appointmentID path string true "Appointment UUID"
// @Success      200 {object} MessageResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/my/appointments/{appointmentID} [delete]
func (h *AppointmentHandler) Cancel(w http.ResponseWriter, r *http.Request) {
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

	if err := h.appointments.Cancel(r.Context(), appointmentID, userID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, MessageResponse{Message: "appointment cancelled"})
}
