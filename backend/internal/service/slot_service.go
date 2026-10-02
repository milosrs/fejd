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
	"log"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SlotService struct {
	appointments     *store.AppointmentStore
	workingHours     *store.WorkingHoursStore
	businessHours    *store.BusinessHoursStore
	overrides        *store.WorkingHoursOverrideStore
	closures         *store.BusinessClosureStore
	services         *store.ServiceStore
	business         *store.BusinessStore
	businessUser     *store.BusinessUserStore
	employeeServices *store.EmployeeServiceStore
	unavailability   *store.EmployeeUnavailabilityStore
	combinations     *store.ServiceCombinationStore
	hub              *sse.Hub
	pool             *pgxpool.Pool
	notifier         *NotificationService
}

func NewSlotService(
	appointments *store.AppointmentStore,
	workingHours *store.WorkingHoursStore,
	businessHours *store.BusinessHoursStore,
	overrides *store.WorkingHoursOverrideStore,
	closures *store.BusinessClosureStore,
	services *store.ServiceStore,
	business *store.BusinessStore,
	businessUser *store.BusinessUserStore,
	employeeServices *store.EmployeeServiceStore,
	unavailability *store.EmployeeUnavailabilityStore,
	combinations *store.ServiceCombinationStore,
	hub *sse.Hub,
	pool *pgxpool.Pool,
) *SlotService {
	return &SlotService{
		appointments:     appointments,
		workingHours:     workingHours,
		businessHours:    businessHours,
		overrides:        overrides,
		closures:         closures,
		services:         services,
		business:         business,
		businessUser:     businessUser,
		employeeServices: employeeServices,
		unavailability:   unavailability,
		combinations:     combinations,
		hub:              hub,
		pool:             pool,
	}
}

// SetNotifier wires the optional notification service used to push a "new
// booking" alert to the selected provider. It is nil by default (no-op).
func (s *SlotService) SetNotifier(n *NotificationService) {
	s.notifier = n
}

// resolveCombinedServices loads and validates the base service plus its add-on
// services for a booking. It ensures every service belongs to the business, is
// active, appears at most once, and (for add-ons) is a configured combinable
// service of the base.
func (s *SlotService) resolveCombinedServices(ctx context.Context, businessID, serviceID uuid.UUID, additionalIDs []uuid.UUID) (*models.Service, []models.Service, error) {
	base, err := s.services.GetByID(ctx, serviceID)
	if err != nil {
		return nil, nil, fmt.Errorf("service not found: %w", err)
	}
	if base.BusinessID != businessID {
		return nil, nil, fmt.Errorf("service does not belong to business")
	}
	if !base.Active {
		return nil, nil, fmt.Errorf("service is inactive")
	}

	seen := map[uuid.UUID]struct{}{serviceID: {}}
	addons := make([]models.Service, 0, len(additionalIDs))
	for _, id := range additionalIDs {
		if _, ok := seen[id]; ok {
			return nil, nil, fmt.Errorf("duplicate service in combination")
		}
		seen[id] = struct{}{}

		addon, err := s.services.GetByID(ctx, id)
		if err != nil {
			return nil, nil, fmt.Errorf("service not found: %w", err)
		}
		if addon.BusinessID != businessID {
			return nil, nil, fmt.Errorf("service does not belong to business")
		}
		if !addon.Active {
			return nil, nil, fmt.Errorf("service is inactive")
		}
		addons = append(addons, *addon)
	}

	if len(addons) > 0 && s.combinations != nil {
		edges, err := s.combinations.ListByService(ctx, serviceID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to list service combinations: %w", err)
		}
		allowed := make(map[uuid.UUID]struct{}, len(edges))
		for _, e := range edges {
			allowed[e.CombinableServiceID] = struct{}{}
		}
		for _, addon := range addons {
			if _, ok := allowed[addon.ID]; !ok {
				return nil, nil, fmt.Errorf("service combination not allowed")
			}
		}
	}

	return base, addons, nil
}

