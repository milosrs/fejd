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
	business         *store.BusinessStore
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
	business *store.BusinessStore,
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
		business:         business,
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
	// The slot grid is divided by the longest active service the employee
	// offers, so a shorter service still fits within a block and the grid stays
	// uniform. Booking still uses the selected service's actual duration.
	longest, err := s.services.LongestDurationByEmployee(ctx, businessUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get longest service duration: %w", err)
	}
	if longest > 0 {
		duration = time.Duration(longest) * time.Minute
	}
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

// ListEmployeeUnavailability returns the blocked-time ranges for the member
// identified by userID within the business (used by the self-service view).
func (s *SlotService) ListEmployeeUnavailability(ctx context.Context, businessID uuid.UUID, userID string) ([]models.EmployeeUnavailability, error) {
	bu, err := s.businessUser.GetByBusinessAndUser(ctx, businessID, userID)
	if err != nil {
		return nil, fmt.Errorf("target user not found in business: %w", err)
	}

	return s.unavailability.ListByBusinessUser(ctx, bu.ID)
}

// DeleteOwnEmployeeUnavailability removes one of the caller's own blocked-time
// ranges. Unlike DeleteEmployeeUnavailability (admin, any employee), this is
// scoped to the caller's own rows.
func (s *SlotService) DeleteOwnEmployeeUnavailability(ctx context.Context, businessID uuid.UUID, userID string, unavailabilityID uuid.UUID) error {
	bu, err := s.businessUser.GetByBusinessAndUser(ctx, businessID, userID)
	if err != nil {
		return fmt.Errorf("target user not found in business: %w", err)
	}

	err = db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		if err := concurrency.XactLock(ctx, tx, bu.ID); err != nil {
			return fmt.Errorf("failed to acquire lock: %w", err)
		}

		if err := s.unavailability.DeleteForBusinessUser(ctx, tx, bu.ID, unavailabilityID); err != nil {
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

// ListOwnAppointments returns the staff member's own reservations (appointments
// where they are the provider) for the given UTC day, in chronological order.
func (s *SlotService) ListOwnAppointments(ctx context.Context, businessID uuid.UUID, userID string, date time.Time) ([]models.Appointment, error) {
	bu, err := s.businessUser.GetByBusinessAndUser(ctx, businessID, userID)
	if err != nil {
		return nil, fmt.Errorf("target user not found in business: %w", err)
	}

	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.AddDate(0, 0, 1)

	return s.appointments.ListByBusinessUser(ctx, bu.ID, dayStart, dayEnd)
}

// CancelOwnAppointment cancels one of the caller's own active reservations with
// a required reason, then notifies slot subscribers so availability refreshes.
func (s *SlotService) CancelOwnAppointment(ctx context.Context, businessID uuid.UUID, userID string, appointmentID uuid.UUID, reason string) error {
	bu, err := s.businessUser.GetByBusinessAndUser(ctx, businessID, userID)
	if err != nil {
		return fmt.Errorf("target user not found in business: %w", err)
	}

	err = db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		if err := concurrency.XactLock(ctx, tx, bu.ID); err != nil {
			return fmt.Errorf("failed to acquire lock: %w", err)
		}

		if err := s.appointments.CancelByBusinessUser(ctx, tx, appointmentID, bu.ID, reason); err != nil {
			return fmt.Errorf("failed to cancel appointment: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	s.publishSlotsChanged(businessID, bu.ID)
	return nil
}

// BookOwnAppointment books an appointment on the caller's own calendar,
// mirroring the customer booking flow. The customer is optional (walk-in);
// created_by is always the calling member.
func (s *SlotService) BookOwnAppointment(ctx context.Context, businessID uuid.UUID, userID string, serviceID uuid.UUID, startTime time.Time, customerUserID string) (*models.Appointment, error) {
	bu, err := s.businessUser.GetByBusinessAndUser(ctx, businessID, userID)
	if err != nil {
		return nil, fmt.Errorf("target user not found in business: %w", err)
	}

	svc, err := s.services.GetByID(ctx, serviceID)
	if err != nil {
		return nil, fmt.Errorf("service not found: %w", err)
	}
	if svc.BusinessID != businessID {
		return nil, fmt.Errorf("service does not belong to business")
	}

	appointment := &models.Appointment{
		BusinessID:     businessID,
		ServiceID:      serviceID,
		BusinessUserID: bu.ID,
		CustomerUserID: customerUserID,
		StartTime:      startTime,
		EndTime:        startTime.Add(time.Duration(svc.DurationMinutes) * time.Minute),
		Status:         models.AppointmentStatusConfirmed,
		CreatedBy:      userID,
	}

	if err := s.BookAppointment(ctx, appointment); err != nil {
		return nil, err
	}
	return appointment, nil
}

// ListMyServices returns the active services the calling member offers.
func (s *SlotService) ListMyServices(ctx context.Context, businessID uuid.UUID, userID string) ([]models.Service, error) {
	bu, err := s.businessUser.GetByBusinessAndUser(ctx, businessID, userID)
	if err != nil {
		return nil, fmt.Errorf("target user not found in business: %w", err)
	}
	return s.services.ListByBusinessUser(ctx, bu.ID)
}

// ErrNoShowTooEarly is returned when a staff member tries to mark an
// appointment as no-show before the salon's configured grace period has passed.
var ErrNoShowTooEarly = errors.New("no-show cannot be marked yet")

// MarkNoShow marks one of the caller's own active reservations as no-show,
// enforcing the salon's policy that a no-show can only be recorded
// business.no_show_after_hours after the appointment start.
func (s *SlotService) MarkNoShow(ctx context.Context, businessID uuid.UUID, userID string, appointmentID uuid.UUID) error {
	bu, err := s.businessUser.GetByBusinessAndUser(ctx, businessID, userID)
	if err != nil {
		return fmt.Errorf("target user not found in business: %w", err)
	}

	appt, err := s.appointments.GetByID(ctx, appointmentID)
	if err != nil {
		return ErrAppointmentNotFound
	}
	if appt.BusinessUserID != bu.ID {
		return ErrAppointmentNotFound
	}

	b, err := s.business.GetByID(ctx, appt.BusinessID)
	if err != nil {
		return fmt.Errorf("business not found: %w", err)
	}

	grace := time.Duration(b.NoShowAfterHours) * time.Hour
	if time.Since(appt.StartTime) < grace {
		return ErrNoShowTooEarly
	}

	err = db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		if err := concurrency.XactLock(ctx, tx, bu.ID); err != nil {
			return fmt.Errorf("failed to acquire lock: %w", err)
		}

		if err := s.appointments.MarkNoShow(ctx, tx, appointmentID, bu.ID); err != nil {
			return fmt.Errorf("failed to mark no-show: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	s.publishSlotsChanged(businessID, bu.ID)
	return nil
}

// ErrCancellationTooLate is returned when a customer tries to cancel inside the
// salon's configured cancellation notice window.
var ErrCancellationTooLate = errors.New("cancellation window has passed")

// ErrAppointmentNotFound is returned when an appointment does not exist or does
// not belong to the caller.
var ErrAppointmentNotFound = errors.New("appointment not found")

// CancelCustomerAppointment cancels one of a customer's own appointments,
// enforcing the salon's cancellation notice policy (at least
// business.cancellation_lead_hours before the start time).
func (s *SlotService) CancelCustomerAppointment(ctx context.Context, appointmentID uuid.UUID, customerUserID, reason string) error {
	appt, err := s.appointments.GetByID(ctx, appointmentID)
	if err != nil {
		return ErrAppointmentNotFound
	}
	if appt.CustomerUserID != customerUserID {
		return ErrAppointmentNotFound
	}

	b, err := s.business.GetByID(ctx, appt.BusinessID)
	if err != nil {
		return fmt.Errorf("business not found: %w", err)
	}

	lead := time.Duration(b.CancellationLeadHours) * time.Hour
	if time.Until(appt.StartTime) < lead {
		return ErrCancellationTooLate
	}

	return s.appointments.Cancel(ctx, appointmentID, customerUserID, reason)
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
