package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"fejd-backend/internal/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func bookOwnReq(businessID, userID, body string) *http.Request {
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/admin/business/"+businessID+"/me/appointments",
		bytes.NewBufferString(body),
	)
	req = withPathParams(req, map[string]string{"businessID": businessID})
	return withUser(req, userID, "approved")
}

func TestAdminHandler_BookOwnAppointment_WalkIn(t *testing.T) {
	h, buStore, businessID, emp1ID, _, svcID, _ := newReservationTestHandler(t)

	start := time.Now().UTC().Add(3 * time.Hour).Truncate(time.Hour)
	body := `{"service_id":"` + svcID.String() + `","start_time":"` + start.Format(time.RFC3339) + `"}`
	req := bookOwnReq(businessID.String(), "emp-1", body)
	rr := httptest.NewRecorder()
	memberGuard(buStore, http.HandlerFunc(h.BookOwnAppointment)).ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
	var appt dto.Appointment
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &appt))
	assert.Equal(t, svcID, appt.ServiceID)
	assert.Equal(t, emp1ID, appt.BusinessUserID)
	assert.Empty(t, appt.CustomerUserID)
	assert.Equal(t, "confirmed", appt.Status)
}

func TestAdminHandler_BookOwnAppointment_WithCustomer(t *testing.T) {
	h, buStore, businessID, _, _, svcID, _ := newReservationTestHandler(t)

	start := time.Now().UTC().Add(4 * time.Hour).Truncate(time.Hour)
	body := `{"service_id":"` + svcID.String() + `","start_time":"` + start.Format(time.RFC3339) + `","customer_user_id":"customer-99"}`
	req := bookOwnReq(businessID.String(), "emp-1", body)
	rr := httptest.NewRecorder()
	memberGuard(buStore, http.HandlerFunc(h.BookOwnAppointment)).ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
	var appt dto.Appointment
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &appt))
	assert.Equal(t, "customer-99", appt.CustomerUserID)
}

func TestAdminHandler_BookOwnAppointment_Past(t *testing.T) {
	h, buStore, businessID, _, _, svcID, _ := newReservationTestHandler(t)

	start := time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Hour)
	body := `{"service_id":"` + svcID.String() + `","start_time":"` + start.Format(time.RFC3339) + `"}`
	req := bookOwnReq(businessID.String(), "emp-1", body)
	rr := httptest.NewRecorder()
	memberGuard(buStore, http.HandlerFunc(h.BookOwnAppointment)).ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
}

func TestAdminHandler_ListMyServices(t *testing.T) {
	h, buStore, businessID, _, _, svcID, _ := newReservationTestHandler(t)

	req := withPathParams(
		httptest.NewRequest(http.MethodGet, "/api/admin/business/"+businessID.String()+"/me/services", nil),
		map[string]string{"businessID": businessID.String()},
	)
	req = withUser(req, "emp-1", "approved")
	rr := httptest.NewRecorder()
	memberGuard(buStore, http.HandlerFunc(h.ListMyServices)).ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var services []dto.Service
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &services))
	require.Len(t, services, 1)
	assert.Equal(t, svcID, services[0].ID)
}

func TestAdminHandler_ListCustomers(t *testing.T) {
	h, buStore, businessID, _, _, _, _ := newReservationTestHandler(t)

	req := withPathParams(
		httptest.NewRequest(http.MethodGet, "/api/admin/business/"+businessID.String()+"/customers", nil),
		map[string]string{"businessID": businessID.String()},
	)
	req = withUser(req, "emp-1", "approved")
	rr := httptest.NewRecorder()
	memberGuard(buStore, http.HandlerFunc(h.ListCustomers)).ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var customers []dto.Customer
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &customers))
	require.Len(t, customers, 1)
	assert.Equal(t, "customer-1", customers[0].UserID)
}
