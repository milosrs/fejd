package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordingClientRedirect struct {
	calls []clientRedirectCall
	err   error
}

type clientRedirectCall struct {
	clientID    string
	redirectURI string
	webOrigin   string
}

func (r *recordingClientRedirect) EnsureClientRedirect(_ context.Context, clientID, redirectURI, webOrigin string) error {
	r.calls = append(r.calls, clientRedirectCall{clientID: clientID, redirectURI: redirectURI, webOrigin: webOrigin})
	return r.err
}

func TestSalonDomainService_RegisterSalon(t *testing.T) {
	t.Run("no-op when app domain is empty", func(t *testing.T) {
		fake := &recordingClientRedirect{}
		s := NewSalonDomainService(fake, "")

		require.NoError(t, s.RegisterSalon(context.Background(), "my-salon"))
		assert.Empty(t, fake.calls)
	})

	t.Run("registers redirect URI and web origin", func(t *testing.T) {
		fake := &recordingClientRedirect{}
		s := NewSalonDomainService(fake, "fejd.fyi")

		require.NoError(t, s.RegisterSalon(context.Background(), "my-salon"))

		require.Len(t, fake.calls, 1)
		assert.Equal(t, webClientID, fake.calls[0].clientID)
		assert.Equal(t, "https://my-salon.fejd.fyi/*", fake.calls[0].redirectURI)
		assert.Equal(t, "https://my-salon.fejd.fyi", fake.calls[0].webOrigin)
	})

	t.Run("propagates keycloak error", func(t *testing.T) {
		fake := &recordingClientRedirect{err: errors.New("boom")}
		s := NewSalonDomainService(fake, "fejd.fyi")

		err := s.RegisterSalon(context.Background(), "my-salon")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "boom")
	})
}
