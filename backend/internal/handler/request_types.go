package handler

import (
	"time"

	"fejd-backend/internal/dto"

	"github.com/google/uuid"
)

type CreateAppointmentRequest struct {
	BusinessID     uuid.UUID `json:"business_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	ServiceID      uuid.UUID `json:"service_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	BusinessUserID uuid.UUID `json:"business_user_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	StartTime      string    `json:"start_time" validate:"required" example:"2024-01-01T09:00:00Z"`
	CustomerUserID string    `json:"customer_user_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type CancelAppointmentRequest struct {
	CancellationReason string `json:"cancellation_reason,omitempty" example:"unexpected absence"`
}

type SetEmployeeServicesRequest struct {
	ServiceIDs []uuid.UUID `json:"service_ids" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type CreateUnavailabilityRequest struct {
	StartTime string `json:"start_time" validate:"required" example:"2024-01-01T09:00:00Z"`
	EndTime   string `json:"end_time" validate:"required" example:"2024-01-01T17:00:00Z"`
	Reason    string `json:"reason,omitempty" example:"Vacation"`
}

type WorkingHoursInput struct {
	DayOfWeek int       `json:"day_of_week" validate:"required" example:"1"`
	StartTime time.Time `json:"start_time" validate:"required" example:"09:00"`
	EndTime   time.Time `json:"end_time" validate:"required" example:"17:00"`
}

type SetWorkingHoursRequest struct {
	WorkingHours []WorkingHoursInput `json:"working_hours" validate:"required"`
}

type ServiceInput struct {
	Name            string     `json:"name" validate:"required" example:"Massage"`
	DurationMinutes int        `json:"duration_minutes" validate:"required" example:"60"`
	Price           float64    `json:"price,omitempty" example:"100.00"`
	Active          bool       `json:"active" validate:"required" example:"true"`
	Description     string     `json:"description,omitempty"`
	PictureID       *uuid.UUID `json:"picture_id,omitempty"`
}

type WorkingHoursOverrideInput struct {
	OverrideDate time.Time  `json:"override_date" validate:"required" example:"2024-12-25"`
	StartTime    *time.Time `json:"start_time,omitempty" example:"10:00"`
	EndTime      *time.Time `json:"end_time,omitempty" example:"14:00"`
	IsOff        bool       `json:"is_off" validate:"required" example:"false"`
	Reason       string     `json:"reason,omitempty" example:"Christmas hours"`
}

type ErrorResponse struct {
	Error string `json:"error" validate:"required" example:"error message"`
}

type MessageResponse struct {
	Message string `json:"message" validate:"required" example:"operation complete"`
}

type BusinessResponse struct {
	Business  dto.Business       `json:"business" validate:"required"`
	Services  []dto.Service      `json:"services" validate:"required"`
	Employees []dto.BusinessUser `json:"employees" validate:"required"`
}

type SlotsResponse struct {
	Slots []dto.TimeSlot `json:"slots" validate:"required"`
	Date  string         `json:"date" validate:"required" example:"2024-01-01"`
}

type WorkingHoursResponse struct {
	WorkingHours []dto.WorkingHours         `json:"working_hours" validate:"required"`
	Overrides    []dto.WorkingHoursOverride `json:"overrides" validate:"required"`
}
