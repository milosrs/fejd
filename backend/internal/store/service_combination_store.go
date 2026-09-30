package store

import (
	"context"
	"fejd-backend/internal/db"
	"fejd-backend/internal/models"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ServiceCombinationStore struct {
	pool *pgxpool.Pool
}

func NewServiceCombinationStore(pool *pgxpool.Pool) *ServiceCombinationStore {
	return &ServiceCombinationStore{pool: pool}
}

// ListByService returns the directed combinability edges whose base service is
// serviceID (i.e. the services that may be added onto serviceID).
func (s *ServiceCombinationStore) ListByService(ctx context.Context, serviceID uuid.UUID) ([]models.ServiceCombination, error) {
	sql, args, err := psql.
		Select("service_id", "combinable_service_id").
		From("service_combinations").
		Where(sq.Eq{"service_id": serviceID}).
		OrderBy("combinable_service_id").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list service combinations: %w", err)
	}
	defer rows.Close()

	var result []models.ServiceCombination
	for rows.Next() {
		var c models.ServiceCombination
		if err := rows.Scan(&c.ServiceID, &c.CombinableServiceID); err != nil {
			return nil, fmt.Errorf("failed to scan service combination: %w", err)
		}
		result = append(result, c)
	}
	return result, nil
}

// ListByBusiness returns every combinability edge configured for a business,
// resolved through the services it belongs to.
func (s *ServiceCombinationStore) ListByBusiness(ctx context.Context, businessID uuid.UUID) ([]models.ServiceCombination, error) {
	sql, args, err := psql.
		Select("sc.service_id", "sc.combinable_service_id").
		From("service_combinations sc").
		Join("services s ON s.id = sc.service_id").
		Where(sq.Eq{"s.business_id": businessID}).
		OrderBy("sc.service_id", "sc.combinable_service_id").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list service combinations: %w", err)
	}
	defer rows.Close()

	var result []models.ServiceCombination
	for rows.Next() {
		var c models.ServiceCombination
		if err := rows.Scan(&c.ServiceID, &c.CombinableServiceID); err != nil {
			return nil, fmt.Errorf("failed to scan service combination: %w", err)
		}
		result = append(result, c)
	}
	return result, nil
}

// ReplaceByService atomically replaces the set of services combinable onto
// serviceID.
func (s *ServiceCombinationStore) ReplaceByService(ctx context.Context, serviceID uuid.UUID, combinableIDs []uuid.UUID) error {
	return db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		delSQL, delArgs, err := psql.
			Delete("service_combinations").
			Where(sq.Eq{"service_id": serviceID}).
			ToSql()
		if err != nil {
			return fmt.Errorf("failed to build delete query: %w", err)
		}
		if _, err := tx.Exec(ctx, delSQL, delArgs...); err != nil {
			return fmt.Errorf("failed to clear service combinations: %w", err)
		}

		for _, cid := range combinableIDs {
			insSQL, insArgs, err := psql.
				Insert("service_combinations").
				Columns("service_id", "combinable_service_id").
				Values(serviceID, cid).
				Suffix("ON CONFLICT DO NOTHING").
				ToSql()
			if err != nil {
				return fmt.Errorf("failed to build insert query: %w", err)
			}
			if _, err := tx.Exec(ctx, insSQL, insArgs...); err != nil {
				return fmt.Errorf("failed to insert service combination: %w", err)
			}
		}

		return nil
	})
}
