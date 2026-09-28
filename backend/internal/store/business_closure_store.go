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

type BusinessClosureStore struct {
	pool *pgxpool.Pool
}

func NewBusinessClosureStore(pool *pgxpool.Pool) *BusinessClosureStore {
	return &BusinessClosureStore{pool: pool}
}

// ListByBusiness returns every non-working-day rule for a business.
func (s *BusinessClosureStore) ListByBusiness(ctx context.Context, businessID uuid.UUID) ([]models.BusinessClosure, error) {
	sql, args, err := psql.
		Select("id", "business_id", "closure_type", "start_date::text", "end_date::text", "day_of_week", "month", "day", "COALESCE(reason, '')").
		From("business_closures").
		Where(sq.Eq{"business_id": businessID}).
		OrderBy("closure_type", "start_date", "day_of_week", "month", "day").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list closures: %w", err)
	}
	defer rows.Close()

	var closures []models.BusinessClosure
	for rows.Next() {
		c, err := scanClosure(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan closure: %w", err)
		}
		closures = append(closures, *c)
	}
	return closures, nil
}

// ReplaceByBusiness atomically replaces the salon's non-working-day rules.
func (s *BusinessClosureStore) ReplaceByBusiness(ctx context.Context, businessID uuid.UUID, closures []models.BusinessClosure) error {
	return db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		delSQL, delArgs, err := psql.
			Delete("business_closures").
			Where(sq.Eq{"business_id": businessID}).
			ToSql()
		if err != nil {
			return fmt.Errorf("failed to build delete query: %w", err)
		}
		if _, err := tx.Exec(ctx, delSQL, delArgs...); err != nil {
			return fmt.Errorf("failed to clear closures: %w", err)
		}

		for _, c := range closures {
			insSQL, insArgs, err := psql.
				Insert("business_closures").
				Columns("business_id", "closure_type", "start_date", "end_date", "day_of_week", "month", "day", "reason").
				Values(businessID, c.ClosureType, nullableDate(c.StartDate), nullableDate(c.EndDate), c.DayOfWeek, c.Month, c.Day, nullableString(c.Reason)).
				ToSql()
			if err != nil {
				return fmt.Errorf("failed to build insert query: %w", err)
			}
			if _, err := tx.Exec(ctx, insSQL, insArgs...); err != nil {
				return fmt.Errorf("failed to insert closure: %w", err)
			}
		}

		return nil
	})
}

// IsClosed reports whether the business is closed on the given date according
// to any of its non-working-day rules.
func (s *BusinessClosureStore) IsClosed(ctx context.Context, businessID uuid.UUID, date time.Time) (bool, error) {
	dateStr := date.Format(time.DateOnly)
	weekday := int(date.Weekday())
	month := int(date.Month())
	day := date.Day()

	sql, args, err := psql.
		Select("1").
		From("business_closures").
		Where(sq.Eq{"business_id": businessID}).
		Where(sq.Or{
			sq.And{sq.Eq{"closure_type": "single"}, sq.Eq{"start_date": dateStr}},
			sq.And{sq.Eq{"closure_type": "range"}, sq.LtOrEq{"start_date": dateStr}, sq.GtOrEq{"end_date": dateStr}},
			sq.And{sq.Eq{"closure_type": "weekly"}, sq.Eq{"day_of_week": weekday}},
			sq.And{sq.Eq{"closure_type": "yearly"}, sq.Eq{"month": month}, sq.Eq{"day": day}},
		}).
		Prefix("SELECT EXISTS (").
		Suffix(")").
		ToSql()
	if err != nil {
		return false, fmt.Errorf("failed to build query: %w", err)
	}

	var exists bool
	if err := s.pool.QueryRow(ctx, sql, args...).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check closure: %w", err)
	}
	return exists, nil
}

func scanClosure(row rowScanner) (*models.BusinessClosure, error) {
	var c models.BusinessClosure
	var startStr, endStr *string
	if err := row.Scan(&c.ID, &c.BusinessID, &c.ClosureType, &startStr, &endStr, &c.DayOfWeek, &c.Month, &c.Day, &c.Reason); err != nil {
		return nil, fmt.Errorf("closure not found: %w", err)
	}

	if startStr != nil {
		t, err := time.Parse(time.DateOnly, *startStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse closure start date: %w", err)
		}
		c.StartDate = &t
	}
	if endStr != nil {
		t, err := time.Parse(time.DateOnly, *endStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse closure end date: %w", err)
		}
		c.EndDate = &t
	}
	return &c, nil
}

func nullableDate(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.Format(time.DateOnly)
}
