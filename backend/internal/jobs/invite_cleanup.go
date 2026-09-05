package jobs

import (
	"context"
	"time"

	"fejd-backend/internal/keycloak"
)

// inviteGrace is the buffer past the invite expiry before a never-completed
// invite is deleted.
const inviteGrace = 24 * time.Hour

type InviteCleanup struct {
	users     keycloak.UserManager
	olderThan time.Duration
}

func NewInviteCleanup(users keycloak.UserManager, inviteExpiryHours int) *InviteCleanup {
	return &InviteCleanup{
		users:     users,
		olderThan: time.Duration(inviteExpiryHours)*time.Hour + inviteGrace,
	}
}

// Run deletes users whose UPDATE_PASSWORD invite has expired past the grace
// window without them completing registration.
func (i *InviteCleanup) Run(ctx context.Context) error {
	expired, err := i.users.ListUsersByRequiredActionAndAge(ctx, "UPDATE_PASSWORD", i.olderThan)
	if err != nil {
		return err
	}

	for _, u := range expired {
		if err := i.users.DeleteUser(ctx, u.ID); err != nil {
			return err
		}
	}
	return nil
}
