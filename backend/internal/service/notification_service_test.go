package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"fejd-backend/internal/models"
	"fejd-backend/internal/push"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeSender struct {
	sent [][]string
	msgs []push.Message
}

func (f *fakeSender) SendMany(_ context.Context, tokens []string, msg push.Message) error {
	f.sent = append(f.sent, tokens)
	f.msgs = append(f.msgs, msg)
	return nil
}

type fakeTokenLister struct {
	tokens map[string][]models.PushToken
}

func (f *fakeTokenLister) ListByUserIDs(_ context.Context, userIDs []string) ([]models.PushToken, error) {
	var out []models.PushToken
	for _, id := range userIDs {
		out = append(out, f.tokens[id]...)
	}
	return out, nil
}

type fakeBusinessUserLookup struct {
	users map[uuid.UUID]*models.BusinessUser
}

func (f *fakeBusinessUserLookup) GetByID(_ context.Context, id uuid.UUID) (*models.BusinessUser, error) {
	return f.users[id], nil
}

type fakeBusinessGetter struct {
	businesses map[uuid.UUID]*models.Business
}

func (f *fakeBusinessGetter) GetByID(_ context.Context, id uuid.UUID) (*models.Business, error) {
	return f.businesses[id], nil
}

type fakeUserGetter struct {
	users map[string]*models.User
}

func (f *fakeUserGetter) GetByID(_ context.Context, id string) (*models.User, error) {
	return f.users[id], nil
}

type fakeLogoGetter struct {
	logo *uuid.UUID
}

func (f *fakeLogoGetter) GetBusinessLogoID(_ context.Context, _ uuid.UUID) (*uuid.UUID, error) {
	return f.logo, nil
}

func newTestService(sender PushSender, business *models.Business, bu *models.BusinessUser, customer *models.User, logo *uuid.UUID) *NotificationService {
	tokens := &fakeTokenLister{tokens: map[string][]models.PushToken{}}
	buLookup := &fakeBusinessUserLookup{}
	businessGetter := &fakeBusinessGetter{}
	userGetter := &fakeUserGetter{}
	if bu != nil {
		buLookup.users = map[uuid.UUID]*models.BusinessUser{bu.ID: bu}
		tokens.tokens[bu.UserID] = []models.PushToken{{Token: "t1"}}
	}
	if business != nil {
		businessGetter.businesses = map[uuid.UUID]*models.Business{business.ID: business}
	}
	if customer != nil {
		userGetter.users = map[string]*models.User{customer.ID: customer}
		tokens.tokens[customer.ID] = []models.PushToken{{Token: "t2"}}
	}
	return NewNotificationService(tokens, buLookup, businessGetter, userGetter, &fakeLogoGetter{logo: logo}, sender, "https://fejd.fyi", "https://fejd.fyi")
}

func TestNotificationService_NotifyUsers_NilSender(t *testing.T) {
	s := newTestService(nil, nil, nil, nil, nil)
	require.NoError(t, s.NotifyUsers(context.Background(), []string{"u1"}, push.Message{Title: "t"}))
}

func TestNotificationService_NotifyUsers_SendsToAllTokens(t *testing.T) {
	tokens := &fakeTokenLister{tokens: map[string][]models.PushToken{
		"u1": {{Token: "t1"}, {Token: "t2"}},
	}}
	sender := &fakeSender{}
	s := NewNotificationService(tokens, &fakeBusinessUserLookup{}, &fakeBusinessGetter{}, &fakeUserGetter{}, &fakeLogoGetter{}, sender, "", "")

	require.NoError(t, s.NotifyUsers(context.Background(), []string{"u1", "u2"}, push.Message{Title: "Hi"}))

	require.Len(t, sender.sent, 1)
	assert.Equal(t, []string{"t1", "t2"}, sender.sent[0])
	assert.Equal(t, "Hi", sender.msgs[0].Title)
}

