package store

import (
	"context"
	"fejd-backend/internal/models"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ImageStore struct {
	pool *pgxpool.Pool
}

func NewImageStore(pool *pgxpool.Pool) *ImageStore {
	return &ImageStore{pool: pool}
}

func (s *ImageStore) Create(ctx context.Context, img *models.Image) error {
	if img.ID == uuid.Nil {
		img.ID = uuid.New()
	}
	sql, args, err := psql.
		Insert("images").
		Columns("id", "storage", "object_key", "data", "url", "content_type").
		Values(img.ID, img.Storage, nullableString(img.ObjectKey), img.Data, nullableString(img.URL), nullableString(img.ContentType)).
		Suffix("RETURNING created_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	return s.pool.QueryRow(ctx, sql, args...).Scan(&img.CreatedAt)
}

func (s *ImageStore) GetByID(ctx context.Context, id uuid.UUID) (*models.Image, error) {
	sql, args, err := psql.
		Select("id", "storage", "COALESCE(object_key, '')", "data", "COALESCE(url, '')", "COALESCE(content_type, '')", "created_at").
		From("images").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var img models.Image
	if err := s.pool.QueryRow(ctx, sql, args...).Scan(&img.ID, &img.Storage, &img.ObjectKey, &img.Data, &img.URL, &img.ContentType, &img.CreatedAt); err != nil {
		return nil, fmt.Errorf("image not found: %w", err)
	}
	return &img, nil
}

func (s *ImageStore) Delete(ctx context.Context, id uuid.UUID) error {
	sql, args, err := psql.
		Delete("images").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = s.pool.Exec(ctx, sql, args...)
	return err
}
