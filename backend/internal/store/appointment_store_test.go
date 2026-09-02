package store

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func seedBooking(t *testing.T, db *testDB) (businessID, buID, serviceID uuid.UUID) {
	t.Helper()
	ctx := context.Background()

	businessID = uuid.New()
	buID = uuid.New()
	serviceID = uuid.New()

	mustExec := func(sql string, args ...any) {
		t.Helper()
		if _, err := db.pool.Exec(ctx, sql, args...); err != nil {
			t.Fatalf("seed failed: %v", err)
		}
	}

	mustExec(`INSERT INTO businesses (id, name, slug) VALUES ($1, 'Concurrency', 'concurrency')`, businessID)
	mustExec(`INSERT INTO business_users (id, business_id, user_id, role, display_name) VALUES ($1, $2, 'emp-1', 'employee', 'Emp')`, buID, businessID)
	mustExec(`INSERT INTO services (id, business_id, name, duration_minutes) VALUES ($1, $2, 'Cut', 30)`, serviceID, businessID)
	mustExec(`INSERT INTO employee_services (business_user_id, service_id) VALUES ($1, $2)`, buID, serviceID)

	return businessID, buID, serviceID
}

// TestConcurrentBookingSameSlotSingleWinner proves the DB-level guarantee that
// two simultaneous bookings for the same employee and overlapping time cannot
// both commit — exactly one wins and the other hits the exclusion constraint.
func TestConcurrentBookingSameSlotSingleWinner(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	businessID, buID, serviceID := seedBooking(t, db)

	start := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Hour)
	end := start.Add(30 * time.Minute)

	const n = 2
	errs := make([]error, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			customerID := fmt.Sprintf("customer-%d", i)
			_, err := db.pool.Exec(ctx,
				`INSERT INTO appointments (business_id, service_id, business_user_id, customer_user_id, start_time, end_time, status, created_by)
				 VALUES ($1, $2, $3, $4, $5, $6, 'confirmed', $4)`,
				businessID, serviceID, buID, customerID, start, end)
			errs[i] = err
		}(i)
	}
	wg.Wait()

	successes := 0
	for _, err := range errs {
		if err == nil {
			successes++
			continue
		}
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) || pgErr.Code != "23P01" {
			t.Errorf("expected exclusion violation (23P01), got %v", err)
		}
	}
	if successes != 1 {
		t.Fatalf("expected exactly 1 successful booking, got %d", successes)
	}
}

// TestAppointmentDailyCap verifies the "max 1 booking per customer per salon
// per day" rule and that cancelling frees the day.
func TestAppointmentDailyCap(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	businessID, buID, serviceID := seedBooking(t, db)

	day := time.Now().UTC().Add(48 * time.Hour).Truncate(24 * time.Hour)
	slot1 := day.Add(9 * time.Hour)
	slot2 := day.Add(10 * time.Hour)

	insert := func(start time.Time, status string) error {
		_, err := db.pool.Exec(ctx,
			`INSERT INTO appointments (business_id, service_id, business_user_id, customer_user_id, start_time, end_time, status, created_by)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $4)`,
			businessID, serviceID, buID, "customer-1", start, start.Add(30*time.Minute), status)
		return err
	}

	if err := insert(slot1, "confirmed"); err != nil {
		t.Fatalf("first booking should succeed: %v", err)
	}

	err := insert(slot2, "confirmed")
	if err == nil {
		t.Fatal("expected second same-day booking to fail with 23505")
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		t.Fatalf("expected 23505, got %v", err)
	}

	// Cancel the first booking; the day is freed for the customer.
	if _, err := db.pool.Exec(ctx,
		`UPDATE appointments SET status = 'cancelled' WHERE customer_user_id = 'customer-1' AND status = 'confirmed'`); err != nil {
		t.Fatalf("cancel failed: %v", err)
	}

	if err := insert(slot2, "confirmed"); err != nil {
		t.Fatalf("rebooking after cancel should succeed: %v", err)
	}
}
