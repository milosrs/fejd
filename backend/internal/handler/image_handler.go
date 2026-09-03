package handler

import (
	"errors"
	"io"
	"net/http"

	"fejd-backend/internal/authutil"
	"fejd-backend/internal/dto"
	"fejd-backend/internal/models"
	"fejd-backend/internal/service"
	"fejd-backend/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ImageHandler struct {
	images       *service.ImageService
	services     *store.ServiceStore
	businessUser *store.BusinessUserStore
}

func NewImageHandler(images *service.ImageService, services *store.ServiceStore, businessUser *store.BusinessUserStore) *ImageHandler {
	return &ImageHandler{
		images:       images,
		services:     services,
		businessUser: businessUser,
	}
}

// UploadBusinessImage godoc
// @Summary      Upload a business image
// @Description  Uploads a business hero/logo/background image (multipart file). Visibility is public.
// @Tags         admin
// @Accept       mpfd
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        purpose formData string false "hero|logo|background (default hero)"
// @Param        file formData file true "Image file"
// @Success      201 {object} dto.Image
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/images [post]
func (h *ImageHandler) UploadBusinessImage(w http.ResponseWriter, r *http.Request) {
	businessID, err := uuid.Parse(chi.URLParam(r, "businessID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errInvalidBusinessID.Error())
		return
	}

	data, contentType, err := h.readImage(w, r)
	if err != nil {
		return
	}

	purpose := r.FormValue("purpose")
	if purpose == "" {
		purpose = "hero"
	}
	switch purpose {
	case "hero", "logo", "background":
	default:
		writeError(w, http.StatusBadRequest, "invalid purpose, must be hero, logo or background")
		return
	}

	img, err := h.images.UploadAndLink(r.Context(), businessID, data, contentType, "business", businessID, purpose, models.VisibilityPublic)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, dto.ImageFromModel(*img))
}

