package models

import (
	"time"

	"github.com/google/uuid"
)

type Business struct {
	ID        uuid.UUID
	Name      string
	Slug      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type BusinessUser struct {
	ID          uuid.UUID
	BusinessID  uuid.UUID
	UserID      string
	Role        string
	DisplayName string
	Active      bool
}

type Service struct {
	ID              uuid.UUID
	BusinessID      uuid.UUID
	Name            string
	DurationMinutes int
	Price           float64
	Active          bool
	Description     string
	PictureID       *uuid.UUID
	CreatedAt       time.Time
}

type WorkingHours struct {
	ID             uuid.UUID
	BusinessUserID uuid.UUID
	DayOfWeek      int
	StartTime      time.Time
	EndTime        time.Time
}

type WorkingHoursOverride struct {
	ID             uuid.UUID
	BusinessUserID uuid.UUID
	OverrideDate   time.Time
	StartTime      *time.Time
	EndTime        *time.Time
	IsOff          bool
	Reason         string
}

type AppointmentStatus string

const (
	AppointmentStatusPending   AppointmentStatus = "pending"
	AppointmentStatusConfirmed AppointmentStatus = "confirmed"
	AppointmentStatusCompleted AppointmentStatus = "completed"
	AppointmentStatusCancelled AppointmentStatus = "cancelled"
	AppointmentStatusNoShow    AppointmentStatus = "no_show"
)

type Appointment struct {
	ID                 uuid.UUID
	BusinessID         uuid.UUID
	ServiceID          uuid.UUID
	BusinessUserID     uuid.UUID
	CustomerUserID     string
	StartTime          time.Time
	EndTime            time.Time
	Status             AppointmentStatus
	CreatedBy          string
	CancellationReason string
	CreatedAt          time.Time
}

type TimeSlot struct {
	StartTime time.Time
	EndTime   time.Time
}

type Image struct {
	ID          uuid.UUID
	Storage     string
	ObjectKey   string
	Data        []byte
	URL         string
	ContentType string
	CreatedAt   time.Time
}

type EmployeeService struct {
	BusinessUserID uuid.UUID
	ServiceID      uuid.UUID
}

type EmployeeUnavailability struct {
	ID             uuid.UUID
	BusinessUserID uuid.UUID
	StartTime      time.Time
	EndTime        time.Time
	Reason         string
}
