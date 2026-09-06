package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fejd-backend/internal/dto"
	"fejd-backend/internal/middleware"
	"fejd-backend/internal/models"
	"fejd-backend/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestAdminHandler(t *testing.T) (*AdminHandler, *store.BusinessStore, *store.BusinessUserStore, *pgxpool.Pool) {
	pool := setupHandlerTestDB(t)
	businessStore := store.NewBusinessStore(pool)
	buStore := store.NewBusinessUserStore(pool)
	serviceStore := store.NewServiceStore(pool)
	pageStore := store.NewPageStore(pool)
	sectionStore := store.NewSectionStore(pool)
	h := NewAdminHandler(businessStore, buStore, serviceStore, pageStore, sectionStore, nil, nil, nil, nil, pool)
	return h, businessStore, buStore, pool
}

func withPathParams(r *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestAdminHandler_CreateSection(t *testing.T) {
	h, businessStore, _, pool := newTestAdminHandler(t)
	ctx := context.Background()

	b := &models.Business{Name: "Salon", Slug: "salon"}
	require.NoError(t, businessStore.Create(ctx, pool, b))

	body := `{"type":"hero","content":{"en":{"headline":"Cuts"}}}`
	req := withPathParams(
		httptest.NewRequest(http.MethodPost, "/api/admin/business/"+b.ID.String()+"/sections", bytes.NewBufferString(body)),
		map[string]string{"businessID": b.ID.String()},
	)
	rr := httptest.NewRecorder()

	h.CreateSection(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
	var sec dto.Section
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &sec))
	assert.Equal(t, "hero", sec.Type)
	assert.JSONEq(t, `{"en":{"headline":"Cuts"}}`, string(sec.Content))
}

func TestAdminHandler_UpdateSection(t *testing.T) {
	h, businessStore, _, pool := newTestAdminHandler(t)
	ctx := context.Background()

	b := &models.Business{Name: "Salon", Slug: "salon"}
	require.NoError(t, businessStore.Create(ctx, pool, b))

	pageID := uuid.New()
	_, err := pool.Exec(ctx, `INSERT INTO pages (id, business_id, name) VALUES ($1, $2, 'landing')`, pageID, b.ID)
	require.NoError(t, err)
	sectionID := uuid.New()
	_, err = pool.Exec(ctx, `INSERT INTO sections (id, page_id, type, content, position) VALUES ($1, $2, 'hero', '{"en":{"headline":"Old"}}'::jsonb, 0)`, sectionID, pageID)
	require.NoError(t, err)

	body := `{"content":{"en":{"headline":"New"}}}`
	req := withPathParams(
		httptest.NewRequest(http.MethodPut, "/api/admin/business/"+b.ID.String()+"/sections/"+sectionID.String(), bytes.NewBufferString(body)),
		map[string]string{"businessID": b.ID.String(), "sectionID": sectionID.String()},
	)
	rr := httptest.NewRecorder()

	h.UpdateSection(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var sec dto.Section
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &sec))
	assert.JSONEq(t, `{"en":{"headline":"New"}}`, string(sec.Content))
}

func TestAdminHandler_DeleteSection(t *testing.T) {
	h, businessStore, _, pool := newTestAdminHandler(t)
	ctx := context.Background()

	b := &models.Business{Name: "Salon", Slug: "salon"}
	require.NoError(t, businessStore.Create(ctx, pool, b))

	pageID := uuid.New()
	_, err := pool.Exec(ctx, `INSERT INTO pages (id, business_id, name) VALUES ($1, $2, 'landing')`, pageID, b.ID)
	require.NoError(t, err)
	sectionID := uuid.New()
	_, err = pool.Exec(ctx, `INSERT INTO sections (id, page_id, type, content, position) VALUES ($1, $2, 'about', '{}'::jsonb, 0)`, sectionID, pageID)
	require.NoError(t, err)

	req := withPathParams(
		httptest.NewRequest(http.MethodDelete, "/api/admin/business/"+b.ID.String()+"/sections/"+sectionID.String(), nil),
		map[string]string{"businessID": b.ID.String(), "sectionID": sectionID.String()},
	)
	rr := httptest.NewRecorder()

	h.DeleteSection(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var count int
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM sections WHERE id = $1`, sectionID).Scan(&count))
	assert.Equal(t, 0, count)
}

func TestAdminHandler_ReorderSections(t *testing.T) {
	h, businessStore, _, pool := newTestAdminHandler(t)
	ctx := context.Background()

	b := &models.Business{Name: "Salon", Slug: "salon"}
	require.NoError(t, businessStore.Create(ctx, pool, b))

	pageID := uuid.New()
	_, err := pool.Exec(ctx, `INSERT INTO pages (id, business_id, name) VALUES ($1, $2, 'landing')`, pageID, b.ID)
	require.NoError(t, err)
	a := uuid.New()
	c := uuid.New()
	_, err = pool.Exec(ctx, `INSERT INTO sections (id, page_id, type, content, position) VALUES ($1, $2, 'hero', '{}'::jsonb, 0), ($3, $2, 'about', '{}'::jsonb, 1)`, a, pageID, c)
	require.NoError(t, err)

	body := `{"section_ids":["` + c.String() + `","` + a.String() + `"]}`
	req := withPathParams(
		httptest.NewRequest(http.MethodPut, "/api/admin/business/"+b.ID.String()+"/sections/reorder", bytes.NewBufferString(body)),
		map[string]string{"businessID": b.ID.String()},
	)
	rr := httptest.NewRecorder()

	h.ReorderSections(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var sections []dto.Section
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &sections))
	require.Len(t, sections, 2)
	assert.Equal(t, c, sections[0].ID)
	assert.Equal(t, a, sections[1].ID)
}

func TestAdminHandler_CreateSection_NonOwnerRejected(t *testing.T) {
	h, businessStore, buStore, pool := newTestAdminHandler(t)
	ctx := context.Background()

	b := &models.Business{Name: "Salon", Slug: "salon"}
	require.NoError(t, businessStore.Create(ctx, pool, b))
	require.NoError(t, buStore.Create(ctx, pool, &models.BusinessUser{
		BusinessID: b.ID,
		UserID:     "owner",
		Role:       "admin",
	}))

	guarded := middleware.RequireBusinessAdmin(buStore)(http.HandlerFunc(h.CreateSection))

	body := `{"type":"hero","content":{"en":{"headline":"Hi"}}}`
	req := withPathParams(
		httptest.NewRequest(http.MethodPost, "/api/admin/business/"+b.ID.String()+"/sections", bytes.NewBufferString(body)),
		map[string]string{"businessID": b.ID.String()},
	)
	req = withUser(req, "not-owner", "approved")
	rr := httptest.NewRecorder()

	guarded.ServeHTTP(rr, req)

	require.Equal(t, http.StatusForbidden, rr.Code)
}
