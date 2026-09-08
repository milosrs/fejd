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

func TestInvitationStore_CreateAndGet(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewInvitationStore(db.pool)

	businessID := insertBusinessHelper(t, ctx, db, "Biz A", "biz-a")

	inv := &models.Invitation{
		BusinessID: businessID,
		TokenHash:  "hash-1",
		CreatedBy:  "owner-sub-1",
		ExpiresAt:  time.Now().UTC().Add(48 * time.Hour),
	}

	err := store.Create(ctx, db.pool, inv)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, inv.ID)
	assert.False(t, inv.CreatedAt.IsZero())
	assert.Equal(t, "employee", inv.Role)
	assert.Equal(t, 1, inv.MaxUses)

	got, err := store.GetByTokenHash(ctx, "hash-1")
	require.NoError(t, err)
	assert.Equal(t, inv.ID, got.ID)
	assert.Equal(t, businessID, got.BusinessID)
	assert.Equal(t, "owner-sub-1", got.CreatedBy)
	assert.Equal(t, 0, got.UseCount)
}

func TestInvitationStore_GetByTokenHash_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewInvitationStore(db.pool)

	_, err := store.GetByTokenHash(ctx, "does-not-exist")
	require.Error(t, err)
}

func TestInvitationStore_ListByBusiness(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewInvitationStore(db.pool)

	businessA := insertBusinessHelper(t, ctx, db, "Biz A", "biz-a")
	businessB := insertBusinessHelper(t, ctx, db, "Biz B", "biz-b")

	now := time.Now().UTC()
	for i, bh := range []string{"hash-a1", "hash-a2"} {
		inv := &models.Invitation{
			BusinessID: businessA,
			TokenHash:  bh,
			CreatedBy:  "owner-a",
			ExpiresAt:  now.Add(time.Duration(i+1) * time.Hour),
		}
		require.NoError(t, store.Create(ctx, db.pool, inv))
		time.Sleep(time.Millisecond) // ensure distinct created_at ordering
	}

	invB := &models.Invitation{
		BusinessID: businessB,
		TokenHash:  "hash-b1",
		CreatedBy:  "owner-b",
		ExpiresAt:  now.Add(time.Hour),
	}
	require.NoError(t, store.Create(ctx, db.pool, invB))

	listA, err := store.ListByBusiness(ctx, businessA)
	require.NoError(t, err)
	require.Len(t, listA, 2)
	assert.Equal(t, "hash-a2", listA[0].TokenHash, "newest first")
	assert.Equal(t, "hash-a1", listA[1].TokenHash)

	listB, err := store.ListByBusiness(ctx, businessB)
	require.NoError(t, err)
	require.Len(t, listB, 1)
	assert.Equal(t, "hash-b1", listB[0].TokenHash)
}

func TestInvitationStore_ListByBusiness_Empty(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewInvitationStore(db.pool)

	list, err := store.ListByBusiness(ctx, uuid.New())
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestInvitationStore_TryConsume_SingleUse(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewInvitationStore(db.pool)

	businessID := insertBusinessHelper(t, ctx, db, "Biz C", "biz-c")

	inv := &models.Invitation{
		BusinessID: businessID,
		TokenHash:  "hash-consume",
		CreatedBy:  "owner-c",
		ExpiresAt:  time.Now().UTC().Add(time.Hour),
	}
	require.NoError(t, store.Create(ctx, db.pool, inv))

	ok, err := store.TryConsume(ctx, db.pool, inv.ID)
	require.NoError(t, err)
	assert.True(t, ok)

	got, err := store.GetByID(ctx, inv.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, got.UseCount)

	ok, err = store.TryConsume(ctx, db.pool, inv.ID)
	require.NoError(t, err)
	assert.False(t, ok, "single-use invite must not be consumable twice")
}

func TestInvitationStore_TryConsume_Expired(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewInvitationStore(db.pool)

	businessID := insertBusinessHelper(t, ctx, db, "Biz D", "biz-d")

	inv := &models.Invitation{
		BusinessID: businessID,
		TokenHash:  "hash-expired",
		CreatedBy:  "owner-d",
		ExpiresAt:  time.Now().UTC().Add(-time.Hour),
	}
	require.NoError(t, store.Create(ctx, db.pool, inv))

	ok, err := store.TryConsume(ctx, db.pool, inv.ID)
	require.NoError(t, err)
	assert.False(t, ok, "expired invite must not be consumable")
}

func TestInvitationStore_TryConsume_Missing(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewInvitationStore(db.pool)

	ok, err := store.TryConsume(ctx, db.pool, uuid.New())
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestInvitationStore_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewInvitationStore(db.pool)

	businessID := insertBusinessHelper(t, ctx, db, "Biz E", "biz-e")

	inv := &models.Invitation{
		BusinessID: businessID,
		TokenHash:  "hash-delete",
		CreatedBy:  "owner-e",
		ExpiresAt:  time.Now().UTC().Add(time.Hour),
	}
	require.NoError(t, store.Create(ctx, db.pool, inv))

	require.NoError(t, store.Delete(ctx, db.pool, inv.ID))

	_, err := store.GetByID(ctx, inv.ID)
	require.Error(t, err)
}

func insertBusinessHelper(t *testing.T, ctx context.Context, db *testDB, name, slug string) uuid.UUID {
	t.Helper()
	businessID := uuid.New()
	_, err := db.pool.Exec(ctx,
		`INSERT INTO businesses (id, name, slug) VALUES ($1, $2, $3)`,
		businessID, name, slug,
	)
	require.NoError(t, err)
	return businessID
}
