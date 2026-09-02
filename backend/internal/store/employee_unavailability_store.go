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
	sql, args, err := psql.
		Insert("employee_unavailability").
		Columns("id", "business_user_id", "start_time", "end_time", "reason").
		Values(u.ID, u.BusinessUserID, u.StartTime, u.EndTime, nullableString(u.Reason)).
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

// ListOverlapping returns unavailability rows overlapping [from, to) for the
// given employee. Used by the slot engine and the booking soft-check.
func (s *EmployeeUnavailabilityStore) ListOverlapping(ctx context.Context, businessUserID uuid.UUID, from, to time.Time) ([]models.EmployeeUnavailability, error) {
	sql, args, err := psql.
		Select("id", "business_user_id", "start_time", "end_time", "COALESCE(reason, '')").
		From("employee_unavailability").
		Where(sq.Eq{"business_user_id": businessUserID}).
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
		if err := rows.Scan(&u.ID, &u.BusinessUserID, &u.StartTime, &u.EndTime, &u.Reason); err != nil {
			return nil, fmt.Errorf("failed to scan unavailability: %w", err)
		}
		result = append(result, u)
	}
	return result, nil
}
