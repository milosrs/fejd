package handler

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"fejd-backend/internal/config"
	"fejd-backend/internal/dto"
	"fejd-backend/internal/models"
	"fejd-backend/internal/service"
	"fejd-backend/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var tinyPNG = mustDecodePNG()

func mustDecodePNG() []byte {
	b, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII=")
	if err != nil {
		panic(err)
	}
	return b
}

func newTestImageHandler(t *testing.T) (*ImageHandler, *store.UserStore) {
	pool := setupHandlerTestDB(t)
	imageStore := store.NewImageStore(pool)
	imageLinkStore := store.NewImageLinkStore(pool)
	buStore := store.NewBusinessUserStore(pool)
	userStore := store.NewUserStore(pool)
	serviceStore := store.NewServiceStore(pool)

	cfg := config.StorageConfig{Backend: config.BackendPostgres, MaxUploadBytes: 10 * 1024 * 1024}
	imageService := service.NewImageService(cfg, nil, imageStore, imageLinkStore, buStore, userStore, pool)

	return NewImageHandler(imageService, serviceStore, buStore), userStore
}

func multipartImageRequest(t *testing.T, data []byte) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("file", "avatar.png")
	require.NoError(t, err)
	_, err = fw.Write(data)
	require.NoError(t, err)
	require.NoError(t, w.Close())

	req := httptest.NewRequest(http.MethodPost, "/api/me/avatar", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

func withImageID(r *http.Request, imageID string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("imageID", imageID)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestImageHandler_UploadAvatar(t *testing.T) {
	h, userStore := newTestImageHandler(t)
	ctx := context.Background()

	require.NoError(t, userStore.Upsert(ctx, &models.User{ID: "user-1", DisplayName: "Sam"}))

	req := withUser(multipartImageRequest(t, tinyPNG), "user-1", "approved")
	rr := httptest.NewRecorder()

	h.UploadAvatar(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
	var img dto.Image
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &img))
	assert.NotEmpty(t, img.ID)
	assert.Equal(t, "/api/images/"+img.ID.String(), img.URL)

	u, err := userStore.GetByID(ctx, "user-1")
	require.NoError(t, err)
	require.NotNil(t, u.AvatarID)
	assert.Equal(t, img.ID, *u.AvatarID)
}

func TestImageHandler_UploadAvatar_RequiresAuth(t *testing.T) {
	h, _ := newTestImageHandler(t)

	req := multipartImageRequest(t, tinyPNG)
	rr := httptest.NewRecorder()

	h.UploadAvatar(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestImageHandler_UploadAvatar_RejectsNonImage(t *testing.T) {
	h, _ := newTestImageHandler(t)

	req := withUser(multipartImageRequest(t, []byte("not an image")), "user-1", "approved")
	rr := httptest.NewRecorder()

	h.UploadAvatar(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestImageHandler_GetImage_ProfilePictureIsPublic(t *testing.T) {
	h, userStore := newTestImageHandler(t)
	ctx := context.Background()

	require.NoError(t, userStore.Upsert(ctx, &models.User{ID: "user-1", DisplayName: "Sam"}))

	req := withUser(multipartImageRequest(t, tinyPNG), "user-1", "approved")
	rr := httptest.NewRecorder()
	h.UploadAvatar(rr, req)
	require.Equal(t, http.StatusCreated, rr.Code)

	var img dto.Image
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &img))

	// The owner can fetch their own avatar.
	getReq := withImageID(withUser(httptest.NewRequest(http.MethodGet, "/api/images/"+img.ID.String(), nil), "user-1", "approved"), img.ID.String())
	getRR := httptest.NewRecorder()
	h.GetImage(getRR, getReq)
	assert.Equal(t, http.StatusOK, getRR.Code)

	// Profile pictures are served publicly (no image_links row, referenced via
	// users.avatar_id), so they render in plain <img> tags without auth.
	anonReq := withImageID(httptest.NewRequest(http.MethodGet, "/api/images/"+img.ID.String(), nil), img.ID.String())
	anonRR := httptest.NewRecorder()
	h.GetImage(anonRR, anonReq)
	assert.Equal(t, http.StatusOK, anonRR.Code)
}

func TestImageHandler_UploadAvatar_ReplacesPrevious(t *testing.T) {
	h, userStore := newTestImageHandler(t)
	ctx := context.Background()

	require.NoError(t, userStore.Upsert(ctx, &models.User{ID: "user-1", DisplayName: "Sam"}))

	first := withUser(multipartImageRequest(t, tinyPNG), "user-1", "approved")
	rr1 := httptest.NewRecorder()
	h.UploadAvatar(rr1, first)
	require.Equal(t, http.StatusCreated, rr1.Code)
	var firstImg dto.Image
	require.NoError(t, json.Unmarshal(rr1.Body.Bytes(), &firstImg))

	second := withUser(multipartImageRequest(t, tinyPNG), "user-1", "approved")
	rr2 := httptest.NewRecorder()
	h.UploadAvatar(rr2, second)
	require.Equal(t, http.StatusCreated, rr2.Code)
	var secondImg dto.Image
	require.NoError(t, json.Unmarshal(rr2.Body.Bytes(), &secondImg))

	require.NotEqual(t, firstImg.ID, secondImg.ID)

	u, err := userStore.GetByID(ctx, "user-1")
	require.NoError(t, err)
	require.NotNil(t, u.AvatarID)
	assert.Equal(t, secondImg.ID, *u.AvatarID)
	assert.NotEqual(t, uuid.Nil, secondImg.ID)
}
