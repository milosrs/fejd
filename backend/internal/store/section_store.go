package store

import (
	"context"
	"fejd-backend/internal/models"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SectionStore struct {
	pool *pgxpool.Pool
}

func NewSectionStore(pool *pgxpool.Pool) *SectionStore {
	return &SectionStore{pool: pool}
}

func (s *SectionStore) ListByPage(ctx context.Context, pageID uuid.UUID) ([]models.Section, error) {
	sql, args, err := psql.
		Select("id", "page_id", "type", "content", "position", "created_at", "updated_at").
		From("sections").
		Where(sq.Eq{"page_id": pageID}).
		OrderBy("position", "created_at").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list sections: %w", err)
	}
	defer rows.Close()

	var sections []models.Section
	for rows.Next() {
		var sec models.Section
		if err := rows.Scan(&sec.ID, &sec.PageID, &sec.Type, &sec.Content, &sec.Position, &sec.CreatedAt, &sec.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan section: %w", err)
		}
		sections = append(sections, sec)
	}
	return sections, nil
}

func (s *SectionStore) GetByID(ctx context.Context, id uuid.UUID) (*models.Section, error) {
	sql, args, err := psql.
		Select("id", "page_id", "type", "content", "position", "created_at", "updated_at").
		From("sections").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var sec models.Section
	err = s.pool.QueryRow(ctx, sql, args...).Scan(&sec.ID, &sec.PageID, &sec.Type, &sec.Content, &sec.Position, &sec.CreatedAt, &sec.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("section not found: %w", err)
	}
	return &sec, nil
}

func (s *SectionStore) Create(ctx context.Context, q Querier, sec *models.Section) error {
	if sec.ID == uuid.Nil {
		sec.ID = uuid.New()
	}
	sql, args, err := psql.
		Insert("sections").
		Columns("id", "page_id", "type", "content", "position").
		Values(sec.ID, sec.PageID, sec.Type, sq.Expr("?::jsonb", string(sec.Content)), sec.Position).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	return q.QueryRow(ctx, sql, args...).Scan(&sec.CreatedAt, &sec.UpdatedAt)
}

func (s *SectionStore) UpdateContent(ctx context.Context, q Querier, id uuid.UUID, content []byte) error {
	sql, args, err := psql.
		Update("sections").
		Set("content", sq.Expr("?::jsonb", string(content))).
		Set("updated_at", sq.Expr("now()")).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = q.Exec(ctx, sql, args...)
	return err
}

func (s *SectionStore) UpdatePosition(ctx context.Context, q Querier, id uuid.UUID, position int) error {
	sql, args, err := psql.
		Update("sections").
		Set("position", position).
		Set("updated_at", sq.Expr("now()")).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = q.Exec(ctx, sql, args...)
	return err
}

func (s *SectionStore) Delete(ctx context.Context, q Querier, id uuid.UUID) error {
	sql, args, err := psql.
		Delete("sections").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = q.Exec(ctx, sql, args...)
	return err
}
