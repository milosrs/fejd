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
		Select("id", "name", "slug", "created_at", "updated_at").
		From("businesses").
		Where(sq.Eq{"slug": slug}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var b models.Business
	err = s.pool.QueryRow(ctx, sql, args...).Scan(&b.ID, &b.Name, &b.Slug, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("business not found: %w", err)
	}
	return &b, nil
}

func (s *BusinessStore) GetByID(ctx context.Context, id uuid.UUID) (*models.Business, error) {
	sql, args, err := psql.
		Select("id", "name", "slug", "created_at", "updated_at").
		From("businesses").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var b models.Business
	err = s.pool.QueryRow(ctx, sql, args...).Scan(&b.ID, &b.Name, &b.Slug, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("business not found: %w", err)
	}
	return &b, nil
}

func (s *BusinessStore) ListByUser(ctx context.Context, userID string) ([]models.Business, error) {
	sql, args, err := psql.
		Select("b.id", "b.name", "b.slug", "b.created_at", "b.updated_at").
		From("businesses b").
		Join("business_users bu ON bu.business_id = b.id").
		Where(sq.Eq{"bu.user_id": userID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list businesses: %w", err)
	}
	defer rows.Close()

	var businesses []models.Business
	for rows.Next() {
		var b models.Business
		if err := rows.Scan(&b.ID, &b.Name, &b.Slug, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan business: %w", err)
		}
		businesses = append(businesses, b)
	}
	return businesses, nil
}

func (s *BusinessStore) Create(ctx context.Context, b *models.Business) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	sql, args, err := psql.
		Insert("businesses").
		Columns("id", "name", "slug").
		Values(b.ID, b.Name, b.Slug).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	return s.pool.QueryRow(ctx, sql, args...).Scan(&b.CreatedAt, &b.UpdatedAt)
}
