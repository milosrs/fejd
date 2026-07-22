package store

import (
	"context"
	"fejd-backend/internal/models"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AppointmentStore struct {
	pool *pgxpool.Pool
}

func NewAppointmentStore(pool *pgxpool.Pool) *AppointmentStore {
	return &AppointmentStore{pool: pool}
}

func (s *AppointmentStore) GetConflictingAppointments(ctx context.Context, businessID, businessUserID uuid.UUID, from, to time.Time) ([]models.Appointment, error) {
	sql, args, err := psql.
		Select("id", "business_id", "service_id", "business_user_id", "customer_user_id",
			"start_time", "end_time", "status", "created_at").
		From("appointments").
		Where(sq.Eq{"business_id": businessID, "business_user_id": businessUserID}).
		Where(sq.NotEq{"status": []string{string(models.AppointmentStatusCancelled), string(models.AppointmentStatusNoShow)}}).
		Where(sq.Lt{"start_time": to}).
		Where(sq.Gt{"end_time": from}).
		OrderBy("start_time").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get appointments: %w", err)
	}
	defer rows.Close()

	var appointments []models.Appointment
	for rows.Next() {
		var a models.Appointment
		if err := rows.Scan(&a.ID, &a.BusinessID, &a.ServiceID, &a.BusinessUserID,
			&a.CustomerUserID, &a.StartTime, &a.EndTime, &a.Status, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan appointment: %w", err)
		}
		appointments = append(appointments, a)
	}
	return appointments, nil
}

func (s *AppointmentStore) Create(ctx context.Context, a *models.Appointment) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	sql, args, err := psql.
		Insert("appointments").
		Columns("id", "business_id", "service_id", "business_user_id", "customer_user_id", "start_time", "end_time", "status").
		Values(a.ID, a.BusinessID, a.ServiceID, a.BusinessUserID, a.CustomerUserID,
			a.StartTime, a.EndTime, a.Status).
		Suffix("RETURNING created_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	return s.pool.QueryRow(ctx, sql, args...).Scan(&a.CreatedAt)
}

func (s *AppointmentStore) ListByCustomer(ctx context.Context, customerUserID string) ([]models.Appointment, error) {
	sql, args, err := psql.
		Select("id", "business_id", "service_id", "business_user_id", "customer_user_id",
			"start_time", "end_time", "status", "created_at").
		From("appointments").
		Where(sq.Eq{"customer_user_id": customerUserID}).
		OrderBy("start_time DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list appointments: %w", err)
	}
	defer rows.Close()

	var appointments []models.Appointment
	for rows.Next() {
		var a models.Appointment
		if err := rows.Scan(&a.ID, &a.BusinessID, &a.ServiceID, &a.BusinessUserID,
			&a.CustomerUserID, &a.StartTime, &a.EndTime, &a.Status, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan appointment: %w", err)
		}
		appointments = append(appointments, a)
	}
	return appointments, nil
}

func (s *AppointmentStore) ListByBusinessUser(ctx context.Context, businessUserID uuid.UUID, from, to time.Time) ([]models.Appointment, error) {
	sql, args, err := psql.
		Select("id", "business_id", "service_id", "business_user_id", "customer_user_id",
			"start_time", "end_time", "status", "created_at").
		From("appointments").
		Where(sq.Eq{"business_user_id": businessUserID}).
		Where(sq.GtOrEq{"start_time": from}).
		Where(sq.Lt{"start_time": to}).
		OrderBy("start_time").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list appointments: %w", err)
	}
	defer rows.Close()

	var appointments []models.Appointment
	for rows.Next() {
		var a models.Appointment
		if err := rows.Scan(&a.ID, &a.BusinessID, &a.ServiceID, &a.BusinessUserID,
			&a.CustomerUserID, &a.StartTime, &a.EndTime, &a.Status, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan appointment: %w", err)
		}
		appointments = append(appointments, a)
	}
	return appointments, nil
}

func (s *AppointmentStore) Cancel(ctx context.Context, id uuid.UUID, customerUserID string) error {
	sql, args, err := psql.
		Update("appointments").
		Set("status", string(models.AppointmentStatusCancelled)).
		Where(sq.Eq{"id": id, "customer_user_id": customerUserID, "status": string(models.AppointmentStatusConfirmed)}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = s.pool.Exec(ctx, sql, args...)
	return err
}
