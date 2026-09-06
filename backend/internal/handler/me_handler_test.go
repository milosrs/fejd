package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fejd-backend/auth"
	"fejd-backend/internal/dto"
	"fejd-backend/internal/models"
	"fejd-backend/internal/store"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestMeHandler(t *testing.T) (*MeHandler, *store.BusinessStore, *store.BusinessUserStore, *pgxpool.Pool) {
	pool := setupHandlerTestDB(t)
	businessStore := store.NewBusinessStore(pool)
	buStore := store.NewBusinessUserStore(pool)
	return NewMeHandler(businessStore, buStore, pool), businessStore, buStore, pool
}

func withUser(r *http.Request, userID, approvalStatus string) *http.Request {
	ctx := r.Context()
	ctx = context.WithValue(ctx, auth.ContextKeyUserID, userID)
	ctx = context.WithValue(ctx, auth.ContextKeyApprovalStatus, approvalStatus)
	ctx = context.WithValue(ctx, auth.ContextKeyClaims, &auth.Claims{ApprovalStatus: approvalStatus})
	return r.WithContext(ctx)
}

func TestMeHandler_GetMe_NoSalon(t *testing.T) {
	h, _, _, _ := newTestMeHandler(t)

	req := withUser(httptest.NewRequest(http.MethodGet, "/api/me", nil), "user-1", "pending")
	rr := httptest.NewRecorder()

	h.GetMe(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var me dto.Me
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &me))
	assert.Equal(t, "pending", me.ApprovalStatus)
	assert.False(t, me.HasSalon)
	assert.Empty(t, me.Businesses)
}

func TestMeHandler_GetMe_HasSalon(t *testing.T) {
	h, businessStore, buStore, pool := newTestMeHandler(t)
	ctx := context.Background()

	b := &models.Business{Name: "My Salon", Slug: "my-salon"}
	require.NoError(t, businessStore.Create(ctx, pool, b))
	require.NoError(t, buStore.Create(ctx, pool, &models.BusinessUser{
		BusinessID: b.ID,
		UserID:     "user-1",
		Role:       "admin",
	}))

	req := withUser(httptest.NewRequest(http.MethodGet, "/api/me", nil), "user-1", "approved")
	rr := httptest.NewRecorder()

	h.GetMe(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var me dto.Me
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &me))
	assert.Equal(t, "approved", me.ApprovalStatus)
	assert.True(t, me.HasSalon)
	require.Len(t, me.Businesses, 1)
	assert.Equal(t, "my-salon", me.Businesses[0].Slug)
	assert.Equal(t, "admin", me.Businesses[0].Role)
}

func TestMeHandler_GetMe_EmployeeMembership(t *testing.T) {
	h, businessStore, buStore, pool := newTestMeHandler(t)
	ctx := context.Background()

	b := &models.Business{Name: "My Salon", Slug: "my-salon"}
	require.NoError(t, businessStore.Create(ctx, pool, b))
	// user-2 is the admin; user-1 is only an employee.
	require.NoError(t, buStore.Create(ctx, pool, &models.BusinessUser{
		BusinessID: b.ID,
		UserID:     "user-2",
		Role:       "admin",
	}))
	require.NoError(t, buStore.Create(ctx, pool, &models.BusinessUser{
		BusinessID: b.ID,
		UserID:     "user-1",
		Role:       "employee",
	}))

	req := withUser(httptest.NewRequest(http.MethodGet, "/api/me", nil), "user-1", "approved")
	rr := httptest.NewRecorder()

	h.GetMe(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var me dto.Me
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &me))
	assert.False(t, me.HasSalon)
	require.Len(t, me.Businesses, 1)
	assert.Equal(t, "my-salon", me.Businesses[0].Slug)
	assert.Equal(t, "employee", me.Businesses[0].Role)
}

func TestMeHandler_CreateBusiness_SlugCollision(t *testing.T) {
	h, businessStore, _, pool := newTestMeHandler(t)
	ctx := context.Background()

	require.NoError(t, businessStore.Create(ctx, pool, &models.Business{Name: "Existing", Slug: "my-salon"}))

	body := `{"name":"My Salon"}`
	req := withUser(httptest.NewRequest(http.MethodPost, "/api/me/business", bytes.NewBufferString(body)), "user-1", "approved")
	rr := httptest.NewRecorder()

	h.CreateBusiness(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
	var b dto.Business
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &b))
	assert.Equal(t, "My Salon", b.Name)
	assert.Equal(t, "my-salon-2", b.Slug)
}

func TestMeHandler_CreateBusiness_DoubleCreate(t *testing.T) {
	h, businessStore, buStore, pool := newTestMeHandler(t)
	ctx := context.Background()

	b := &models.Business{Name: "Existing", Slug: "existing"}
	require.NoError(t, businessStore.Create(ctx, pool, b))
	require.NoError(t, buStore.Create(ctx, pool, &models.BusinessUser{
		BusinessID: b.ID,
		UserID:     "user-1",
		Role:       "admin",
	}))

	body := `{"name":"New Salon"}`
	req := withUser(httptest.NewRequest(http.MethodPost, "/api/me/business", bytes.NewBufferString(body)), "user-1", "approved")
	rr := httptest.NewRecorder()

	h.CreateBusiness(rr, req)

	require.Equal(t, http.StatusConflict, rr.Code)
}

func TestMeHandler_TwoTierAuth(t *testing.T) {
	h, _, _, _ := newTestMeHandler(t)

	// A pending user can read their own status via GET /api/me.
	req := withUser(httptest.NewRequest(http.MethodGet, "/api/me", nil), "user-1", "pending")
	rr := httptest.NewRecorder()
	h.GetMe(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	// POST /api/me/business is guarded by RequireApproved: pending is rejected.
	m := &auth.Middleware{}
	guarded := m.RequireApproved(http.HandlerFunc(h.CreateBusiness))

	req2 := withUser(httptest.NewRequest(http.MethodPost, "/api/me/business", bytes.NewBufferString(`{"name":"X"}`)), "user-1", "pending")
	rr2 := httptest.NewRecorder()
	guarded.ServeHTTP(rr2, req2)
	assert.Equal(t, http.StatusForbidden, rr2.Code)

	// Approved passes through to the handler.
	req3 := withUser(httptest.NewRequest(http.MethodPost, "/api/me/business", bytes.NewBufferString(`{"name":"X"}`)), "user-1", "approved")
	rr3 := httptest.NewRecorder()
	guarded.ServeHTTP(rr3, req3)
	assert.Equal(t, http.StatusCreated, rr3.Code)
}

func TestSlugify(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"My Salon", "my-salon"},
		{"  My  Salon  ", "my-salon"},
		{"Hello World!", "hello-world"},
		{"Café", "caf"},
		{"", "salon"},
		{"!!!", "salon"},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.want, slugify(tc.in), "slugify(%q)", tc.in)
	}
}
