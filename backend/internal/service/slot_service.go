package service

import (
	"context"
	"fejd-backend/internal/models"
	"fejd-backend/internal/sse"
	"fejd-backend/internal/store"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type SlotService struct {
	appointments    *store.AppointmentStore
	workingHours    *store.WorkingHoursStore
	overrides       *store.WorkingHoursOverrideStore
	services        *store.ServiceStore
	businessUser    *store.BusinessUserStore
	hub             *sse.Hub
}

func NewSlotService(
	appointments *store.AppointmentStore,
	workingHours *store.WorkingHoursStore,
	overrides *store.WorkingHoursOverrideStore,
	services *store.ServiceStore,
	businessUser *store.BusinessUserStore,
	hub *sse.Hub,
) *SlotService {
	return &SlotService{
		appointments: appointments,
		workingHours: workingHours,
		overrides:    overrides,
		services:     services,
		businessUser: businessUser,
		hub:          hub,
	}
}

func (s *SlotService) GetAvailableSlots(
	ctx context.Context,
	businessID uuid.UUID,
	serviceID uuid.UUID,
	businessUserID uuid.UUID,
	date time.Time,
) ([]models.TimeSlot, error) {
	svc, err := s.services.GetByID(ctx, serviceID)
	if err != nil {
		return nil, fmt.Errorf("service not found: %w", err)
	}

	if svc.BusinessID != businessID {
		return nil, fmt.Errorf("service does not belong to business")
	}

	dayOfWeek := int(date.Weekday())
	override, _ := s.overrides.GetByBusinessUserAndDate(ctx, businessUserID, date)

	if override != nil && override.IsOff {
		return nil, nil
	}

	var startTime, endTime string
	if override != nil && override.StartTime != nil && override.EndTime != nil {
		startTime = *override.StartTime
		endTime = *override.EndTime
	} else {
		hours, err := s.workingHours.GetByBusinessUser(ctx, businessUserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get working hours: %w", err)
		}

		found := false
		for _, wh := range hours {
			if wh.DayOfWeek == dayOfWeek {
				startTime = wh.StartTime
				endTime = wh.EndTime
				found = true
				break
			}
		}
		if !found {
			return nil, nil
		}
	}

	dateStr := date.Format("2006-01-02")
	dayStart, err := time.Parse("2006-01-02 15:04:05", dateStr+" "+startTime+":00")
	if err != nil {
		return nil, fmt.Errorf("invalid start time: %w", err)
	}
	dayEnd, err := time.Parse("2006-01-02 15:04:05", dateStr+" "+endTime+":00")
	if err != nil {
		return nil, fmt.Errorf("invalid end time: %w", err)
	}

	duration := time.Duration(svc.DurationMinutes) * time.Minute
	existing, err := s.appointments.GetConflictingAppointments(ctx, businessID, businessUserID, dayStart, dayEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing appointments: %w", err)
	}

	busySlots := make([]models.TimeSlot, 0, len(existing))
	for _, a := range existing {
		busySlots = append(busySlots, models.TimeSlot{StartTime: a.StartTime, EndTime: a.EndTime})
	}

	slots := computeSlots(dayStart, dayEnd, duration, busySlots)
	return slots, nil
}

func (s *SlotService) BookAppointment(ctx context.Context, appointment *models.Appointment) error {
	if appointment.Status == "" {
		appointment.Status = models.AppointmentStatusConfirmed
	}

	svc, err := s.services.GetByID(ctx, appointment.ServiceID)
	if err != nil {
		return fmt.Errorf("service not found: %w", err)
	}

	expectedEnd := appointment.StartTime.Add(time.Duration(svc.DurationMinutes) * time.Minute)
	if !appointment.EndTime.Equal(expectedEnd) {
		return fmt.Errorf("appointment end time does not match service duration")
	}

	existing, err := s.appointments.GetConflictingAppointments(ctx,
		appointment.BusinessID, appointment.BusinessUserID,
		appointment.StartTime, appointment.EndTime,
	)
	if err != nil {
		return fmt.Errorf("failed to check conflicts: %w", err)
	}
	if len(existing) > 0 {
		return fmt.Errorf("time slot is no longer available")
	}

	if err := s.appointments.Create(ctx, appointment); err != nil {
		return fmt.Errorf("failed to create appointment: %w", err)
	}

	s.hub.Publish(appointment.BusinessID.String(), map[string]any{
		"type":            "appointment_booked",
		"business_user_id": appointment.BusinessUserID.String(),
		"start_time":       appointment.StartTime.Format(time.RFC3339),
		"end_time":         appointment.EndTime.Format(time.RFC3339),
	})

	return nil
}

func (s *SlotService) PublishSlotUpdate(businessID uuid.UUID, businessUserID uuid.UUID, date time.Time) {
	s.hub.Publish(businessID.String(), map[string]any{
		"type":             "slots_updated",
		"business_user_id": businessUserID.String(),
		"date":             date.Format("2006-01-02"),
	})
}

func computeSlots(dayStart, dayEnd time.Time, slotDuration time.Duration, busySlots []models.TimeSlot) []models.TimeSlot {
	var slots []models.TimeSlot
	current := dayStart

	for current.Add(slotDuration).Compare(dayEnd) <= 0 {
		slotEnd := current.Add(slotDuration)

		conflict := false
		for _, busy := range busySlots {
			if current.Before(busy.EndTime) && slotEnd.After(busy.StartTime) {
				conflict = true
				break
			}
		}

		if !conflict && current.After(time.Now()) {
			slots = append(slots, models.TimeSlot{StartTime: current, EndTime: slotEnd})
		}

		current = slotEnd
	}

	return slots
}
