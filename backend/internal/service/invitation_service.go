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
	ErrInvitationNotFound = errors.New("invitation not found")
	ErrInvitationExpired  = errors.New("invitation expired")
	ErrInvitationUsed     = errors.New("invitation already used")
	ErrCannotInviteOwner  = errors.New("cannot invite the salon owner")
)

// InvitationUserManager is the Keycloak admin surface the invitation service
// needs for role/approval grants only — never for reading user attributes.
// *keycloak.Client satisfies it; tests provide a fake.
type InvitationUserManager interface {
	AddRealmRole(ctx context.Context, userID, role string) error
	UpdateUserAttributes(ctx context.Context, userID string, attrs map[string][]string) error
}

// InvitationService creates shareable link/QR invitations that link a user to a
// business as an employee, and redeems them. Only the token hash is persisted;
// the raw token is returned to the caller to embed in the invite URL.
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
// returns the raw token + invite URL. ttl<=0 falls back to the caller's
// default (callers always pass an explicit ttl).
func (s *InvitationService) CreateInvitation(ctx context.Context, businessID uuid.UUID, createdBy string, ttl time.Duration) (*InvitationOut, error) {
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
		Role:       "employee",
		CreatedBy:  createdBy,
		MaxUses:    1,
		ExpiresAt:  time.Now().UTC().Add(ttl),
	}
	if err := s.invitations.Create(ctx, s.pool, inv); err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	return &InvitationOut{
		ID:        inv.ID,
		Token:     raw,
		URL:       s.inviteURL(raw),
		ExpiresAt: inv.ExpiresAt,
	}, nil
}

// GetInvitation resolves a raw token into public info for the landing page.
// It maps missing/expired/used tokens to the corresponding sentinel errors.
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

	business, err := s.businessStore.GetByID(ctx, inv.BusinessID)
	if err != nil {
		return nil, fmt.Errorf("failed to load business: %w", err)
	}

	return &InvitationPublic{
		SalonName: business.Name,
		SalonSlug: business.Slug,
		Role:      inv.Role,
		ExpiresAt: inv.ExpiresAt,
	}, nil
}

// AcceptInvitation redeems a raw token for the calling user: it links them to
// the salon as an employee and grants the Employee Keycloak realm role. It is
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

	business, err := s.businessStore.GetByID(ctx, inv.BusinessID)
	if err != nil {
		return nil, fmt.Errorf("failed to load business: %w", err)
	}

	existing, err := s.businessUsers.GetByBusinessAndUser(ctx, business.ID, userID)
	if err == nil {
		if existing.Role == "admin" {
			return nil, ErrCannotInviteOwner
		}
		if existing.Role == "employee" {
			// Already linked: (re-)grant roles idempotently and succeed without
			// consuming the invite, so a retry after a partial failure works.
			if err := s.grantEmployee(ctx, userID); err != nil {
				return nil, err
			}
			return business, nil
		}
	}

	displayName := ""
	if u, err := s.userStore.GetByID(ctx, userID); err == nil {
		displayName = u.DisplayName
	}

	err = db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
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

func validateInvitation(inv *models.Invitation) error {
	if time.Now().UTC().After(inv.ExpiresAt) {
		return ErrInvitationExpired
	}
	if inv.UseCount >= inv.MaxUses {
		return ErrInvitationUsed
	}
	return nil
}

func (s *InvitationService) inviteURL(token string) string {
	if s.baseURL == "" {
		return "/invite/" + token
	}
	return strings.TrimRight(s.baseURL, "/") + "/invite/" + token
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
