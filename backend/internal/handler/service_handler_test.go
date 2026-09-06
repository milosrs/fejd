package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"fejd-backend/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminHandler_DeleteService_WithAppointments(t *testing.T) {
	h, businessStore, _, pool := newTestAdminHandler(t)
	ctx := context.Background()

	b := &models.Business{Name: "Salon", Slug: "salon"}
	require.NoError(t, businessStore.Create(ctx, pool, b))

	buID := uuid.New()
	_, err := pool.Exec(ctx,
		`INSERT INTO business_users (id, business_id, user_id, role) VALUES ($1, $2, 'emp-1', 'employee')`,
		buID, b.ID,
	)
	require.NoError(t, err)

	serviceID := uuid.New()
	_, err = pool.Exec(ctx,
		`INSERT INTO services (id, business_id, name, duration_minutes, active) VALUES ($1, $2, 'Haircut', 30, true)`,
		serviceID, b.ID,
	)
	require.NoError(t, err)

	_, err = pool.Exec(ctx,
		`INSERT INTO employee_services (business_user_id, service_id) VALUES ($1, $2)`,
		buID, serviceID,
	)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO appointments (business_id, service_id, business_user_id, customer_user_id, start_time, end_time, status, created_by)
		VALUES ($1, $2, $3, 'cust-1', now() + interval '1 day', now() + interval '1 day' + interval '30 minutes', 'confirmed', 'cust-1')`,
		b.ID, serviceID, buID,
	)
	require.NoError(t, err)

	req := withPathParams(
		httptest.NewRequest(http.MethodDelete, "/api/admin/business/"+b.ID.String()+"/services/"+serviceID.String(), nil),
		map[string]string{"businessID": b.ID.String(), "serviceID": serviceID.String()},
	)
	rr := httptest.NewRecorder()

	h.DeleteService(rr, req)

	require.Equal(t, http.StatusConflict, rr.Code)
	assert.Contains(t, rr.Body.String(), "cannot be deleted")
}
