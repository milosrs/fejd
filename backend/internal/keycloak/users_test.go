package keycloak

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsersExpired(t *testing.T) {
	cutoff := time.UnixMilli(1700000000000)

	users := []User{
		{ID: "a", RequiredActions: []string{"UPDATE_PASSWORD"}, CreatedTimestamp: 1600000000000},
		{ID: "b", RequiredActions: []string{"UPDATE_PASSWORD"}, CreatedTimestamp: 1800000000000},
		{ID: "c", RequiredActions: []string{"VERIFY_EMAIL"}, CreatedTimestamp: 1600000000000},
		{ID: "d", RequiredActions: []string{"UPDATE_PASSWORD"}, CreatedTimestamp: 0},
	}

	got := usersExpired(users, "UPDATE_PASSWORD", cutoff)

	require.Len(t, got, 1)
	assert.Equal(t, "a", got[0].ID)
}

func TestContainsString(t *testing.T) {
	assert.True(t, containsString([]string{"a", "b"}, "b"))
	assert.False(t, containsString([]string{"a", "b"}, "c"))
}
