package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"fejd-backend/internal/middleware"
	"fejd-backend/internal/models"
	"fejd-backend/internal/service"
	"fejd-backend/internal/sse"
	"fejd-backend/internal/store"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type cancelTestEnv struct {
	apptHandler      *AppointmentHandler
	adminHandler     *AdminHandler
	appointmentStore *store.AppointmentStore
	businessStore    *store.BusinessStore
	buStore          *store.BusinessUserStore
	businessID       uuid.UUID
	farApptID        uuid.UUID
	nearApptID       uuid.UUID
	otherApptID      uuid.UUID
	policyApptID     uuid.UUID
	pastApptID       uuid.UUID
	recentApptID     uuid.UUID
}

func newCancelTestEnv(t *testing.T) *cancelTestEnv {
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

	b := &models.Business{Name: "Salon", Slug: "salon"}
	require.NoError(t, businessStore.Create(ctx, pool, b))

	require.NoError(t, buStore.Create(ctx, pool, &models.BusinessUser{BusinessID: b.ID, UserID: "owner-1", Role: "admin"}))
	require.NoError(t, buStore.Create(ctx, pool, &models.BusinessUser{BusinessID: b.ID, UserID: "emp-1", Role: "employee"}))

	emp1, err := buStore.GetByBusinessAndUser(ctx, b.ID, "emp-1")
	require.NoError(t, err)

	svc := &models.Service{BusinessID: b.ID, Name: "Haircut", DurationMinutes: 30, Active: true}
	require.NoError(t, serviceStore.Create(ctx, svc))
	require.NoError(t, employeeServiceStore.Assign(ctx, emp1.ID, svc.ID))

	insertAppt := func(customerID string, start time.Time) uuid.UUID {
		id := uuid.New()
		_, err := pool.Exec(ctx,
			`INSERT INTO appointments (id, business_id, service_id, business_user_id, customer_user_id, start_time, end_time, status, created_by)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, 'confirmed', $5)`,
			id, b.ID, svc.ID, emp1.ID, customerID, start, start.Add(30*time.Minute))
		require.NoError(t, err)
		return id
	}

	// farAppt and nearAppt are 24h apart, so they fall on different UTC days
	// (the daily booking cap is one per customer per UTC day).
	farApptID := insertAppt("customer-1", time.Now().UTC().Add(25*time.Hour))
	nearApptID := insertAppt("customer-1", time.Now().UTC().Add(1*time.Hour))
	otherApptID := insertAppt("customer-2", time.Now().UTC().Add(5*time.Hour))
	policyApptID := insertAppt("customer-3", time.Now().UTC().Add(6*time.Hour))
	pastApptID := insertAppt("customer-4", time.Now().UTC().Add(-3*time.Hour))
	recentApptID := insertAppt("customer-5", time.Now().UTC().Add(-1*time.Hour))

	return &cancelTestEnv{
		apptHandler:      NewAppointmentHandler(appointmentStore, serviceStore, businessStore, buStore, slotService),
		adminHandler:     NewAdminHandler(businessStore, buStore, serviceStore, nil, nil, businessHoursStore, nil, appointmentStore, slotService, nil, nil, pool),
		appointmentStore: appointmentStore,
		businessStore:    businessStore,
		buStore:          buStore,
		businessID:       b.ID,
		farApptID:        farApptID,
		nearApptID:       nearApptID,
		otherApptID:      otherApptID,
		policyApptID:     policyApptID,
		pastApptID:       pastApptID,
		recentApptID:     recentApptID,
	}
}

func customerCancelReq(appointmentID uuid.UUID, customerID, body string) *http.Request {
	req := httptest.NewRequest(
		http.MethodDelete,
		"/api/my/appointments/"+appointmentID.String(),
		bytes.NewBufferString(body),
	)
	req = withPathParams(req, map[string]string{"appointmentID": appointmentID.String()})
	return withUser(req, customerID, "approved")
}

