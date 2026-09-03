package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"

	"fejd-backend/internal/config"
	"fejd-backend/internal/db"
	"fejd-backend/internal/models"
	"fejd-backend/internal/storage"
	"fejd-backend/internal/store"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ImageService orchestrates image byte storage (object store or postgres) with
// image metadata and polymorphic linking. The active backend is chosen by
// config.ImageStorage.Backend; object stores go through storage.ImageStorage
// while postgres bytes live inline in the images.data column.
type ImageService struct {
	cfg           config.StorageConfig
	storage       storage.ImageStorage
	images        *store.ImageStore
	links         *store.ImageLinkStore
	businessUsers *store.BusinessUserStore
	pool          *pgxpool.Pool
}

func NewImageService(
	cfg config.StorageConfig,
	st storage.ImageStorage,
	images *store.ImageStore,
	links *store.ImageLinkStore,
	businessUsers *store.BusinessUserStore,
	pool *pgxpool.Pool,
) *ImageService {
	return &ImageService{
		cfg:           cfg,
		storage:       st,
		images:        images,
		links:         links,
		businessUsers: businessUsers,
		pool:          pool,
	}
}

func (s *ImageService) MaxUploadBytes() int64 {
	return s.cfg.MaxUploadBytes
}

// UploadAndLink stores image bytes and links them to an entity. For singleton
// purposes (hero/logo/background/avatar/picture) an existing link is replaced
// in place; otherwise a new link is appended.
func (s *ImageService) UploadAndLink(
	ctx context.Context,
	businessID uuid.UUID,
	data []byte,
	contentType string,
	entityType string,
	entityID uuid.UUID,
	purpose string,
	visibility models.Visibility,
) (*models.Image, error) {
	if int64(len(data)) > s.cfg.MaxUploadBytes {
		return nil, fmt.Errorf("image exceeds maximum upload size of %d bytes", s.cfg.MaxUploadBytes)
	}

	img := &models.Image{
		ID:          uuid.New(),
		BusinessID:  businessID,
		ContentType: contentType,
	}

	objectKey := ""
	switch s.cfg.Backend {
	case config.BackendPostgres:
		img.Storage = string(config.BackendPostgres)
		img.Data = data
	default:
		if s.storage == nil {
			return nil, fmt.Errorf("object storage is not configured")
		}
		objectKey = fmt.Sprintf("%s/%s", businessID.String(), uuid.NewString())
		img.Storage = string(s.cfg.Backend)
		img.ObjectKey = objectKey
		if err := s.storage.Put(ctx, objectKey, data, contentType); err != nil {
			return nil, fmt.Errorf("failed to store image bytes: %w", err)
		}
	}

	oldImageID, err := s.upsertLink(ctx, img, entityType, entityID, purpose, visibility)
	if err != nil {
		if objectKey != "" {
			s.removeObject(ctx, objectKey)
		}
		return nil, err
	}

	if oldImageID != uuid.Nil {
		if err := s.deleteImageIfOrphaned(ctx, oldImageID); err != nil {
			log.Printf("[images] failed to clean up replaced image %s: %v", oldImageID, err)
		}
	}

	return img, nil
}

func (s *ImageService) upsertLink(ctx context.Context, img *models.Image, entityType string, entityID uuid.UUID, purpose string, visibility models.Visibility) (uuid.UUID, error) {
	var oldImageID uuid.UUID
	var err error

	// Two attempts: the FOR UPDATE guards the existing-row case, and the
	// partial unique index guards the create-first race; on a unique
	// violation we retry once, which lands on the update path.
	for attempt := 0; attempt < 2; attempt++ {
		oldImageID, err = s.tryUpsert(ctx, img, entityType, entityID, purpose, visibility)
		if err == nil {
			return oldImageID, nil
		}
		if !isUniqueViolation(err) {
			return uuid.Nil, err
		}
	}
	return uuid.Nil, err
}

func (s *ImageService) tryUpsert(ctx context.Context, img *models.Image, entityType string, entityID uuid.UUID, purpose string, visibility models.Visibility) (uuid.UUID, error) {
	var oldImageID uuid.UUID

	err := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		if err := s.images.Create(ctx, tx, img); err != nil {
			return fmt.Errorf("failed to create image: %w", err)
		}

		existing, err := s.links.GetByEntityPurposeForUpdate(ctx, tx, entityType, entityID, purpose)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("failed to look up link: %w", err)
		}

		if existing != nil {
			oldImageID = existing.ImageID
			return s.links.UpdateImageID(ctx, tx, existing.ID, img.ID)
		}

		link := &models.ImageLink{
			ImageID:    img.ID,
			EntityType: entityType,
			EntityID:   entityID,
			Purpose:    purpose,
			Visibility: visibility,
		}
		return s.links.Create(ctx, tx, link)
	})
	if err != nil {
		return uuid.Nil, err
	}

	return oldImageID, nil
}

