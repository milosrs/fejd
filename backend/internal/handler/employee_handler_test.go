package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"fejd-backend/internal/keycloak"
	"fejd-backend/internal/middleware"
	"fejd-backend/internal/models"
	"fejd-backend/internal/service"
	"fejd-backend/internal/sse"
	"fejd-backend/internal/store"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeInviter struct {
	userID          string
	created         keycloak.CreateUserInput
	executedActions []string
	clientID        string
	redirectURI     string
	lifespan        int
	addedRoles      []string
	removedRoles    []string
	err             error
}

func (f *fakeInviter) CreateUser(ctx context.Context, in keycloak.CreateUserInput) (string, error) {
	f.created = in
	if f.err != nil {
		return "", f.err
	}
	return f.userID, nil
}

func (f *fakeInviter) ExecuteActionsEmail(ctx context.Context, userID string, actions []string, clientID, redirectURI string, lifespan int) error {
	f.executedActions = actions
	f.clientID = clientID
	f.redirectURI = redirectURI
	f.lifespan = lifespan
	return f.err
}

func (f *fakeInviter) AddRealmRole(ctx context.Context, userID, roleName string) error {
	f.addedRoles = append(f.addedRoles, roleName)
	return f.err
}

func (f *fakeInviter) RemoveRealmRole(ctx context.Context, userID, roleName string) error {
	f.removedRoles = append(f.removedRoles, roleName)
	return f.err
}

func TestAdminHandler_CreateEmployee(t *testing.T) {
	pool := setupHandlerTestDB(t)
	ctx := context.Background()

	businessStore := store.NewBusinessStore(pool)
	buStore := store.NewBusinessUserStore(pool)
	serviceStore := store.NewServiceStore(pool)
	employeeServiceStore := store.NewEmployeeServiceStore(pool)

	inviter := &fakeInviter{userID: "kc-user-123"}
	employeeService := service.NewEmployeeService(inviter, buStore, employeeServiceStore, pool, "fejd://callback", 48*3600)
	h := NewAdminHandler(businessStore, buStore, serviceStore, nil, nil, nil, nil, nil, nil, nil, employeeService, pool)

	b := &models.Business{Name: "Salon", Slug: "salon"}
	require.NoError(t, businessStore.Create(ctx, pool, b))

	svcID := uuid.New()
	_, err := pool.Exec(ctx,
		`INSERT INTO services (id, business_id, name, duration_minutes, active) VALUES ($1, $2, 'Haircut', 30, true)`,
		svcID, b.ID,
	)
	require.NoError(t, err)

	body := `{"name":"Sam","email":"sam@example.com","service_ids":["` + svcID.String() + `"]}`
	req := withPathParams(
		httptest.NewRequest(http.MethodPost, "/api/admin/business/"+b.ID.String()+"/employees", bytes.NewBufferString(body)),
		map[string]string{"businessID": b.ID.String()},
	)
	rr := httptest.NewRecorder()

	h.CreateEmployee(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)

	assert.Equal(t, "sam@example.com", inviter.created.Email)
	assert.Equal(t, "sam@example.com", inviter.created.Username)
	assert.Equal(t, []string{"UPDATE_PASSWORD"}, inviter.executedActions)
	assert.Equal(t, "salon-mobile", inviter.clientID)
	assert.Equal(t, "fejd://callback", inviter.redirectURI)

	bu, err := buStore.GetByBusinessAndUser(ctx, b.ID, "kc-user-123")
	require.NoError(t, err)
	assert.Equal(t, "employee", bu.Role)
	assert.Equal(t, "Sam", bu.DisplayName)

	offers, err := employeeServiceStore.OffersService(ctx, bu.ID, svcID)
	require.NoError(t, err)
	assert.True(t, offers)
}

