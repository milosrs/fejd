package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"fejd-backend/internal/db"
	"fejd-backend/internal/models"
	"fejd-backend/internal/store"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Sentinel errors returned by invitation lookups and redemption. Handlers map
// these to HTTP status codes.
var (
	ErrInvitationNotFound   = errors.New("invitation not found")
	ErrInvitationExpired    = errors.New("invitation expired")
	ErrInvitationUsed       = errors.New("invitation already used")
	ErrCannotInviteOwner    = errors.New("cannot invite the salon owner")
	ErrInvalidInviteRole    = errors.New("invalid invitation role")
	ErrCannotInviteEmployee = errors.New("only the salon owner can invite employees")
	ErrInviteBaseURLUnset   = errors.New("INVITE_BASE_URL is not configured")
)

// Invitation roles. They mirror the realm roles (Employee/Customer/Owner) and,
// for employees, the business_users.role value. Customers and owners never get
// a business_users row — customers are referenced by their Keycloak sub, and
// owners create their salon later via the onboarding flow.
const (
	inviteRoleEmployee   = "employee"
	inviteRoleCustomer   = "customer"
	inviteRoleOwner      = "owner"
	inviteRoleRealmAdmin = "realm-admin"
)

// InvitationUserManager is the Keycloak admin surface the invitation service
// needs for role/approval grants only — never for reading user attributes.
// *keycloak.Client satisfies it; tests provide a fake.
type InvitationUserManager interface {
	AddRealmRole(ctx context.Context, userID, role string) error
	AddClientRole(ctx context.Context, userID, clientID, role string) error
	UpdateUserAttributes(ctx context.Context, userID string, attrs map[string][]string) error
}

// InvitationService creates shareable link/QR invitations that link a user to a
// business as an employee or customer, or — for realm admins — to the platform
// as a customer or owner. Only the token hash is persisted; the raw token is
// returned to the caller to embed in the invite URL.
type InvitationService struct {
	invitations   *store.InvitationStore
	businessStore *store.BusinessStore
	businessUsers *store.BusinessUserStore
	userStore     *store.UserStore
	users         InvitationUserManager
	pool          *pgxpool.Pool
	baseURL       string
}

func NewInvitationService(
	invitations *store.InvitationStore,
	businessStore *store.BusinessStore,
	businessUsers *store.BusinessUserStore,
	userStore *store.UserStore,
	users InvitationUserManager,
	pool *pgxpool.Pool,
	baseURL string,
) *InvitationService {
	return &InvitationService{
		invitations:   invitations,
		businessStore: businessStore,
		businessUsers: businessUsers,
		userStore:     userStore,
		users:         users,
		pool:          pool,
		baseURL:       baseURL,
	}
}

// InvitationOut is the shareable result of creating an invitation.
type InvitationOut struct {
	ID        uuid.UUID
	Token     string
	URL       string
	ExpiresAt time.Time
}

// InvitationPublic is the non-secret view of an invitation exposed to the
// landing page / app before registration. It never echoes the token.
type InvitationPublic struct {
	SalonName string
	SalonSlug string
	Role      string
	ExpiresAt time.Time
}

// CreateInvitation mints a single-use, expiring invite for the business and
// returns the raw token + invite URL. role must be inviteRoleEmployee or
// inviteRoleCustomer; an empty role defaults to employee. Only the salon owner
// (business_users admin) may invite employees; members may only invite
// customers. ttl<=0 falls back to the caller's default.
func (s *InvitationService) CreateInvitation(ctx context.Context, businessID uuid.UUID, createdBy, role string, ttl time.Duration) (*InvitationOut, error) {
	if role == "" {
		role = inviteRoleEmployee
	}
	if role != inviteRoleEmployee && role != inviteRoleCustomer {
		return nil, ErrInvalidInviteRole
	}

	if role == inviteRoleEmployee {
		bu, err := s.businessUsers.GetByBusinessAndUser(ctx, businessID, createdBy)
		if err != nil || bu.Role != "admin" {
			return nil, ErrCannotInviteEmployee
		}
	}

	if ttl <= 0 {
		ttl = 48 * time.Hour
	}

	raw, hash, err := newInviteToken()
	if err != nil {
		return nil, err
	}

	inv := &models.Invitation{
		BusinessID: businessID,
		TokenHash:  hash,
		Role:       role,
		CreatedBy:  createdBy,
		MaxUses:    1,
		ExpiresAt:  time.Now().UTC().Add(ttl),
	}
	if err := s.invitations.Create(ctx, s.pool, inv); err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	inviteURL, err := s.inviteURL(raw)
	if err != nil {
		return nil, err
	}

	return &InvitationOut{
		ID:        inv.ID,
		Token:     raw,
		URL:       inviteURL,
		ExpiresAt: inv.ExpiresAt,
	}, nil
}

// CreatePlatformInvitation mints a single-use, expiring invite to the platform
// (no salon) as a customer, owner, or realm admin. It is intended for realm
// administrators.
func (s *InvitationService) CreatePlatformInvitation(ctx context.Context, createdBy, role string, ttl time.Duration) (*InvitationOut, error) {
	if role != inviteRoleCustomer && role != inviteRoleOwner && role != inviteRoleRealmAdmin {
		return nil, ErrInvalidInviteRole
	}
	return s.mintPlatformInvite(ctx, createdBy, role, ttl)
}

