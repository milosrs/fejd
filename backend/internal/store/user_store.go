package store

import (
	"context"
	"fmt"

	"fejd-backend/internal/models"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
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
		Select("id", "COALESCE(display_name, '')", "COALESCE(email, '')", "avatar_id", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var u models.User
	if err := s.pool.QueryRow(ctx, sql, args...).Scan(&u.ID, &u.DisplayName, &u.Email, &u.AvatarID, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &u, nil
}

// GetAvatarIDForUpdate returns the user's current avatar reference under a row
// lock, for use inside an avatar-replacement transaction. A nil return means
// the user has no avatar yet (or no local users row).
func (s *UserStore) GetAvatarIDForUpdate(ctx context.Context, q Querier, id string) (*uuid.UUID, error) {
	sql, args, err := psql.
		Select("avatar_id").
		From("users").
		Where(sq.Eq{"id": id}).
		Suffix("FOR UPDATE").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var avatarID *uuid.UUID
	if err := q.QueryRow(ctx, sql, args...).Scan(&avatarID); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get avatar: %w", err)
	}
	return avatarID, nil
}

// SetAvatar updates a user's profile picture reference. Pass a nil avatarID to
// clear it.
func (s *UserStore) SetAvatar(ctx context.Context, q Querier, userID string, avatarID *uuid.UUID) error {
	sql, args, err := psql.
		Update("users").
		Set("avatar_id", nullableUUID(avatarID)).
		Where(sq.Eq{"id": userID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = q.Exec(ctx, sql, args...)
	return err
}

// AvatarImageExists reports whether any user references the image as their
// profile picture.
func (s *UserStore) AvatarImageExists(ctx context.Context, imageID uuid.UUID) (bool, error) {
	sql, args, err := psql.
		Select("1").
		From("users").
		Where(sq.Eq{"avatar_id": imageID}).
		Prefix("SELECT EXISTS (").
		Suffix(")").
		ToSql()
	if err != nil {
		return false, fmt.Errorf("failed to build query: %w", err)
	}

	var exists bool
	if err := s.pool.QueryRow(ctx, sql, args...).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check avatar image: %w", err)
	}
	return exists, nil
}
