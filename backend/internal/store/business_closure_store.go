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

// ListByBusiness returns every closure for a business ordered by date.
func (s *BusinessClosureStore) ListByBusiness(ctx context.Context, businessID uuid.UUID) ([]models.BusinessClosure, error) {
	sql, args, err := psql.
		Select("id", "business_id", "closure_date::text", "COALESCE(reason, '')").
		From("business_closures").
		Where(sq.Eq{"business_id": businessID}).
		OrderBy("closure_date").
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

// ReplaceByBusiness atomically replaces the salon's non-working days.
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
				Columns("business_id", "closure_date", "reason").
				Values(businessID, c.ClosureDate.Format(time.DateOnly), nullableString(c.Reason)).
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

// IsClosed reports whether the business is closed on the given date.
func (s *BusinessClosureStore) IsClosed(ctx context.Context, businessID uuid.UUID, date time.Time) (bool, error) {
	sql, args, err := psql.
		Select("1").
		From("business_closures").
		Where(sq.Eq{"business_id": businessID, "closure_date": date.Format(time.DateOnly)}).
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
	var dateStr string
	if err := row.Scan(&c.ID, &c.BusinessID, &dateStr, &c.Reason); err != nil {
		return nil, fmt.Errorf("closure not found: %w", err)
	}

	var err error
	if c.ClosureDate, err = time.Parse(time.DateOnly, dateStr); err != nil {
		return nil, fmt.Errorf("failed to parse closure date: %w", err)
	}
	return &c, nil
}
