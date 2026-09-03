package storage

import (
	"context"
	"io"
)

// ImageStorage abstracts where image bytes live. MinIO and S3 (via minio-go)
// are the object-store backends. The postgres backend stores bytes in the
// images.data column and is handled directly by the image service, so it has
// no adapter here.
type ImageStorage interface {
	Put(ctx context.Context, key string, data []byte, contentType string) error
	Open(ctx context.Context, key string) (io.ReadCloser, string, error)
	Delete(ctx context.Context, key string) error
	URL(ctx context.Context, key string) string
}
