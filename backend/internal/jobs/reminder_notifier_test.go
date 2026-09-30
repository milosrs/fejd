package jobs

import (
	"context"
	"testing"
	"time"

	"fejd-backend/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeBusinessLister struct {
	businesses []models.Business
}

func (f *fakeBusinessLister) List(context.Context) ([]models.Business, error) {
	return f.businesses, nil
}

type fakeReminderAppointments struct {
	due       []models.Appointment
	marked    []uuid.UUID
	cutoffArg time.Time
}

func (f *fakeReminderAppointments) ListDueForReminder(_ context.Context, _ uuid.UUID, _, cutoff time.Time) ([]models.Appointment, error) {
	f.cutoffArg = cutoff
	return f.due, nil
}

func (f *fakeReminderAppointments) MarkReminderSent(_ context.Context, id uuid.UUID) error {
	f.marked = append(f.marked, id)
	return nil
}

type fakeReminderSender struct {
	sent []uuid.UUID
}

func (f *fakeReminderSender) NotifyAppointmentReminder(_ context.Context, appt *models.Appointment) error {
	f.sent = append(f.sent, appt.ID)
	return nil
}

func TestReminderNotifier_SendsAndMarks(t *testing.T) {
	enabled := models.Business{ID: uuid.New(), AppointmentReminderEnabled: true, AppointmentReminderLeadMinutes: 30}
	disabled := models.Business{ID: uuid.New(), AppointmentReminderEnabled: false}

	due := []models.Appointment{
		{ID: uuid.New(), CustomerUserID: "customer-1"},
		{ID: uuid.New(), CustomerUserID: "customer-2"},
	}

	businesses := &fakeBusinessLister{businesses: []models.Business{enabled, disabled}}
	appointments := &fakeReminderAppointments{due: due}
	sender := &fakeReminderSender{}

	now := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)
	n := NewReminderNotifier(businesses, appointments, sender)
	n.now = func() time.Time { return now }

	require.NoError(t, n.Run(context.Background()))

	assert.Equal(t, []uuid.UUID{due[0].ID, due[1].ID}, sender.sent)
	assert.Equal(t, []uuid.UUID{due[0].ID, due[1].ID}, appointments.marked)
	assert.Equal(t, now.Add(30*time.Minute), appointments.cutoffArg)
}

func TestReminderNotifier_SkipsDisabled(t *testing.T) {
	businesses := &fakeBusinessLister{businesses: []models.Business{
		{ID: uuid.New(), AppointmentReminderEnabled: false},
	}}
	appointments := &fakeReminderAppointments{}
	sender := &fakeReminderSender{}

	n := NewReminderNotifier(businesses, appointments, sender)
	require.NoError(t, n.Run(context.Background()))

	assert.Empty(t, sender.sent)
	assert.Empty(t, appointments.marked)
}
