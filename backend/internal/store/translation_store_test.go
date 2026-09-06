package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranslationStore_ListByLocale(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewTranslationStore(db.pool)

	_, err := db.pool.Exec(ctx, `
		INSERT INTO translations (key, locale, value) VALUES
		('a.key', 'de', 'A'),
		('b.key', 'de', 'B')`)
	require.NoError(t, err)

	de, err := store.ListByLocale(ctx, "de")
	require.NoError(t, err)
	require.Len(t, de, 2)
	assert.Equal(t, "a.key", de[0].Key)
	assert.Equal(t, "b.key", de[1].Key)
}

func TestTranslationStore_ListByLocale_Seeded(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewTranslationStore(db.pool)

	en, err := store.ListByLocale(ctx, "en")
	require.NoError(t, err)
	require.Len(t, en, 3)
	assert.Equal(t, "landing.empty.body", en[0].Key)
	assert.Equal(t, "landing.empty.title", en[1].Key)
	assert.Equal(t, "services.book.requiresAuth", en[2].Key)

	rs, err := store.ListByLocale(ctx, "rs")
	require.NoError(t, err)
	require.Len(t, rs, 3)

	none, err := store.ListByLocale(ctx, "zz")
	require.NoError(t, err)
	assert.Empty(t, none)
}
