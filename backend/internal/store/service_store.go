package store

import (
	"context"
	"fejd-backend/internal/models"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ServiceStore struct {
	pool *pgxpool.Pool
}

func NewServiceStore(pool *pgxpool.Pool) *ServiceStore {
	return &ServiceStore{pool: pool}
}

func (s *ServiceStore) ListByBusiness(ctx context.Context, businessID uuid.UUID) ([]models.Service, error) {
	sql, args, err := psql.
		Select("id", "business_id", "name", "duration_minutes", "price", "active", "COALESCE(description, '')", "picture_id", "created_at").
		From("services").
		Where(sq.Eq{"business_id": businessID}).
		OrderBy("name").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}
	defer rows.Close()

	var services []models.Service
	for rows.Next() {
		svc, err := scanService(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan service: %w", err)
		}
		services = append(services, *svc)
	}
	return services, nil
}

func (s *ServiceStore) GetByID(ctx context.Context, id uuid.UUID) (*models.Service, error) {
	sql, args, err := psql.
		Select("id", "business_id", "name", "duration_minutes", "price", "active", "COALESCE(description, '')", "picture_id", "created_at").
		From("services").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	return scanService(s.pool.QueryRow(ctx, sql, args...))
}

func (s *ServiceStore) Create(ctx context.Context, svc *models.Service) error {
	if svc.ID == uuid.Nil {
		svc.ID = uuid.New()
	}
	sql, args, err := psql.
		Insert("services").
		Columns("id", "business_id", "name", "duration_minutes", "price", "active", "description", "picture_id").
		Values(svc.ID, svc.BusinessID, svc.Name, svc.DurationMinutes, svc.Price, svc.Active, svc.Description, nullableUUID(svc.PictureID)).
		Suffix("RETURNING created_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	return s.pool.QueryRow(ctx, sql, args...).Scan(&svc.CreatedAt)
}

func (s *ServiceStore) Update(ctx context.Context, svc *models.Service) error {
	sql, args, err := psql.
		Update("services").
		Set("name", svc.Name).
		Set("duration_minutes", svc.DurationMinutes).
		Set("price", svc.Price).
		Set("active", svc.Active).
		Set("description", svc.Description).
		Set("picture_id", nullableUUID(svc.PictureID)).
		Where(sq.Eq{"id": svc.ID, "business_id": svc.BusinessID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = s.pool.Exec(ctx, sql, args...)
	return err
}

func (s *ServiceStore) Delete(ctx context.Context, id, businessID uuid.UUID) error {
	sql, args, err := psql.
		Delete("services").
		Where(sq.Eq{"id": id, "business_id": businessID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = s.pool.Exec(ctx, sql, args...)
	return err
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanService(row rowScanner) (*models.Service, error) {
	var svc models.Service
	var price *float64
	var pictureID *uuid.UUID
	if err := row.Scan(&svc.ID, &svc.BusinessID, &svc.Name, &svc.DurationMinutes, &price, &svc.Active, &svc.Description, &pictureID, &svc.CreatedAt); err != nil {
		return nil, fmt.Errorf("service not found: %w", err)
	}
	if price != nil {
		svc.Price = *price
	}
	svc.PictureID = pictureID
	return &svc, nil
}

func nullableUUID(id *uuid.UUID) any {
	if id == nil {
		return nil
	}
	return *id
}
