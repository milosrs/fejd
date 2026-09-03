package dto

import (
	"time"

	"github.com/google/uuid"
)

type Business struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BusinessUser struct {
	ID          uuid.UUID `json:"id"`
	BusinessID  uuid.UUID `json:"business_id"`
	UserID      string    `json:"user_id"`
	Role        string    `json:"role"`
	DisplayName string    `json:"display_name"`
	Active      bool      `json:"active"`
}

type Service struct {
	ID              uuid.UUID  `json:"id"`
	BusinessID      uuid.UUID  `json:"business_id"`
	Name            string     `json:"name"`
	DurationMinutes int        `json:"duration_minutes"`
	Price           float64    `json:"price,omitempty"`
	Active          bool       `json:"active"`
	Description     string     `json:"description,omitempty"`
	PictureID       *uuid.UUID `json:"picture_id,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type WorkingHours struct {
	ID             uuid.UUID `json:"id"`
	BusinessUserID uuid.UUID `json:"business_user_id"`
	DayOfWeek      int       `json:"day_of_week"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
}

type WorkingHoursOverride struct {
	ID             uuid.UUID  `json:"id"`
	BusinessUserID uuid.UUID  `json:"business_user_id"`
	OverrideDate   time.Time  `json:"override_date"`
	StartTime      *time.Time `json:"start_time,omitempty"`
	EndTime        *time.Time `json:"end_time,omitempty"`
	IsOff          bool       `json:"is_off"`
	Reason         string     `json:"reason,omitempty"`
}

type Appointment struct {
	ID                 uuid.UUID `json:"id"`
	BusinessID         uuid.UUID `json:"business_id"`
	ServiceID          uuid.UUID `json:"service_id"`
	BusinessUserID     uuid.UUID `json:"business_user_id"`
	CustomerUserID     string    `json:"customer_user_id"`
	StartTime          time.Time `json:"start_time"`
	EndTime            time.Time `json:"end_time"`
	Status             string    `json:"status"`
	CreatedBy          string    `json:"created_by"`
	CancellationReason string    `json:"cancellation_reason,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
}

type TimeSlot struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

type EmployeeUnavailability struct {
	ID             uuid.UUID `json:"id"`
	BusinessUserID uuid.UUID `json:"business_user_id"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	Reason         string    `json:"reason,omitempty"`
}

type Image struct {
	ID          uuid.UUID `json:"id"`
	URL         string    `json:"url"`
	ContentType string    `json:"content_type,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}
