package jobs

import (
	"context"
	"log"
	"time"

	"fejd-backend/internal/models"

	"github.com/google/uuid"
)

// reminderBusinessLister lists businesses (the job filters to those with the
// reminder policy enabled). *store.BusinessStore satisfies it.
type reminderBusinessLister interface {
	List(ctx context.Context) ([]models.Business, error)
}

// reminderAppointmentStore finds appointments due for a reminder and stamps
// them once sent. *store.AppointmentStore satisfies it.
type reminderAppointmentStore interface {
	ListDueForReminder(ctx context.Context, businessID uuid.UUID, now, cutoff time.Time) ([]models.Appointment, error)
	MarkReminderSent(ctx context.Context, id uuid.UUID) error
}

// reminderSender delivers an "appointment about to begin" notification.
// *service.NotificationService satisfies it.
type reminderSender interface {
	NotifyAppointmentReminder(ctx context.Context, appt *models.Appointment) error
}

// ReminderNotifier is the scheduled job that reminds customers when their
// confirmed appointment is about to begin, per each salon's reminder policy.
type ReminderNotifier struct {
	businesses   reminderBusinessLister
	appointments reminderAppointmentStore
	sender       reminderSender
	now          func() time.Time
}

func NewReminderNotifier(businesses reminderBusinessLister, appointments reminderAppointmentStore, sender reminderSender) *ReminderNotifier {
	return &ReminderNotifier{
		businesses:   businesses,
		appointments: appointments,
		sender:       sender,
		now:          time.Now,
	}
}

// Run scans every salon with the reminder policy enabled and notifies the
// customer of each confirmed appointment whose start time has entered the
// salon's reminder window.
func (r *ReminderNotifier) Run(ctx context.Context) error {
	businesses, err := r.businesses.List(ctx)
	if err != nil {
		return err
	}

	now := r.now().UTC()
	for _, b := range businesses {
		if !b.AppointmentReminderEnabled {
			continue
		}

		cutoff := now.Add(time.Duration(b.AppointmentReminderLeadMinutes) * time.Minute)
		due, err := r.appointments.ListDueForReminder(ctx, b.ID, now, cutoff)
		if err != nil {
			log.Printf("[reminder] failed to list due appointments for business %s: %v", b.ID, err)
			continue
		}

		for _, appt := range due {
			if err := r.sender.NotifyAppointmentReminder(ctx, &appt); err != nil {
				log.Printf("[reminder] failed to notify appointment %s: %v", appt.ID, err)
				continue
			}
			if err := r.appointments.MarkReminderSent(ctx, appt.ID); err != nil {
				log.Printf("[reminder] failed to mark reminder sent for appointment %s: %v", appt.ID, err)
			}
		}
	}

	return nil
}