// GetWithLinks returns an image and all of its links. An image with no links
// is treated as private by callers.
func (s *ImageService) GetWithLinks(ctx context.Context, imageID uuid.UUID) (*models.Image, []models.ImageLink, error) {
	img, err := s.images.GetByID(ctx, imageID)
	if err != nil {
		return nil, nil, err
	}
	links, err := s.links.ListByImage(ctx, imageID)
	if err != nil {
		return nil, nil, err
	}
	return img, links, nil
}

// Serve opens the image bytes for streaming through the API. The caller must
// close the returned reader.
func (s *ImageService) Serve(ctx context.Context, img *models.Image) (io.ReadCloser, string, error) {
	if img.Storage == string(config.BackendPostgres) {
		return io.NopCloser(bytes.NewReader(img.Data)), img.ContentType, nil
	}
	if s.storage == nil {
		return nil, "", fmt.Errorf("object storage is not configured")
	}
	return s.storage.Open(ctx, img.ObjectKey)
}

// PublicURL returns a presigned URL for public object-store images, or "" when
// the image must be streamed through the API (postgres backend or no storage).
func (s *ImageService) PublicURL(ctx context.Context, img *models.Image) string {
	if img.Storage == string(config.BackendPostgres) || s.storage == nil {
		return ""
	}
	return s.storage.URL(ctx, img.ObjectKey)
}

// UserCanAccess reports whether the user may view a private image: the
// business admin always can, and the employee may view their own avatar.
func (s *ImageService) UserCanAccess(ctx context.Context, userID string, img *models.Image, links []models.ImageLink) bool {
	if userID == "" {
		return false
	}

	if isAdmin, err := s.businessUsers.IsAdmin(ctx, img.BusinessID, userID); err == nil && isAdmin {
		return true
	}

	_, owns := s.OwnsAvatarLink(ctx, userID, links)
	return owns
}

// OwnsAvatarLink returns the link ID of the caller's own avatar link, if any.
func (s *ImageService) OwnsAvatarLink(ctx context.Context, userID string, links []models.ImageLink) (uuid.UUID, bool) {
	if userID == "" {
		return uuid.Nil, false
	}
	for _, l := range links {
		if l.EntityType == "business_user" && l.Purpose == "avatar" {
			if bu, err := s.businessUsers.GetByID(ctx, l.EntityID); err == nil && bu.UserID == userID {
				return l.ID, true
			}
		}
	}
	return uuid.Nil, false
}

// Delete removes an image and its links (cascading). Object-store bytes are
// deleted first; on failure the DB row is kept so the object stays traceable.
func (s *ImageService) Delete(ctx context.Context, img *models.Image) error {
	if img.Storage != string(config.BackendPostgres) {
		if s.storage == nil {
			return fmt.Errorf("object storage is not configured")
		}
		if err := s.removeObject(ctx, img.ObjectKey); err != nil {
			return err
		}
	}
	return s.images.Delete(ctx, s.pool, img.ID)
}

// DeleteImageScoped removes only the link the caller is authorized against
// when the image is still linked elsewhere; the last link falls through to a
// full image delete.
func (s *ImageService) DeleteImageScoped(ctx context.Context, img *models.Image, authorizedLinkID uuid.UUID) error {
	count, err := s.links.CountByImage(ctx, img.ID)
	if err != nil {
		return err
	}
	if count <= 1 {
		return s.Delete(ctx, img)
	}
	return s.links.DeleteByID(ctx, s.pool, authorizedLinkID)
}

// UnlinkAndMaybeDeleteAll removes every link for an entity and deletes images
// that are left with no remaining links.
func (s *ImageService) UnlinkAndMaybeDeleteAll(ctx context.Context, entityType string, entityID uuid.UUID) error {
	links, err := s.links.ListByEntity(ctx, entityType, entityID)
	if err != nil {
		return err
	}

	for _, l := range links {
		if err := s.links.DeleteByID(ctx, s.pool, l.ID); err != nil {
			return fmt.Errorf("failed to delete link: %w", err)
		}
	}

	for _, l := range links {
		if err := s.deleteImageIfOrphaned(ctx, l.ImageID); err != nil {
			log.Printf("[images] failed to clean up orphaned image %s: %v", l.ImageID, err)
		}
	}
	return nil
}

func (s *ImageService) deleteImageIfOrphaned(ctx context.Context, imageID uuid.UUID) error {
	count, err := s.links.CountByImage(ctx, imageID)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	img, err := s.images.GetByID(ctx, imageID)
	if err != nil {
		return err
	}
	return s.Delete(ctx, img)
}

func (s *ImageService) removeObject(ctx context.Context, key string) error {
	if err := s.storage.Delete(ctx, key); err != nil {
		log.Printf("[images] orphaned object %q could not be deleted: %v", key, err)
		return err
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
