package store

import (
	"context"
	"testing"

	"github.com/google/uuid"
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
	if err != nil {
		t.Fatalf("failed to insert business: %v", err)
	}

	buID := uuid.New()
	_, err = db.pool.Exec(ctx,
		`INSERT INTO business_users (id, business_id, user_id, role, display_name) VALUES ($1, $2, $3, $4, $5)`,
		buID, businessID, userID, "admin", "Test User",
	)
	if err != nil {
		t.Fatalf("failed to insert business user: %v", err)
	}

	bu, err := store.GetByBusinessAndUser(ctx, businessID, userID)
	if err != nil {
		t.Fatalf("GetByBusinessAndUser failed: %v", err)
	}

	if bu.ID != buID {
		t.Errorf("expected ID %v, got %v", buID, bu.ID)
	}
	if bu.BusinessID != businessID {
		t.Errorf("expected BusinessID %v, got %v", businessID, bu.BusinessID)
	}
	if bu.UserID != userID {
		t.Errorf("expected UserID %q, got %q", userID, bu.UserID)
	}
	if bu.Role != "admin" {
		t.Errorf("expected role 'admin', got %q", bu.Role)
	}
	if bu.DisplayName != "Test User" {
		t.Errorf("expected DisplayName 'Test User', got %q", bu.DisplayName)
	}
}

func TestBusinessUserStore_GetByBusinessAndUser_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewBusinessUserStore(db.pool)

	_, err := store.GetByBusinessAndUser(ctx, uuid.New(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
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
	if err != nil {
		t.Fatalf("failed to insert business: %v", err)
	}

	_, err = db.pool.Exec(ctx,
		`INSERT INTO business_users (id, business_id, user_id, role, display_name) VALUES ($1, $2, $3, $4, $5)`,
		buID, businessID, userID, "employee", "Employee One",
	)
	if err != nil {
		t.Fatalf("failed to insert business user: %v", err)
	}

	bu, err := store.GetByID(ctx, buID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if bu.ID != buID {
		t.Errorf("expected ID %v, got %v", buID, bu.ID)
	}
	if bu.DisplayName != "Employee One" {
		t.Errorf("expected DisplayName 'Employee One', got %q", bu.DisplayName)
	}
}

func TestBusinessUserStore_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewBusinessUserStore(db.pool)

	_, err := store.GetByID(ctx, uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
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
	if err != nil {
		t.Fatalf("failed to insert business: %v", err)
	}

	_, err = db.pool.Exec(ctx,
		`INSERT INTO business_users (id, business_id, user_id, role, display_name) VALUES
		 ($1, $2, 'user-a', 'admin', 'Alice'),
		 ($3, $2, 'user-b', 'employee', 'Bob'),
		 ($4, $2, 'user-c', 'employee', 'Charlie')`,
		uuid.New(), businessID, uuid.New(), uuid.New(),
	)
	if err != nil {
		t.Fatalf("failed to insert business users: %v", err)
	}

	users, err := store.ListByBusiness(ctx, businessID)
	if err != nil {
		t.Fatalf("ListByBusiness failed: %v", err)
	}

	if len(users) != 3 {
		t.Fatalf("expected 3 users, got %d", len(users))
	}

	if users[0].DisplayName != "Alice" {
		t.Errorf("expected first user to be Alice, got %q", users[0].DisplayName)
	}
	if users[1].DisplayName != "Bob" {
		t.Errorf("expected second user to be Bob, got %q", users[1].DisplayName)
	}
	if users[2].DisplayName != "Charlie" {
		t.Errorf("expected third user to be Charlie, got %q", users[2].DisplayName)
	}
}

func TestBusinessUserStore_ListByBusiness_Empty(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewBusinessUserStore(db.pool)

	users, err := store.ListByBusiness(ctx, uuid.New())
	if err != nil {
		t.Fatalf("ListByBusiness failed: %v", err)
	}

	if len(users) != 0 {
		t.Errorf("expected 0 users, got %d", len(users))
	}
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
	if err != nil {
		t.Fatalf("failed to insert business: %v", err)
	}

	_, err = db.pool.Exec(ctx,
		`INSERT INTO business_users (id, business_id, user_id, role, display_name) VALUES
		 ($1, $2, 'user-admin', 'admin', 'Admin'),
		 ($3, $2, 'user-emp1', 'employee', 'Dave'),
		 ($4, $2, 'user-emp2', 'employee', 'Eve')`,
		uuid.New(), businessID, uuid.New(), uuid.New(),
	)
	if err != nil {
		t.Fatalf("failed to insert business users: %v", err)
	}

	users, err := store.ListEmployeesByBusiness(ctx, businessID)
	if err != nil {
		t.Fatalf("ListEmployeesByBusiness failed: %v", err)
	}

	if len(users) != 2 {
		t.Fatalf("expected 2 employees, got %d", len(users))
	}

	for _, u := range users {
		if u.Role != "employee" {
			t.Errorf("expected role 'employee', got %q for user %q", u.Role, u.DisplayName)
		}
		if u.DisplayName == "Admin" {
			t.Error("admin should not appear in employee list")
		}
	}

	if users[0].DisplayName != "Dave" {
		t.Errorf("expected first employee to be Dave, got %q", users[0].DisplayName)
	}
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
	if err != nil {
		t.Fatalf("failed to insert business: %v", err)
	}

	_, err = db.pool.Exec(ctx,
		`INSERT INTO business_users (id, business_id, user_id, role, display_name) VALUES
		 ($1, $2, $3, 'admin', 'Admin User'),
		 ($4, $2, $5, 'employee', 'Regular User')`,
		uuid.New(), businessID, adminUserID, uuid.New(), nonAdminUserID,
	)
	if err != nil {
		t.Fatalf("failed to insert business users: %v", err)
	}

	isAdmin, err := store.IsAdmin(ctx, businessID, adminUserID)
	if err != nil {
		t.Fatalf("IsAdmin failed: %v", err)
	}
	if !isAdmin {
		t.Error("expected IsAdmin to return true for admin user")
	}

	isAdmin, err = store.IsAdmin(ctx, businessID, nonAdminUserID)
	if err != nil {
		t.Fatalf("IsAdmin failed: %v", err)
	}
	if isAdmin {
		t.Error("expected IsAdmin to return false for non-admin user")
	}

	isAdmin, err = store.IsAdmin(ctx, businessID, "nonexistent")
	if err != nil {
		t.Fatalf("IsAdmin failed: %v", err)
	}
	if isAdmin {
		t.Error("expected IsAdmin to return false for nonexistent user")
	}
}
