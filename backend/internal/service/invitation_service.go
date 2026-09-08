package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"fejd-backend/internal/models"
	"fejd-backend/internal/store"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// InvitationService creates shareable link/QR invitations that link a user to a
// business as an employee. Only the token hash is persisted; the raw token is
// returned to the caller to embed in the invite URL.
type InvitationService struct {
	invitations *store.InvitationStore
	pool        *pgxpool.Pool
	baseURL     string
}

func NewInvitationService(invitations *store.InvitationStore, pool *pgxpool.Pool, baseURL string) *InvitationService {
	return &InvitationService{invitations: invitations, pool: pool, baseURL: baseURL}
}

// InvitationOut is the shareable result of creating an invitation.
type InvitationOut struct {
	ID        uuid.UUID
	Token     string
	URL       string
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
	sum := sha256.Sum256([]byte(raw))
	return raw, hex.EncodeToString(sum[:]), nil
}
