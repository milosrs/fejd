package jobs

import (
	"context"
	"testing"

	"fejd-backend/internal/keycloak"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsersToNotify(t *testing.T) {
	users := []keycloak.User{
		{ID: "1", Attributes: map[string][]string{"approval_status": {"pending"}}},
		{ID: "2", Attributes: map[string][]string{"approval_status": {"pending"}, "notified_at": {"2024-01-01T00:00:00Z"}}},
		{ID: "3", Attributes: nil},
	}

	got := usersToNotify(users)

	require.Len(t, got, 2)
	assert.Equal(t, "1", got[0].ID)
	assert.Equal(t, "3", got[1].ID)
}

func TestPendingNotifier_Run(t *testing.T) {
	users := &fakeUserManager{
		searchFn: func(ctx context.Context, attr, value string) ([]keycloak.User, error) {
			assert.Equal(t, "approval_status", attr)
			assert.Equal(t, "pending", value)
			return []keycloak.User{
				{ID: "1", Username: "alice", Email: "alice@example.com"},
				{ID: "2", Username: "bob", Email: "bob@example.com", Attributes: map[string][]string{"notified_at": {"2024-01-01T00:00:00Z"}}},
			}, nil
		},
	}
	mailer := &fakeMailer{}

	p := NewPendingNotifier(users, mailer, "admin@example.com")
	require.NoError(t, p.Run(context.Background()))

	require.Len(t, mailer.sent, 1)
	assert.Equal(t, "admin@example.com", mailer.sent[0].to)
	assert.Equal(t, "New salon awaiting approval", mailer.sent[0].subject)
	assert.Contains(t, mailer.sent[0].body, "alice@example.com")

	require.Len(t, users.updated, 1)
	assert.Equal(t, "1", users.updated[0])
}