func TestAdminHandler_CreateEmployee_NonOwnerRejected(t *testing.T) {
	pool := setupHandlerTestDB(t)
	ctx := context.Background()

	businessStore := store.NewBusinessStore(pool)
	buStore := store.NewBusinessUserStore(pool)
	serviceStore := store.NewServiceStore(pool)
	employeeServiceStore := store.NewEmployeeServiceStore(pool)

	inviter := &fakeInviter{userID: "kc-user-123"}
	employeeService := service.NewEmployeeService(inviter, buStore, employeeServiceStore, pool, "fejd://callback", 48*3600)
	h := NewAdminHandler(businessStore, buStore, serviceStore, nil, nil, nil, nil, nil, nil, nil, employeeService, pool)

	b := &models.Business{Name: "Salon", Slug: "salon"}
	require.NoError(t, businessStore.Create(ctx, pool, b))
	require.NoError(t, buStore.Create(ctx, pool, &models.BusinessUser{
		BusinessID: b.ID,
		UserID:     "owner",
		Role:       "admin",
	}))

	guarded := middleware.RequireBusinessAdmin(buStore)(http.HandlerFunc(h.CreateEmployee))

	body := `{"name":"Sam","email":"sam@example.com","service_ids":[]}`
	req := withPathParams(
		httptest.NewRequest(http.MethodPost, "/api/admin/business/"+b.ID.String()+"/employees", bytes.NewBufferString(body)),
		map[string]string{"businessID": b.ID.String()},
	)
	req = withUser(req, "not-owner", "approved")
	rr := httptest.NewRecorder()

	guarded.ServeHTTP(rr, req)

	require.Equal(t, http.StatusForbidden, rr.Code)
}

func TestAdminHandler_RemoveEmployee_RevokesRole(t *testing.T) {
	pool := setupHandlerTestDB(t)
	ctx := context.Background()

	businessStore := store.NewBusinessStore(pool)
	buStore := store.NewBusinessUserStore(pool)
	serviceStore := store.NewServiceStore(pool)
	appointmentStore := store.NewAppointmentStore(pool)
	workingHoursStore := store.NewWorkingHoursStore(pool)
	overrideStore := store.NewWorkingHoursOverrideStore(pool)
	businessHoursStore := store.NewBusinessHoursStore(pool)
	employeeServiceStore := store.NewEmployeeServiceStore(pool)
	unavailabilityStore := store.NewEmployeeUnavailabilityStore(pool)
	hub := sse.NewHub()

	slotService := service.NewSlotService(
		appointmentStore, workingHoursStore, businessHoursStore, overrideStore,
		serviceStore, businessStore, buStore, employeeServiceStore, unavailabilityStore, hub, pool,
	)

	inviter := &fakeInviter{userID: "emp-1"}
	employeeService := service.NewEmployeeService(inviter, buStore, employeeServiceStore, pool, "fejd://callback", 48*3600)

	h := NewAdminHandler(businessStore, buStore, serviceStore, nil, nil, businessHoursStore, nil, appointmentStore, slotService, nil, employeeService, pool)

	b := &models.Business{Name: "Salon", Slug: "salon"}
	require.NoError(t, businessStore.Create(ctx, pool, b))
	require.NoError(t, buStore.Create(ctx, pool, &models.BusinessUser{BusinessID: b.ID, UserID: "owner", Role: "admin"}))
	require.NoError(t, buStore.Create(ctx, pool, &models.BusinessUser{BusinessID: b.ID, UserID: "emp-1", Role: "employee"}))

	req := withPathParams(
		httptest.NewRequest(http.MethodDelete, "/api/admin/business/"+b.ID.String()+"/employees/emp-1", nil),
		map[string]string{"businessID": b.ID.String(), "userID": "emp-1"},
	)
	rr := httptest.NewRecorder()
	h.RemoveEmployee(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, inviter.removedRoles, "Employee")

	bu, err := buStore.GetByBusinessAndUser(ctx, b.ID, "emp-1")
	require.NoError(t, err)
	assert.False(t, bu.Active)
}