// CreateCustomerInvitation mints a single-use, expiring invite to the platform
// (no salon) that grants the Customer role. Any authenticated user may create
// one — this is the "invite a friend" path.
func (s *InvitationService) CreateCustomerInvitation(ctx context.Context, createdBy string, ttl time.Duration) (*InvitationOut, error) {
	return s.mintPlatformInvite(ctx, createdBy, inviteRoleCustomer, ttl)
}

func (s *InvitationService) mintPlatformInvite(ctx context.Context, createdBy, role string, ttl time.Duration) (*InvitationOut, error) {
	if ttl <= 0 {
		ttl = 48 * time.Hour
	}

	raw, hash, err := newInviteToken()
	if err != nil {
		return nil, err
	}

	inv := &models.Invitation{
		TokenHash: hash,
		Role:      role,
		CreatedBy: createdBy,
		MaxUses:   1,
		ExpiresAt: time.Now().UTC().Add(ttl),
	}
	if err := s.invitations.Create(ctx, s.pool, inv); err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	inviteURL, err := s.inviteURL(raw)
	if err != nil {
		return nil, err
	}

	return &InvitationOut{
		ID:        inv.ID,
		Token:     raw,
		URL:       inviteURL,
		ExpiresAt: inv.ExpiresAt,
	}, nil
}

// GetInvitation resolves a raw token into public info for the landing page.
// It maps missing/expired/used tokens to the corresponding sentinel errors.
// Platform invitations (owner/customer) have no salon, so the salon fields are
// left empty for them.
func (s *InvitationService) GetInvitation(ctx context.Context, rawToken string) (*InvitationPublic, error) {
	inv, err := s.invitations.GetByTokenHash(ctx, hashToken(rawToken))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvitationNotFound
		}
		return nil, err
	}
	if err := validateInvitation(inv); err != nil {
		return nil, err
	}

	salonName, salonSlug := "", ""
	if inv.BusinessID != uuid.Nil {
		business, err := s.businessStore.GetByID(ctx, inv.BusinessID)
		if err != nil {
			return nil, fmt.Errorf("failed to load business: %w", err)
		}
		salonName, salonSlug = business.Name, business.Slug
	}

	return &InvitationPublic{
		SalonName: salonName,
		SalonSlug: salonSlug,
		Role:      inv.Role,
		ExpiresAt: inv.ExpiresAt,
	}, nil
}

// AcceptInvitation redeems a raw token for the calling user. Salon invites
// either link them to the salon as an employee (business_users row + Employee
// realm role + approval) or grant the Customer realm role; platform invites
// grant the Owner or Customer realm role (no business_users row). It is
// idempotent and safe to retry.
func (s *InvitationService) AcceptInvitation(ctx context.Context, rawToken, userID string) (*models.Business, error) {
	inv, err := s.invitations.GetByTokenHash(ctx, hashToken(rawToken))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvitationNotFound
		}
		return nil, err
	}
	if time.Now().UTC().After(inv.ExpiresAt) {
		return nil, ErrInvitationExpired
	}

	// Platform invitations have no business.
	if inv.BusinessID == uuid.Nil {
		return s.acceptPlatform(ctx, inv, userID)
	}

	business, err := s.businessStore.GetByID(ctx, inv.BusinessID)
	if err != nil {
		return nil, fmt.Errorf("failed to load business: %w", err)
	}

	existing, err := s.businessUsers.GetByBusinessAndUser(ctx, business.ID, userID)
	if err == nil && existing.Role == "admin" {
		return nil, ErrCannotInviteOwner
	}

	switch inv.Role {
	case inviteRoleEmployee:
		return s.acceptEmployee(ctx, inv, business, userID, existing)
	case inviteRoleCustomer:
		return s.acceptCustomer(ctx, inv, business, userID)
	default:
		return nil, ErrInvalidInviteRole
	}
}