// combinedDuration is the total reservation duration of a base service plus its
// add-on services.
func combinedDuration(base *models.Service, addons []models.Service) time.Duration {
	d := time.Duration(base.DurationMinutes) * time.Minute
	for _, a := range addons {
		d += time.Duration(a.DurationMinutes) * time.Minute
	}
	return d
}

// CombinedDuration resolves and validates a base service plus its add-ons and
// returns the total reservation duration, reusing the same validation as
// booking and slot computation.
func (s *SlotService) CombinedDuration(ctx context.Context, businessID, serviceID uuid.UUID, additionalIDs []uuid.UUID) (time.Duration, error) {
	base, addons, err := s.resolveCombinedServices(ctx, businessID, serviceID, additionalIDs)
	if err != nil {
		return 0, err
	}
	return combinedDuration(base, addons), nil
}

func (s *SlotService) GetAvailableSlots(
	ctx context.Context,
	businessID uuid.UUID,
	serviceID uuid.UUID,
	additionalServiceIDs []uuid.UUID,
	businessUserID uuid.UUID,
	date time.Time,
) ([]models.TimeSlot, error) {
	base, addons, err := s.resolveCombinedServices(ctx, businessID, serviceID, additionalServiceIDs)
	if err != nil {
		return nil, err
	}

	if s.closures != nil {
		closed, err := s.closures.IsClosed(ctx, businessID, date)
		if err != nil {
			return nil, fmt.Errorf("failed to check salon closures: %w", err)
		}
		if closed {
			return nil, nil
		}
	}

	bu, err := s.businessUser.GetByID(ctx, businessUserID)
	if err != nil || !bu.Active {
		return nil, nil
	}

	for _, svc := range append([]models.Service{*base}, addons...) {
		offers, err := s.employeeServices.OffersService(ctx, businessUserID, svc.ID)
		if err != nil || !offers {
			return nil, nil
		}
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

		// Fall back to the salon's default working hours when the employee has
		// none of their own.
		if !found {
			if salonHours, err := s.businessHours.ListByBusiness(ctx, businessID); err == nil {
				for _, bh := range salonHours {
					if bh.DayOfWeek == dayOfWeek {
						startTime = bh.StartTime
						endTime = bh.EndTime
						found = true
						break
					}
				}
			}
		}
	}
	if !found {
		return nil, nil
	}

	dayStart := time.Date(date.Year(), date.Month(), date.Day(), startTime.Hour(), startTime.Minute(), startTime.Second(), 0, date.Location())
	dayEnd := time.Date(date.Year(), date.Month(), date.Day(), endTime.Hour(), endTime.Minute(), endTime.Second(), 0, date.Location())

	// Time slots are computed dynamically from the combined services' total
	// duration and the employee's existing reservations and unavailability, so
	// the service is squeezed into the day's free windows. The salon's slot
	// interval no longer drives the grid; it is only a UI default for new
	// service durations.
	serviceDuration := combinedDuration(base, addons)

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

	slots := computeSlots(dayStart, dayEnd, serviceDuration, busySlots)
	return slots, nil
}

