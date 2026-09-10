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

// ReplaceByService atomically replaces the set of employees mapped to a
// service. Deletion is restricted by the appointments composite FK, so removing
// an employee that existing appointments reference will fail.
func (s *EmployeeServiceStore) ReplaceByService(ctx context.Context, serviceID uuid.UUID, businessUserIDs []uuid.UUID) error {
	return db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		delSQL, delArgs, err := psql.
			Delete("employee_services").
			Where(sq.Eq{"service_id": serviceID}).
			ToSql()
		if err != nil {
			return fmt.Errorf("failed to build delete query: %w", err)
		}
		if _, err := tx.Exec(ctx, delSQL, delArgs...); err != nil {
			return fmt.Errorf("failed to clear service employees: %w", err)
		}

		for _, buID := range businessUserIDs {
			insSQL, insArgs, err := psql.
				Insert("employee_services").
				Columns("business_user_id", "service_id").
				Values(buID, serviceID).
				Suffix("ON CONFLICT DO NOTHING").
				ToSql()
			if err != nil {
				return fmt.Errorf("failed to build insert query: %w", err)
			}
			if _, err := tx.Exec(ctx, insSQL, insArgs...); err != nil {
				return fmt.Errorf("failed to assign employee: %w", err)
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

// ListByService returns every employee-service link for a service.
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

// ListServiceProviderIDs returns the distinct business_user IDs that offer at
// least one service in the business.
func (s *EmployeeServiceStore) ListServiceProviderIDs(ctx context.Context, businessID uuid.UUID) ([]uuid.UUID, error) {
	sql, args, err := psql.
		Select("DISTINCT es.business_user_id").
		From("employee_services es").
		Join("services s ON s.id = es.service_id").
		Where(sq.Eq{"s.business_id": businessID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list service providers: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan service provider: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}
