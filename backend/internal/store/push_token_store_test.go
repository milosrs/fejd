package store

import (
	"context"
	"testing"

	"fejd-backend/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPushTokenStore_UpsertAndListByUserIDs(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewPushTokenStore(db.pool)

	token := &models.PushToken{UserID: "user-1", Token: "fcm-token-1", Platform: "android"}
	require.NoError(t, store.Upsert(ctx, token))

	require.NoError(t, store.Upsert(ctx, &models.PushToken{UserID: "user-2", Token: "fcm-token-2", Platform: "ios"}))

	tokens, err := store.ListByUserIDs(ctx, []string{"user-1"})
	require.NoError(t, err)
	require.Len(t, tokens, 1)
	assert.Equal(t, "user-1", tokens[0].UserID)
	assert.Equal(t, "fcm-token-1", tokens[0].Token)
	assert.Equal(t, "android", tokens[0].Platform)
}

func TestPushTokenStore_UpsertReassignsExistingToken(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewPushTokenStore(db.pool)

	require.NoError(t, store.Upsert(ctx, &models.PushToken{UserID: "user-a", Token: "shared-token", Platform: "web"}))

	// A token belongs to a single device; re-registering it for another user
	// (e.g. the device changed hands) must move the row, not duplicate it.
	require.NoError(t, store.Upsert(ctx, &models.PushToken{UserID: "user-b", Token: "shared-token", Platform: "web"}))

	tokens, err := store.ListByUserIDs(ctx, []string{"user-b"})
	require.NoError(t, err)
	require.Len(t, tokens, 1)
	assert.Equal(t, "user-b", tokens[0].UserID)

	old, err := store.ListByUserIDs(ctx, []string{"user-a"})
	require.NoError(t, err)
	assert.Empty(t, old)
}

func TestPushTokenStore_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewPushTokenStore(db.pool)

	require.NoError(t, store.Upsert(ctx, &models.PushToken{UserID: "user-1", Token: "fcm-token-1", Platform: "android"}))
	require.NoError(t, store.Upsert(ctx, &models.PushToken{UserID: "user-1", Token: "fcm-token-2", Platform: "web"}))

	require.NoError(t, store.Delete(ctx, "user-1", "fcm-token-1"))

	tokens, err := store.ListByUserIDs(ctx, []string{"user-1"})
	require.NoError(t, err)
	require.Len(t, tokens, 1)
	assert.Equal(t, "fcm-token-2", tokens[0].Token)
}

func TestPushTokenStore_ListByUserIDs_Empty(t *testing.T) {
	db := setupTestDB(t)
	defer db.teardown()

	ctx := context.Background()
	store := NewPushTokenStore(db.pool)

	tokens, err := store.ListByUserIDs(ctx, nil)
	require.NoError(t, err)
	assert.Empty(t, tokens)
}
