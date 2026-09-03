package store

import (
	"context"
	"fejd-backend/internal/models"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ImageLinkStore struct {
	pool *pgxpool.Pool
}

func NewImageLinkStore(pool *pgxpool.Pool) *ImageLinkStore {
	return &ImageLinkStore{pool: pool}
}

const imageLinkColumns = "id, image_id, entity_type, entity_id, purpose, visibility, created_at"

func (s *ImageLinkStore) Create(ctx context.Context, q Querier, link *models.ImageLink) error {
	if link.ID == uuid.Nil {
		link.ID = uuid.New()
	}
	sql, args, err := psql.
		Insert("image_links").
		Columns("id", "image_id", "entity_type", "entity_id", "purpose", "visibility").
		Values(link.ID, link.ImageID, link.EntityType, link.EntityID, link.Purpose, string(link.Visibility)).
		Suffix("RETURNING created_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	return q.QueryRow(ctx, sql, args...).Scan(&link.CreatedAt)
}

func (s *ImageLinkStore) GetByEntityPurposeForUpdate(ctx context.Context, q Querier, entityType string, entityID uuid.UUID, purpose string) (*models.ImageLink, error) {
	sql, args, err := psql.
		Select(imageLinkColumns).
		From("image_links").
		Where(sq.Eq{"entity_type": entityType, "entity_id": entityID, "purpose": purpose}).
		Suffix("FOR UPDATE").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	return scanImageLink(q.QueryRow(ctx, sql, args...))
}

func (s *ImageLinkStore) UpdateImageID(ctx context.Context, q Querier, id, imageID uuid.UUID) error {
	sql, args, err := psql.
		Update("image_links").
		Set("image_id", imageID).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = q.Exec(ctx, sql, args...)
	return err
}

func (s *ImageLinkStore) DeleteByID(ctx context.Context, q Querier, id uuid.UUID) error {
	sql, args, err := psql.
		Delete("image_links").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = q.Exec(ctx, sql, args...)
	return err
}

func (s *ImageLinkStore) ListByEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]models.ImageLink, error) {
	sql, args, err := psql.
		Select(imageLinkColumns).
		From("image_links").
		Where(sq.Eq{"entity_type": entityType, "entity_id": entityID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	return s.list(ctx, sql, args...)
}

func (s *ImageLinkStore) ListByImage(ctx context.Context, imageID uuid.UUID) ([]models.ImageLink, error) {
	sql, args, err := psql.
		Select(imageLinkColumns).
		From("image_links").
		Where(sq.Eq{"image_id": imageID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	return s.list(ctx, sql, args...)
}

func (s *ImageLinkStore) CountByImage(ctx context.Context, imageID uuid.UUID) (int, error) {
	sql, args, err := psql.
		Select("COUNT(*)").
		From("image_links").
		Where(sq.Eq{"image_id": imageID}).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build query: %w", err)
	}

	var count int
	if err := s.pool.QueryRow(ctx, sql, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count links: %w", err)
	}
	return count, nil
}

func (s *ImageLinkStore) list(ctx context.Context, sql string, args ...any) ([]models.ImageLink, error) {
	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list links: %w", err)
	}
	defer rows.Close()

	var links []models.ImageLink
	for rows.Next() {
		l, err := scanImageLink(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan link: %w", err)
		}
		links = append(links, *l)
	}
	return links, nil
}

func scanImageLink(row rowScanner) (*models.ImageLink, error) {
	var l models.ImageLink
	var visibility string
	if err := row.Scan(&l.ID, &l.ImageID, &l.EntityType, &l.EntityID, &l.Purpose, &visibility, &l.CreatedAt); err != nil {
		return nil, err
	}
	l.Visibility = models.Visibility(visibility)
	return &l, nil
}
