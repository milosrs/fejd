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
	keys := map[string]string{}
	for _, tr := range en {
		keys[tr.Key] = tr.Value
	}
	assert.Equal(t, "This page isn't set up yet", keys["landing.empty.title"])
	assert.Equal(t, "To book, you have to register.", keys["services.book.requiresAuth"])
	assert.Equal(t, "Choose a service", keys["booking.step.service"])

	rs, err := store.ListByLocale(ctx, "rs")
	require.NoError(t, err)
	require.NotEmpty(t, rs)

	none, err := store.ListByLocale(ctx, "zz")
	require.NoError(t, err)
	assert.Empty(t, none)
}
