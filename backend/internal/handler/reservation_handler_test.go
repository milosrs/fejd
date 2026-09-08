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
	"fejd-backend/internal/sse"
	"fejd-backend/internal/store"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newReservationTestHandler(t *testing.T) (*AdminHandler, *store.BusinessUserStore, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) {
	pool := setupHandlerTestDB(t)
	ctx := context.Background()

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

	b := &models.Business{Name: "Salon", Slug: "salon"}
	require.NoError(t, businessStore.Create(ctx, pool, b))

	require.NoError(t, buStore.Create(ctx, pool, &models.BusinessUser{BusinessID: b.ID, UserID: "owner-1", Role: "admin"}))
	require.NoError(t, buStore.Create(ctx, pool, &models.BusinessUser{BusinessID: b.ID, UserID: "emp-1", Role: "employee"}))
	require.NoError(t, buStore.Create(ctx, pool, &models.BusinessUser{BusinessID: b.ID, UserID: "emp-2", Role: "employee"}))

	emp1, err := buStore.GetByBusinessAndUser(ctx, b.ID, "emp-1")
	require.NoError(t, err)
	emp2, err := buStore.GetByBusinessAndUser(ctx, b.ID, "emp-2")
	require.NoError(t, err)

	svc := &models.Service{BusinessID: b.ID, Name: "Haircut", DurationMinutes: 30, Active: true}
	require.NoError(t, serviceStore.Create(ctx, svc))
	require.NoError(t, employeeServiceStore.Assign(ctx, emp1.ID, svc.ID))

	apptID := uuid.New()
	day := time.Now().UTC().AddDate(0, 0, 1)
	start := time.Date(day.Year(), day.Month(), day.Day(), 9, 0, 0, 0, time.UTC)
	_, err = pool.Exec(ctx,
		`INSERT INTO appointments (id, business_id, service_id, business_user_id, customer_user_id, start_time, end_time, status, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, 'confirmed', $5)`,
		apptID, b.ID, svc.ID, emp1.ID, "customer-1", start, start.Add(30*time.Minute))
	require.NoError(t, err)

	return h, buStore, b.ID, emp1.ID, emp2.ID, svc.ID, apptID
}

func reservationListURL(businessID uuid.UUID, date time.Time) string {
	return "/api/admin/business/" + businessID.String() + "/me/appointments?date=" + date.Format("2006-01-02")
}

func TestAdminHandler_MyReservations_ListAndCancel(t *testing.T) {
	h, buStore, businessID, _, _, _, apptID := newReservationTestHandler(t)

	day := time.Now().UTC().AddDate(0, 0, 1)

	listReq := withPathParams(
		httptest.NewRequest(http.MethodGet, reservationListURL(businessID, day), nil),
		map[string]string{"businessID": businessID.String()},
	)
	listReq = withUser(listReq, "emp-1", "approved")

	rr := httptest.NewRecorder()
	memberGuard(buStore, http.HandlerFunc(h.ListMyReservations)).ServeHTTP(rr, listReq)
	require.Equal(t, http.StatusOK, rr.Code)

	var reservations []dto.Appointment
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &reservations))
	require.Len(t, reservations, 1)
	assert.Equal(t, apptID, reservations[0].ID)
	assert.Equal(t, "Haircut", reservations[0].ServiceName)
	assert.Equal(t, "confirmed", reservations[0].Status)

	// Cancelling without a reason is rejected.
	cancelReq := withPathParams(
		httptest.NewRequest(http.MethodDelete, "/api/admin/business/"+businessID.String()+"/me/appointments/"+apptID.String(), bytes.NewBufferString(`{"cancellation_reason":""}`)),
		map[string]string{"businessID": businessID.String(), "appointmentID": apptID.String()},
	)
	cancelReq = withUser(cancelReq, "emp-1", "approved")
	rr = httptest.NewRecorder()
	memberGuard(buStore, http.HandlerFunc(h.CancelMyReservation)).ServeHTTP(rr, cancelReq)
	require.Equal(t, http.StatusBadRequest, rr.Code)

	// Cancelling with a reason succeeds.
	cancelReq = withPathParams(
		httptest.NewRequest(http.MethodDelete, "/api/admin/business/"+businessID.String()+"/me/appointments/"+apptID.String(), bytes.NewBufferString(`{"cancellation_reason":"car broke down"}`)),
		map[string]string{"businessID": businessID.String(), "appointmentID": apptID.String()},
	)
	cancelReq = withUser(cancelReq, "emp-1", "approved")
	rr = httptest.NewRecorder()
	memberGuard(buStore, http.HandlerFunc(h.CancelMyReservation)).ServeHTTP(rr, cancelReq)
	require.Equal(t, http.StatusOK, rr.Code)

	// The reservation is now listed as cancelled.
	listReq = withUser(listReq, "emp-1", "approved")
	rr = httptest.NewRecorder()
	memberGuard(buStore, http.HandlerFunc(h.ListMyReservations)).ServeHTTP(rr, listReq)
	require.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &reservations))
	require.Len(t, reservations, 1)
	assert.Equal(t, "cancelled", reservations[0].Status)
	assert.Equal(t, "car broke down", reservations[0].CancellationReason)
}

func TestAdminHandler_MyReservations_CannotCancelOthers(t *testing.T) {
	h, buStore, businessID, _, _, _, apptID := newReservationTestHandler(t)

	cancelReq := withPathParams(
		httptest.NewRequest(http.MethodDelete, "/api/admin/business/"+businessID.String()+"/me/appointments/"+apptID.String(), bytes.NewBufferString(`{"cancellation_reason":"not mine"}`)),
		map[string]string{"businessID": businessID.String(), "appointmentID": apptID.String()},
	)
	cancelReq = withUser(cancelReq, "emp-2", "approved")
	rr := httptest.NewRecorder()
	memberGuard(buStore, http.HandlerFunc(h.CancelMyReservation)).ServeHTTP(rr, cancelReq)

	// The call succeeds (no error) but is scoped to the caller, so emp-1's
	// appointment stays confirmed.
	require.Equal(t, http.StatusOK, rr.Code)

	day := time.Now().UTC().AddDate(0, 0, 1)
	listReq := withPathParams(
		httptest.NewRequest(http.MethodGet, reservationListURL(businessID, day), nil),
		map[string]string{"businessID": businessID.String()},
	)
	listReq = withUser(listReq, "emp-1", "approved")
	rr = httptest.NewRecorder()
	memberGuard(buStore, http.HandlerFunc(h.ListMyReservations)).ServeHTTP(rr, listReq)
	require.Equal(t, http.StatusOK, rr.Code)

	var reservations []dto.Appointment
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &reservations))
	require.Len(t, reservations, 1)
	assert.Equal(t, "confirmed", reservations[0].Status)
}

func TestAdminHandler_MyReservations_NonMemberForbidden(t *testing.T) {
	h, buStore, businessID, _, _, _, _ := newReservationTestHandler(t)

	day := time.Now().UTC().AddDate(0, 0, 1)
	listReq := withPathParams(
		httptest.NewRequest(http.MethodGet, reservationListURL(businessID, day), nil),
		map[string]string{"businessID": businessID.String()},
	)
	listReq = withUser(listReq, "stranger", "approved")
	rr := httptest.NewRecorder()
	memberGuard(buStore, http.HandlerFunc(h.ListMyReservations)).ServeHTTP(rr, listReq)

	require.Equal(t, http.StatusForbidden, rr.Code)
}
