package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordingRoleManager struct {
	roles []string
	attrs []map[string][]string
}

func (r *recordingRoleManager) AddRealmRole(ctx context.Context, userID, role string) error {
	r.roles = append(r.roles, role)
	return nil
}

func (r *recordingRoleManager) UpdateUserAttributes(ctx context.Context, userID string, attrs map[string][]string) error {
	r.attrs = append(r.attrs, attrs)
	return nil
}

func TestRegistrationService_ClaimRegistrationRole(t *testing.T) {
	t.Run("Customer is approved automatically", func(t *testing.T) {
		m := &recordingRoleManager{}
		s := NewRegistrationService(m)

		require.NoError(t, s.ClaimRegistrationRole(context.Background(), "user-1", "Customer"))

		assert.Equal(t, []string{"Customer"}, m.roles)
		require.Len(t, m.attrs, 1)
		assert.Equal(t, []string{"approved"}, m.attrs[0]["approval_status"])
		_, cleared := m.attrs[0]["registration_role"]
		assert.True(t, cleared)
	})

	t.Run("Employee is approved automatically", func(t *testing.T) {
		m := &recordingRoleManager{}
		s := NewRegistrationService(m)

		require.NoError(t, s.ClaimRegistrationRole(context.Background(), "user-1", "Employee"))

		assert.Equal(t, []string{"Employee"}, m.roles)
		require.Len(t, m.attrs, 1)
		assert.Equal(t, []string{"approved"}, m.attrs[0]["approval_status"])
		_, cleared := m.attrs[0]["registration_role"]
		assert.True(t, cleared)
	})

	t.Run("Owner stays pending for admin verification", func(t *testing.T) {
		m := &recordingRoleManager{}
		s := NewRegistrationService(m)

		require.NoError(t, s.ClaimRegistrationRole(context.Background(), "user-1", "Owner"))

		assert.Equal(t, []string{"Owner"}, m.roles)
		require.Len(t, m.attrs, 1)
		assert.NotContains(t, m.attrs[0], "approval_status")
		_, cleared := m.attrs[0]["registration_role"]
		assert.True(t, cleared)
	})
}

func TestRegistrationService_ClaimRegistrationRole_RejectsInvalid(t *testing.T) {
	m := &recordingRoleManager{}
	s := NewRegistrationService(m)

	err := s.ClaimRegistrationRole(context.Background(), "user-1", "admin")

	require.Error(t, err)
	assert.Empty(t, m.roles)
	assert.Empty(t, m.attrs)
}

func TestRegistrationService_ClaimRegistrationRole_PropagatesGrantError(t *testing.T) {
	m := &stubRoleManager{grantErr: errors.New("boom")}
	s := NewRegistrationService(m)

	err := s.ClaimRegistrationRole(context.Background(), "user-1", "Customer")

	require.Error(t, err)
	assert.Empty(t, m.attrs)
}

type stubRoleManager struct {
	grantErr error
	attrs    []map[string][]string
}

func (s *stubRoleManager) AddRealmRole(ctx context.Context, userID, role string) error {
	return s.grantErr
}

func (s *stubRoleManager) UpdateUserAttributes(ctx context.Context, userID string, attrs map[string][]string) error {
	s.attrs = append(s.attrs, attrs)
	return nil
}
