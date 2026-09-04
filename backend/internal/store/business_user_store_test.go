package store

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBusinessUserStore_GetByBusinessAndUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewBusinessUserStore(db.pool)

	businessID := uuid.New()
	userID := "user-1"

	_, err := db.pool.Exec(ctx,
		`INSERT INTO businesses (id, name, slug) VALUES ($1, 'Test Biz', 'test-biz')`,
		businessID,
	)
	require.NoError(t, err)

	buID := uuid.New()
	_, err = db.pool.Exec(ctx,
		`INSERT INTO business_users (id, business_id, user_id, role, display_name) VALUES ($1, $2, $3, $4, $5)`,
		buID, businessID, userID, "admin", "Test User",
	)
	require.NoError(t, err)

	bu, err := store.GetByBusinessAndUser(ctx, businessID, userID)
	require.NoError(t, err)

	assert.Equal(t, buID, bu.ID)
	assert.Equal(t, businessID, bu.BusinessID)
	assert.Equal(t, userID, bu.UserID)
	assert.Equal(t, "admin", bu.Role)
	assert.Equal(t, "Test User", bu.DisplayName)
}

func TestBusinessUserStore_GetByBusinessAndUser_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewBusinessUserStore(db.pool)

	_, err := store.GetByBusinessAndUser(ctx, uuid.New(), "nonexistent")
	require.Error(t, err)
}

func TestBusinessUserStore_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewBusinessUserStore(db.pool)

	businessID := uuid.New()
	buID := uuid.New()
	userID := "user-2"

	_, err := db.pool.Exec(ctx,
		`INSERT INTO businesses (id, name, slug) VALUES ($1, 'Biz B', 'biz-b')`,
		businessID,
	)
	require.NoError(t, err)

	_, err = db.pool.Exec(ctx,
		`INSERT INTO business_users (id, business_id, user_id, role, display_name) VALUES ($1, $2, $3, $4, $5)`,
		buID, businessID, userID, "employee", "Employee One",
	)
	require.NoError(t, err)

	bu, err := store.GetByID(ctx, buID)
	require.NoError(t, err)

	assert.Equal(t, buID, bu.ID)
	assert.Equal(t, "Employee One", bu.DisplayName)
}

func TestBusinessUserStore_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewBusinessUserStore(db.pool)

	_, err := store.GetByID(ctx, uuid.New())
	require.Error(t, err)
}

func TestBusinessUserStore_ListByBusiness(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewBusinessUserStore(db.pool)

	businessID := uuid.New()

	_, err := db.pool.Exec(ctx,
		`INSERT INTO businesses (id, name, slug) VALUES ($1, 'Biz C', 'biz-c')`,
		businessID,
	)
	require.NoError(t, err)

	_, err = db.pool.Exec(ctx,
		`INSERT INTO business_users (id, business_id, user_id, role, display_name) VALUES
		 ($1, $2, 'user-a', 'admin', 'Alice'),
		 ($3, $2, 'user-b', 'employee', 'Bob'),
		 ($4, $2, 'user-c', 'employee', 'Charlie')`,
		uuid.New(), businessID, uuid.New(), uuid.New(),
	)
	require.NoError(t, err)

	users, err := store.ListByBusiness(ctx, businessID)
	require.NoError(t, err)

	require.Len(t, users, 3)
	assert.Equal(t, "Alice", users[0].DisplayName)
	assert.Equal(t, "Bob", users[1].DisplayName)
	assert.Equal(t, "Charlie", users[2].DisplayName)
}

func TestBusinessUserStore_ListByBusiness_Empty(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewBusinessUserStore(db.pool)

	users, err := store.ListByBusiness(ctx, uuid.New())
	require.NoError(t, err)
	assert.Empty(t, users)
}

func TestBusinessUserStore_ListEmployeesByBusiness(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewBusinessUserStore(db.pool)

	businessID := uuid.New()

	_, err := db.pool.Exec(ctx,
		`INSERT INTO businesses (id, name, slug) VALUES ($1, 'Biz D', 'biz-d')`,
		businessID,
	)
	require.NoError(t, err)

	_, err = db.pool.Exec(ctx,
		`INSERT INTO business_users (id, business_id, user_id, role, display_name) VALUES
		 ($1, $2, 'user-admin', 'admin', 'Admin'),
		 ($3, $2, 'user-emp1', 'employee', 'Dave'),
		 ($4, $2, 'user-emp2', 'employee', 'Eve')`,
		uuid.New(), businessID, uuid.New(), uuid.New(),
	)
	require.NoError(t, err)

	users, err := store.ListEmployeesByBusiness(ctx, businessID)
	require.NoError(t, err)

	require.Len(t, users, 2)
	for _, u := range users {
		assert.Equal(t, "employee", u.Role, "user %q should be an employee", u.DisplayName)
		assert.NotEqual(t, "Admin", u.DisplayName, "admin should not appear in employee list")
	}
	assert.Equal(t, "Dave", users[0].DisplayName)
}

func TestBusinessUserStore_IsAdmin(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewBusinessUserStore(db.pool)

	businessID := uuid.New()
	adminUserID := "admin-user"
	nonAdminUserID := "non-admin-user"

	_, err := db.pool.Exec(ctx,
		`INSERT INTO businesses (id, name, slug) VALUES ($1, 'Biz E', 'biz-e')`,
		businessID,
	)
	require.NoError(t, err)

	_, err = db.pool.Exec(ctx,
		`INSERT INTO business_users (id, business_id, user_id, role, display_name) VALUES
		 ($1, $2, $3, 'admin', 'Admin User'),
		 ($4, $2, $5, 'employee', 'Regular User')`,
		uuid.New(), businessID, adminUserID, uuid.New(), nonAdminUserID,
	)
	require.NoError(t, err)

	isAdmin, err := store.IsAdmin(ctx, businessID, adminUserID)
	require.NoError(t, err)
	assert.True(t, isAdmin)

	isAdmin, err = store.IsAdmin(ctx, businessID, nonAdminUserID)
	require.NoError(t, err)
	assert.False(t, isAdmin)

	isAdmin, err = store.IsAdmin(ctx, businessID, "nonexistent")
	require.NoError(t, err)
	assert.False(t, isAdmin)
}

func TestBusinessUserStore_HasAdminBusiness(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewBusinessUserStore(db.pool)

	businessID := uuid.New()
	adminUserID := "admin-user"
	employeeUserID := "employee-user"

	_, err := db.pool.Exec(ctx,
		`INSERT INTO businesses (id, name, slug) VALUES ($1, 'Biz F', 'biz-f')`,
		businessID,
	)
	require.NoError(t, err)

	_, err = db.pool.Exec(ctx,
		`INSERT INTO business_users (id, business_id, user_id, role, display_name) VALUES
		 ($1, $2, $3, 'admin', 'Admin'),
		 ($4, $2, $5, 'employee', 'Employee')`,
		uuid.New(), businessID, adminUserID, uuid.New(), employeeUserID,
	)
	require.NoError(t, err)

	has, err := store.HasAdminBusiness(ctx, db.pool, adminUserID)
	require.NoError(t, err)
	assert.True(t, has)

	has, err = store.HasAdminBusiness(ctx, db.pool, employeeUserID)
	require.NoError(t, err)
	assert.False(t, has)

	has, err = store.HasAdminBusiness(ctx, db.pool, "nonexistent")
	require.NoError(t, err)
	assert.False(t, has)
}
