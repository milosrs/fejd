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
