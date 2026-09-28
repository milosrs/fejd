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

func intPtr(v int) *int { return &v }

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
	dec31 := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)

	rules := []models.BusinessClosure{
		{BusinessID: businessID, ClosureType: models.BusinessClosureSingle, StartDate: &dec25, Reason: "Christmas"},
		{BusinessID: businessID, ClosureType: models.BusinessClosureRange, StartDate: &dec25, EndDate: &dec31},
		{BusinessID: businessID, ClosureType: models.BusinessClosureWeekly, DayOfWeek: intPtr(6)},
		{BusinessID: businessID, ClosureType: models.BusinessClosureYearly, Month: intPtr(1), Day: intPtr(1)},
	}

	err = store.ReplaceByBusiness(ctx, businessID, rules)
	require.NoError(t, err)

	closures, err := store.ListByBusiness(ctx, businessID)
	require.NoError(t, err)
	require.Len(t, closures, 4)

	// Single day.
	closed, err := store.IsClosed(ctx, businessID, dec25)
	require.NoError(t, err)
	assert.True(t, closed)

	// Range.
	closed, err = store.IsClosed(ctx, businessID, time.Date(2026, 12, 28, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	assert.True(t, closed)

	// Weekly: 2026-12-26 is a Saturday.
	closed, err = store.IsClosed(ctx, businessID, time.Date(2026, 12, 26, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	assert.True(t, closed)

	// Yearly: any January 1st.
	closed, err = store.IsClosed(ctx, businessID, time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	assert.True(t, closed)

	// An open day.
	closed, err = store.IsClosed(ctx, businessID, time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	assert.False(t, closed)

	// Replacing with an empty list clears every rule.
	err = store.ReplaceByBusiness(ctx, businessID, nil)
	require.NoError(t, err)

	closures, err = store.ListByBusiness(ctx, businessID)
	require.NoError(t, err)
	assert.Empty(t, closures)

	closed, err = store.IsClosed(ctx, businessID, dec25)
	require.NoError(t, err)
	assert.False(t, closed)
}
