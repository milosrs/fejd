package store

import (
	"context"
	"fmt"

	"fejd-backend/internal/models"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserStore struct {
	pool *pgxpool.Pool
}

func NewUserStore(pool *pgxpool.Pool) *UserStore {
	return &UserStore{pool: pool}
}

// Upsert inserts or refreshes a user's locally-cached identity attributes. The
// update only fires when a value actually changed so per-request syncs stay
// cheap.
func (s *UserStore) Upsert(ctx context.Context, u *models.User) error {
	sql, args, err := psql.
		Insert("users").
		Columns("id", "display_name", "email").
		Values(u.ID, u.DisplayName, nullableString(u.Email)).
		Suffix(`ON CONFLICT (id) DO UPDATE SET
			display_name = EXCLUDED.display_name,
			email = EXCLUDED.email,
			updated_at = now()
			WHERE users.display_name IS DISTINCT FROM EXCLUDED.display_name
			   OR users.email IS DISTINCT FROM EXCLUDED.email`).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = s.pool.Exec(ctx, sql, args...)
	return err
}

// GetByID returns a user by Keycloak subject, or pgx.ErrNoRows when absent.
func (s *UserStore) GetByID(ctx context.Context, id string) (*models.User, error) {
	sql, args, err := psql.
		Select("id", "COALESCE(display_name, '')", "COALESCE(email, '')", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var u models.User
	if err := s.pool.QueryRow(ctx, sql, args...).Scan(&u.ID, &u.DisplayName, &u.Email, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &u, nil
}
