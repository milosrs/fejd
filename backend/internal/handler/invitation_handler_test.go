package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"fejd-backend/internal/dto"
	"fejd-backend/internal/models"
	"fejd-backend/internal/service"
	"fejd-backend/internal/store"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestInvitationHandlerWith(t *testing.T, users service.InvitationUserManager) (*InvitationHandler, *store.BusinessStore, *store.BusinessUserStore, *pgxpool.Pool) {
	pool := setupHandlerTestDB(t)
	businessStore := store.NewBusinessStore(pool)
	buStore := store.NewBusinessUserStore(pool)
	userStore := store.NewUserStore(pool)
	invitationStore := store.NewInvitationStore(pool)
	invitationService := service.NewInvitationService(invitationStore, businessStore, buStore, userStore, users, pool, "https://app.example.com")
	h := NewInvitationHandler(invitationService, 48*time.Hour)
	return h, businessStore, buStore, pool
}

func newTestInvitationHandler(t *testing.T) (*InvitationHandler, *store.BusinessStore, *store.BusinessUserStore, *pgxpool.Pool) {
	return newTestInvitationHandlerWith(t, &noopInvitationUsers{})
}

// noopInvitationUsers satisfies service.InvitationUserManager for tests that
// only exercise invitation creation or public resolution.
type noopInvitationUsers struct{}

func (noopInvitationUsers) AddRealmRole(context.Context, string, string) error { return nil }
func (noopInvitationUsers) AddClientRole(context.Context, string, string, string) error {
	return nil
}
func (noopInvitationUsers) UpdateUserAttributes(context.Context, string, map[string][]string) error {
	return nil
}

// recordingUsers records role grants and attribute updates for assertions.
type recordingUsers struct {
	roles       []string
	clientRoles []clientRoleGrant
	attrs       []map[string][]string
}

type clientRoleGrant struct {
	clientID string
	role     string
}

func (r *recordingUsers) AddRealmRole(ctx context.Context, userID, role string) error {
	r.roles = append(r.roles, role)
	return nil
}
func (r *recordingUsers) AddClientRole(ctx context.Context, userID, clientID, role string) error {
	r.clientRoles = append(r.clientRoles, clientRoleGrant{clientID: clientID, role: role})
	return nil
}
func (r *recordingUsers) UpdateUserAttributes(ctx context.Context, userID string, attrs map[string][]string) error {
	r.attrs = append(r.attrs, attrs)
	return nil
}

func createInvite(t *testing.T, h *InvitationHandler, businessID uuid.UUID, createdBy string) string {
	t.Helper()
	out, err := h.invitations.CreateInvitation(context.Background(), businessID, createdBy, "employee", time.Hour)
	require.NoError(t, err)
	return out.Token
}

func createCustomerInvite(t *testing.T, h *InvitationHandler, businessID uuid.UUID, createdBy string) string {
	t.Helper()
	out, err := h.invitations.CreateInvitation(context.Background(), businessID, createdBy, "customer", time.Hour)
	require.NoError(t, err)
	return out.Token
}

