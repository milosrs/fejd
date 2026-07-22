package models

import (
	"time"

	"github.com/google/uuid"
)

type Business struct {
	ID        uuid.UUID `json:"id" validate:"required"`
	Name      string    `json:"name" validate:"required" example:"Acme Spa"`
	Slug      string    `json:"slug" validate:"required" example:"acme-spa"`
	CreatedAt time.Time `json:"created_at" validate:"required"`
	UpdatedAt time.Time `json:"updated_at" validate:"required"`
}

type BusinessUser struct {
	ID          uuid.UUID `json:"id" validate:"required"`
	BusinessID  uuid.UUID `json:"business_id" validate:"required"`
	UserID      string    `json:"user_id" validate:"required"`
	Role        string    `json:"role" validate:"required" example:"admin"`
	DisplayName string    `json:"display_name" validate:"required" example:"John Doe"`
}

type Service struct {
	ID              uuid.UUID `json:"id" validate:"required"`
	BusinessID      uuid.UUID `json:"business_id" validate:"required"`
	Name            string    `json:"name" validate:"required" example:"Massage"`
	DurationMinutes int       `json:"duration_minutes" validate:"required" example:"60"`
	Price           float64   `json:"price,omitempty" example:"100.00"`
	Active          bool      `json:"active" validate:"required" example:"true"`
	CreatedAt       time.Time `json:"created_at" validate:"required"`
}

type WorkingHours struct {
	ID             uuid.UUID `json:"id"`
	BusinessUserID uuid.UUID `json:"business_user_id"`
	DayOfWeek      int       `json:"day_of_week" validate:"required" example:"1"`
	StartTime      string    `json:"start_time" validate:"required" example:"09:00"`
	EndTime        string    `json:"end_time" validate:"required" example:"17:00"`
}

type WorkingHoursOverride struct {
	ID             uuid.UUID `json:"id"`
	BusinessUserID uuid.UUID `json:"business_user_id"`
	OverrideDate   string    `json:"override_date" validate:"required" example:"2024-12-25"`
	StartTime      *string   `json:"start_time,omitempty" example:"10:00"`
	EndTime        *string   `json:"end_time,omitempty" example:"14:00"`
	IsOff          bool      `json:"is_off" validate:"required" example:"false"`
	Reason         string    `json:"reason,omitempty" example:"Christmas hours"`
}

type AppointmentStatus string

const (
	AppointmentStatusConfirmed AppointmentStatus = "confirmed"
	AppointmentStatusCancelled AppointmentStatus = "cancelled"
	AppointmentStatusCompleted AppointmentStatus = "completed"
	AppointmentStatusNoShow    AppointmentStatus = "no_show"
)

type Appointment struct {
	ID             uuid.UUID         `json:"id" validate:"required"`
	BusinessID     uuid.UUID         `json:"business_id" validate:"required"`
	ServiceID      uuid.UUID         `json:"service_id" validate:"required"`
	BusinessUserID uuid.UUID         `json:"business_user_id" validate:"required"`
	CustomerUserID string            `json:"customer_user_id" validate:"required"`
	StartTime      time.Time         `json:"start_time" validate:"required"`
	EndTime        time.Time         `json:"end_time" validate:"required"`
	Status         AppointmentStatus `json:"status" validate:"required" example:"confirmed"`
	CreatedAt      time.Time         `json:"created_at" validate:"required"`
}

type TimeSlot struct {
	StartTime time.Time `json:"start_time" validate:"required"`
	EndTime   time.Time `json:"end_time" validate:"required"`
}
