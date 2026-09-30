package store

import (
	"context"
	"fmt"

	"fejd-backend/internal/models"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PushTokenStore struct {
	pool *pgxpool.Pool
}

func NewPushTokenStore(pool *pgxpool.Pool) *PushTokenStore {
	return &PushTokenStore{pool: pool}
}

// Upsert registers a device token for a user. Re-registering an existing token
// (e.g. the user logs in again on the same device) refreshes the user link and
// updated_at instead of duplicating the row.
func (s *PushTokenStore) Upsert(ctx context.Context, t *models.PushToken) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	sql, args, err := psql.
		Insert("push_tokens").
		Columns("id", "user_id", "token", "platform").
		Values(t.ID, t.UserID, t.Token, t.Platform).
		Suffix(`ON CONFLICT (token) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			platform = EXCLUDED.platform,
			updated_at = now()`).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = s.pool.Exec(ctx, sql, args...)
	return err
}

// Delete removes a specific device token belonging to a user. Used when a user
// signs out or revokes notification permission on a device.
func (s *PushTokenStore) Delete(ctx context.Context, userID, token string) error {
	sql, args, err := psql.
		Delete("push_tokens").
		Where(sq.Eq{"user_id": userID, "token": token}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = s.pool.Exec(ctx, sql, args...)
	return err
}

// ListByUserIDs returns every device token for the given users, used when a
// notification must reach a user on all of their devices.
func (s *PushTokenStore) ListByUserIDs(ctx context.Context, userIDs []string) ([]models.PushToken, error) {
	if len(userIDs) == 0 {
		return []models.PushToken{}, nil
	}

	sql, args, err := psql.
		Select("id", "user_id", "token", "platform", "created_at", "updated_at").
		From("push_tokens").
		Where(sq.Eq{"user_id": userIDs}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list push tokens: %w", err)
	}
	defer rows.Close()

	var tokens []models.PushToken
	for rows.Next() {
		var t models.PushToken
		if err := rows.Scan(&t.ID, &t.UserID, &t.Token, &t.Platform, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan push token: %w", err)
		}
		tokens = append(tokens, t)
	}
	if tokens == nil {
		tokens = []models.PushToken{}
	}
	return tokens, nil
}