func TestAppointmentHandler_CancelWithinWindow(t *testing.T) {
	env := newCancelTestEnv(t)

	req := customerCancelReq(env.farApptID, "customer-1", `{"cancellation_reason":"changed plans"}`)
	rr := httptest.NewRecorder()
	env.apptHandler.Cancel(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	appt, err := env.appointmentStore.GetByID(context.Background(), env.farApptID)
	require.NoError(t, err)
	assert.Equal(t, models.AppointmentStatusCancelled, appt.Status)
	assert.Equal(t, "changed plans", appt.CancellationReason)
}

func TestAppointmentHandler_CancelTooLate(t *testing.T) {
	env := newCancelTestEnv(t)

	req := customerCancelReq(env.nearApptID, "customer-1", `{"cancellation_reason":"oops"}`)
	rr := httptest.NewRecorder()
	env.apptHandler.Cancel(rr, req)

	require.Equal(t, http.StatusConflict, rr.Code)

	appt, err := env.appointmentStore.GetByID(context.Background(), env.nearApptID)
	require.NoError(t, err)
	assert.Equal(t, models.AppointmentStatusConfirmed, appt.Status)
}

func TestAppointmentHandler_CancelOthersAppointment(t *testing.T) {
	env := newCancelTestEnv(t)

	req := customerCancelReq(env.otherApptID, "customer-1", `{"cancellation_reason":"mine"}`)
	rr := httptest.NewRecorder()
	env.apptHandler.Cancel(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)

	appt, err := env.appointmentStore.GetByID(context.Background(), env.otherApptID)
	require.NoError(t, err)
	assert.Equal(t, models.AppointmentStatusConfirmed, appt.Status)
}

func TestAdminHandler_SalonPolicy(t *testing.T) {
	env := newCancelTestEnv(t)

	getReq := withPathParams(
		httptest.NewRequest(http.MethodGet, "/api/admin/business/"+env.businessID.String()+"/policy", nil),
		map[string]string{"businessID": env.businessID.String()},
	)
	getReq = withUser(getReq, "owner-1", "approved")
	rr := httptest.NewRecorder()
	env.adminHandler.GetSalonPolicy(rr, getReq)
	require.Equal(t, http.StatusOK, rr.Code)
	var policy SalonPolicyResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &policy))
	assert.Equal(t, 2, policy.CancellationLeadHours)
	assert.Equal(t, 2, policy.NoShowAfterHours)
	assert.Equal(t, 30, policy.SlotIntervalMinutes)

	putReq := withPathParams(
		httptest.NewRequest(http.MethodPut, "/api/admin/business/"+env.businessID.String()+"/policy", bytes.NewBufferString(`{"cancellation_lead_hours":24,"no_show_after_hours":4,"slot_interval_minutes":45,"working_hours":[{"day_of_week":1,"start_time":"09:00","end_time":"17:00"}]}`)),
		map[string]string{"businessID": env.businessID.String()},
	)
	putReq = withUser(putReq, "owner-1", "approved")
	rr = httptest.NewRecorder()
	env.adminHandler.UpdateSalonPolicy(rr, putReq)
	require.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &policy))
	assert.Equal(t, 24, policy.CancellationLeadHours)
	assert.Equal(t, 4, policy.NoShowAfterHours)
	assert.Equal(t, 45, policy.SlotIntervalMinutes)
	assert.Len(t, policy.WorkingHours, 1)
	assert.Equal(t, "09:00", policy.WorkingHours[0].StartTime)

	// A 5-hour-away appointment is now inside the 24-hour window.
	req := customerCancelReq(env.policyApptID, "customer-3", `{"cancellation_reason":"too late now"}`)
	rr = httptest.NewRecorder()
	env.apptHandler.Cancel(rr, req)
	require.Equal(t, http.StatusConflict, rr.Code)
}

func TestAdminHandler_SalonPolicy_NonAdminForbidden(t *testing.T) {
	env := newCancelTestEnv(t)

	guarded := middleware.RequireBusinessAdmin(env.buStore)(http.HandlerFunc(env.adminHandler.GetSalonPolicy))
	req := withPathParams(
		httptest.NewRequest(http.MethodGet, "/api/admin/business/"+env.businessID.String()+"/policy", nil),
		map[string]string{"businessID": env.businessID.String()},
	)
	req = withUser(req, "emp-1", "approved")
	rr := httptest.NewRecorder()
	guarded.ServeHTTP(rr, req)

	require.Equal(t, http.StatusForbidden, rr.Code)
}

func noShowReq(businessID, appointmentID uuid.UUID, userID string) *http.Request {
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/admin/business/"+businessID.String()+"/me/appointments/"+appointmentID.String()+"/no-show",
		nil,
	)
	req = withPathParams(req, map[string]string{"businessID": businessID.String(), "appointmentID": appointmentID.String()})
	return withUser(req, userID, "approved")
}

func TestAdminHandler_MarkNoShow(t *testing.T) {
	env := newCancelTestEnv(t)

	req := noShowReq(env.businessID, env.pastApptID, "emp-1")
	rr := httptest.NewRecorder()
	env.adminHandler.MarkNoShow(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	appt, err := env.appointmentStore.GetByID(context.Background(), env.pastApptID)
	require.NoError(t, err)
	assert.Equal(t, models.AppointmentStatusNoShow, appt.Status)
}

func TestAdminHandler_MarkNoShowTooEarly(t *testing.T) {
	env := newCancelTestEnv(t)

	req := noShowReq(env.businessID, env.recentApptID, "emp-1")
	rr := httptest.NewRecorder()
	env.adminHandler.MarkNoShow(rr, req)

	require.Equal(t, http.StatusConflict, rr.Code)

	appt, err := env.appointmentStore.GetByID(context.Background(), env.recentApptID)
	require.NoError(t, err)
	assert.Equal(t, models.AppointmentStatusConfirmed, appt.Status)
}

func TestAdminHandler_MarkNoShowOthers(t *testing.T) {
	env := newCancelTestEnv(t)

	req := noShowReq(env.businessID, env.pastApptID, "owner-1")
	rr := httptest.NewRecorder()
	env.adminHandler.MarkNoShow(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)

	appt, err := env.appointmentStore.GetByID(context.Background(), env.pastApptID)
	require.NoError(t, err)
	assert.Equal(t, models.AppointmentStatusConfirmed, appt.Status)
}
