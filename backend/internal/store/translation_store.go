package store

import (
	"context"
	"fejd-backend/internal/models"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TranslationStore struct {
	pool *pgxpool.Pool
}

func NewTranslationStore(pool *pgxpool.Pool) *TranslationStore {
	return &TranslationStore{pool: pool}
}

func (s *TranslationStore) ListByLocale(ctx context.Context, locale string) ([]models.Translation, error) {
	sql, args, err := psql.
		Select("id", "key", "locale", "value", "created_at", "updated_at").
		From("translations").
		Where(sq.Eq{"locale": locale}).
		OrderBy("key").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list translations: %w", err)
	}
	defer rows.Close()

	var translations []models.Translation
	for rows.Next() {
		var tr models.Translation
		if err := rows.Scan(&tr.ID, &tr.Key, &tr.Locale, &tr.Value, &tr.CreatedAt, &tr.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan translation: %w", err)
		}
		translations = append(translations, tr)
	}
	return translations, nil
}