func TestInvitationHandler_CreateInvitation_AsEmployee_CustomerOnly(t *testing.T) {
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
		withUser(httptest.NewRequest(http.MethodPost, "/api/admin/business/"+b.ID.String()+"/invitations", bytes.NewBufferString(`{"role":"customer"}`)), "user-1", "approved"),
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

func TestInvitationHandler_CreateInvitation_EmployeeCannotInviteEmployee(t *testing.T) {
	h, businessStore, buStore, pool := newTestInvitationHandler(t)
	ctx := context.Background()

	b := &models.Business{Name: "Salon", Slug: "salon"}
	require.NoError(t, businessStore.Create(ctx, pool, b))
	require.NoError(t, buStore.Create(ctx, pool, &models.BusinessUser{
		BusinessID: b.ID, UserID: "user-2", Role: "admin",
	}))
	require.NoError(t, buStore.Create(ctx, pool, &models.BusinessUser{
		BusinessID: b.ID, UserID: "user-1", Role: "employee",
	}))

	req := withPathParams(
		withUser(httptest.NewRequest(http.MethodPost, "/api/admin/business/"+b.ID.String()+"/invitations", bytes.NewBufferString(`{"role":"employee"}`)), "user-1", "approved"),
		map[string]string{"businessID": b.ID.String()},
	)
	rr := httptest.NewRecorder()

	h.CreateInvitation(rr, req)

	require.Equal(t, http.StatusForbidden, rr.Code)
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

func TestInvitationHandler_GetInvitation_Public(t *testing.T) {
	h, businessStore, buStore, pool := newTestInvitationHandler(t)
	ctx := context.Background()

	b := &models.Business{Name: "Salon", Slug: "salon"}
	require.NoError(t, businessStore.Create(ctx, pool, b))
	require.NoError(t, buStore.Create(ctx, pool, &models.BusinessUser{
		BusinessID: b.ID, UserID: "owner", Role: "admin",
	}))
	token := createInvite(t, h, b.ID, "owner")

	req := withPathParams(
		httptest.NewRequest(http.MethodGet, "/api/invitations/"+token, nil),
		map[string]string{"token": token},
	)
	rr := httptest.NewRecorder()

	h.GetInvitation(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var resp PublicInvitationResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "Salon", resp.SalonName)
	assert.Equal(t, "salon", resp.SalonSlug)
	assert.Equal(t, "employee", resp.Role)
}

func TestInvitationHandler_GetInvitation_NotFound(t *testing.T) {
	h, _, _, _ := newTestInvitationHandler(t)

	req := withPathParams(
		httptest.NewRequest(http.MethodGet, "/api/invitations/does-not-exist", nil),
		map[string]string{"token": "does-not-exist"},
	)
	rr := httptest.NewRecorder()

	h.GetInvitation(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)
}

func TestInvitationHandler_AcceptInvitation_LinksEmployee(t *testing.T) {
	users := &recordingUsers{}
	h, businessStore, buStore, pool := newTestInvitationHandlerWith(t, users)
	ctx := context.Background()

	b := &models.Business{Name: "Salon", Slug: "salon"}
	require.NoError(t, businessStore.Create(ctx, pool, b))
	require.NoError(t, buStore.Create(ctx, pool, &models.BusinessUser{
		BusinessID: b.ID, UserID: "owner", Role: "admin",
	}))

	token := createInvite(t, h, b.ID, "owner")

	req := withPathParams(
		withUser(httptest.NewRequest(http.MethodPost, "/api/invitations/"+token+"/accept", nil), "emp-1", "pending"),
		map[string]string{"token": token},
	)
	rr := httptest.NewRecorder()

	h.AcceptInvitation(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var business dto.Business
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &business))
	assert.Equal(t, "salon", business.Slug)

	bu, err := buStore.GetByBusinessAndUser(ctx, b.ID, "emp-1")
	require.NoError(t, err)
	assert.Equal(t, "employee", bu.Role)
	assert.True(t, bu.Active)

	require.Len(t, users.roles, 1)
	assert.Equal(t, "Employee", users.roles[0])
	require.Len(t, users.attrs, 1)
	assert.Equal(t, []string{"approved"}, users.attrs[0]["approval_status"])
}

func TestInvitationHandler_AcceptInvitation_LinksCustomer(t *testing.T) {
	users := &recordingUsers{}
	h, businessStore, buStore, pool := newTestInvitationHandlerWith(t, users)
	ctx := context.Background()

	b := &models.Business{Name: "Salon", Slug: "salon"}
	require.NoError(t, businessStore.Create(ctx, pool, b))
	require.NoError(t, buStore.Create(ctx, pool, &models.BusinessUser{
		BusinessID: b.ID, UserID: "owner", Role: "admin",
	}))

	token := createCustomerInvite(t, h, b.ID, "owner")

	req := withPathParams(
		withUser(httptest.NewRequest(http.MethodPost, "/api/invitations/"+token+"/accept", nil), "cust-1", "pending"),
		map[string]string{"token": token},
	)
	rr := httptest.NewRecorder()

	h.AcceptInvitation(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var business dto.Business
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &business))
	assert.Equal(t, "salon", business.Slug)

	_, err := buStore.GetByBusinessAndUser(ctx, b.ID, "cust-1")
	require.Error(t, err, "customer must not get a business_users row")

	require.Len(t, users.roles, 1)
	assert.Equal(t, "Customer", users.roles[0])
	require.Len(t, users.attrs, 1)
	assert.Equal(t, []string{"approved"}, users.attrs[0]["approval_status"])
}

