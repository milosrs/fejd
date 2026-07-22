package store

import (
	"context"
	"fejd-backend/internal/models"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WorkingHoursStore struct {
	pool *pgxpool.Pool
}

func NewWorkingHoursStore(pool *pgxpool.Pool) *WorkingHoursStore {
	return &WorkingHoursStore{pool: pool}
}

func (s *WorkingHoursStore) GetByBusinessUser(ctx context.Context, businessUserID uuid.UUID) ([]models.WorkingHours, error) {
	sql, args, err := psql.
		Select("id", "business_user_id", "day_of_week", "start_time::text", "end_time::text").
		From("working_hours").
		Where(sq.Eq{"business_user_id": businessUserID}).
		OrderBy("day_of_week").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get working hours: %w", err)
	}
	defer rows.Close()

	var hours []models.WorkingHours
	for rows.Next() {
		var wh models.WorkingHours
		if err := rows.Scan(&wh.ID, &wh.BusinessUserID, &wh.DayOfWeek, &wh.StartTime, &wh.EndTime); err != nil {
			return nil, fmt.Errorf("failed to scan working hours: %w", err)
		}
		hours = append(hours, wh)
	}
	return hours, nil
}

func (s *WorkingHoursStore) Upsert(ctx context.Context, wh *models.WorkingHours) error {
	sql, args, err := psql.
		Insert("working_hours").
		Columns("business_user_id", "day_of_week", "start_time", "end_time").
		Values(wh.BusinessUserID, wh.DayOfWeek, wh.StartTime, wh.EndTime).
		Suffix("ON CONFLICT (business_user_id, day_of_week) DO UPDATE SET start_time = EXCLUDED.start_time, end_time = EXCLUDED.end_time").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = s.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to upsert working hours: %w", err)
	}
	return nil
}

func (s *WorkingHoursStore) DeleteByBusinessUser(ctx context.Context, businessUserID uuid.UUID) error {
	sql, args, err := psql.
		Delete("working_hours").
		Where(sq.Eq{"business_user_id": businessUserID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = s.pool.Exec(ctx, sql, args...)
	return err
}
