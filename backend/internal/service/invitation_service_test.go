package service

import (
	"testing"
	"time"

	"fejd-backend/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashToken(t *testing.T) {
	hash := hashToken("abc123")

	require.Len(t, hash, 64, "sha256 hex is 64 chars")
	assert.Equal(t, hash, hashToken("abc123"), "hashing must be deterministic")
	assert.NotEqual(t, hash, hashToken("abc124"))
}

func TestNewInviteToken(t *testing.T) {
	raw, hash, err := newInviteToken()

	require.NoError(t, err)
	assert.NotEmpty(t, raw)
	assert.Equal(t, hashToken(raw), hash)
	assert.NotEqual(t, raw, hash)
}

func TestValidateInvitation(t *testing.T) {
	now := time.Now().UTC()

	t.Run("expired", func(t *testing.T) {
		inv := &models.Invitation{ExpiresAt: now.Add(-time.Minute), MaxUses: 1, UseCount: 0}
		assert.ErrorIs(t, validateInvitation(inv), ErrInvitationExpired)
	})

	t.Run("used", func(t *testing.T) {
		inv := &models.Invitation{ExpiresAt: now.Add(time.Hour), MaxUses: 1, UseCount: 1}
		assert.ErrorIs(t, validateInvitation(inv), ErrInvitationUsed)
	})

	t.Run("valid", func(t *testing.T) {
		inv := &models.Invitation{ExpiresAt: now.Add(time.Hour), MaxUses: 1, UseCount: 0}
		assert.NoError(t, validateInvitation(inv))
	})
}
