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

// RecordInviteRegistration marks a user as having registered via a QR/invite
// link and records the salon they were invited to (nil for platform invites).
// It upserts so it is safe whether or not the user row already exists.
func (s *UserStore) RecordInviteRegistration(ctx context.Context, q Querier, userID string, businessID *uuid.UUID) error {
	sql, args, err := psql.
		Insert("users").
		Columns("id", "display_name", "registration_source", "invited_business_id").
		Values(userID, "", "invite", nullableUUID(businessID)).
		Suffix(`ON CONFLICT (id) DO UPDATE SET
			registration_source = EXCLUDED.registration_source,
			invited_business_id = EXCLUDED.invited_business_id,
			updated_at = now()`).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = q.Exec(ctx, sql, args...)
	return err
}

// ListInvitedCustomers returns every customer invited to a salon (users whose
// invited_business_id is set), joined with the salon's name/slug/logo and the
// customer's display name/avatar, ordered by salon name then customer name.
func (s *UserStore) ListInvitedCustomers(ctx context.Context) ([]models.InvitedCustomer, error) {
	sql, args, err := psql.
		Select("b.id", "b.name", "b.slug", "il.image_id", "u.id", "COALESCE(u.display_name, '')", "u.avatar_id").
		From("users u").
		Join("businesses b ON b.id = u.invited_business_id").
		LeftJoin("image_links il ON il.entity_type = 'business' AND il.entity_id = b.id AND il.purpose = 'logo'").
		OrderBy("b.name", "u.display_name").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list invited customers: %w", err)
	}
	defer rows.Close()

	var customers []models.InvitedCustomer
	for rows.Next() {
		var c models.InvitedCustomer
		if err := rows.Scan(&c.BusinessID, &c.BusinessName, &c.BusinessSlug, &c.BusinessLogo, &c.UserID, &c.DisplayName, &c.AvatarID); err != nil {
			return nil, fmt.Errorf("failed to scan invited customer: %w", err)
		}
		customers = append(customers, c)
	}
	if customers == nil {
		customers = []models.InvitedCustomer{}
	}
	return customers, nil
}

// ListSelfRegisteredUsers returns every user who registered directly (their
// registration_source is still the "self" default), ordered by display name.
func (s *UserStore) ListSelfRegisteredUsers(ctx context.Context) ([]models.SelfRegisteredUser, error) {
	sql, args, err := psql.
		Select("id", "COALESCE(display_name, '')", "avatar_id").
		From("users").
		Where(sq.Eq{"registration_source": "self"}).
		OrderBy("display_name").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list self-registered users: %w", err)
	}
	defer rows.Close()

	var users []models.SelfRegisteredUser
	for rows.Next() {
		var u models.SelfRegisteredUser
		if err := rows.Scan(&u.UserID, &u.DisplayName, &u.AvatarID); err != nil {
			return nil, fmt.Errorf("failed to scan self-registered user: %w", err)
		}
		users = append(users, u)
	}
	if users == nil {
		users = []models.SelfRegisteredUser{}
	}
	return users, nil
}

// GetByID returns a user by Keycloak subject, or pgx.ErrNoRows when absent.
func (s *UserStore) GetByID(ctx context.Context, id string) (*models.User, error) {
	sql, args, err := psql.
		Select("id", "COALESCE(display_name, '')", "COALESCE(email, '')", "avatar_id", "COALESCE(registration_source, '')", "invited_business_id", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var u models.User
	if err := s.pool.QueryRow(ctx, sql, args...).Scan(&u.ID, &u.DisplayName, &u.Email, &u.AvatarID, &u.RegistrationSource, &u.InvitedBusinessID, &u.CreatedAt, &u.UpdatedAt); err != nil {
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
