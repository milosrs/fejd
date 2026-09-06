package store

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPageStore_GetByBusinessAndName(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewPageStore(db.pool)

	businessID := uuid.New()
	pageID := uuid.New()
	_, err := db.pool.Exec(ctx,
		`INSERT INTO businesses (id, name, slug) VALUES ($1, 'Salon', 'salon')`,
		businessID,
	)
	require.NoError(t, err)
	_, err = db.pool.Exec(ctx,
		`INSERT INTO pages (id, business_id, name, position) VALUES ($1, $2, 'landing', 0)`,
		pageID, businessID,
	)
	require.NoError(t, err)

	page, err := store.GetByBusinessAndName(ctx, businessID, "landing")
	require.NoError(t, err)
	assert.Equal(t, pageID, page.ID)
	assert.Equal(t, "landing", page.Name)

	_, err = store.GetByBusinessAndName(ctx, businessID, "missing")
	assert.Error(t, err)
}
