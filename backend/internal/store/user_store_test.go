package store

import (
	"context"
	"testing"
	"time"

	"fejd-backend/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserStore_UpsertAndGet(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()
	ctx := context.Background()
	store := NewUserStore(db.pool)

	require.NoError(t, store.Upsert(ctx, &models.User{ID: "user-1", DisplayName: "Sam Barber", Email: "sam@example.com"}))

	got, err := store.GetByID(ctx, "user-1")
	require.NoError(t, err)
	assert.Equal(t, "Sam Barber", got.DisplayName)
	assert.Equal(t, "sam@example.com", got.Email)

	require.NoError(t, store.Upsert(ctx, &models.User{ID: "user-1", DisplayName: "Sammy B", Email: "sam@example.com"}))
	got, err = store.GetByID(ctx, "user-1")
	require.NoError(t, err)
	assert.Equal(t, "Sammy B", got.DisplayName)
}

func TestUserStore_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()
	_, err := NewUserStore(db.pool).GetByID(context.Background(), "missing")
	require.Error(t, err)
}

func TestAppointmentStore_ListCustomers(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()
	ctx := context.Background()

	businessID, buID, serviceID := seedBooking(t, db)
	userStore := NewUserStore(db.pool)

	require.NoError(t, userStore.Upsert(ctx, &models.User{ID: "cust-1", DisplayName: "Alice", Email: "alice@example.com"}))

	start := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Hour)
	insertAppt := func(customerID string, at time.Time) {
		_, err := db.pool.Exec(ctx,
			`INSERT INTO appointments (business_id, service_id, business_user_id, customer_user_id, start_time, end_time, status, created_by)
			 VALUES ($1, $2, $3, $4, $5, $6, 'confirmed', $4)`,
			businessID, serviceID, buID, customerID, at, at.Add(30*time.Minute))
		require.NoError(t, err)
	}
	insertAppt("cust-1", start)
	insertAppt("cust-2", start.Add(2*time.Hour))

	customers, err := NewAppointmentStore(db.pool).ListCustomers(ctx, businessID)
	require.NoError(t, err)
	require.Len(t, customers, 2)

	byID := map[string]string{}
	for _, c := range customers {
		byID[c.UserID] = c.DisplayName
	}
	assert.Equal(t, "Alice", byID["cust-1"])
	assert.Equal(t, "", byID["cust-2"])
}
