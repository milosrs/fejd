package store

import (
	"context"
	"fejd-backend/internal/models"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LandingPageName is the well-known page name for a salon's landing page.
const LandingPageName = "landing"

type PageStore struct {
	pool *pgxpool.Pool
}

func NewPageStore(pool *pgxpool.Pool) *PageStore {
	return &PageStore{pool: pool}
}

func (s *PageStore) GetByBusinessAndName(ctx context.Context, businessID uuid.UUID, name string) (*models.Page, error) {
	sql, args, err := psql.
		Select("id", "business_id", "name", "position", "created_at", "updated_at").
		From("pages").
		Where(sq.Eq{"business_id": businessID, "name": name}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var p models.Page
	err = s.pool.QueryRow(ctx, sql, args...).Scan(&p.ID, &p.BusinessID, &p.Name, &p.Position, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("page not found: %w", err)
	}
	return &p, nil
}

func (s *PageStore) Create(ctx context.Context, q Querier, p *models.Page) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	sql, args, err := psql.
		Insert("pages").
		Columns("id", "business_id", "name", "position").
		Values(p.ID, p.BusinessID, p.Name, p.Position).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	return q.QueryRow(ctx, sql, args...).Scan(&p.CreatedAt, &p.UpdatedAt)
}
