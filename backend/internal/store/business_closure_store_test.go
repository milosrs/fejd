package store

import (
	"context"
	"testing"
	"time"

	"fejd-backend/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBusinessClosureStore_ReplaceListAndIsClosed(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	businessID := uuid.New()
	_, err := db.pool.Exec(ctx,
		`INSERT INTO businesses (id, name, slug) VALUES ($1, 'Closures', 'closures')`,
		businessID,
	)
	require.NoError(t, err)

	store := NewBusinessClosureStore(db.pool)

	dec25 := time.Date(2026, 12, 25, 0, 0, 0, 0, time.UTC)
	jan1 := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)

	err = store.ReplaceByBusiness(ctx, businessID, []models.BusinessClosure{
		{BusinessID: businessID, ClosureDate: jan1, Reason: "New Year"},
		{BusinessID: businessID, ClosureDate: dec25, Reason: "Christmas"},
	})
	require.NoError(t, err)

	closures, err := store.ListByBusiness(ctx, businessID)
	require.NoError(t, err)
	require.Len(t, closures, 2)

	// Ordered by date: Christmas before New Year.
	assert.Equal(t, dec25, closures[0].ClosureDate)
	assert.Equal(t, "Christmas", closures[0].Reason)
	assert.Equal(t, jan1, closures[1].ClosureDate)

	closed, err := store.IsClosed(ctx, businessID, dec25)
	require.NoError(t, err)
	assert.True(t, closed)

	closed, err = store.IsClosed(ctx, businessID, time.Date(2026, 12, 26, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	assert.False(t, closed)

	// Replacing with an empty list clears every closure.
	err = store.ReplaceByBusiness(ctx, businessID, nil)
	require.NoError(t, err)

	closures, err = store.ListByBusiness(ctx, businessID)
	require.NoError(t, err)
	assert.Empty(t, closures)
}
