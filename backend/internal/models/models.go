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
	// CancellationLeadHours is the minimum notice (hours) a customer must give
	// before cancelling an appointment. Soft policy, configurable by the owner.
	CancellationLeadHours int
	// NoShowAfterHours is how many hours after an appointment start a staff
	// member may mark it as no-show. Soft policy, configurable by the owner.
	NoShowAfterHours int
	// SlotIntervalMinutes is the granularity (in minutes) of the booking slot
	// grid. Soft policy, configurable by the owner.
	SlotIntervalMinutes int
}

type BusinessUser struct {
	ID          uuid.UUID
	BusinessID  uuid.UUID
	UserID      string
	Role        string
	DisplayName string
	Active      bool
}

// BusinessMembership pairs a business with the calling user's role in it.
type BusinessMembership struct {
	BusinessID uuid.UUID
	Name       string
	Slug       string
	Role       string
}

// Section is a landing-page content block. Content is a locale-keyed JSONB
// object ({ "<locale>": { ...type-specific fields } }) so it is
// internationalization-ready without schema changes.
type Section struct {
	ID        uuid.UUID
	PageID    uuid.UUID
	Type      string
	Content   []byte
	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Page is a named page of a business's public site (e.g. "landing").
type Page struct {
	ID         uuid.UUID
	BusinessID uuid.UUID
	Name       string
	Position   int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Translation is a single app UI label in a given locale.
type Translation struct {
	ID        uuid.UUID
	Key       string
	Locale    string
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
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

// BusinessHours is a salon-level default working-hours row, used as a fallback
// for staff who have no per-employee working hours.
type BusinessHours struct {
	ID         uuid.UUID
	BusinessID uuid.UUID
	DayOfWeek  int
	StartTime  time.Time
	EndTime    time.Time
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

// User is the local cache of a Keycloak user's identity attributes. It is keyed
// by the Keycloak subject (sub) and populated from JWT claims at authentication,
// so the app never needs to query Keycloak for display attributes.
type User struct {
	ID          string
	DisplayName string
	Email       string
	// AvatarID references the user's profile picture in the images table.
	AvatarID  *uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Customer is a user who has booked with a business (resolved from local users).
type Customer struct {
	UserID      string
	DisplayName string
}

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
	BusinessID  uuid.UUID
	Storage     string
	ObjectKey   string
	Data        []byte
	URL         string
	ContentType string
	CreatedAt   time.Time
}

type Visibility string

const (
	VisibilityPublic  Visibility = "public"
	VisibilityPrivate Visibility = "private"
)

type ImageLink struct {
	ID         uuid.UUID
	ImageID    uuid.UUID
	EntityType string
	EntityID   uuid.UUID
	Purpose    string
	Visibility Visibility
	CreatedAt  time.Time
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

// Invitation is a shareable link/QR invite that links a user to a business as
// an employee. The raw token is never stored; only its hash is persisted.
type Invitation struct {
	ID         uuid.UUID
	BusinessID uuid.UUID
	TokenHash  string
	Role       string
	CreatedBy  string
	MaxUses    int
	UseCount   int
	ExpiresAt  time.Time
	CreatedAt  time.Time
}
