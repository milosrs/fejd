package store

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBusinessStore_SlugExists(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewBusinessStore(db.pool)

	_, err := db.pool.Exec(ctx,
		`INSERT INTO businesses (id, name, slug) VALUES ($1, 'Slug Biz', 'slug-biz')`,
		uuid.New(),
	)
	require.NoError(t, err)

	exists, err := store.SlugExists(ctx, db.pool, "slug-biz")
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = store.SlugExists(ctx, db.pool, "missing-slug")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestBusinessStore_List(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewBusinessStore(db.pool)

	_, err := db.pool.Exec(ctx,
		`INSERT INTO businesses (id, name, slug) VALUES
			($1, 'First Salon', 'first-salon'),
			($2, 'Second Salon', 'second-salon')`,
		uuid.New(), uuid.New(),
	)
	require.NoError(t, err)

	businesses, err := store.List(ctx)
	require.NoError(t, err)

	require.Len(t, businesses, 2)
	slugs := map[string]bool{}
	for _, b := range businesses {
		slugs[b.Slug] = true
	}
	assert.True(t, slugs["first-salon"])
	assert.True(t, slugs["second-salon"])
}
