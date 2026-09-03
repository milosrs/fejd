package storage

import (
	"fmt"

	"fejd-backend/internal/config"
)

// NewFromConfig builds the object-store backend selected by the configuration.
// The postgres backend stores bytes in the images.data column (handled by the
// image service), so it has no object-store adapter and is not supported here.
func NewFromConfig(cfg config.StorageConfig) (ImageStorage, error) {
	switch cfg.Backend {
	case config.BackendMinio:
		return NewMinioImageStorage(cfg.Minio.Endpoint, cfg.Minio.AccessKey, cfg.Minio.SecretKey, cfg.Minio.Bucket, cfg.Minio.UseSSL)
	case config.BackendS3:
		return NewS3ImageStorage(cfg.S3.Region, cfg.S3.Endpoint, cfg.S3.AccessKey, cfg.S3.SecretKey, cfg.S3.Bucket, cfg.S3.UseSSL)
	default:
		return nil, fmt.Errorf("unsupported object-store backend %q", cfg.Backend)
	}
}
