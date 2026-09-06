package service

import (
	"context"
	"fmt"

	"fejd-backend/internal/keycloak"
	"fejd-backend/internal/models"
	"fejd-backend/internal/store"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const inviteClientID = "salon-mobile"

// UserInviter abstracts the Keycloak admin operations needed to invite an
// employee, so the service can be tested without a live Keycloak.
type UserInviter interface {
	CreateUser(ctx context.Context, in keycloak.CreateUserInput) (string, error)
	ExecuteActionsEmail(ctx context.Context, userID string, actions []string, clientID, redirectURI string, lifespan int) error
}

type EmployeeService struct {
	inviter           UserInviter
	businessUser      *store.BusinessUserStore
	employeeServices  *store.EmployeeServiceStore
	pool              *pgxpool.Pool
	inviteRedirectURI string
	inviteLifespan    int
}

func NewEmployeeService(
	inviter UserInviter,
	businessUser *store.BusinessUserStore,
	employeeServices *store.EmployeeServiceStore,
	pool *pgxpool.Pool,
	inviteRedirectURI string,
	inviteLifespan int,
) *EmployeeService {
	return &EmployeeService{
		inviter:           inviter,
		businessUser:      businessUser,
		employeeServices:  employeeServices,
		pool:              pool,
		inviteRedirectURI: inviteRedirectURI,
		inviteLifespan:    inviteLifespan,
	}
}

// Invite creates a Keycloak user (with an execute-actions invite email), a
// business_user row, and assigns the employee's services.
func (s *EmployeeService) Invite(ctx context.Context, businessID uuid.UUID, name, email string, serviceIDs []uuid.UUID) (*models.BusinessUser, error) {
	userID, err := s.inviter.CreateUser(ctx, keycloak.CreateUserInput{
		Username:        email,
		Email:           email,
		FirstName:       name,
		EmailVerified:   true,
		Enabled:         true,
		RequiredActions: []string{"UPDATE_PASSWORD"},
		Attributes:      map[string][]string{"approval_status": {"approved"}},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if err := s.inviter.ExecuteActionsEmail(ctx, userID, []string{"UPDATE_PASSWORD"}, inviteClientID, s.inviteRedirectURI, s.inviteLifespan); err != nil {
		return nil, fmt.Errorf("failed to send invite email: %w", err)
	}

	bu := &models.BusinessUser{
		BusinessID:  businessID,
		UserID:      userID,
		Role:        "employee",
		DisplayName: name,
		Active:      true,
	}
	if err := s.businessUser.Create(ctx, s.pool, bu); err != nil {
		return nil, fmt.Errorf("failed to create employee: %w", err)
	}

	if err := s.employeeServices.ReplaceByBusinessUser(ctx, bu.ID, serviceIDs); err != nil {
		return nil, fmt.Errorf("failed to assign services: %w", err)
	}

	return bu, nil
}
