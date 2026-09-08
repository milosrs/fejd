package keycloak

import (
	"context"
	"time"
)

// User is the subset of Keycloak's UserRepresentation the jobs need.
type User struct {
	ID               string              `json:"id"`
	Username         string              `json:"username"`
	Email            string              `json:"email"`
	FirstName        string              `json:"firstName"`
	LastName         string              `json:"lastName"`
	Attributes       map[string][]string `json:"attributes"`
	RequiredActions  []string            `json:"requiredActions"`
	CreatedTimestamp int64               `json:"createdTimestamp"`
}

// UserManager is the admin-facing surface the jobs depend on. The concrete
// *Client implements it; tests provide a fake.
type UserManager interface {
	SearchUsersByAttribute(ctx context.Context, attr, value string) ([]User, error)
	UpdateUserAttributes(ctx context.Context, userID string, attrs map[string][]string) error
	ListUsersByRequiredActionAndAge(ctx context.Context, action string, olderThan time.Duration) ([]User, error)
	DeleteUser(ctx context.Context, userID string) error
}
