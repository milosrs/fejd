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
	h := NewBusinessHandler(businessStore, buStore, serviceStore, pageStore, sectionStore, imageLinkStore, nil)
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