// UploadServiceImage godoc
// @Summary      Upload a service picture
// @Description  Uploads a service picture (multipart file). Visibility is public; services.picture_id is updated.
// @Tags         admin
// @Accept       mpfd
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        serviceID path string true "Service UUID"
// @Param        file formData file true "Image file"
// @Success      201 {object} dto.Image
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/services/{serviceID}/image [post]
func (h *ImageHandler) UploadServiceImage(w http.ResponseWriter, r *http.Request) {
	businessID, err := uuid.Parse(chi.URLParam(r, "businessID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errInvalidBusinessID.Error())
		return
	}

	serviceID, err := uuid.Parse(chi.URLParam(r, "serviceID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errInvalidServiceID.Error())
		return
	}

	svc, err := h.services.GetByID(r.Context(), serviceID)
	if err != nil || svc.BusinessID != businessID {
		writeError(w, http.StatusNotFound, "service not found")
		return
	}

	data, contentType, err := h.readImage(w, r)
	if err != nil {
		return
	}

	img, err := h.images.UploadAndLink(r.Context(), businessID, data, contentType, "service", serviceID, "picture", models.VisibilityPublic)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := h.services.SetPicture(r.Context(), serviceID, businessID, img.ID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, dto.ImageFromModel(*img))
}

// UploadEmployeeImage godoc
// @Summary      Upload an employee avatar
// @Description  Uploads an employee avatar (multipart file). Visibility is private; allowed for the employee or an admin.
// @Tags         admin
// @Accept       mpfd
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        userID path string true "User ID (Keycloak sub)"
// @Param        file formData file true "Image file"
// @Success      201 {object} dto.Image
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/employees/{userID}/image [post]
func (h *ImageHandler) UploadEmployeeImage(w http.ResponseWriter, r *http.Request) {
	businessID, err := uuid.Parse(chi.URLParam(r, "businessID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errInvalidBusinessID.Error())
		return
	}

	targetUserID := chi.URLParam(r, "userID")

	callerID, err := authutil.GetUserID(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	if callerID != targetUserID {
		isAdmin, err := h.businessUser.IsAdmin(r.Context(), businessID, callerID)
		if err != nil || !isAdmin {
			writeError(w, http.StatusForbidden, "forbidden")
			return
		}
	}

	bu, err := h.businessUser.GetByBusinessAndUser(r.Context(), businessID, targetUserID)
	if err != nil {
		writeError(w, http.StatusNotFound, "employee not found")
		return
	}

	data, contentType, err := h.readImage(w, r)
	if err != nil {
		return
	}

	img, err := h.images.UploadAndLink(r.Context(), businessID, data, contentType, "business_user", bu.ID, "avatar", models.VisibilityPrivate)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, dto.ImageFromModel(*img))
}

// DeleteImage godoc
// @Summary      Delete an image
// @Description  Deletes an image. Admins delete the whole image; an employee may remove only their own avatar link.
// @Tags         admin
// @Produce      json
// @Param        businessID path string true "Business UUID"
// @Param        imageID path string true "Image UUID"
// @Success      200 {object} MessageResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/business/{businessID}/images/{imageID} [delete]
func (h *ImageHandler) DeleteImage(w http.ResponseWriter, r *http.Request) {
	businessID, err := uuid.Parse(chi.URLParam(r, "businessID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errInvalidBusinessID.Error())
		return
	}

	imageID, err := uuid.Parse(chi.URLParam(r, "imageID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid image ID")
		return
	}

	callerID, err := authutil.GetUserID(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	img, links, err := h.images.GetWithLinks(r.Context(), imageID)
	if err != nil {
		writeError(w, http.StatusNotFound, "image not found")
		return
	}
	if img.BusinessID != businessID {
		writeError(w, http.StatusNotFound, "image not found")
		return
	}

	if isAdmin, _ := h.businessUser.IsAdmin(r.Context(), businessID, callerID); isAdmin {
		if err := h.images.Delete(r.Context(), img); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, MessageResponse{Message: "image deleted"})
		return
	}

	if linkID, owns := h.images.OwnsAvatarLink(r.Context(), callerID, links); owns {
		if err := h.images.DeleteImageScoped(r.Context(), img, linkID); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, MessageResponse{Message: "image deleted"})
		return
	}

	writeError(w, http.StatusForbidden, "forbidden")
}

// GetImage godoc
// @Summary      Retrieve an image
// @Description  Public images are served directly; private images require auth and access (streamed through the API).
// @Tags         public
// @Produce      image/*
// @Param        imageID path string true "Image UUID"
// @Success      200 {file} binary
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Router       /api/images/{imageID} [get]
func (h *ImageHandler) GetImage(w http.ResponseWriter, r *http.Request) {
	imageID, err := uuid.Parse(chi.URLParam(r, "imageID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid image ID")
		return
	}

	img, links, err := h.images.GetWithLinks(r.Context(), imageID)
	if err != nil {
		writeError(w, http.StatusNotFound, "image not found")
		return
	}

	if h.isPublic(links) {
		h.servePublic(w, r, img)
		return
	}

	ar := authutil.AuthStatusOf(r)
	switch ar.Status {
	case authutil.AuthInvalid:
		writeError(w, http.StatusUnauthorized, "invalid token")
		return
	case authutil.AuthAnonymous:
		writeError(w, http.StatusNotFound, "image not found")
		return
	}

	if !h.images.UserCanAccess(r.Context(), ar.UserID, img, links) {
		writeError(w, http.StatusNotFound, "image not found")
		return
	}

	h.serveProxied(w, r, img)
}

func (h *ImageHandler) readImage(w http.ResponseWriter, r *http.Request) ([]byte, string, error) {
	maxBytes := h.images.MaxUploadBytes()
	if err := r.ParseMultipartForm(maxBytes); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "image exceeds maximum upload size")
			return nil, "", err
		}
		writeError(w, http.StatusBadRequest, "invalid multipart form")
		return nil, "", err
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing file field")
		return nil, "", err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read file")
		return nil, "", err
	}
	if len(data) == 0 {
		writeError(w, http.StatusBadRequest, "empty file")
		return nil, "", err
	}

	if int64(len(data)) > maxBytes {
		writeError(w, http.StatusRequestEntityTooLarge, "image exceeds maximum upload size")
		return nil, "", errors.New("image exceeds maximum upload size")
	}

	contentType := http.DetectContentType(data)
	switch contentType {
	case "image/jpeg", "image/png", "image/webp", "image/gif":
	default:
		writeError(w, http.StatusBadRequest, "unsupported image type: "+contentType)
		return nil, "", err
	}

	return data, contentType, nil
}

func (h *ImageHandler) isPublic(links []models.ImageLink) bool {
	for _, l := range links {
		if l.Visibility == models.VisibilityPublic {
			return true
		}
	}
	return false
}

func (h *ImageHandler) servePublic(w http.ResponseWriter, r *http.Request, img *models.Image) {
	if url := h.images.PublicURL(r.Context(), img); url != "" {
		http.Redirect(w, r, url, http.StatusFound)
		return
	}
	h.serveProxied(w, r, img)
}

func (h *ImageHandler) serveProxied(w http.ResponseWriter, r *http.Request, img *models.Image) {
	rc, contentType, err := h.images.Serve(r.Context(), img)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load image")
		return
	}
	defer rc.Close()
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, rc)
}
