package store

import (
	"context"
	"errors"
	"fmt"

	"fejd-backend/internal/models"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SalonPublishStore struct {
	pool *pgxpool.Pool
}

func NewSalonPublishStore(pool *pgxpool.Pool) *SalonPublishStore {
	return &SalonPublishStore{pool: pool}
}

// GetByBusiness returns the salon's last published state, or pgx.ErrNoRows when
// it has never been published.
func (s *SalonPublishStore) GetByBusiness(ctx context.Context, businessID uuid.UUID) (*models.SalonPublish, error) {
	sql, args, err := psql.
		Select("business_id", "content_hash", "published_at").
		From("salon_publish").
		Where(sq.Eq{"business_id": businessID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var p models.SalonPublish
	if err := s.pool.QueryRow(ctx, sql, args...).Scan(&p.BusinessID, &p.ContentHash, &p.PublishedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get salon publish state: %w", err)
	}
	return &p, nil
}

// Upsert inserts or replaces the salon's published content hash.
func (s *SalonPublishStore) Upsert(ctx context.Context, q Querier, p *models.SalonPublish) error {
	sql, args, err := psql.
		Insert("salon_publish").
		Columns("business_id", "content_hash", "published_at").
		Values(p.BusinessID, p.ContentHash, p.PublishedAt).
		Suffix("ON CONFLICT (business_id) DO UPDATE SET content_hash = EXCLUDED.content_hash, published_at = EXCLUDED.published_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = q.Exec(ctx, sql, args...)
	return err
}