func TestInvitationHandler_CreateInvitation_WithCustomerRole(t *testing.T) {
	h, businessStore, buStore, pool := newTestInvitationHandler(t)
	ctx := context.Background()

	b := &models.Business{Name: "Salon", Slug: "salon"}
	require.NoError(t, businessStore.Create(ctx, pool, b))
	require.NoError(t, buStore.Create(ctx, pool, &models.BusinessUser{
		BusinessID: b.ID, UserID: "user-1", Role: "admin",
	}))

	req := withPathParams(
		withUser(httptest.NewRequest(http.MethodPost, "/api/admin/business/"+b.ID.String()+"/invitations", bytes.NewBufferString(`{"role":"customer"}`)), "user-1", "approved"),
		map[string]string{"businessID": b.ID.String()},
	)
	rr := httptest.NewRecorder()

	h.CreateInvitation(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
	var resp InvitationResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

	getReq := withPathParams(
		httptest.NewRequest(http.MethodGet, "/api/invitations/"+resp.Token, nil),
		map[string]string{"token": resp.Token},
	)
	grr := httptest.NewRecorder()
	h.GetInvitation(grr, getReq)

	require.Equal(t, http.StatusOK, grr.Code)
	var pub PublicInvitationResponse
	require.NoError(t, json.Unmarshal(grr.Body.Bytes(), &pub))
	assert.Equal(t, "customer", pub.Role)
}

func TestInvitationHandler_CreateInvitation_InvalidRole(t *testing.T) {
	h, businessStore, buStore, pool := newTestInvitationHandler(t)
	ctx := context.Background()

	b := &models.Business{Name: "Salon", Slug: "salon"}
	require.NoError(t, businessStore.Create(ctx, pool, b))
	require.NoError(t, buStore.Create(ctx, pool, &models.BusinessUser{
		BusinessID: b.ID, UserID: "user-1", Role: "admin",
	}))

	req := withPathParams(
		withUser(httptest.NewRequest(http.MethodPost, "/api/admin/business/"+b.ID.String()+"/invitations", bytes.NewBufferString(`{"role":"owner"}`)), "user-1", "approved"),
		map[string]string{"businessID": b.ID.String()},
	)
	rr := httptest.NewRecorder()

	h.CreateInvitation(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestInvitationHandler_AcceptInvitation_SingleUse(t *testing.T) {
	h, businessStore, buStore, pool := newTestInvitationHandler(t)
	ctx := context.Background()

	b := &models.Business{Name: "Salon", Slug: "salon"}
	require.NoError(t, businessStore.Create(ctx, pool, b))
	require.NoError(t, buStore.Create(ctx, pool, &models.BusinessUser{
		BusinessID: b.ID, UserID: "owner", Role: "admin",
	}))

	token := createInvite(t, h, b.ID, "owner")

	accept := func(userID string) *httptest.ResponseRecorder {
		req := withPathParams(
			withUser(httptest.NewRequest(http.MethodPost, "/api/invitations/"+token+"/accept", nil), userID, "approved"),
			map[string]string{"token": token},
		)
		rr := httptest.NewRecorder()
		h.AcceptInvitation(rr, req)
		return rr
	}

	require.Equal(t, http.StatusOK, accept("emp-1").Code)
	require.Equal(t, http.StatusGone, accept("emp-2").Code)

	_, err := buStore.GetByBusinessAndUser(ctx, b.ID, "emp-2")
	require.Error(t, err, "second user must not be linked")
}

func TestInvitationHandler_AcceptInvitation_OwnerForbidden(t *testing.T) {
	h, businessStore, buStore, pool := newTestInvitationHandler(t)
	ctx := context.Background()

	b := &models.Business{Name: "Salon", Slug: "salon"}
	require.NoError(t, businessStore.Create(ctx, pool, b))
	require.NoError(t, buStore.Create(ctx, pool, &models.BusinessUser{
		BusinessID: b.ID, UserID: "owner", Role: "admin",
	}))

	token := createInvite(t, h, b.ID, "owner")

	req := withPathParams(
		withUser(httptest.NewRequest(http.MethodPost, "/api/invitations/"+token+"/accept", nil), "owner", "approved"),
		map[string]string{"token": token},
	)
	rr := httptest.NewRecorder()

	h.AcceptInvitation(rr, req)

	require.Equal(t, http.StatusConflict, rr.Code)
}

func TestInvitationHandler_CreatePlatformInvitation_Owner(t *testing.T) {
	h, _, _, _ := newTestInvitationHandler(t)

	req := withUser(
		httptest.NewRequest(http.MethodPost, "/api/admin/invitations", bytes.NewBufferString(`{"role":"owner"}`)),
		"admin-1", "approved",
	)
	rr := httptest.NewRecorder()

	h.CreatePlatformInvitation(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
	var resp InvitationResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp.Token)
	assert.Contains(t, resp.URL, "/invite/"+resp.Token)

	getReq := withPathParams(
		httptest.NewRequest(http.MethodGet, "/api/invitations/"+resp.Token, nil),
		map[string]string{"token": resp.Token},
	)
	grr := httptest.NewRecorder()
	h.GetInvitation(grr, getReq)

	require.Equal(t, http.StatusOK, grr.Code)
	var pub PublicInvitationResponse
	require.NoError(t, json.Unmarshal(grr.Body.Bytes(), &pub))
	assert.Equal(t, "owner", pub.Role)
	assert.Empty(t, pub.SalonName, "platform invites have no salon")
	assert.Empty(t, pub.SalonSlug)
}

func TestInvitationHandler_CreatePlatformInvitation_InvalidRole(t *testing.T) {
	h, _, _, _ := newTestInvitationHandler(t)

	req := withUser(
		httptest.NewRequest(http.MethodPost, "/api/admin/invitations", bytes.NewBufferString(`{"role":"employee"}`)),
		"admin-1", "approved",
	)
	rr := httptest.NewRecorder()

	h.CreatePlatformInvitation(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestInvitationHandler_CreateCustomerInvitation(t *testing.T) {
	h, _, _, _ := newTestInvitationHandler(t)

	req := withUser(
		httptest.NewRequest(http.MethodPost, "/api/invitations", bytes.NewBufferString(`{}`)),
		"cust-1", "approved",
	)
	rr := httptest.NewRecorder()

	h.CreateCustomerInvitation(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
	var resp InvitationResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp.Token)
	assert.Contains(t, resp.URL, "/invite/"+resp.Token)

	getReq := withPathParams(
		httptest.NewRequest(http.MethodGet, "/api/invitations/"+resp.Token, nil),
		map[string]string{"token": resp.Token},
	)
	grr := httptest.NewRecorder()
	h.GetInvitation(grr, getReq)

	require.Equal(t, http.StatusOK, grr.Code)
	var pub PublicInvitationResponse
	require.NoError(t, json.Unmarshal(grr.Body.Bytes(), &pub))
	assert.Equal(t, "customer", pub.Role)
	assert.Empty(t, pub.SalonName, "customer invites have no salon")
}

func TestInvitationHandler_AcceptInvitation_PlatformOwner(t *testing.T) {
	users := &recordingUsers{}
	h, _, buStore, _ := newTestInvitationHandlerWith(t, users)
	ctx := context.Background()

	out, err := h.invitations.CreatePlatformInvitation(ctx, "admin-1", "owner", time.Hour)
	require.NoError(t, err)

	req := withPathParams(
		withUser(httptest.NewRequest(http.MethodPost, "/api/invitations/"+out.Token+"/accept", nil), "owner-1", "pending"),
		map[string]string{"token": out.Token},
	)
	rr := httptest.NewRecorder()

	h.AcceptInvitation(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Empty(t, rr.Body.Bytes(), "platform accept returns no business body")

	require.Len(t, users.roles, 1)
	assert.Equal(t, "Owner", users.roles[0])
	require.Len(t, users.attrs, 1)
	assert.Equal(t, []string{"approved"}, users.attrs[0]["approval_status"])

	// No business_users row for a platform owner (they create their salon later).
	_, err = buStore.GetByBusinessAndUser(ctx, uuid.Nil, "owner-1")
	require.Error(t, err)
}

func TestInvitationHandler_AcceptInvitation_PlatformCustomer(t *testing.T) {
	users := &recordingUsers{}
	h, _, _, _ := newTestInvitationHandlerWith(t, users)
	ctx := context.Background()

	out, err := h.invitations.CreatePlatformInvitation(ctx, "admin-1", "customer", time.Hour)
	require.NoError(t, err)

	req := withPathParams(
		withUser(httptest.NewRequest(http.MethodPost, "/api/invitations/"+out.Token+"/accept", nil), "cust-1", "pending"),
		map[string]string{"token": out.Token},
	)
	rr := httptest.NewRecorder()

	h.AcceptInvitation(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Empty(t, rr.Body.Bytes())

	require.Len(t, users.roles, 1)
	assert.Equal(t, "Customer", users.roles[0])
	require.Len(t, users.attrs, 1)
	assert.Equal(t, []string{"approved"}, users.attrs[0]["approval_status"])
}

func TestInvitationHandler_AcceptInvitation_PlatformRealmAdmin(t *testing.T) {
	users := &recordingUsers{}
	h, _, _, _ := newTestInvitationHandlerWith(t, users)
	ctx := context.Background()

	out, err := h.invitations.CreatePlatformInvitation(ctx, "admin-1", "realm-admin", time.Hour)
	require.NoError(t, err)

	req := withPathParams(
		withUser(httptest.NewRequest(http.MethodPost, "/api/invitations/"+out.Token+"/accept", nil), "admin-2", "pending"),
		map[string]string{"token": out.Token},
	)
	rr := httptest.NewRecorder()

	h.AcceptInvitation(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Empty(t, rr.Body.Bytes())

	require.Empty(t, users.roles, "realm-admin is a client role, not a realm role")
	require.Len(t, users.clientRoles, 1)
	assert.Equal(t, "realm-management", users.clientRoles[0].clientID)
	assert.Equal(t, "realm-admin", users.clientRoles[0].role)
	require.Len(t, users.attrs, 1)
	assert.Equal(t, []string{"approved"}, users.attrs[0]["approval_status"])
}
