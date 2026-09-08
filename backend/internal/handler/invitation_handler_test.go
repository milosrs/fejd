package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"fejd-backend/internal/keycloak"
	"fejd-backend/internal/models"
	"fejd-backend/internal/service"
	"fejd-backend/internal/store"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestInvitationHandler(t *testing.T) (*InvitationHandler, *store.BusinessStore, *store.BusinessUserStore, *pgxpool.Pool) {
	pool := setupHandlerTestDB(t)
	businessStore := store.NewBusinessStore(pool)
	buStore := store.NewBusinessUserStore(pool)
	invitationStore := store.NewInvitationStore(pool)
	invitationService := service.NewInvitationService(invitationStore, businessStore, buStore, &noopInvitationUsers{}, pool, "https://app.example.com")
	h := NewInvitationHandler(invitationService, 48*time.Hour)
	return h, businessStore, buStore, pool
}

// noopInvitationUsers satisfies service.InvitationUserManager for handler
// tests that only exercise invitation creation.
type noopInvitationUsers struct{}

func (noopInvitationUsers) GetUser(context.Context, string) (*keycloak.User, error) {
	return &keycloak.User{}, nil
}
func (noopInvitationUsers) AddRealmRole(context.Context, string, string) error { return nil }
func (noopInvitationUsers) UpdateUserAttributes(context.Context, string, map[string][]string) error {
	return nil
}

func TestInvitationHandler_CreateInvitation_AsEmployee(t *testing.T) {
	h, businessStore, buStore, pool := newTestInvitationHandler(t)
	ctx := context.Background()

	b := &models.Business{Name: "Salon", Slug: "salon"}
	require.NoError(t, businessStore.Create(ctx, pool, b))
	// user-2 is the admin; user-1 is the employee member generating the invite.
	require.NoError(t, buStore.Create(ctx, pool, &models.BusinessUser{
		BusinessID: b.ID, UserID: "user-2", Role: "admin",
	}))
	require.NoError(t, buStore.Create(ctx, pool, &models.BusinessUser{
		BusinessID: b.ID, UserID: "user-1", Role: "employee",
	}))

	req := withPathParams(
		withUser(httptest.NewRequest(http.MethodPost, "/api/admin/business/"+b.ID.String()+"/invitations", bytes.NewBufferString(`{}`)), "user-1", "approved"),
		map[string]string{"businessID": b.ID.String()},
	)
	rr := httptest.NewRecorder()

	h.CreateInvitation(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
	var resp InvitationResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp.Token)
	assert.NotEmpty(t, resp.URL)
	assert.Contains(t, resp.URL, "/invite/"+resp.Token)
	assert.False(t, resp.ExpiresAt.IsZero())
	assert.NotEqual(t, "", resp.ID.String())
}

func TestInvitationHandler_CreateInvitation_CustomExpiry(t *testing.T) {
	h, businessStore, buStore, pool := newTestInvitationHandler(t)
	ctx := context.Background()

	b := &models.Business{Name: "Salon", Slug: "salon"}
	require.NoError(t, businessStore.Create(ctx, pool, b))
	require.NoError(t, buStore.Create(ctx, pool, &models.BusinessUser{
		BusinessID: b.ID, UserID: "user-1", Role: "admin",
	}))

	req := withPathParams(
		withUser(httptest.NewRequest(http.MethodPost, "/api/admin/business/"+b.ID.String()+"/invitations", bytes.NewBufferString(`{"expires_in_hours":72}`)), "user-1", "approved"),
		map[string]string{"businessID": b.ID.String()},
	)
	rr := httptest.NewRecorder()

	h.CreateInvitation(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
	var resp InvitationResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.WithinDuration(t, time.Now().UTC().Add(72*time.Hour), resp.ExpiresAt, time.Minute)
}

func TestInvitationHandler_CreateInvitation_InvalidBusinessID(t *testing.T) {
	h, _, _, _ := newTestInvitationHandler(t)

	req := withUser(httptest.NewRequest(http.MethodPost, "/api/admin/business/not-a-uuid/invitations", bytes.NewBufferString(`{}`)), "user-1", "approved")
	rr := httptest.NewRecorder()

	h.CreateInvitation(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}
