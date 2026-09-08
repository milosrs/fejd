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

func seedUnavailability(t *testing.T, db *testDB) (businessID, buID, otherBuID uuid.UUID) {
	t.Helper()
	ctx := context.Background()

	businessID = uuid.New()
	buID = uuid.New()
	otherBuID = uuid.New()

	mustExec := func(sql string, args ...any) {
		t.Helper()
		_, err := db.pool.Exec(ctx, sql, args...)
		require.NoError(t, err)
	}

	mustExec(`INSERT INTO businesses (id, name, slug) VALUES ($1, 'Blocked', 'blocked')`, businessID)
	mustExec(`INSERT INTO business_users (id, business_id, user_id, role, display_name) VALUES ($1, $2, 'emp-1', 'employee', 'Emp')`, buID, businessID)
	mustExec(`INSERT INTO business_users (id, business_id, user_id, role, display_name) VALUES ($1, $2, 'emp-2', 'employee', 'Other')`, otherBuID, businessID)

	return businessID, buID, otherBuID
}

func TestEmployeeUnavailabilityStore_ListByBusinessUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	_, buID, _ := seedUnavailability(t, db)
	ctx := context.Background()
	store := NewEmployeeUnavailabilityStore(db.pool)

	base := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Hour)
	first := &models.EmployeeUnavailability{
		BusinessUserID: buID,
		StartTime:      base.Add(2 * time.Hour),
		EndTime:        base.Add(3 * time.Hour),
		Reason:         "errand",
	}
	second := &models.EmployeeUnavailability{
		BusinessUserID: buID,
		StartTime:      base,
		EndTime:        base.Add(time.Hour),
	}
	require.NoError(t, store.Create(ctx, db.pool, first))
	require.NoError(t, store.Create(ctx, db.pool, second))

	list, err := store.ListByBusinessUser(ctx, buID)
	require.NoError(t, err)
	require.Len(t, list, 2)
	// Ordered by start_time ascending.
	assert.True(t, list[0].StartTime.Before(list[1].StartTime))
	assert.Equal(t, second.ID, list[0].ID)
	assert.Equal(t, "errand", list[1].Reason)
}

func TestEmployeeUnavailabilityStore_DeleteForBusinessUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	_, buID, otherBuID := seedUnavailability(t, db)
	ctx := context.Background()
	store := NewEmployeeUnavailabilityStore(db.pool)

	base := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Hour)
	mine := &models.EmployeeUnavailability{
		BusinessUserID: buID,
		StartTime:      base,
		EndTime:        base.Add(time.Hour),
	}
	other := &models.EmployeeUnavailability{
		BusinessUserID: otherBuID,
		StartTime:      base,
		EndTime:        base.Add(time.Hour),
	}
	require.NoError(t, store.Create(ctx, db.pool, mine))
	require.NoError(t, store.Create(ctx, db.pool, other))

	// Deleting with the wrong owner leaves the row untouched.
	require.NoError(t, store.DeleteForBusinessUser(ctx, db.pool, otherBuID, mine.ID))

	list, err := store.ListByBusinessUser(ctx, buID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, mine.ID, list[0].ID)

	// Deleting with the right owner removes it.
	require.NoError(t, store.DeleteForBusinessUser(ctx, db.pool, buID, mine.ID))

	list, err = store.ListByBusinessUser(ctx, buID)
	require.NoError(t, err)
	assert.Empty(t, list)

	// The other employee's row is unaffected.
	otherList, err := store.ListByBusinessUser(ctx, otherBuID)
	require.NoError(t, err)
	require.Len(t, otherList, 1)
}
