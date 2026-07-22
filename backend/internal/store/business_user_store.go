package store

import (
	"context"
	"fejd-backend/internal/models"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BusinessUserStore struct {
	pool *pgxpool.Pool
}

func NewBusinessUserStore(pool *pgxpool.Pool) *BusinessUserStore {
	return &BusinessUserStore{pool: pool}
}

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func (s *BusinessUserStore) GetByBusinessAndUser(ctx context.Context, businessID uuid.UUID, userID string) (*models.BusinessUser, error) {
	sql, args, err := psql.
		Select("id", "business_id", "user_id", "role", "display_name").
		From("business_users").
		Where(sq.Eq{"business_id": businessID, "user_id": userID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var bu models.BusinessUser
	err = s.pool.QueryRow(ctx, sql, args...).Scan(&bu.ID, &bu.BusinessID, &bu.UserID, &bu.Role, &bu.DisplayName)
	if err != nil {
		return nil, fmt.Errorf("business user not found: %w", err)
	}
	return &bu, nil
}

func (s *BusinessUserStore) GetByID(ctx context.Context, id uuid.UUID) (*models.BusinessUser, error) {
	sql, args, err := psql.
		Select("id", "business_id", "user_id", "role", "display_name").
		From("business_users").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var bu models.BusinessUser
	err = s.pool.QueryRow(ctx, sql, args...).Scan(&bu.ID, &bu.BusinessID, &bu.UserID, &bu.Role, &bu.DisplayName)
	if err != nil {
		return nil, fmt.Errorf("business user not found: %w", err)
	}
	return &bu, nil
}

func (s *BusinessUserStore) ListByBusiness(ctx context.Context, businessID uuid.UUID) ([]models.BusinessUser, error) {
	sql, args, err := psql.
		Select("id", "business_id", "user_id", "role", "display_name").
		From("business_users").
		Where(sq.Eq{"business_id": businessID}).
		OrderBy("display_name").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list business users: %w", err)
	}
	defer rows.Close()

	var users []models.BusinessUser
	for rows.Next() {
		var bu models.BusinessUser
		if err := rows.Scan(&bu.ID, &bu.BusinessID, &bu.UserID, &bu.Role, &bu.DisplayName); err != nil {
			return nil, fmt.Errorf("failed to scan business user: %w", err)
		}
		users = append(users, bu)
	}
	return users, nil
}

func (s *BusinessUserStore) ListEmployeesByBusiness(ctx context.Context, businessID uuid.UUID) ([]models.BusinessUser, error) {
	sql, args, err := psql.
		Select("id", "business_id", "user_id", "role", "display_name").
		From("business_users").
		Where(sq.Eq{"business_id": businessID, "role": "employee"}).
		OrderBy("display_name").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list employees: %w", err)
	}
	defer rows.Close()

	var users []models.BusinessUser
	for rows.Next() {
		var bu models.BusinessUser
		if err := rows.Scan(&bu.ID, &bu.BusinessID, &bu.UserID, &bu.Role, &bu.DisplayName); err != nil {
			return nil, fmt.Errorf("failed to scan employee: %w", err)
		}
		users = append(users, bu)
	}
	return users, nil
}

func (s *BusinessUserStore) IsAdmin(ctx context.Context, businessID uuid.UUID, userID string) (bool, error) {
	sql, args, err := psql.
		Select("role").
		From("business_users").
		Where(sq.Eq{"business_id": businessID, "user_id": userID, "role": "admin"}).
		ToSql()
	if err != nil {
		return false, fmt.Errorf("failed to build query: %w", err)
	}

	var role string
	err = s.pool.QueryRow(ctx, sql, args...).Scan(&role)
	if err != nil {
		return false, nil
	}
	return true, nil
}
