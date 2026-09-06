package store

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceStore_LongestDurationByEmployee(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewServiceStore(db.pool)

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

	short := uuid.New()
	long := uuid.New()
	inactive := uuid.New()
	_, err = db.pool.Exec(ctx, `
		INSERT INTO services (id, business_id, name, duration_minutes, active) VALUES
		($1, $4, 'Haircut', 30, true),
		($2, $4, 'Coloring', 90, true),
		($3, $4, 'Retired', 120, false)`,
		short, long, inactive, businessID,
	)
	require.NoError(t, err)

	_, err = db.pool.Exec(ctx, `
		INSERT INTO employee_services (business_user_id, service_id) VALUES
		($1, $2), ($1, $3), ($1, $4)`,
		employeeID, short, long, inactive,
	)
	require.NoError(t, err)

	longest, err := store.LongestDurationByEmployee(ctx, employeeID)
	require.NoError(t, err)
	assert.Equal(t, 90, longest)
}

func TestServiceStore_LongestDurationByEmployee_None(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewServiceStore(db.pool)

	employeeID := uuid.New()

	longest, err := store.LongestDurationByEmployee(ctx, employeeID)
	require.NoError(t, err)
	assert.Equal(t, 0, longest)
}
