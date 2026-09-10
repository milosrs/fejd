package store

import (
	"context"
	"fejd-backend/internal/db"
	"fejd-backend/internal/models"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BusinessHoursStore struct {
	pool *pgxpool.Pool
}

func NewBusinessHoursStore(pool *pgxpool.Pool) *BusinessHoursStore {
	return &BusinessHoursStore{pool: pool}
}

func (s *BusinessHoursStore) ListByBusiness(ctx context.Context, businessID uuid.UUID) ([]models.BusinessHours, error) {
	sql, args, err := psql.
		Select("id", "business_id", "day_of_week", "start_time::text", "end_time::text").
		From("business_hours").
		Where(sq.Eq{"business_id": businessID}).
		OrderBy("day_of_week").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list business hours: %w", err)
	}
	defer rows.Close()

	var hours []models.BusinessHours
	for rows.Next() {
		var bh models.BusinessHours
		var startStr, endStr string
		if err := rows.Scan(&bh.ID, &bh.BusinessID, &bh.DayOfWeek, &startStr, &endStr); err != nil {
			return nil, fmt.Errorf("failed to scan business hours: %w", err)
		}
		if bh.StartTime, err = time.Parse(time.TimeOnly, startStr); err != nil {
			return nil, fmt.Errorf("failed to parse start time: %w", err)
		}
		if bh.EndTime, err = time.Parse(time.TimeOnly, endStr); err != nil {
			return nil, fmt.Errorf("failed to parse end time: %w", err)
		}
		hours = append(hours, bh)
	}
	return hours, nil
}

// ReplaceByBusiness atomically replaces the salon's default working hours.
func (s *BusinessHoursStore) ReplaceByBusiness(ctx context.Context, businessID uuid.UUID, hours []models.BusinessHours) error {
	return db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		delSQL, delArgs, err := psql.
			Delete("business_hours").
			Where(sq.Eq{"business_id": businessID}).
			ToSql()
		if err != nil {
			return fmt.Errorf("failed to build delete query: %w", err)
		}
		if _, err := tx.Exec(ctx, delSQL, delArgs...); err != nil {
			return fmt.Errorf("failed to clear business hours: %w", err)
		}

		for _, h := range hours {
			insSQL, insArgs, err := psql.
				Insert("business_hours").
				Columns("business_id", "day_of_week", "start_time", "end_time").
				Values(businessID, h.DayOfWeek, h.StartTime.Format(time.TimeOnly), h.EndTime.Format(time.TimeOnly)).
				ToSql()
			if err != nil {
				return fmt.Errorf("failed to build insert query: %w", err)
			}
			if _, err := tx.Exec(ctx, insSQL, insArgs...); err != nil {
				return fmt.Errorf("failed to insert business hours: %w", err)
			}
		}

		return nil
	})
}

// SeedDefaults inserts a full week of default working hours (all 7 days,
// 09:00–17:00) for a business, so a newly created salon is bookable
// immediately.
func (s *BusinessHoursStore) SeedDefaults(ctx context.Context, q Querier, businessID uuid.UUID) error {
	start := time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)
	end := time.Date(0, 1, 1, 17, 0, 0, 0, time.UTC)
	for day := 0; day < 7; day++ {
		sql, args, err := psql.
			Insert("business_hours").
			Columns("business_id", "day_of_week", "start_time", "end_time").
			Values(businessID, day, start.Format(time.TimeOnly), end.Format(time.TimeOnly)).
			ToSql()
		if err != nil {
			return fmt.Errorf("failed to build query: %w", err)
		}
		if _, err := q.Exec(ctx, sql, args...); err != nil {
			return fmt.Errorf("failed to seed business hours: %w", err)
		}
	}
	return nil
}
