package service

import (
	"context"
	"fmt"
)

// Valid registration roles. They mirror the realm roles a user can select on
// the registration form. The whitelist prevents privilege escalation: the
// registration_role attribute is user-writable, so it must never grant roles
// beyond this set.
var validRegistrationRoles = map[string]struct{}{
	"Customer": {},
	"Employee": {},
	"Owner":    {},
}

// RegistrationRoleManager is the Keycloak admin surface needed to finalize a
// self-registration: grant the chosen realm role and clear the temporary
// registration_role attribute. *keycloak.Client satisfies it.
type RegistrationRoleManager interface {
	AddRealmRole(ctx context.Context, userID, role string) error
	UpdateUserAttributes(ctx context.Context, userID string, attrs map[string][]string) error
}

// RegistrationService finalizes self-registrations by turning the role chosen
// on the registration form (carried as the registration_role user attribute /
// JWT claim) into an actual realm role.
type RegistrationService struct {
	users RegistrationRoleManager
}

func NewRegistrationService(users RegistrationRoleManager) *RegistrationService {
	return &RegistrationService{users: users}
}

// ClaimRegistrationRole grants the realm role recorded at registration and
// clears the registration_role attribute so it disappears from future tokens.
// Customers and employees are approved automatically; owners stay pending so
// an admin must verify them. It is idempotent. role must be Customer, Employee,
// or Owner; any other value is rejected.
func (s *RegistrationService) ClaimRegistrationRole(ctx context.Context, userID, role string) error {
	if _, ok := validRegistrationRoles[role]; !ok {
		return fmt.Errorf("invalid registration role: %s", role)
	}

	if err := s.users.AddRealmRole(ctx, userID, role); err != nil {
		return fmt.Errorf("failed to grant %s role: %w", role, err)
	}

	attrs := map[string][]string{"registration_role": {}}
	if role != "Owner" {
		attrs["approval_status"] = []string{"approved"}
	}

	if err := s.users.UpdateUserAttributes(ctx, userID, attrs); err != nil {
		return fmt.Errorf("failed to finalize registration: %w", err)
	}

	return nil
}