func TestNotificationService_NotifyNewBooking(t *testing.T) {
	businessID := uuid.New()
	bu := &models.BusinessUser{ID: uuid.New(), UserID: "provider-sub"}
	business := &models.Business{ID: businessID, Slug: "salon", StaffNotificationsEnabled: true}
	avatarID := uuid.New()
	customer := &models.User{ID: "customer-sub", DisplayName: "John Doe", AvatarID: &avatarID}
	sender := &fakeSender{}

	s := newTestService(sender, business, bu, customer, nil)

	start := time.Date(2026, 9, 30, 14, 30, 0, 0, time.UTC)
	appt := &models.Appointment{ID: uuid.New(), BusinessID: businessID, BusinessUserID: bu.ID, CustomerUserID: "customer-sub", StartTime: start}
	require.NoError(t, s.NotifyNewBooking(context.Background(), appt))

	require.Len(t, sender.msgs, 1)
	assert.Equal(t, "John Doe", sender.msgs[0].Title)
	assert.True(t, strings.HasPrefix(sender.msgs[0].Body, "Booked for "))
	assert.Equal(t, "https://fejd.fyi/api/images/"+avatarID.String(), sender.msgs[0].ImageURL)
	assert.Equal(t, "https://fejd.fyi/api/images/"+avatarID.String(), sender.msgs[0].IconURL)
	assert.Equal(t, "https://fejd.fyi/admin/business/"+businessID.String()+"/my-reservations", sender.msgs[0].Link)
	assert.Equal(t, "appointment_booked", sender.msgs[0].Data["type"])
}

func TestNotificationService_NotifyNewBooking_StaffDisabled(t *testing.T) {
	business := &models.Business{ID: uuid.New(), StaffNotificationsEnabled: false}
	bu := &models.BusinessUser{ID: uuid.New(), UserID: "provider-sub"}
	sender := &fakeSender{}

	s := newTestService(sender, business, bu, nil, nil)

	appt := &models.Appointment{ID: uuid.New(), BusinessID: business.ID, BusinessUserID: bu.ID}
	require.NoError(t, s.NotifyNewBooking(context.Background(), appt))

	assert.Empty(t, sender.sent)
}

func TestNotificationService_NotifyAppointmentReminder(t *testing.T) {
	business := &models.Business{ID: uuid.New(), Slug: "salon", ReminderTitle: "Almost time", ReminderBody: "See you soon"}
	logoID := uuid.New()
	customer := &models.User{ID: "customer-sub", DisplayName: "John Doe"}
	sender := &fakeSender{}

	s := newTestService(sender, business, nil, customer, &logoID)

	appt := &models.Appointment{ID: uuid.New(), BusinessID: business.ID, CustomerUserID: "customer-sub"}
	require.NoError(t, s.NotifyAppointmentReminder(context.Background(), appt))

	require.Len(t, sender.msgs, 1)
	assert.Equal(t, "Almost time", sender.msgs[0].Title)
	assert.Equal(t, "See you soon", sender.msgs[0].Body)
	assert.Equal(t, "https://fejd.fyi/api/images/"+logoID.String(), sender.msgs[0].ImageURL)
	assert.Equal(t, "https://fejd.fyi/salon", sender.msgs[0].Link)
	assert.Equal(t, "appointment_reminder", sender.msgs[0].Data["type"])
}

func TestNotificationService_NotifyAppointmentReminder_FallsBackToDefaults(t *testing.T) {
	business := &models.Business{ID: uuid.New(), Slug: "salon"}
	customer := &models.User{ID: "customer-sub"}
	sender := &fakeSender{}

	s := newTestService(sender, business, nil, customer, nil)

	appt := &models.Appointment{ID: uuid.New(), BusinessID: business.ID, CustomerUserID: "customer-sub"}
	require.NoError(t, s.NotifyAppointmentReminder(context.Background(), appt))

	require.Len(t, sender.msgs, 1)
	assert.Equal(t, "Upcoming appointment", sender.msgs[0].Title)
	assert.Equal(t, "Your appointment is starting soon.", sender.msgs[0].Body)
}

func TestNotificationService_NotifyAppointmentReminder_NoCustomer(t *testing.T) {
	s := newTestService(&fakeSender{}, nil, nil, nil, nil)
	require.NoError(t, s.NotifyAppointmentReminder(context.Background(), &models.Appointment{ID: uuid.New()}))
}
