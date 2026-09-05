package jobs

import (
	"context"
	"testing"
	"time"

	"fejd-backend/internal/keycloak"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInviteCleanup_Run(t *testing.T) {
	users := &fakeUserManager{
		listFn: func(ctx context.Context, action string, olderThan time.Duration) ([]keycloak.User, error) {
			assert.Equal(t, "UPDATE_PASSWORD", action)
			assert.Equal(t, 72*time.Hour, olderThan)
			return []keycloak.User{{ID: "1"}, {ID: "2"}}, nil
		},
	}

	i := NewInviteCleanup(users, 48)
	require.NoError(t, i.Run(context.Background()))

	require.Len(t, users.deleted, 2)
	assert.Equal(t, "1", users.deleted[0])
	assert.Equal(t, "2", users.deleted[1])
}
