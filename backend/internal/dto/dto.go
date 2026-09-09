package dto

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Business struct {
	ID                    uuid.UUID `json:"id" validate:"required"`
	Name                  string    `json:"name" validate:"required"`
	Slug                  string    `json:"slug" validate:"required"`
	CreatedAt             time.Time `json:"created_at" validate:"required"`
	UpdatedAt             time.Time `json:"updated_at" validate:"required"`
	CancellationLeadHours int       `json:"cancellation_lead_hours" validate:"required"`
	NoShowAfterHours      int       `json:"no_show_after_hours" validate:"required"`
}

// Me is the authenticated user's own onboarding state, synthesized from the
// JWT approval claim and a lookup rather than a DB row.
type Me struct {
	ApprovalStatus string       `json:"approval_status" validate:"required"`
	HasSalon       bool         `json:"has_salon" validate:"required"`
	Businesses     []MeBusiness `json:"businesses" validate:"required"`
}

// MeBusiness is a business the caller belongs to (as admin or employee).
type MeBusiness struct {
	ID   uuid.UUID `json:"id" validate:"required"`
	Name string    `json:"name" validate:"required"`
	Slug string    `json:"slug" validate:"required"`
	Role string    `json:"role" validate:"required"`
}

type BusinessUser struct {
	ID          uuid.UUID `json:"id" validate:"required"`
	BusinessID  uuid.UUID `json:"business_id" validate:"required"`
	UserID      string    `json:"user_id" validate:"required"`
	Role        string    `json:"role" validate:"required"`
	DisplayName string    `json:"display_name"`
	Active      bool      `json:"active" validate:"required"`
	Avatar      string    `json:"avatar,omitempty"`
}

// Customer is an existing customer of a business, with the display name cached
// in the local users table (never fetched from Keycloak).
type Customer struct {
	UserID      string `json:"user_id" validate:"required"`
	DisplayName string `json:"display_name"`
}

type Service struct {
	ID              uuid.UUID  `json:"id" validate:"required"`
	BusinessID      uuid.UUID  `json:"business_id" validate:"required"`
	Name            string     `json:"name" validate:"required"`
	DurationMinutes int        `json:"duration_minutes" validate:"required"`
	Price           float64    `json:"price,omitempty"`
	Active          bool       `json:"active" validate:"required"`
	Description     string     `json:"description,omitempty"`
	PictureID       *uuid.UUID `json:"picture_id,omitempty"`
	CreatedAt       time.Time  `json:"created_at" validate:"required"`
}

// Section is a landing-page content block. Content is a locale-keyed object
// ({ "<locale>": { ...type-specific fields } }).
type Section struct {
	ID       uuid.UUID `json:"id" validate:"required"`
	PageID   uuid.UUID `json:"page_id" validate:"required"`
	Type     string    `json:"type" validate:"required"`
	// Content is a locale-keyed JSON object.
	Content  json.RawMessage `json:"content" validate:"required" swaggertype:"object"`
	Position int             `json:"position" validate:"required"`
}

type WorkingHours struct {
	ID             uuid.UUID `json:"id" validate:"required"`
	BusinessUserID uuid.UUID `json:"business_user_id" validate:"required"`
	DayOfWeek      int       `json:"day_of_week" validate:"required"`
	StartTime      time.Time `json:"start_time" validate:"required"`
	EndTime        time.Time `json:"end_time" validate:"required"`
}

type WorkingHoursOverride struct {
	ID             uuid.UUID  `json:"id" validate:"required"`
	BusinessUserID uuid.UUID  `json:"business_user_id" validate:"required"`
	OverrideDate   time.Time  `json:"override_date" validate:"required"`
	StartTime      *time.Time `json:"start_time,omitempty"`
	EndTime        *time.Time `json:"end_time,omitempty"`
	IsOff          bool       `json:"is_off" validate:"required"`
	Reason         string     `json:"reason,omitempty"`
}

type Appointment struct {
	ID                 uuid.UUID `json:"id" validate:"required"`
	BusinessID         uuid.UUID `json:"business_id" validate:"required"`
	ServiceID          uuid.UUID `json:"service_id" validate:"required"`
	BusinessUserID     uuid.UUID `json:"business_user_id" validate:"required"`
	CustomerUserID     string    `json:"customer_user_id,omitempty"`
	StartTime          time.Time `json:"start_time" validate:"required"`
	EndTime            time.Time `json:"end_time" validate:"required"`
	Status             string    `json:"status" validate:"required"`
	CreatedBy          string    `json:"created_by" validate:"required"`
	CancellationReason string    `json:"cancellation_reason,omitempty"`
	CreatedAt          time.Time `json:"created_at" validate:"required"`
	// ServiceName is populated for staff-facing reservation lists only.
	ServiceName string `json:"service_name,omitempty"`
	// CancellationLeadHours is populated for customer appointment lists so the
	// UI can surface the salon's cancellation notice window.
	CancellationLeadHours int `json:"cancellation_lead_hours,omitempty"`
	// NoShowAfterHours is populated for staff reservation lists so the UI can
	// surface when a no-show may be recorded.
	NoShowAfterHours int `json:"no_show_after_hours,omitempty"`
}

type TimeSlot struct {
	StartTime time.Time `json:"start_time" validate:"required"`
	EndTime   time.Time `json:"end_time" validate:"required"`
}

type EmployeeUnavailability struct {
	ID             uuid.UUID `json:"id" validate:"required"`
	BusinessUserID uuid.UUID `json:"business_user_id" validate:"required"`
	StartTime      time.Time `json:"start_time" validate:"required"`
	EndTime        time.Time `json:"end_time" validate:"required"`
	Reason         string    `json:"reason,omitempty"`
}

type Image struct {
	ID          uuid.UUID `json:"id" validate:"required"`
	URL         string    `json:"url" validate:"required"`
	ContentType string    `json:"content_type,omitempty"`
	CreatedAt   time.Time `json:"created_at" validate:"required"`
}
