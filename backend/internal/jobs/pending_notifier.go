package jobs

import (
	"context"
	"fmt"
	"time"

	"fejd-backend/internal/keycloak"
)

// Mailer is the minimal send surface PendingNotifier needs. *email.Sender
// satisfies it.
type Mailer interface {
	Send(to, subject, body string) error
}

type PendingNotifier struct {
	users    keycloak.UserManager
	mailer   Mailer
	notifyTo string
}

func NewPendingNotifier(users keycloak.UserManager, mailer Mailer, superadminEmail string) *PendingNotifier {
	return &PendingNotifier{users: users, mailer: mailer, notifyTo: superadminEmail}
}

// Run notifies the superadmin about pending users that have not been
// notified yet, then stamps them with notified_at.
func (p *PendingNotifier) Run(ctx context.Context) error {
	pending, err := p.users.SearchUsersByAttribute(ctx, "approval_status", "pending")
	if err != nil {
		return err
	}

	for _, u := range usersToNotify(pending) {
		body := fmt.Sprintf("A new salon owner registered and is awaiting approval.\n\nUser: %s\nEmail: %s\n", u.Username, u.Email)
		if err := p.mailer.Send(p.notifyTo, "New salon awaiting approval", body); err != nil {
			return err
		}
		attrs := map[string][]string{"notified_at": {time.Now().UTC().Format(time.RFC3339)}}
		if err := p.users.UpdateUserAttributes(ctx, u.ID, attrs); err != nil {
			return err
		}
	}
	return nil
}

// usersToNotify drops pending users that already carry a notified_at attribute.
func usersToNotify(users []keycloak.User) []keycloak.User {
	out := make([]keycloak.User, 0, len(users))
	for _, u := range users {
		if _, ok := u.Attributes["notified_at"]; ok {
			continue
		}
		out = append(out, u)
	}
	return out
}
