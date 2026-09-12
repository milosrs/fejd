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

func TestBusinessStore_Rename(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewBusinessStore(db.pool)

	businessID := uuid.New()
	_, err := db.pool.Exec(ctx,
		`INSERT INTO businesses (id, name, slug) VALUES ($1, 'Old Name', 'old-name')`,
		businessID,
	)
	require.NoError(t, err)

	err = store.Rename(ctx, db.pool, businessID, "New Name", "new-name")
	require.NoError(t, err)

	b, err := store.GetByID(ctx, businessID)
	require.NoError(t, err)
	assert.Equal(t, "New Name", b.Name)
	assert.Equal(t, "new-name", b.Slug)
}

func TestBusinessStore_SlugTakenByOther(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewBusinessStore(db.pool)

	_, err := db.pool.Exec(ctx,
		`INSERT INTO businesses (id, name, slug) VALUES ($1, 'A', 'taken')`,
		uuid.New(),
	)
	require.NoError(t, err)

	taken, err := store.SlugTakenByOther(ctx, db.pool, "taken", uuid.New())
	require.NoError(t, err)
	assert.True(t, taken)

	taken, err = store.SlugTakenByOther(ctx, db.pool, "free", uuid.New())
	require.NoError(t, err)
	assert.False(t, taken)
}
