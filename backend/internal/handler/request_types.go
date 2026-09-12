package handler

import (
	"encoding/json"
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

type RejectRequest struct {
	Reason string `json:"reason" validate:"required" example:"slot no longer available"`
}

type CreateOwnAppointmentRequest struct {
	ServiceID      uuid.UUID `json:"service_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	StartTime      string    `json:"start_time" validate:"required" example:"2024-01-01T09:00:00Z"`
	CustomerUserID string    `json:"customer_user_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type SetEmployeeServicesRequest struct {
	ServiceIDs []uuid.UUID `json:"service_ids" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type SetServiceEmployeesRequest struct {
	BusinessUserIDs []uuid.UUID `json:"business_user_ids" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type CreateEmployeeRequest struct {
	Name       string      `json:"name" validate:"required" example:"Sam Barber"`
	Email      string      `json:"email" validate:"required" example:"sam@example.com"`
	ServiceIDs []uuid.UUID `json:"service_ids" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type CreateInvitationRequest struct {
	ExpiresInHours int `json:"expires_in_hours,omitempty" example:"48"`
}

type InvitationResponse struct {
	ID        uuid.UUID `json:"id" validate:"required"`
	URL       string    `json:"url" validate:"required"`
	Token     string    `json:"token" validate:"required"`
	ExpiresAt time.Time `json:"expires_at" validate:"required"`
}

type PublicInvitationResponse struct {
	SalonName string    `json:"salon_name" validate:"required"`
	SalonSlug string    `json:"salon_slug" validate:"required"`
	Role      string    `json:"role" validate:"required"`
	ExpiresAt time.Time `json:"expires_at" validate:"required"`
}

type RemoveEmployeeResponse struct {
	Reassigned int    `json:"reassigned" validate:"required"`
	Cancelled  int    `json:"cancelled" validate:"required"`
	Message    string `json:"message" validate:"required"`
}

type CreateUnavailabilityRequest struct {
	StartTime string `json:"start_time" validate:"required" example:"2024-01-01T09:00:00Z"`
	EndTime   string `json:"end_time" validate:"required" example:"2024-01-01T17:00:00Z"`
	Reason    string `json:"reason,omitempty" example:"Vacation"`
}

type SalonPolicyResponse struct {
	CancellationLeadHours int                `json:"cancellation_lead_hours" validate:"required" example:"2"`
	NoShowAfterHours      int                `json:"no_show_after_hours" validate:"required" example:"2"`
	SlotIntervalMinutes   int                `json:"slot_interval_minutes" validate:"required" example:"30"`
	WorkingHours          []dto.BusinessHours `json:"working_hours" validate:"required"`
}

type UpdateSalonPolicyRequest struct {
	CancellationLeadHours int                  `json:"cancellation_lead_hours" validate:"required" example:"2"`
	NoShowAfterHours      int                  `json:"no_show_after_hours" validate:"required" example:"2"`
	SlotIntervalMinutes   int                  `json:"slot_interval_minutes" validate:"required" example:"30"`
	WorkingHours          []BusinessHoursInput `json:"working_hours" validate:"required"`
}

type BusinessHoursInput struct {
	DayOfWeek int    `json:"day_of_week" validate:"required" example:"1"`
	StartTime string `json:"start_time" validate:"required" example:"09:00"`
	EndTime   string `json:"end_time" validate:"required" example:"17:00"`
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

type BusinessCreateInput struct {
	Name string `json:"name" validate:"required" example:"My Salon"`
}

type RenameBusinessRequest struct {
	Name string `json:"name" validate:"required" example:"My Salon"`
}

type CreateSectionRequest struct {
	Type     string          `json:"type" validate:"required" example:"hero"`
	Content  json.RawMessage `json:"content" swaggertype:"object"`
	Position int             `json:"position,omitempty"`
}

type UpdateSectionRequest struct {
	Content json.RawMessage `json:"content" validate:"required" swaggertype:"object"`
}

type ReorderSectionsRequest struct {
	SectionIDs []uuid.UUID `json:"section_ids" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type MessageResponse struct {
	Message string `json:"message" validate:"required" example:"operation complete"`
}

// Translations is a flat map of translation key -> localized value.
type Translations map[string]string

type BusinessResponse struct {
	Business  dto.Business       `json:"business" validate:"required"`
	Services  []dto.Service      `json:"services" validate:"required"`
	Employees []dto.BusinessUser `json:"employees" validate:"required"`
	Images    BusinessImages     `json:"images" validate:"required"`
}

// BusinessImages exposes the salon's public business images as URLs.
type BusinessImages struct {
	Hero       string `json:"hero,omitempty"`
	Logo       string `json:"logo,omitempty"`
	Background string `json:"background,omitempty"`
}

type SlotsResponse struct {
	Slots []dto.TimeSlot `json:"slots" validate:"required"`
	Date  string         `json:"date" validate:"required" example:"2024-01-01"`
}

type WorkingHoursResponse struct {
	WorkingHours []dto.WorkingHours         `json:"working_hours" validate:"required"`
	Overrides    []dto.WorkingHoursOverride `json:"overrides" validate:"required"`
}
