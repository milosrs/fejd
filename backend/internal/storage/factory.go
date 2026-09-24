package storage

import (
	"context"
	"fmt"

	"fejd-backend/internal/config"
	"fejd-backend/internal/retry"
)

// NewFromConfig builds the object-store backend selected by the configuration.
// The postgres backend stores bytes in the images.data column (handled by the
// image service), so it has no object-store adapter and is not supported here.
func NewFromConfig(ctx context.Context, cfg config.StorageConfig) (ImageStorage, error) {
	var storage ImageStorage
	var objectStore *objectStore
	var err error

	switch cfg.Backend {
	case config.BackendSeaweedfs:
		var seaweedfsStorage *SeaweedfsImageStorage
		seaweedfsStorage, err = NewSeaweedfsImageStorage(cfg.Seaweedfs.Endpoint, cfg.Seaweedfs.AccessKey, cfg.Seaweedfs.SecretKey, cfg.Seaweedfs.Bucket, cfg.Seaweedfs.UseSSL)
		storage = seaweedfsStorage
		if seaweedfsStorage != nil {
			objectStore = seaweedfsStorage.objectStore
		}
	case config.BackendS3:
		var s3Storage *S3ImageStorage
		s3Storage, err = NewS3ImageStorage(cfg.S3.Region, cfg.S3.Endpoint, cfg.S3.AccessKey, cfg.S3.SecretKey, cfg.S3.Bucket, cfg.S3.UseSSL, cfg.S3.ForcePathStyle)
		storage = s3Storage
		if s3Storage != nil {
			objectStore = s3Storage.objectStore
		}
	default:
		return nil, fmt.Errorf("unsupported object-store backend %q", cfg.Backend)
	}
	if err != nil {
		return nil, err
	}

	err = retry.Do(ctx, "object store", func() error {
		_, err := objectStore.client.BucketExists(ctx, objectStore.bucket)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to object store: %w", err)
	}
	return storage, nil
}
