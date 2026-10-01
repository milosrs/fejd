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
	// NoShowAfterMinutes is how many minutes after an appointment start a staff
	// member may mark it as no-show. Soft policy, configurable by the owner.
	NoShowAfterMinutes int
	// SlotIntervalMinutes is the granularity (in minutes) of the booking slot
	// grid. Soft policy, configurable by the owner.
	SlotIntervalMinutes int
	// AutoApprove is whether new appointments are approved automatically
	// instead of waiting for a barber or owner to approve them.
	AutoApprove bool
	// AppointmentReminderEnabled is whether customers receive a push reminder
	// before their appointment starts. Soft policy, configurable by the owner.
	AppointmentReminderEnabled bool
	// AppointmentReminderLeadMinutes is how many minutes before an appointment
	// the reminder is sent. Soft policy, configurable by the owner.
	AppointmentReminderLeadMinutes int
	// StaffNotificationsEnabled is whether the owner/employees receive a push
	// when a customer books an appointment with them. Soft policy.
	StaffNotificationsEnabled bool
	// ReminderTitle is the owner's custom reminder header. Empty falls back to
	// the default title.
	ReminderTitle string
	// ReminderBody is the owner's custom reminder description. Empty falls back
	// to the default body.
	ReminderBody string
	// AddressLine is the salon's street address (structured for local SEO).
	AddressLine string
	// City is the salon's locality (e.g. "Sremska Mitrovica"); the key local
	// search signal. Empty until the owner fills the location form.
	City string
	// PostalCode is the salon's postal code.
	PostalCode string
	// Country is the ISO country code (e.g. "RS").
	Country string
	// Latitude/Longitude are the geo coordinates, geocoded from the address.
	Latitude  *float64
	Longitude *float64
	// Phone is the salon's public phone number.
	Phone string
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
	// Slug is the stable, human-readable identifier used in public URLs
	// (/services/{slug}). Immutable after creation.
	Slug            string
	DurationMinutes int
	Price           float64
	Active          bool
	Description     string
	PictureID       *uuid.UUID
	CreatedAt       time.Time
}

// ServiceCombination declares a directed combinability edge: a customer may add
// CombinableServiceID onto a booking whose base service is ServiceID. The edge
// is directional (beard -> fade does not imply fade -> beard).
type ServiceCombination struct {
	ServiceID           uuid.UUID
	CombinableServiceID uuid.UUID
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

// BusinessClosureType discriminates the kinds of salon non-working-day rules.
type BusinessClosureType string

const (
	BusinessClosureSingle BusinessClosureType = "single"
	BusinessClosureRange  BusinessClosureType = "range"
	BusinessClosureWeekly BusinessClosureType = "weekly"
	BusinessClosureYearly BusinessClosureType = "yearly"
)

// BusinessClosure marks when the salon is closed (a non-working-day rule), so
// no appointments are offered for any employee on matching days. Exactly one
// set of fields is populated per ClosureType:
//   - single: StartDate
//   - range:  StartDate + EndDate
//   - weekly: DayOfWeek (0=Sun .. 6=Sat)
//   - yearly: Month + Day
type BusinessClosure struct {
	ID          uuid.UUID
	BusinessID  uuid.UUID
	ClosureType BusinessClosureType
	StartDate   *time.Time
	EndDate     *time.Time
	DayOfWeek   *int
	Month       *int
	Day         *int
	Reason      string
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
	AvatarID *uuid.UUID
	// RegistrationSource records how the user registered: "self" (direct
	// registration) or "invite" (via a QR/invite link).
	RegistrationSource string
	// InvitedBusinessID is the salon a customer was invited to, if the invite
	// was salon-scoped; nil for platform ("invite a friend") invites.
	InvitedBusinessID *uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// Customer is a user who has booked with a business (resolved from local users).
type Customer struct {
	UserID      string
	DisplayName string
	// AvatarID references the user's profile picture in the images table.
	AvatarID *uuid.UUID
}

// InvitedCustomer is a customer who was invited to a salon via a QR/invite link,
// together with the salon's identifying info for the realm-admin dashboard.
type InvitedCustomer struct {
	BusinessID   uuid.UUID
	BusinessName string
	BusinessSlug string
	// BusinessLogo references the salon's logo image, if set.
	BusinessLogo *uuid.UUID
	UserID       string
	DisplayName  string
	// AvatarID references the customer's profile picture in the images table.
	AvatarID *uuid.UUID
}

// SelfRegisteredUser is a user who registered directly (not via a QR/invite
// link), shown on the realm-admin dashboard under "own registration".
type SelfRegisteredUser struct {
	UserID      string
	DisplayName string
	// AvatarID references the user's profile picture in the images table.
	AvatarID *uuid.UUID
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
	// AdditionalServiceIDs are the add-on services combined onto the base
	// ServiceID, in display order. The reservation's duration and price are the
	// sum of the base service and these add-ons.
	AdditionalServiceIDs []uuid.UUID
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

type UnavailabilityStatus string

const (
	UnavailabilityStatusPending   UnavailabilityStatus = "pending"
	UnavailabilityStatusConfirmed UnavailabilityStatus = "confirmed"
	UnavailabilityStatusRejected  UnavailabilityStatus = "rejected"
)

type EmployeeUnavailability struct {
	ID              uuid.UUID
	BusinessUserID  uuid.UUID
	StartTime       time.Time
	EndTime         time.Time
	Reason          string
	Status          UnavailabilityStatus
	RejectionReason string
}

// Invitation is a shareable link/QR invite that links a user to a business as
// an employee or customer. The raw token is never stored; only its hash is
// persisted. A nil ExpiresAt means the invitation never expires (used for
// customer QR codes that owners print and stick up).
type Invitation struct {
	ID         uuid.UUID
	BusinessID uuid.UUID
	TokenHash  string
	Role       string
	CreatedBy  string
	MaxUses    int
	UseCount   int
	ExpiresAt  *time.Time
	CreatedAt  time.Time
}

// PushToken is a device's FCM registration token, linked to a user's Keycloak
// subject so notifications can target a user across their devices.
type PushToken struct {
	ID        uuid.UUID
	UserID    string
	Token     string
	Platform  string
	CreatedAt time.Time
	UpdatedAt time.Time
}
