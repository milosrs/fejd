package service

import (
	"context"
	"fmt"

	"fejd-backend/internal/models"
	"fejd-backend/internal/push"

	"github.com/google/uuid"
)

// PushSender sends push notifications to device tokens. *push.Sender satisfies
// it; tests provide a fake.
type PushSender interface {
	SendMany(ctx context.Context, tokens []string, msg push.Message) error
}

// tokenLister looks up device tokens for a set of users.
type tokenLister interface {
	ListByUserIDs(ctx context.Context, userIDs []string) ([]models.PushToken, error)
}

// businessUserLookup resolves a business user (provider) by ID.
type businessUserLookup interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.BusinessUser, error)
}

// businessGetter resolves a business (salon) by ID.
type businessGetter interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.Business, error)
}

// userGetter resolves a user by Keycloak subject.
type userGetter interface {
	GetByID(ctx context.Context, id string) (*models.User, error)
}

// businessLogoGetter resolves a salon's logo image ID.
type businessLogoGetter interface {
	GetBusinessLogoID(ctx context.Context, businessID uuid.UUID) (*uuid.UUID, error)
}

// NotificationService turns domain events (bookings, upcoming appointments)
// into push notifications. It is safe to construct with a nil sender: sends
// become no-ops.
type NotificationService struct {
	tokens       tokenLister
	bu           businessUserLookup
	business     businessGetter
	user         userGetter
	logo         businessLogoGetter
	sender       PushSender
	imageBaseURL string
	appURL       string
}

func NewNotificationService(
	tokens tokenLister,
	bu businessUserLookup,
	business businessGetter,
	user userGetter,
	logo businessLogoGetter,
	sender PushSender,
	imageBaseURL string,
	appURL string,
) *NotificationService {
	return &NotificationService{
		tokens:       tokens,
		bu:           bu,
		business:     business,
		user:         user,
		logo:         logo,
		sender:       sender,
		imageBaseURL: imageBaseURL,
		appURL:       appURL,
	}
}

// NotifyUsers sends msg to every device token of the given users. It is a no-op
// when no sender is configured or no tokens exist.
func (s *NotificationService) NotifyUsers(ctx context.Context, userIDs []string, msg push.Message) error {
	if s.sender == nil || len(userIDs) == 0 {
		return nil
	}

	tokens, err := s.tokens.ListByUserIDs(ctx, userIDs)
	if err != nil {
		return fmt.Errorf("list push tokens: %w", err)
	}
	if len(tokens) == 0 {
		return nil
	}

	deviceTokens := make([]string, len(tokens))
	for i, t := range tokens {
		deviceTokens[i] = t.Token
	}
	return s.sender.SendMany(ctx, deviceTokens, msg)
}

// NotifyNewBooking notifies the selected provider (owner or employee) that a
// customer booked them, showing the customer's name and avatar and the
// appointment time. It is gated by the salon's staff-notifications policy.
func (s *NotificationService) NotifyNewBooking(ctx context.Context, appt *models.Appointment) error {
	if s.sender == nil {
		return nil
	}

	business, err := s.business.GetByID(ctx, appt.BusinessID)
	if err != nil {
		return fmt.Errorf("resolve business: %w", err)
	}
	if !business.StaffNotificationsEnabled {
		return nil
	}

	bu, err := s.bu.GetByID(ctx, appt.BusinessUserID)
	if err != nil {
		return fmt.Errorf("resolve provider: %w", err)
	}

	title := "New booking"
	var imageURL string
	if appt.CustomerUserID != "" {
		if customer, err := s.user.GetByID(ctx, appt.CustomerUserID); err == nil && customer.DisplayName != "" {
			title = customer.DisplayName
			imageURL = s.imageURL(customer.AvatarID)
		}
	}

	return s.NotifyUsers(ctx, []string{bu.UserID}, push.Message{
		Title:    title,
		Body:     "Booked for " + appt.StartTime.UTC().Format("Mon, 02 Jan 15:04"),
		ImageURL: imageURL,
		IconURL:  imageURL,
		Link:     s.reservationsURL(appt.BusinessID),
		Data: map[string]string{
			"type":           "appointment_booked",
			"appointment_id": appt.ID.String(),
			"business_id":    appt.BusinessID.String(),
		},
	})
}

// NotifyAppointmentReminder notifies a customer that their appointment is about
// to begin, using the salon's custom reminder content and logo.
func (s *NotificationService) NotifyAppointmentReminder(ctx context.Context, appt *models.Appointment) error {
	if s.sender == nil || appt.CustomerUserID == "" {
		return nil
	}

	business, err := s.business.GetByID(ctx, appt.BusinessID)
	if err != nil {
		return fmt.Errorf("resolve business: %w", err)
	}

	title := business.ReminderTitle
	if title == "" {
		title = "Upcoming appointment"
	}
	body := business.ReminderBody
	if body == "" {
		body = "Your appointment is starting soon."
	}

	imageURL := ""
	if logoID, err := s.logo.GetBusinessLogoID(ctx, appt.BusinessID); err == nil && logoID != nil {
		imageURL = s.imageURL(logoID)
	}

	return s.NotifyUsers(ctx, []string{appt.CustomerUserID}, push.Message{
		Title:    title,
		Body:     body,
		ImageURL: imageURL,
		IconURL:  imageURL,
		Link:     s.salonURL(business.Slug),
		Data: map[string]string{
			"type":           "appointment_reminder",
			"appointment_id": appt.ID.String(),
			"business_id":    appt.BusinessID.String(),
		},
	})
}

// imageURL builds an absolute public URL for an image, or "" when the image is
// absent.
func (s *NotificationService) imageURL(id *uuid.UUID) string {
	if s.imageBaseURL == "" || id == nil {
		return ""
	}
	return s.imageBaseURL + "/api/images/" + id.String()
}

// reservationsURL builds the web URL of the salon's staff reservation list.
func (s *NotificationService) reservationsURL(businessID uuid.UUID) string {
	if s.appURL == "" {
		return ""
	}
	return s.appURL + "/admin/business/" + businessID.String() + "/my-reservations"
}

// salonURL builds the web URL of a salon's site.
func (s *NotificationService) salonURL(slug string) string {
	if s.appURL == "" {
		return ""
	}
	return s.appURL + "/" + slug
}
