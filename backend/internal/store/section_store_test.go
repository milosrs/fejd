package store

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSectionStore_ListByPage(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewSectionStore(db.pool)

	businessID := uuid.New()
	pageID := uuid.New()
	_, err := db.pool.Exec(ctx,
		`INSERT INTO businesses (id, name, slug) VALUES ($1, 'Salon', 'salon')`,
		businessID,
	)
	require.NoError(t, err)
	_, err = db.pool.Exec(ctx,
		`INSERT INTO pages (id, business_id, name) VALUES ($1, $2, 'landing')`,
		pageID, businessID,
	)
	require.NoError(t, err)

	_, err = db.pool.Exec(ctx, `
		INSERT INTO sections (page_id, type, content, position) VALUES
		($1, 'about',   '{"en":{"heading":"About","body":"Hello"}}'::jsonb, 2),
		($1, 'hero',    '{"en":{"headline":"Cuts","cta_text":"Book"}}'::jsonb, 0),
		($1, 'contact', '{}'::jsonb, 1)`,
		pageID,
	)
	require.NoError(t, err)

	sections, err := store.ListByPage(ctx, pageID)
	require.NoError(t, err)
	require.Len(t, sections, 3)

	assert.Equal(t, "hero", sections[0].Type)
	assert.Equal(t, "contact", sections[1].Type)
	assert.Equal(t, "about", sections[2].Type)
	assert.JSONEq(t, `{"en":{"headline":"Cuts","cta_text":"Book"}}`, string(sections[0].Content))
}

func TestSectionStore_ListByPage_Empty(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewSectionStore(db.pool)

	businessID := uuid.New()
	pageID := uuid.New()
	_, err := db.pool.Exec(ctx,
		`INSERT INTO businesses (id, name, slug) VALUES ($1, 'Salon', 'salon')`,
		businessID,
	)
	require.NoError(t, err)
	_, err = db.pool.Exec(ctx,
		`INSERT INTO pages (id, business_id, name) VALUES ($1, $2, 'landing')`,
		pageID, businessID,
	)
	require.NoError(t, err)

	sections, err := store.ListByPage(ctx, pageID)
	require.NoError(t, err)
	assert.Empty(t, sections)
}
