package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlatformRole(t *testing.T) {
	t.Run("customer", func(t *testing.T) {
		role, approved, err := platformRole("customer")
		require.NoError(t, err)
		assert.Equal(t, "Customer", role)
		assert.False(t, approved)
	})

	t.Run("owner", func(t *testing.T) {
		role, approved, err := platformRole("owner")
		require.NoError(t, err)
		assert.Equal(t, "Owner", role)
		assert.True(t, approved)
	})

	t.Run("invalid", func(t *testing.T) {
		role, approved, err := platformRole("employee")
		require.Error(t, err)
		assert.Empty(t, role)
		assert.False(t, approved)
	})

	t.Run("empty", func(t *testing.T) {
		role, approved, err := platformRole("")
		require.Error(t, err)
		assert.Empty(t, role)
		assert.False(t, approved)
	})
}
