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
	"fejd-backend/internal/middleware"
	"fejd-backend/internal/models"
	"fejd-backend/internal/service"
	"fejd-backend/internal/sse"
	"fejd-backend/internal/store"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newUnavailabilityTestHandler(t *testing.T) (*AdminHandler, *store.BusinessStore, *store.BusinessUserStore, uuid.UUID) {
	pool := setupHandlerTestDB(t)
	businessStore := store.NewBusinessStore(pool)
	buStore := store.NewBusinessUserStore(pool)
	serviceStore := store.NewServiceStore(pool)
	appointmentStore := store.NewAppointmentStore(pool)
	workingHoursStore := store.NewWorkingHoursStore(pool)
	overrideStore := store.NewWorkingHoursOverrideStore(pool)
	employeeServiceStore := store.NewEmployeeServiceStore(pool)
	unavailabilityStore := store.NewEmployeeUnavailabilityStore(pool)
	hub := sse.NewHub()

	slotService := service.NewSlotService(
		appointmentStore, workingHoursStore, overrideStore,
		serviceStore, buStore, employeeServiceStore, unavailabilityStore, hub, pool,
	)

	h := NewAdminHandler(businessStore, buStore, serviceStore, nil, nil, nil, appointmentStore, slotService, nil, nil, pool)

	ctx := context.Background()
	b := &models.Business{Name: "Salon", Slug: "salon"}
	require.NoError(t, businessStore.Create(ctx, pool, b))
	require.NoError(t, buStore.Create(ctx, pool, &models.BusinessUser{
		BusinessID: b.ID,
		UserID:     "owner-1",
		Role:       "admin",
	}))
	require.NoError(t, buStore.Create(ctx, pool, &models.BusinessUser{
		BusinessID: b.ID,
		UserID:     "emp-1",
		Role:       "employee",
	}))

	return h, businessStore, buStore, b.ID
}

func memberGuard(buStore *store.BusinessUserStore, next http.HandlerFunc) http.Handler {
	return middleware.RequireBusinessMember(buStore)(next)
}

func TestAdminHandler_MyUnavailability_Lifecycle(t *testing.T) {
	h, _, buStore, businessID := newUnavailabilityTestHandler(t)

	guard := memberGuard(buStore, http.HandlerFunc(h.ListMyUnavailability))

	// Initially empty.
	req := withPathParams(
		httptest.NewRequest(http.MethodGet, "/api/admin/business/"+businessID.String()+"/me/unavailability", nil),
		map[string]string{"businessID": businessID.String()},
	)
	req = withUser(req, "emp-1", "approved")
	rr := httptest.NewRecorder()
	guard.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	var initial []dto.EmployeeUnavailability
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &initial))
	assert.Empty(t, initial)

	// Reserve a slot.
	start := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Hour)
	end := start.Add(time.Hour)
	body := `{"start_time":"` + start.Format(time.RFC3339) + `","end_time":"` + end.Format(time.RFC3339) + `","reason":"errand"}`
	postReq := withPathParams(
		httptest.NewRequest(http.MethodPost, "/api/admin/business/"+businessID.String()+"/me/unavailability", bytes.NewBufferString(body)),
		map[string]string{"businessID": businessID.String()},
	)
	postReq = withUser(postReq, "emp-1", "approved")
	postRR := httptest.NewRecorder()
	memberGuard(buStore, http.HandlerFunc(h.AddMyUnavailability)).ServeHTTP(postRR, postReq)
	require.Equal(t, http.StatusCreated, postRR.Code)
	var created dto.EmployeeUnavailability
	require.NoError(t, json.Unmarshal(postRR.Body.Bytes(), &created))
	assert.Equal(t, "errand", created.Reason)
	require.NotEqual(t, uuid.Nil, created.ID)

	// Now listed.
	rr = httptest.NewRecorder()
	guard.ServeHTTP(rr, withUser(withPathParams(
		httptest.NewRequest(http.MethodGet, "/api/admin/business/"+businessID.String()+"/me/unavailability", nil),
		map[string]string{"businessID": businessID.String()},
	), "emp-1", "approved"))
	require.Equal(t, http.StatusOK, rr.Code)
	var list []dto.EmployeeUnavailability
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &list))
	require.Len(t, list, 1)
	assert.Equal(t, created.ID, list[0].ID)

	// Delete it.
	delReq := withPathParams(
		httptest.NewRequest(http.MethodDelete, "/api/admin/business/"+businessID.String()+"/me/unavailability/"+created.ID.String(), nil),
		map[string]string{"businessID": businessID.String(), "unavailabilityID": created.ID.String()},
	)
	delReq = withUser(delReq, "emp-1", "approved")
	delRR := httptest.NewRecorder()
	memberGuard(buStore, http.HandlerFunc(h.DeleteMyUnavailability)).ServeHTTP(delRR, delReq)
	require.Equal(t, http.StatusOK, delRR.Code)

	rr = httptest.NewRecorder()
	guard.ServeHTTP(rr, withUser(withPathParams(
		httptest.NewRequest(http.MethodGet, "/api/admin/business/"+businessID.String()+"/me/unavailability", nil),
		map[string]string{"businessID": businessID.String()},
	), "emp-1", "approved"))
	require.Equal(t, http.StatusOK, rr.Code)
	var after []dto.EmployeeUnavailability
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &after))
	assert.Empty(t, after)
}

func TestAdminHandler_MyUnavailability_OwnerCanReserve(t *testing.T) {
	h, _, buStore, businessID := newUnavailabilityTestHandler(t)

	start := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Hour)
	body := `{"start_time":"` + start.Format(time.RFC3339) + `","end_time":"` + start.Add(time.Hour).Format(time.RFC3339) + `"}`
	req := withPathParams(
		httptest.NewRequest(http.MethodPost, "/api/admin/business/"+businessID.String()+"/me/unavailability", bytes.NewBufferString(body)),
		map[string]string{"businessID": businessID.String()},
	)
	req = withUser(req, "owner-1", "approved")
	rr := httptest.NewRecorder()
	memberGuard(buStore, http.HandlerFunc(h.AddMyUnavailability)).ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
}

func TestAdminHandler_MyUnavailability_NonMemberForbidden(t *testing.T) {
	h, _, buStore, businessID := newUnavailabilityTestHandler(t)

	req := withPathParams(
		httptest.NewRequest(http.MethodGet, "/api/admin/business/"+businessID.String()+"/me/unavailability", nil),
		map[string]string{"businessID": businessID.String()},
	)
	req = withUser(req, "stranger", "approved")
	rr := httptest.NewRecorder()
	memberGuard(buStore, http.HandlerFunc(h.ListMyUnavailability)).ServeHTTP(rr, req)

	require.Equal(t, http.StatusForbidden, rr.Code)
}
