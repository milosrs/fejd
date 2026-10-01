package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fejd-backend/internal/dto"
	"fejd-backend/internal/models"
	"fejd-backend/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestBusinessHandler(t *testing.T) (*BusinessHandler, *store.BusinessStore, *pgxpool.Pool) {
	pool := setupHandlerTestDB(t)
	businessStore := store.NewBusinessStore(pool)
	buStore := store.NewBusinessUserStore(pool)
	serviceStore := store.NewServiceStore(pool)
	pageStore := store.NewPageStore(pool)
	sectionStore := store.NewSectionStore(pool)
	imageLinkStore := store.NewImageLinkStore(pool)
	employeeServiceStore := store.NewEmployeeServiceStore(pool)
	businessClosureStore := store.NewBusinessClosureStore(pool)
	businessHoursStore := store.NewBusinessHoursStore(pool)
	h := NewBusinessHandler(businessStore, buStore, nil, serviceStore, pageStore, sectionStore, imageLinkStore, employeeServiceStore, businessClosureStore, businessHoursStore, nil, "fejd.fyi")
	return h, businessStore, pool
}

func withSlug(r *http.Request, slug string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("slug", slug)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestBusinessHandler_GetSections(t *testing.T) {
	h, businessStore, pool := newTestBusinessHandler(t)
	ctx := context.Background()

	b := &models.Business{Name: "Salon", Slug: "salon"}
	require.NoError(t, businessStore.Create(ctx, pool, b))

	pageID := uuid.New()
	_, err := pool.Exec(ctx,
		`INSERT INTO pages (id, business_id, name) VALUES ($1, $2, 'landing')`,
		pageID, b.ID,
	)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO sections (page_id, type, content, position) VALUES
		($1, 'about', '{"en":{"heading":"About","body":"Hello"}}'::jsonb, 1),
		($1, 'hero',  '{"en":{"headline":"Cuts","cta_text":"Book"}}'::jsonb, 0)`,
		pageID,
	)
	require.NoError(t, err)

	req := withSlug(httptest.NewRequest(http.MethodGet, "/api/business/salon/sections", nil), "salon")
	rr := httptest.NewRecorder()

	h.GetSections(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var sections []dto.Section
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &sections))
	require.Len(t, sections, 2)
	assert.Equal(t, "hero", sections[0].Type)
	assert.Equal(t, 0, sections[0].Position)
	assert.JSONEq(t, `{"en":{"headline":"Cuts","cta_text":"Book"}}`, string(sections[0].Content))
}

func TestBusinessHandler_GetSections_NoPage(t *testing.T) {
	h, businessStore, pool := newTestBusinessHandler(t)
	ctx := context.Background()

	b := &models.Business{Name: "Salon", Slug: "salon"}
	require.NoError(t, businessStore.Create(ctx, pool, b))

	req := withSlug(httptest.NewRequest(http.MethodGet, "/api/business/salon/sections", nil), "salon")
	rr := httptest.NewRecorder()

	h.GetSections(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.JSONEq(t, "[]", rr.Body.String())
}

func TestBusinessHandler_GetBusiness_Images(t *testing.T) {
	h, businessStore, pool := newTestBusinessHandler(t)
	ctx := context.Background()

	b := &models.Business{Name: "Salon", Slug: "salon"}
	require.NoError(t, businessStore.Create(ctx, pool, b))

	imageID := uuid.New()
	_, err := pool.Exec(ctx,
		`INSERT INTO images (id, storage, url, content_type, business_id) VALUES ($1, 'external', 'https://example.com/hero.jpg', 'image/jpeg', $2)`,
		imageID, b.ID,
	)
	require.NoError(t, err)

	_, err = pool.Exec(ctx,
		`INSERT INTO image_links (image_id, entity_type, entity_id, purpose, visibility) VALUES ($1, 'business', $2, 'hero', 'public')`,
		imageID, b.ID,
	)
	require.NoError(t, err)

	req := withSlug(httptest.NewRequest(http.MethodGet, "/api/business/salon", nil), "salon")
	rr := httptest.NewRecorder()

	h.GetBusiness(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var resp BusinessResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "/api/images/"+imageID.String(), resp.Images.Hero)
	assert.Empty(t, resp.Images.Logo)
}

func TestBusinessHandler_GetServiceEmployees(t *testing.T) {
	h, businessStore, pool := newTestBusinessHandler(t)
	ctx := context.Background()

	b := &models.Business{Name: "Salon", Slug: "salon"}
	require.NoError(t, businessStore.Create(ctx, pool, b))

	emp1 := uuid.New()
	emp2 := uuid.New()
	_, err := pool.Exec(ctx,
		`INSERT INTO business_users (id, business_id, user_id, role, display_name) VALUES ($1, $2, 'emp-1', 'employee', 'Sam'), ($3, $2, 'emp-2', 'employee', 'Alex')`,
		emp1, b.ID, emp2,
	)
	require.NoError(t, err)

	svcA := uuid.New()
	svcB := uuid.New()
	_, err = pool.Exec(ctx,
		`INSERT INTO services (id, business_id, name, slug, duration_minutes, active) VALUES ($1, $2, 'Haircut', 'haircut', 30, true), ($3, $2, 'Beard', 'beard', 20, true)`,
		svcA, b.ID, svcB,
	)
	require.NoError(t, err)

	_, err = pool.Exec(ctx,
		`INSERT INTO employee_services (business_user_id, service_id) VALUES ($1, $2), ($3, $4)`,
		emp1, svcA, emp2, svcB,
	)
	require.NoError(t, err)

	req := withPathParams(
		httptest.NewRequest(http.MethodGet, "/api/business/salon/services/"+svcA.String()+"/employees", nil),
		map[string]string{"slug": "salon", "serviceID": svcA.String()},
	)
	rr := httptest.NewRecorder()

	h.GetServiceEmployees(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var employees []dto.BusinessUser
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &employees))
	require.Len(t, employees, 1)
	assert.Equal(t, emp1, employees[0].ID)
}

func TestBusinessHandler_GetClosures(t *testing.T) {
	h, businessStore, pool := newTestBusinessHandler(t)
	ctx := context.Background()

	b := &models.Business{Name: "Salon", Slug: "salon"}
	require.NoError(t, businessStore.Create(ctx, pool, b))

	_, err := pool.Exec(ctx,
		`INSERT INTO business_closures (business_id, closure_type, start_date, reason) VALUES ($1, 'single', '2026-12-25', 'Christmas')`,
		b.ID,
	)
	require.NoError(t, err)
	_, err = pool.Exec(ctx,
		`INSERT INTO business_closures (business_id, closure_type, day_of_week) VALUES ($1, 'weekly', 6)`,
		b.ID,
	)
	require.NoError(t, err)

	req := withSlug(httptest.NewRequest(http.MethodGet, "/api/business/salon/closures", nil), "salon")
	rr := httptest.NewRecorder()

	h.GetClosures(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var closures []dto.BusinessClosure
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &closures))
	require.Len(t, closures, 2)
	assert.Equal(t, "single", closures[0].Type)
	assert.Equal(t, "2026-12-25", closures[0].StartDate)
	assert.Equal(t, "Christmas", closures[0].Reason)
	assert.Equal(t, "weekly", closures[1].Type)
	require.NotNil(t, closures[1].DayOfWeek)
	assert.Equal(t, 6, *closures[1].DayOfWeek)
}

func TestBusinessHandler_GetClosures_Empty(t *testing.T) {
	h, businessStore, pool := newTestBusinessHandler(t)
	ctx := context.Background()

	b := &models.Business{Name: "Salon", Slug: "salon"}
	require.NoError(t, businessStore.Create(ctx, pool, b))

	req := withSlug(httptest.NewRequest(http.MethodGet, "/api/business/salon/closures", nil), "salon")
	rr := httptest.NewRecorder()

	h.GetClosures(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.JSONEq(t, "[]", rr.Body.String())
}

func TestBusinessHandler_GetWorkingHours(t *testing.T) {
	h, businessStore, pool := newTestBusinessHandler(t)
	ctx := context.Background()

	b := &models.Business{Name: "Salon", Slug: "salon"}
	require.NoError(t, businessStore.Create(ctx, pool, b))

	_, err := pool.Exec(ctx,
		`INSERT INTO business_hours (business_id, day_of_week, start_time, end_time) VALUES
			($1, 1, '09:00:00'::time, '17:00:00'::time),
			($1, 2, '10:00:00'::time, '18:30:00'::time)`,
		b.ID,
	)
	require.NoError(t, err)

	req := withSlug(httptest.NewRequest(http.MethodGet, "/api/business/salon/working-hours", nil), "salon")
	rr := httptest.NewRecorder()

	h.GetWorkingHours(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var hours []dto.BusinessHours
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &hours))
	require.Len(t, hours, 2)
	assert.Equal(t, 1, hours[0].DayOfWeek)
	assert.Equal(t, "09:00", hours[0].StartTime)
	assert.Equal(t, "17:00", hours[0].EndTime)
	assert.Equal(t, 2, hours[1].DayOfWeek)
	assert.Equal(t, "18:30", hours[1].EndTime)
}
