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
