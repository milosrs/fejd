package store

import (
	"context"
	"errors"
	"fejd-backend/internal/db"
	"fejd-backend/internal/models"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EmployeeServiceStore struct {
	pool *pgxpool.Pool
}

func NewEmployeeServiceStore(pool *pgxpool.Pool) *EmployeeServiceStore {
	return &EmployeeServiceStore{pool: pool}
}

func (s *EmployeeServiceStore) Assign(ctx context.Context, businessUserID, serviceID uuid.UUID) error {
	sql, args, err := psql.
		Insert("employee_services").
		Columns("business_user_id", "service_id").
		Values(businessUserID, serviceID).
		Suffix("ON CONFLICT DO NOTHING").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = s.pool.Exec(ctx, sql, args...)
	return err
}

func (s *EmployeeServiceStore) Unassign(ctx context.Context, businessUserID, serviceID uuid.UUID) error {
	sql, args, err := psql.
		Delete("employee_services").
		Where(sq.Eq{"business_user_id": businessUserID, "service_id": serviceID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = s.pool.Exec(ctx, sql, args...)
	return err
}

// ReplaceByBusinessUser atomically replaces the set of services an employee
// offers. Deletion is restricted by the appointments composite FK, so removing
// a service that existing appointments reference will fail.
func (s *EmployeeServiceStore) ReplaceByBusinessUser(ctx context.Context, businessUserID uuid.UUID, serviceIDs []uuid.UUID) error {
	return db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		delSQL, delArgs, err := psql.
			Delete("employee_services").
			Where(sq.Eq{"business_user_id": businessUserID}).
			ToSql()
		if err != nil {
			return fmt.Errorf("failed to build delete query: %w", err)
		}
		if _, err := tx.Exec(ctx, delSQL, delArgs...); err != nil {
			return fmt.Errorf("failed to clear employee services: %w", err)
		}

		for _, sid := range serviceIDs {
			insSQL, insArgs, err := psql.
				Insert("employee_services").
				Columns("business_user_id", "service_id").
				Values(businessUserID, sid).
				Suffix("ON CONFLICT DO NOTHING").
				ToSql()
			if err != nil {
				return fmt.Errorf("failed to build insert query: %w", err)
			}
			if _, err := tx.Exec(ctx, insSQL, insArgs...); err != nil {
				return fmt.Errorf("failed to assign service: %w", err)
			}
		}

		return nil
	})
}

func (s *EmployeeServiceStore) OffersService(ctx context.Context, businessUserID, serviceID uuid.UUID) (bool, error) {
	sql, args, err := psql.
		Select("1").
		From("employee_services").
		Where(sq.Eq{"business_user_id": businessUserID, "service_id": serviceID}).
		Limit(1).
		ToSql()
	if err != nil {
		return false, fmt.Errorf("failed to build query: %w", err)
	}

	var one int
	err = s.pool.QueryRow(ctx, sql, args...).Scan(&one)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *EmployeeServiceStore) ListByBusinessUser(ctx context.Context, businessUserID uuid.UUID) ([]models.EmployeeService, error) {
	sql, args, err := psql.
		Select("business_user_id", "service_id").
		From("employee_services").
		Where(sq.Eq{"business_user_id": businessUserID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list employee services: %w", err)
	}
	defer rows.Close()

	var result []models.EmployeeService
	for rows.Next() {
		var es models.EmployeeService
		if err := rows.Scan(&es.BusinessUserID, &es.ServiceID); err != nil {
			return nil, fmt.Errorf("failed to scan employee service: %w", err)
		}
		result = append(result, es)
	}
	return result, nil
}

func (s *EmployeeServiceStore) ListByService(ctx context.Context, serviceID uuid.UUID) ([]models.EmployeeService, error) {
	sql, args, err := psql.
		Select("business_user_id", "service_id").
		From("employee_services").
		Where(sq.Eq{"service_id": serviceID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list employees for service: %w", err)
	}
	defer rows.Close()

	var result []models.EmployeeService
	for rows.Next() {
		var es models.EmployeeService
		if err := rows.Scan(&es.BusinessUserID, &es.ServiceID); err != nil {
			return nil, fmt.Errorf("failed to scan employee service: %w", err)
		}
		result = append(result, es)
	}
	return result, nil
}
