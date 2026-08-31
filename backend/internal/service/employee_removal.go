package service

import (
	"context"
	"fejd-backend/internal/models"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// RemoveEmployee soft-deletes an employee and relocates their future
// reservations. Past reservations are left untouched (the row is retained with
// active=false). Future pending/confirmed reservations are reassigned to the
// first available active employee offering the service; if none can take a
// reservation it is cancelled with a reason and the customer is notified via SSE.
//
// Note: "fairly split" is currently implemented as first-available; a balanced
// strategy can be layered on later in the service/API plan without a migration.
func (s *SlotService) RemoveEmployee(ctx context.Context, businessID, businessUserID uuid.UUID) error {
	bu, err := s.businessUser.GetByID(ctx, businessUserID)
	if err != nil {
		return fmt.Errorf("employee not found: %w", err)
	}
	if bu.BusinessID != businessID {
		return fmt.Errorf("employee does not belong to business")
	}

	if err := s.businessUser.SetActive(ctx, businessUserID, false); err != nil {
		return fmt.Errorf("failed to deactivate employee: %w", err)
	}

	now := time.Now()
	future, err := s.appointments.ListByBusinessUser(ctx, businessUserID, now, now.AddDate(1, 0, 0))
	if err != nil {
		return fmt.Errorf("failed to list future appointments: %w", err)
	}

	employees, err := s.businessUser.ListEmployeesByBusiness(ctx, businessID)
	if err != nil {
		return fmt.Errorf("failed to list employees: %w", err)
	}

	for _, appt := range future {
		if appt.Status != models.AppointmentStatusPending && appt.Status != models.AppointmentStatusConfirmed {
			continue
		}

		reassigned := false
		for _, emp := range employees {
			if emp.ID == businessUserID {
				continue
			}

			offers, err := s.employeeServices.OffersService(ctx, emp.ID, appt.ServiceID)
			if err != nil || !offers {
				continue
			}

			conflicts, err := s.appointments.GetConflictingAppointments(ctx, appt.BusinessID, emp.ID, appt.StartTime, appt.EndTime)
			if err != nil || len(conflicts) > 0 {
				continue
			}

			unavail, err := s.unavailability.ListOverlapping(ctx, emp.ID, appt.StartTime, appt.EndTime)
			if err != nil || len(unavail) > 0 {
				continue
			}

			if err := s.appointments.Reassign(ctx, appt.ID, emp.ID); err != nil {
				continue
			}

			s.hub.Publish(businessID.String(), map[string]any{
				"type":             "appointment_reassigned",
				"appointment_id":   appt.ID.String(),
				"business_user_id": emp.ID.String(),
				"start_time":       appt.StartTime.Format(time.RFC3339),
			})
			reassigned = true
			break
		}

		if !reassigned {
			if err := s.appointments.CancelByID(ctx, appt.ID, "employee unavailable"); err != nil {
				return fmt.Errorf("failed to cancel appointment: %w", err)
			}
			s.hub.Publish(businessID.String(), map[string]any{
				"type":             "appointment_cancelled",
				"appointment_id":   appt.ID.String(),
				"customer_user_id": appt.CustomerUserID,
				"start_time":       appt.StartTime.Format(time.RFC3339),
			})
		}
	}

	return nil
}
