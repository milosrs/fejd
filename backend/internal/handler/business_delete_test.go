package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"fejd-backend/internal/models"
	"fejd-backend/internal/service"
	"fejd-backend/internal/store"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminHandler_DeleteBusiness(t *testing.T) {
	pool := setupHandlerTestDB(t)
	ctx := context.Background()

	businessStore := store.NewBusinessStore(pool)
	buStore := store.NewBusinessUserStore(pool)
	serviceStore := store.NewServiceStore(pool)
	appointmentStore := store.NewAppointmentStore(pool)
	userStore := store.NewUserStore(pool)
	invitationStore := store.NewInvitationStore(pool)
	employeeServiceStore := store.NewEmployeeServiceStore(pool)

	inviter := &fakeInviter{userID: "owner"}
	employeeService := service.NewEmployeeService(inviter, buStore, employeeServiceStore, pool, "fejd://callback", 48*3600)

	h := NewAdminHandler(businessStore, buStore, serviceStore, nil, nil, nil, nil, appointmentStore, nil, nil, employeeService, pool)

	b := &models.Business{Name: "Salon", Slug: "salon"}
	require.NoError(t, businessStore.Create(ctx, pool, b))

	// Platform users: the owner, employees, and invitees all exist as Keycloak
	// accounts cached in the local users table.
	userIDs := []string{"owner", "emp-1", "emp-2", "customer-1", "customer-2"}
	for _, id := range userIDs {
		require.NoError(t, userStore.Upsert(ctx, &models.User{ID: id, DisplayName: id}))
	}

	// Employees are linked to the business via business_users rows.
	require.NoError(t, buStore.Create(ctx, pool, &models.BusinessUser{BusinessID: b.ID, UserID: "owner", Role: "admin"}))
	require.NoError(t, buStore.Create(ctx, pool, &models.BusinessUser{BusinessID: b.ID, UserID: "emp-1", Role: "employee"}))
	require.NoError(t, buStore.Create(ctx, pool, &models.BusinessUser{BusinessID: b.ID, UserID: "emp-2", Role: "employee"}))

	// Users the business has invited to the platform.
	require.NoError(t, invitationStore.Create(ctx, pool, &models.Invitation{
		BusinessID: b.ID,
		TokenHash:  "hash-invite-customer-1",
		Role:       "customer",
		CreatedBy:  "owner",
		MaxUses:    1,
		ExpiresAt:  time.Now().UTC().Add(48 * time.Hour),
	}))
	require.NoError(t, invitationStore.Create(ctx, pool, &models.Invitation{
		BusinessID: b.ID,
		TokenHash:  "hash-invite-customer-2",
		Role:       "customer",
		CreatedBy:  "owner",
		MaxUses:    1,
		ExpiresAt:  time.Now().UTC().Add(48 * time.Hour),
	}))

	// A name that doesn't match is rejected before anything is deleted.
	req := withPathParams(
		httptest.NewRequest(http.MethodDelete, "/api/admin/business/"+b.ID.String(), bytes.NewBufferString(`{"name":"wrong"}`)),
		map[string]string{"businessID": b.ID.String()},
	)
	rr := httptest.NewRecorder()
	h.DeleteBusiness(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)

	// The correct name deletes the business.
	req = withPathParams(
		httptest.NewRequest(http.MethodDelete, "/api/admin/business/"+b.ID.String(), bytes.NewBufferString(`{"name":"Salon"}`)),
		map[string]string{"businessID": b.ID.String()},
	)
	rr = httptest.NewRecorder()
	h.DeleteBusiness(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	// The business is gone.
	_, err := businessStore.GetByID(ctx, b.ID)
	require.Error(t, err)

	// Employees lose their employee status (their business_users rows cascade).
	for _, id := range []string{"owner", "emp-1", "emp-2"} {
		_, err := buStore.GetByBusinessAndUser(ctx, b.ID, id)
		require.Error(t, err)
	}

	// The business's invitations are cascade-deleted.
	invites, err := invitationStore.ListByBusiness(ctx, b.ID)
	require.NoError(t, err)
	require.Empty(t, invites)

	// All users remain on the platform.
	for _, id := range userIDs {
		u, err := userStore.GetByID(ctx, id)
		require.NoError(t, err)
		require.NotNil(t, u)
	}

	// Employees were switched to customers: Employee role dropped, Customer
	// role granted. The owner's role is untouched.
	assert.ElementsMatch(t, []string{"Employee", "Employee"}, inviter.removedRoles)
	assert.ElementsMatch(t, []string{"Customer", "Customer"}, inviter.addedRoles)
	assert.NotContains(t, inviter.removedRoles, "Owner")
	assert.NotContains(t, inviter.addedRoles, "Owner")
}