func (s *SlotService) BookAppointment(ctx context.Context, appointment *models.Appointment) error {
	if appointment.CreatedBy == "" {
		appointment.CreatedBy = appointment.CustomerUserID
	}

	// New appointments start as pending unless the salon has automatic approval
	// enabled, in which case they are confirmed straight away.
	b, err := s.business.GetByID(ctx, appointment.BusinessID)
	if err != nil {
		return fmt.Errorf("business not found: %w", err)
	}
	if b.AutoApprove {
		appointment.Status = models.AppointmentStatusConfirmed
	} else {
		appointment.Status = models.AppointmentStatusPending
	}

	svc, addons, err := s.resolveCombinedServices(ctx, appointment.BusinessID, appointment.ServiceID, appointment.AdditionalServiceIDs)
	if err != nil {
		return err
	}

	expectedEnd := appointment.StartTime.Add(combinedDuration(svc, addons))
	if !appointment.EndTime.Equal(expectedEnd) {
		return fmt.Errorf("appointment end time does not match service duration")
	}

	if s.closures != nil {
		closed, err := s.closures.IsClosed(ctx, appointment.BusinessID, appointment.StartTime)
		if err != nil {
			return fmt.Errorf("failed to check salon closures: %w", err)
		}
		if closed {
			return fmt.Errorf("salon is closed on this day")
		}
	}

	bu, err := s.businessUser.GetByID(ctx, appointment.BusinessUserID)
	if err != nil {
		return fmt.Errorf("employee not found: %w", err)
	}
	if !bu.Active {
		return fmt.Errorf("employee is inactive")
	}

	for _, svc := range append([]models.Service{*svc}, addons...) {
		offers, err := s.employeeServices.OffersService(ctx, appointment.BusinessUserID, svc.ID)
		if err != nil {
			return fmt.Errorf("failed to check employee services: %w", err)
		}
		if !offers {
			return fmt.Errorf("employee does not offer this service")
		}
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

		if len(addons) > 0 {
			addonIDs := make([]uuid.UUID, len(addons))
			for i, a := range addons {
				addonIDs[i] = a.ID
			}
			if err := s.appointments.InsertAdditionalServices(ctx, tx, appointment.ID, addonIDs); err != nil {
				return fmt.Errorf("failed to save combined services: %w", err)
			}
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

	if s.notifier != nil {
		if err := s.notifier.NotifyNewBooking(ctx, appointment); err != nil {
			log.Printf("[booking] failed to notify provider: %v", err)
		}
	}

	return nil
}

func (s *SlotService) PublishSlotUpdate(businessID uuid.UUID, businessUserID uuid.UUID, date time.Time) {
	s.hub.Publish(businessID.String(), map[string]any{
		"type":             "slots_updated",
		"business_user_id": businessUserID.String(),
		"date":             date.Format(time.DateOnly),
	})
}

// PublishSlotsChangedForBusiness publishes a slots_updated event for every
// subscriber of the business. Used when a business-wide policy (e.g. the slot
// interval) changes and every provider's availability should refresh.
func (s *SlotService) PublishSlotsChangedForBusiness(businessID uuid.UUID) {
	s.hub.Publish(businessID.String(), map[string]any{"type": "slots_updated"})
}

// PublishClosuresChangedForBusiness publishes a closures_updated event for every
// subscriber of the business, so booking pages disable newly closed days
// immediately.
func (s *SlotService) PublishClosuresChangedForBusiness(businessID uuid.UUID) {
	s.hub.Publish(businessID.String(), map[string]any{"type": "closures_updated"})
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

// SetServiceEmployees replaces the set of employees mapped to a service.
// Employees removed from the service have their future reservations cancelled
// (with a localized reason) before the mapping is replaced.
func (s *SlotService) SetServiceEmployees(ctx context.Context, businessID, serviceID uuid.UUID, businessUserIDs []uuid.UUID) error {
	svc, err := s.services.GetByID(ctx, serviceID)
	if err != nil || svc.BusinessID != businessID {
		return fmt.Errorf("service not found in business")
	}

	old, err := s.employeeServices.ListByService(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("failed to list service employees: %w", err)
	}

	for _, buID := range businessUserIDs {
		bu, err := s.businessUser.GetByID(ctx, buID)
		if err != nil || bu.BusinessID != businessID {
			return fmt.Errorf("employee not found in business")
		}
	}

	keep := make(map[uuid.UUID]struct{}, len(businessUserIDs))
	for _, buID := range businessUserIDs {
		keep[buID] = struct{}{}
	}

	affected := map[uuid.UUID]struct{}{}
	var removed []uuid.UUID
	for _, link := range old {
		affected[link.BusinessUserID] = struct{}{}
		if _, ok := keep[link.BusinessUserID]; !ok {
			removed = append(removed, link.BusinessUserID)
		}
	}
	for _, buID := range businessUserIDs {
		affected[buID] = struct{}{}
	}

	for _, buID := range removed {
		if err := s.cancelServiceAppointments(ctx, businessID, buID, serviceID); err != nil {
			return err
		}
	}

	if err := s.employeeServices.ReplaceByService(ctx, serviceID, businessUserIDs); err != nil {
		return fmt.Errorf("failed to set service employees: %w", err)
	}

	for buID := range affected {
		s.publishSlotsChanged(businessID, buID)
	}
	return nil
}

// cancellationReasonNoService is the translation key stored as the reason when
// a service is unmapped from an employee; the frontend appends the unmapping
// date and localizes the label.
const cancellationReasonNoService = "cancellation.reason.noService"

// cancelServiceAppointments cancels an employee's future pending/confirmed
// reservations for a service, used when the service is unmapped from them.
func (s *SlotService) cancelServiceAppointments(ctx context.Context, businessID, businessUserID, serviceID uuid.UUID) error {
	now := time.Now()
	future, err := s.appointments.ListByBusinessUser(ctx, businessUserID, now, now.AddDate(1, 0, 0))
	if err != nil {
		return fmt.Errorf("failed to list future appointments: %w", err)
	}

	reason := fmt.Sprintf("%s|%s", cancellationReasonNoService, now.UTC().Format(time.DateOnly))
	for _, appt := range future {
		if appt.ServiceID != serviceID {
			continue
		}
		if appt.Status != models.AppointmentStatusPending && appt.Status != models.AppointmentStatusConfirmed {
			continue
		}

		if err := s.appointments.CancelByID(ctx, appt.ID, reason); err != nil {
			return fmt.Errorf("failed to cancel appointment: %w", err)
		}

		s.hub.Publish(businessID.String(), map[string]any{
			"type":             "appointment_cancelled",
			"appointment_id":   appt.ID.String(),
			"customer_user_id": appt.CustomerUserID,
			"start_time":       appt.StartTime.Format(time.RFC3339),
		})
	}
	return nil
}

// ListServiceEmployees returns the business users mapped to a service.
func (s *SlotService) ListServiceEmployees(ctx context.Context, businessID, serviceID uuid.UUID) ([]models.BusinessUser, error) {
	svc, err := s.services.GetByID(ctx, serviceID)
	if err != nil || svc.BusinessID != businessID {
		return nil, fmt.Errorf("service not found in business")
	}

	links, err := s.employeeServices.ListByService(ctx, serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list service employees: %w", err)
	}

	users := make([]models.BusinessUser, 0, len(links))
	for _, link := range links {
		bu, err := s.businessUser.GetByID(ctx, link.BusinessUserID)
		if err != nil {
			continue
		}
		users = append(users, *bu)
	}
	return users, nil
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

// ListBusinessUnavailability returns all blocked-time ranges across a business
// (used by the owner to acknowledge employee reservations).
func (s *SlotService) ListBusinessUnavailability(ctx context.Context, businessID uuid.UUID) ([]models.EmployeeUnavailability, error) {
	return s.unavailability.ListByBusiness(ctx, businessID)
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
// where they are the provider) whose start time falls within the given
// half-open [from, to) UTC interval, in chronological order.
func (s *SlotService) ListOwnAppointments(ctx context.Context, businessID uuid.UUID, userID string, from, to time.Time) ([]models.Appointment, error) {
	bu, err := s.businessUser.GetByBusinessAndUser(ctx, businessID, userID)
	if err != nil {
		return nil, fmt.Errorf("target user not found in business: %w", err)
	}

	return s.appointments.ListByBusinessUser(ctx, bu.ID, from, to)
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

// ListCombinableServices returns the active services a base service may be
// combined with, ordered by name.
func (s *SlotService) ListCombinableServices(ctx context.Context, businessID, serviceID uuid.UUID) ([]models.Service, error) {
	base, err := s.services.GetByID(ctx, serviceID)
	if err != nil || base.BusinessID != businessID {
		return nil, fmt.Errorf("service not found in business")
	}

	if s.combinations == nil {
		return nil, nil
	}

	edges, err := s.combinations.ListByService(ctx, serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list service combinations: %w", err)
	}

	result := make([]models.Service, 0, len(edges))
	for _, e := range edges {
		svc, err := s.services.GetByID(ctx, e.CombinableServiceID)
		if err != nil {
			continue
		}
		if svc.Active && svc.BusinessID == businessID {
			result = append(result, *svc)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

// SetServiceCombinations replaces the set of services combinable onto a base
// service. Every combinable service must belong to the same business and differ
// from the base.
func (s *SlotService) SetServiceCombinations(ctx context.Context, businessID, serviceID uuid.UUID, combinableIDs []uuid.UUID) error {
	base, err := s.services.GetByID(ctx, serviceID)
	if err != nil || base.BusinessID != businessID {
		return fmt.Errorf("service not found in business")
	}

	seen := map[uuid.UUID]struct{}{serviceID: {}}
	for _, id := range combinableIDs {
		if _, ok := seen[id]; ok {
			return fmt.Errorf("service cannot be combined with itself or duplicated")
		}
		seen[id] = struct{}{}

		svc, err := s.services.GetByID(ctx, id)
		if err != nil || svc.BusinessID != businessID {
			return fmt.Errorf("combinable service not found in business")
		}
	}

	if s.combinations == nil {
		return fmt.Errorf("service combinations unavailable")
	}

	if err := s.combinations.ReplaceByService(ctx, serviceID, combinableIDs); err != nil {
		return fmt.Errorf("failed to set service combinations: %w", err)
	}
	return nil
}

// ErrNoShowTooEarly is returned when a staff member tries to mark an
// appointment as no-show before the salon's configured grace period has passed.
var ErrNoShowTooEarly = errors.New("no-show cannot be marked yet")

// MarkNoShow marks one of the caller's own active reservations as no-show,
// enforcing the salon's policy that a no-show can only be recorded
// business.no_show_after_minutes after the appointment start.
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

	grace := time.Duration(b.NoShowAfterMinutes) * time.Minute
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

// ErrForbidden is returned when the caller is not allowed to perform an action.
var ErrForbidden = errors.New("not allowed to perform this action")

// ErrReasonRequired is returned when a non-owner rejects an appointment without
// a reason. Owners may decline without one.
var ErrReasonRequired = errors.New("reason is required")

// AcceptAppointment acknowledges a pending customer appointment. The
// appointment's provider or the business owner may accept it.
func (s *SlotService) AcceptAppointment(ctx context.Context, businessID uuid.UUID, callerUserID string, appointmentID uuid.UUID) error {
	caller, err := s.businessUser.GetByBusinessAndUser(ctx, businessID, callerUserID)
	if err != nil {
		return fmt.Errorf("target user not found in business: %w", err)
	}

	appt, err := s.appointments.GetByID(ctx, appointmentID)
	if err != nil {
		return ErrAppointmentNotFound
	}
	if appt.BusinessID != businessID {
		return ErrAppointmentNotFound
	}

	if caller.ID != appt.BusinessUserID && caller.Role != "admin" {
		return ErrForbidden
	}

	if err := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		if err := concurrency.XactLock(ctx, tx, appt.BusinessUserID); err != nil {
			return fmt.Errorf("failed to acquire lock: %w", err)
		}
		return s.appointments.Acknowledge(ctx, tx, appointmentID, appt.BusinessUserID)
	}); err != nil {
		return err
	}

	s.publishSlotsChanged(businessID, appt.BusinessUserID)
	return nil
}

// AcceptUnavailability acknowledges an employee's pending blocked-time
// reservation. Only the business owner may accept it.
func (s *SlotService) AcceptUnavailability(ctx context.Context, businessID uuid.UUID, callerUserID string, unavailabilityID uuid.UUID) error {
	caller, err := s.businessUser.GetByBusinessAndUser(ctx, businessID, callerUserID)
	if err != nil {
		return fmt.Errorf("target user not found in business: %w", err)
	}
	if caller.Role != "admin" {
		return ErrForbidden
	}

	u, err := s.unavailability.GetByID(ctx, unavailabilityID)
	if err != nil {
		return err
	}

	if err := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		if err := concurrency.XactLock(ctx, tx, u.BusinessUserID); err != nil {
			return fmt.Errorf("failed to acquire lock: %w", err)
		}
		return s.unavailability.Acknowledge(ctx, tx, unavailabilityID)
	}); err != nil {
		return err
	}

	s.publishSlotsChanged(businessID, u.BusinessUserID)
	return nil
}

// RejectAppointment cancels a pending customer appointment with a reason. The
// appointment's provider or the business owner may reject it.
func (s *SlotService) RejectAppointment(ctx context.Context, businessID uuid.UUID, callerUserID string, appointmentID uuid.UUID, reason string) error {
	caller, err := s.businessUser.GetByBusinessAndUser(ctx, businessID, callerUserID)
	if err != nil {
		return fmt.Errorf("target user not found in business: %w", err)
	}

	appt, err := s.appointments.GetByID(ctx, appointmentID)
	if err != nil {
		return ErrAppointmentNotFound
	}
	if appt.BusinessID != businessID {
		return ErrAppointmentNotFound
	}

	if caller.ID != appt.BusinessUserID && caller.Role != "admin" {
		return ErrForbidden
	}

	// Employees must provide a reason when declining; the owner may decline
	// without one.
	if reason == "" && caller.Role != "admin" {
		return ErrReasonRequired
	}

	if err := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		if err := concurrency.XactLock(ctx, tx, appt.BusinessUserID); err != nil {
			return fmt.Errorf("failed to acquire lock: %w", err)
		}
		return s.appointments.Reject(ctx, tx, appointmentID, appt.BusinessUserID, reason)
	}); err != nil {
		return err
	}

	s.publishSlotsChanged(businessID, appt.BusinessUserID)
	return nil
}

// RejectUnavailability rejects an employee's pending blocked-time reservation
// with a reason. Only the business owner may reject it.
func (s *SlotService) RejectUnavailability(ctx context.Context, businessID uuid.UUID, callerUserID string, unavailabilityID uuid.UUID, reason string) error {
	caller, err := s.businessUser.GetByBusinessAndUser(ctx, businessID, callerUserID)
	if err != nil {
		return fmt.Errorf("target user not found in business: %w", err)
	}
	if caller.Role != "admin" {
		return ErrForbidden
	}

	u, err := s.unavailability.GetByID(ctx, unavailabilityID)
	if err != nil {
		return err
	}

	if err := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		if err := concurrency.XactLock(ctx, tx, u.BusinessUserID); err != nil {
			return fmt.Errorf("failed to acquire lock: %w", err)
		}
		return s.unavailability.Reject(ctx, tx, unavailabilityID, reason)
	}); err != nil {
		return err
	}

	s.publishSlotsChanged(businessID, u.BusinessUserID)
	return nil
}

func computeSlots(dayStart, dayEnd time.Time, serviceDuration time.Duration, busySlots []models.TimeSlot) []models.TimeSlot {
	if serviceDuration <= 0 {
		return nil
	}

	type interval struct{ start, end time.Time }

	// Clip busy intervals to the working day and drop empty ones.
	busy := make([]interval, 0, len(busySlots))
	for _, b := range busySlots {
		start, end := b.StartTime, b.EndTime
		if start.Before(dayStart) {
			start = dayStart
		}
		if end.After(dayEnd) {
			end = dayEnd
		}
		if !start.Before(end) {
			continue
		}
		busy = append(busy, interval{start: start, end: end})
	}

	sort.Slice(busy, func(i, j int) bool { return busy[i].start.Before(busy[j].start) })

	// Merge overlapping and adjacent busy intervals.
	var merged []interval
	for _, b := range busy {
		if len(merged) == 0 {
			merged = append(merged, b)
			continue
		}
		last := &merged[len(merged)-1]
		if !b.start.After(last.end) {
			if b.end.After(last.end) {
				last.end = b.end
			}
		} else {
			merged = append(merged, b)
		}
	}

	now := time.Now()
	var slots []models.TimeSlot

	// Emit service-sized slots packed into a free window, skipping any that
	// start in the past.
	appendWindow := func(from, to time.Time) {
		for t := from; !t.Add(serviceDuration).After(to); t = t.Add(serviceDuration) {
			if t.After(now) {
				slots = append(slots, models.TimeSlot{StartTime: t, EndTime: t.Add(serviceDuration)})
			}
		}
	}

	cursor := dayStart
	for _, b := range merged {
		appendWindow(cursor, b.start)
		if b.end.After(cursor) {
			cursor = b.end
		}
	}
	appendWindow(cursor, dayEnd)

	return slots
}
