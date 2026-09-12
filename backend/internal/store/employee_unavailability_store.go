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

type EmployeeUnavailabilityStore struct {
	pool *pgxpool.Pool
}

func NewEmployeeUnavailabilityStore(pool *pgxpool.Pool) *EmployeeUnavailabilityStore {
	return &EmployeeUnavailabilityStore{pool: pool}
}

func (s *EmployeeUnavailabilityStore) Create(ctx context.Context, q Querier, u *models.EmployeeUnavailability) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	if u.Status == "" {
		u.Status = models.UnavailabilityStatusConfirmed
	}
	sql, args, err := psql.
		Insert("employee_unavailability").
		Columns("id", "business_user_id", "start_time", "end_time", "reason", "status").
		Values(u.ID, u.BusinessUserID, u.StartTime, u.EndTime, nullableString(u.Reason), u.Status).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = q.Exec(ctx, sql, args...)
	return err
}

func (s *EmployeeUnavailabilityStore) Delete(ctx context.Context, q Querier, id uuid.UUID) error {
	sql, args, err := psql.
		Delete("employee_unavailability").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = q.Exec(ctx, sql, args...)
	return err
}

// DeleteForBusinessUser removes an unavailability row only if it belongs to the
// given employee. Used by the self-service path so a member can only delete
// their own blocks.
func (s *EmployeeUnavailabilityStore) DeleteForBusinessUser(ctx context.Context, q Querier, businessUserID, id uuid.UUID) error {
	sql, args, err := psql.
		Delete("employee_unavailability").
		Where(sq.Eq{"id": id, "business_user_id": businessUserID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = q.Exec(ctx, sql, args...)
	return err
}

// ListByBusinessUser returns all unavailability rows for an employee ordered by
// start time.
func (s *EmployeeUnavailabilityStore) ListByBusinessUser(ctx context.Context, businessUserID uuid.UUID) ([]models.EmployeeUnavailability, error) {
	sql, args, err := psql.
		Select("id", "business_user_id", "start_time", "end_time", "COALESCE(reason, '')", "status", "COALESCE(rejection_reason, '')").
		From("employee_unavailability").
		Where(sq.Eq{"business_user_id": businessUserID}).
		OrderBy("start_time").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list unavailability: %w", err)
	}
	defer rows.Close()

	var result []models.EmployeeUnavailability
	for rows.Next() {
		var u models.EmployeeUnavailability
		if err := rows.Scan(&u.ID, &u.BusinessUserID, &u.StartTime, &u.EndTime, 		&u.Reason, &u.Status, &u.RejectionReason); err != nil {
			return nil, fmt.Errorf("failed to scan unavailability: %w", err)
		}
		result = append(result, u)
	}
	return result, nil
}

// ListOverlapping returns acknowledged (confirmed) unavailability rows
// overlapping [from, to) for the given employee. Pending reservations do not
// block the calendar until the owner accepts them.
func (s *EmployeeUnavailabilityStore) ListOverlapping(ctx context.Context, businessUserID uuid.UUID, from, to time.Time) ([]models.EmployeeUnavailability, error) {
	sql, args, err := psql.
		Select("id", "business_user_id", "start_time", "end_time", "COALESCE(reason, '')", "status", "COALESCE(rejection_reason, '')").
		From("employee_unavailability").
		Where(sq.Eq{"business_user_id": businessUserID}).
		Where(sq.Eq{"status": string(models.UnavailabilityStatusConfirmed)}).
		Where(sq.Lt{"start_time": to}).
		Where(sq.Gt{"end_time": from}).
		OrderBy("start_time").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list unavailability: %w", err)
	}
	defer rows.Close()

	var result []models.EmployeeUnavailability
	for rows.Next() {
		var u models.EmployeeUnavailability
		if err := rows.Scan(&u.ID, &u.BusinessUserID, &u.StartTime, &u.EndTime, 		&u.Reason, &u.Status, &u.RejectionReason); err != nil {
			return nil, fmt.Errorf("failed to scan unavailability: %w", err)
		}
		result = append(result, u)
	}
	return result, nil
}

// Acknowledge transitions an unavailability row from pending to confirmed.
func (s *EmployeeUnavailabilityStore) Acknowledge(ctx context.Context, q Querier, id uuid.UUID) error {
	sql, args, err := psql.
		Update("employee_unavailability").
		Set("status", string(models.UnavailabilityStatusConfirmed)).
		Where(sq.Eq{"id": id, "status": string(models.UnavailabilityStatusPending)}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = q.Exec(ctx, sql, args...)
	return err
}

// Reject transitions an unavailability row from pending to rejected with a
// stated reason.
func (s *EmployeeUnavailabilityStore) Reject(ctx context.Context, q Querier, id uuid.UUID, reason string) error {
	sql, args, err := psql.
		Update("employee_unavailability").
		Set("status", string(models.UnavailabilityStatusRejected)).
		Set("rejection_reason", reason).
		Where(sq.Eq{"id": id, "status": string(models.UnavailabilityStatusPending)}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = q.Exec(ctx, sql, args...)
	return err
}

// GetByID returns a single unavailability row.
func (s *EmployeeUnavailabilityStore) GetByID(ctx context.Context, id uuid.UUID) (*models.EmployeeUnavailability, error) {
	sql, args, err := psql.
		Select("id", "business_user_id", "start_time", "end_time", "COALESCE(reason, '')", "status", "COALESCE(rejection_reason, '')").
		From("employee_unavailability").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var u models.EmployeeUnavailability
	if err := s.pool.QueryRow(ctx, sql, args...).Scan(&u.ID, &u.BusinessUserID, &u.StartTime, &u.EndTime, 		&u.Reason, &u.Status, &u.RejectionReason); err != nil {
		return nil, fmt.Errorf("unavailability not found: %w", err)
	}
	return &u, nil
}

// ListByBusiness returns all unavailability rows for a business, ordered by
// start time.
func (s *EmployeeUnavailabilityStore) ListByBusiness(ctx context.Context, businessID uuid.UUID) ([]models.EmployeeUnavailability, error) {
	sql, args, err := psql.
		Select("u.id", "u.business_user_id", "u.start_time", "u.end_time", "COALESCE(u.reason, '')", "u.status", "COALESCE(u.rejection_reason, '')").
		From("employee_unavailability u").
		Join("business_users bu ON bu.id = u.business_user_id").
		Where(sq.Eq{"bu.business_id": businessID}).
		OrderBy("u.start_time").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list unavailability: %w", err)
	}
	defer rows.Close()

	var result []models.EmployeeUnavailability
	for rows.Next() {
		var u models.EmployeeUnavailability
		if err := rows.Scan(&u.ID, &u.BusinessUserID, &u.StartTime, &u.EndTime, 		&u.Reason, &u.Status, &u.RejectionReason); err != nil {
			return nil, fmt.Errorf("failed to scan unavailability: %w", err)
		}
		result = append(result, u)
	}
	return result, nil
}
