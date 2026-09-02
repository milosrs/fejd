package service

import (
	"context"
	"errors"
	"fejd-backend/internal/concurrency"
	"fejd-backend/internal/db"
	"fejd-backend/internal/models"
	"fejd-backend/internal/sse"
	"fejd-backend/internal/store"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SlotService struct {
	appointments     *store.AppointmentStore
	workingHours     *store.WorkingHoursStore
	overrides        *store.WorkingHoursOverrideStore
	services         *store.ServiceStore
	businessUser     *store.BusinessUserStore
	employeeServices *store.EmployeeServiceStore
	unavailability   *store.EmployeeUnavailabilityStore
	hub              *sse.Hub
	pool             *pgxpool.Pool
}

func NewSlotService(
	appointments *store.AppointmentStore,
	workingHours *store.WorkingHoursStore,
	overrides *store.WorkingHoursOverrideStore,
	services *store.ServiceStore,
	businessUser *store.BusinessUserStore,
	employeeServices *store.EmployeeServiceStore,
	unavailability *store.EmployeeUnavailabilityStore,
	hub *sse.Hub,
	pool *pgxpool.Pool,
) *SlotService {
	return &SlotService{
		appointments:     appointments,
		workingHours:     workingHours,
		overrides:        overrides,
		services:         services,
		businessUser:     businessUser,
		employeeServices: employeeServices,
		unavailability:   unavailability,
		hub:              hub,
		pool:             pool,
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

	bu, err := s.businessUser.GetByID(ctx, businessUserID)
	if err != nil || !bu.Active {
		return nil, nil
	}

	offers, err := s.employeeServices.OffersService(ctx, businessUserID, serviceID)
	if err != nil || !offers {
		return nil, nil
	}

	dayOfWeek := int(date.Weekday())
	override, _ := s.overrides.GetByBusinessUserAndDate(ctx, businessUserID, date)

	if override != nil && override.IsOff {
		return nil, nil
	}

	var startTime, endTime time.Time
	found := false
	if override != nil && override.StartTime != nil && override.EndTime != nil {
		startTime = *override.StartTime
		endTime = *override.EndTime
		found = true
	} else {
		hours, err := s.workingHours.GetByBusinessUser(ctx, businessUserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get working hours: %w", err)
		}

		for _, wh := range hours {
			if wh.DayOfWeek == dayOfWeek {
				startTime = wh.StartTime
				endTime = wh.EndTime
				found = true
				break
			}
		}
	}
	if !found {
		return nil, nil
	}

	dayStart := time.Date(date.Year(), date.Month(), date.Day(), startTime.Hour(), startTime.Minute(), startTime.Second(), 0, date.Location())
	dayEnd := time.Date(date.Year(), date.Month(), date.Day(), endTime.Hour(), endTime.Minute(), endTime.Second(), 0, date.Location())

	duration := time.Duration(svc.DurationMinutes) * time.Minute
	existing, err := s.appointments.GetConflictingAppointments(ctx, businessID, businessUserID, dayStart, dayEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing appointments: %w", err)
	}

	unavail, err := s.unavailability.ListOverlapping(ctx, businessUserID, dayStart, dayEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get unavailability: %w", err)
	}

	busySlots := make([]models.TimeSlot, 0, len(existing)+len(unavail))
	for _, a := range existing {
		busySlots = append(busySlots, models.TimeSlot{StartTime: a.StartTime, EndTime: a.EndTime})
	}
	for _, u := range unavail {
		busySlots = append(busySlots, models.TimeSlot{StartTime: u.StartTime, EndTime: u.EndTime})
	}

	slots := computeSlots(dayStart, dayEnd, duration, busySlots)
	return slots, nil
}

func (s *SlotService) BookAppointment(ctx context.Context, appointment *models.Appointment) error {
	if appointment.Status == "" {
		appointment.Status = models.AppointmentStatusConfirmed
	}
	if appointment.CreatedBy == "" {
		appointment.CreatedBy = appointment.CustomerUserID
	}

	svc, err := s.services.GetByID(ctx, appointment.ServiceID)
	if err != nil {
		return fmt.Errorf("service not found: %w", err)
	}

	expectedEnd := appointment.StartTime.Add(time.Duration(svc.DurationMinutes) * time.Minute)
	if !appointment.EndTime.Equal(expectedEnd) {
		return fmt.Errorf("appointment end time does not match service duration")
	}

	bu, err := s.businessUser.GetByID(ctx, appointment.BusinessUserID)
	if err != nil {
		return fmt.Errorf("employee not found: %w", err)
	}
	if !bu.Active {
		return fmt.Errorf("employee is inactive")
	}

	offers, err := s.employeeServices.OffersService(ctx, appointment.BusinessUserID, appointment.ServiceID)
	if err != nil {
		return fmt.Errorf("failed to check employee services: %w", err)
	}
	if !offers {
		return fmt.Errorf("employee does not offer this service")
	}

	unavail, err := s.unavailability.ListOverlapping(ctx, appointment.BusinessUserID, appointment.StartTime, appointment.EndTime)
	if err != nil {
		return fmt.Errorf("failed to check unavailability: %w", err)
	}
	if len(unavail) > 0 {
		return fmt.Errorf("employee unavailable at this time")
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

	// Serialize per-employee writes. The exclusion constraint is the real
	// booking-vs-booking guard; the advisory lock closes the race with
	// unavailability/deactivation writes, which take the same lock.
	err = db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		if err := concurrency.XactLock(ctx, tx, appointment.BusinessUserID); err != nil {
			return fmt.Errorf("failed to acquire booking lock: %w", err)
		}

		if err := s.appointments.Create(ctx, tx, appointment); err != nil {
			return mapAppointmentError(err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	s.hub.Publish(appointment.BusinessID.String(), map[string]any{
		"type":             "appointment_booked",
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
		"date":             date.Format(time.DateOnly),
	})
}

func (s *SlotService) publishSlotsChanged(businessID, businessUserID uuid.UUID) {
	s.hub.Publish(businessID.String(), map[string]any{
		"type":             "slots_updated",
		"business_user_id": businessUserID.String(),
	})
}

// SetEmployeeServices replaces the set of services an employee offers.
func (s *SlotService) SetEmployeeServices(ctx context.Context, businessID uuid.UUID, userID string, serviceIDs []uuid.UUID) error {
	bu, err := s.businessUser.GetByBusinessAndUser(ctx, businessID, userID)
	if err != nil {
		return fmt.Errorf("target user not found in business: %w", err)
	}

	if err := s.employeeServices.ReplaceByBusinessUser(ctx, bu.ID, serviceIDs); err != nil {
		return fmt.Errorf("failed to set employee services: %w", err)
	}

	s.publishSlotsChanged(businessID, bu.ID)
	return nil
}

// AddEmployeeUnavailability marks an employee unavailable (e.g. vacation). The
// write takes the same per-employee advisory lock as booking so it serializes
// against concurrent bookings.
func (s *SlotService) AddEmployeeUnavailability(ctx context.Context, businessID uuid.UUID, userID string, u *models.EmployeeUnavailability) error {
	bu, err := s.businessUser.GetByBusinessAndUser(ctx, businessID, userID)
	if err != nil {
		return fmt.Errorf("target user not found in business: %w", err)
	}
	u.BusinessUserID = bu.ID

	err = db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		if err := concurrency.XactLock(ctx, tx, bu.ID); err != nil {
			return fmt.Errorf("failed to acquire lock: %w", err)
		}

		if err := s.unavailability.Create(ctx, tx, u); err != nil {
			return fmt.Errorf("failed to create unavailability: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	s.publishSlotsChanged(businessID, bu.ID)
	return nil
}

// DeleteEmployeeUnavailability removes an unavailability block.
func (s *SlotService) DeleteEmployeeUnavailability(ctx context.Context, businessID uuid.UUID, userID string, unavailabilityID uuid.UUID) error {
	bu, err := s.businessUser.GetByBusinessAndUser(ctx, businessID, userID)
	if err != nil {
		return fmt.Errorf("target user not found in business: %w", err)
	}

	err = db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		if err := concurrency.XactLock(ctx, tx, bu.ID); err != nil {
			return fmt.Errorf("failed to acquire lock: %w", err)
		}

		if err := s.unavailability.Delete(ctx, tx, unavailabilityID); err != nil {
			return fmt.Errorf("failed to delete unavailability: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	s.publishSlotsChanged(businessID, bu.ID)
	return nil
}

func mapAppointmentError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23P01":
			return fmt.Errorf("time slot is no longer available")
		case "23505":
			return fmt.Errorf("already booked today")
		case "23503":
			return fmt.Errorf("employee does not offer this service")
		}
	}
	return err
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