// acceptPlatform redeems a platform invitation: it consumes the token
// (single-use) and grants the Owner or Customer realm role. Owners are also
// pre-approved so they can create their salon; there is no business to link.
func (s *InvitationService) acceptPlatform(ctx context.Context, inv *models.Invitation, userID string) (*models.Business, error) {
	err := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		ok, err := s.invitations.TryConsume(ctx, tx, inv.ID)
		if err != nil {
			return err
		}
		if !ok {
			return ErrInvitationUsed
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	switch inv.Role {
	case inviteRoleOwner:
		if err := s.grantOwner(ctx, userID); err != nil {
			return nil, err
		}
	case inviteRoleCustomer:
		if err := s.grantCustomer(ctx, userID); err != nil {
			return nil, err
		}
	case inviteRoleRealmAdmin:
		if err := s.grantRealmAdmin(ctx, userID); err != nil {
			return nil, err
		}
	default:
		return nil, ErrInvalidInviteRole
	}

	return nil, nil
}

// acceptEmployee redeems an employee invite: it consumes the token and creates
// the business_users row in one transaction, then grants the Employee realm
// role and approves the account. Idempotent for an already-linked employee.
func (s *InvitationService) acceptEmployee(ctx context.Context, inv *models.Invitation, business *models.Business, userID string, existing *models.BusinessUser) (*models.Business, error) {
	if existing != nil && existing.Role == "employee" {
		// Already linked: (re-)grant roles idempotently and succeed without
		// consuming the invite, so a retry after a partial failure works.
		if err := s.grantEmployee(ctx, userID); err != nil {
			return nil, err
		}
		return business, nil
	}

	displayName := ""
	if u, err := s.userStore.GetByID(ctx, userID); err == nil {
		displayName = u.DisplayName
	}

	err := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		ok, err := s.invitations.TryConsume(ctx, tx, inv.ID)
		if err != nil {
			return err
		}
		if !ok {
			return ErrInvitationUsed
		}

		bu := &models.BusinessUser{
			BusinessID:  business.ID,
			UserID:      userID,
			Role:        "employee",
			DisplayName: displayName,
			Active:      true,
		}
		return s.businessUsers.Create(ctx, tx, bu)
	})
	if err != nil {
		return nil, err
	}

	if err := s.grantEmployee(ctx, userID); err != nil {
		return nil, err
	}

	return business, nil
}

// acceptCustomer redeems a customer invite: it consumes the token (single-use)
// and grants the Customer realm role + approval. No business_users row is
// created because customers are referenced only by their Keycloak sub.
func (s *InvitationService) acceptCustomer(ctx context.Context, inv *models.Invitation, business *models.Business, userID string) (*models.Business, error) {
	err := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		ok, err := s.invitations.TryConsume(ctx, tx, inv.ID)
		if err != nil {
			return err
		}
		if !ok {
			return ErrInvitationUsed
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if err := s.grantCustomer(ctx, userID); err != nil {
		return nil, err
	}

	return business, nil
}

// grantEmployee assigns the Employee realm role and approves the account. It is
// idempotent so it can be re-run safely on retries.
func (s *InvitationService) grantEmployee(ctx context.Context, userID string) error {
	if err := s.users.AddRealmRole(ctx, userID, "Employee"); err != nil {
		return fmt.Errorf("failed to assign employee role: %w", err)
	}
	if err := s.users.UpdateUserAttributes(ctx, userID, map[string][]string{"approval_status": {"approved"}}); err != nil {
		return fmt.Errorf("failed to approve user: %w", err)
	}
	return nil
}

// grantCustomer assigns the Customer realm role and approves the account. QR
// invites are an owner/admin voucher, so invitees are pre-approved on Keycloak.
func (s *InvitationService) grantCustomer(ctx context.Context, userID string) error {
	if err := s.users.AddRealmRole(ctx, userID, "Customer"); err != nil {
		return fmt.Errorf("failed to assign customer role: %w", err)
	}
	if err := s.users.UpdateUserAttributes(ctx, userID, map[string][]string{"approval_status": {"approved"}}); err != nil {
		return fmt.Errorf("failed to approve user: %w", err)
	}
	return nil
}

// grantOwner assigns the Owner realm role and approves the account so the
// invited owner can create their salon. It is idempotent.
func (s *InvitationService) grantOwner(ctx context.Context, userID string) error {
	if err := s.users.AddRealmRole(ctx, userID, "Owner"); err != nil {
		return fmt.Errorf("failed to assign owner role: %w", err)
	}
	if err := s.users.UpdateUserAttributes(ctx, userID, map[string][]string{"approval_status": {"approved"}}); err != nil {
		return fmt.Errorf("failed to approve user: %w", err)
	}
	return nil
}

// grantRealmAdmin assigns the realm-management realm-admin client role and
// approves the account. It is idempotent.
func (s *InvitationService) grantRealmAdmin(ctx context.Context, userID string) error {
	if err := s.users.AddClientRole(ctx, userID, "realm-management", "realm-admin"); err != nil {
		return fmt.Errorf("failed to assign realm-admin role: %w", err)
	}
	if err := s.users.UpdateUserAttributes(ctx, userID, map[string][]string{"approval_status": {"approved"}}); err != nil {
		return fmt.Errorf("failed to approve user: %w", err)
	}
	return nil
}

func validateInvitation(inv *models.Invitation) error {
	if time.Now().UTC().After(inv.ExpiresAt) {
		return ErrInvitationExpired
	}
	if inv.UseCount >= inv.MaxUses {
		return ErrInvitationUsed
	}
	return nil
}

func (s *InvitationService) inviteURL(token string) (string, error) {
	if s.baseURL == "" {
		return "", ErrInviteBaseURLUnset
	}
	return strings.TrimRight(s.baseURL, "/") + "/invite/" + token, nil
}

// newInviteToken returns a 256-bit random URL-safe token and its SHA-256 hash.
func newInviteToken() (raw, hash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", fmt.Errorf("failed to generate invite token: %w", err)
	}
	raw = base64.RawURLEncoding.EncodeToString(b)
	return raw, hashToken(raw), nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
