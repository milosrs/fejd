package store

import (
	"context"
	"fejd-backend/internal/models"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BusinessStore struct {
	pool *pgxpool.Pool
}

func NewBusinessStore(pool *pgxpool.Pool) *BusinessStore {
	return &BusinessStore{pool: pool}
}

func (s *BusinessStore) GetBySlug(ctx context.Context, slug string) (*models.Business, error) {
	sql, args, err := psql.
		Select("id", "name", "slug", "created_at", "updated_at", "cancellation_lead_hours", "no_show_after_hours", "slot_interval_minutes").
		From("businesses").
		Where(sq.Eq{"slug": slug}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var b models.Business
	err = s.pool.QueryRow(ctx, sql, args...).Scan(&b.ID, &b.Name, &b.Slug, &b.CreatedAt, &b.UpdatedAt, &b.CancellationLeadHours, &b.NoShowAfterHours, &b.SlotIntervalMinutes)
	if err != nil {
		return nil, fmt.Errorf("business not found: %w", err)
	}
	return &b, nil
}

func (s *BusinessStore) GetByID(ctx context.Context, id uuid.UUID) (*models.Business, error) {
	sql, args, err := psql.
		Select("id", "name", "slug", "created_at", "updated_at", "cancellation_lead_hours", "no_show_after_hours", "slot_interval_minutes").
		From("businesses").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var b models.Business
	err = s.pool.QueryRow(ctx, sql, args...).Scan(&b.ID, &b.Name, &b.Slug, &b.CreatedAt, &b.UpdatedAt, &b.CancellationLeadHours, &b.NoShowAfterHours, &b.SlotIntervalMinutes)
	if err != nil {
		return nil, fmt.Errorf("business not found: %w", err)
	}
	return &b, nil
}

func (s *BusinessStore) ListMembershipsByUser(ctx context.Context, userID string) ([]models.BusinessMembership, error) {
	sql, args, err := psql.
		Select("b.id", "b.name", "b.slug", "bu.role").
		From("businesses b").
		Join("business_users bu ON bu.business_id = b.id").
		Where(sq.Eq{"bu.user_id": userID}).
		OrderBy("b.created_at").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list memberships: %w", err)
	}
	defer rows.Close()

	var memberships []models.BusinessMembership
	for rows.Next() {
		var m models.BusinessMembership
		if err := rows.Scan(&m.BusinessID, &m.Name, &m.Slug, &m.Role); err != nil {
			return nil, fmt.Errorf("failed to scan membership: %w", err)
		}
		memberships = append(memberships, m)
	}
	return memberships, nil
}

func (s *BusinessStore) Create(ctx context.Context, q Querier, b *models.Business) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	sql, args, err := psql.
		Insert("businesses").
		Columns("id", "name", "slug").
		Values(b.ID, b.Name, b.Slug).
		Suffix("RETURNING created_at, updated_at, cancellation_lead_hours, no_show_after_hours, slot_interval_minutes").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	return q.QueryRow(ctx, sql, args...).Scan(&b.CreatedAt, &b.UpdatedAt, &b.CancellationLeadHours, &b.NoShowAfterHours, &b.SlotIntervalMinutes)
}

// UpdatePolicy updates a business's cancellation, no-show and slot policies.
func (s *BusinessStore) UpdatePolicy(ctx context.Context, businessID uuid.UUID, cancellationLeadHours, noShowAfterHours, slotIntervalMinutes int) error {
	sql, args, err := psql.
		Update("businesses").
		Set("cancellation_lead_hours", cancellationLeadHours).
		Set("no_show_after_hours", noShowAfterHours).
		Set("slot_interval_minutes", slotIntervalMinutes).
		Where(sq.Eq{"id": businessID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = s.pool.Exec(ctx, sql, args...)
	return err
}

func (s *BusinessStore) SlugExists(ctx context.Context, q Querier, slug string) (bool, error) {
	sql, args, err := psql.
		Select("1").
		From("businesses").
		Where(sq.Eq{"slug": slug}).
		Prefix("SELECT EXISTS (").
		Suffix(")").
		ToSql()
	if err != nil {
		return false, fmt.Errorf("failed to build query: %w", err)
	}

	var exists bool
	if err := q.QueryRow(ctx, sql, args...).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check slug: %w", err)
	}
	return exists, nil
}
