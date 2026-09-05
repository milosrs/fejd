package jobs

import (
	"context"
	"time"

	"fejd-backend/internal/keycloak"
)

type fakeUserManager struct {
	searchFn func(ctx context.Context, attr, value string) ([]keycloak.User, error)
	listFn   func(ctx context.Context, action string, olderThan time.Duration) ([]keycloak.User, error)
	updated  []string
	deleted  []string
}

func (f *fakeUserManager) SearchUsersByAttribute(ctx context.Context, attr, value string) ([]keycloak.User, error) {
	if f.searchFn != nil {
		return f.searchFn(ctx, attr, value)
	}
	return nil, nil
}

func (f *fakeUserManager) UpdateUserAttributes(ctx context.Context, userID string, attrs map[string][]string) error {
	f.updated = append(f.updated, userID)
	return nil
}

func (f *fakeUserManager) ListUsersByRequiredActionAndAge(ctx context.Context, action string, olderThan time.Duration) ([]keycloak.User, error) {
	if f.listFn != nil {
		return f.listFn(ctx, action, olderThan)
	}
	return nil, nil
}

func (f *fakeUserManager) DeleteUser(ctx context.Context, userID string) error {
	f.deleted = append(f.deleted, userID)
	return nil
}

type sentMail struct {
	to      string
	subject string
	body    string
}

type fakeMailer struct {
	sent []sentMail
}

func (f *fakeMailer) Send(to, subject, body string) error {
	f.sent = append(f.sent, sentMail{to: to, subject: subject, body: body})
	return nil
}
