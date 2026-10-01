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

func TestServiceCombinationStore_ReplaceAndList(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewServiceCombinationStore(db.pool)

	businessID := uuid.New()
	_, err := db.pool.Exec(ctx,
		`INSERT INTO businesses (id, name, slug) VALUES ($1, 'Salon', 'salon')`,
		businessID,
	)
	require.NoError(t, err)

	beard := uuid.New()
	fade := uuid.New()
	dye := uuid.New()
	_, err = db.pool.Exec(ctx, `
		INSERT INTO services (id, business_id, name, slug, duration_minutes, active) VALUES
		($1, $4, 'Beard', 'beard', 30, true),
		($2, $4, 'Fade', 'fade', 45, true),
		($3, $4, 'Hair Dye', 'hair-dye', 60, true)`,
		beard, fade, dye, businessID,
	)
	require.NoError(t, err)

	err = store.ReplaceByService(ctx, beard, []uuid.UUID{fade, dye})
	require.NoError(t, err)

	edges, err := store.ListByService(ctx, beard)
	require.NoError(t, err)
	require.Len(t, edges, 2)

	combinable := map[uuid.UUID]bool{}
	for _, e := range edges {
		combinable[e.CombinableServiceID] = true
	}
	assert.True(t, combinable[fade])
	assert.True(t, combinable[dye])

	// Directionality: fade is not combinable with anything yet.
	reverse, err := store.ListByService(ctx, fade)
	require.NoError(t, err)
	assert.Empty(t, reverse)

	// Replace narrows the set.
	err = store.ReplaceByService(ctx, beard, []uuid.UUID{fade})
	require.NoError(t, err)
	edges, err = store.ListByService(ctx, beard)
	require.NoError(t, err)
	require.Len(t, edges, 1)
	assert.Equal(t, fade, edges[0].CombinableServiceID)
}

func TestServiceCombinationStore_ListByBusiness(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewServiceCombinationStore(db.pool)

	businessID := uuid.New()
	_, err := db.pool.Exec(ctx,
		`INSERT INTO businesses (id, name, slug) VALUES ($1, 'Salon', 'salon')`,
		businessID,
	)
	require.NoError(t, err)

	beard := uuid.New()
	fade := uuid.New()
	_, err = db.pool.Exec(ctx, `
		INSERT INTO services (id, business_id, name, slug, duration_minutes, active) VALUES
		($1, $3, 'Beard', 'beard', 30, true),
		($2, $3, 'Fade', 'fade', 45, true)`,
		beard, fade, businessID,
	)
	require.NoError(t, err)

	err = store.ReplaceByService(ctx, beard, []uuid.UUID{fade})
	require.NoError(t, err)

	edges, err := store.ListByBusiness(ctx, businessID)
	require.NoError(t, err)
	require.Len(t, edges, 1)
	assert.Equal(t, beard, edges[0].ServiceID)
	assert.Equal(t, fade, edges[0].CombinableServiceID)
}

func TestAppointmentStore_AdditionalServices(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	appointmentStore := NewAppointmentStore(db.pool)

	businessID := uuid.New()
	employeeID := uuid.New()
	_, err := db.pool.Exec(ctx,
		`INSERT INTO businesses (id, name, slug) VALUES ($1, 'Salon', 'salon')`,
		businessID,
	)
	require.NoError(t, err)
	_, err = db.pool.Exec(ctx,
		`INSERT INTO business_users (id, business_id, user_id, role) VALUES ($1, $2, 'emp-1', 'employee')`,
		employeeID, businessID,
	)
	require.NoError(t, err)

	beard := uuid.New()
	fade := uuid.New()
	_, err = db.pool.Exec(ctx, `
		INSERT INTO services (id, business_id, name, slug, duration_minutes, active) VALUES
		($1, $3, 'Beard', 'beard', 30, true),
		($2, $3, 'Fade', 'fade', 45, true)`,
		beard, fade, businessID,
	)
	require.NoError(t, err)

	appt := &models.Appointment{
		BusinessID:     businessID,
		ServiceID:      beard,
		BusinessUserID: employeeID,
		CustomerUserID: "customer-1",
		StartTime:      time.Now().UTC().Add(24 * time.Hour),
		EndTime:        time.Now().UTC().Add(25 * time.Hour),
		Status:         models.AppointmentStatusConfirmed,
		CreatedBy:      "customer-1",
	}
	err = appointmentStore.Create(ctx, db.pool, appt)
	require.NoError(t, err)

	err = appointmentStore.InsertAdditionalServices(ctx, db.pool, appt.ID, []uuid.UUID{fade})
	require.NoError(t, err)

	got, err := appointmentStore.ListAdditionalServiceIDs(ctx, []uuid.UUID{appt.ID})
	require.NoError(t, err)
	assert.Equal(t, []uuid.UUID{fade}, got[appt.ID])
}
