package storage

import (
	"context"
	"io"
)

// ImageStorage abstracts where image bytes live. S3-compatible object stores
// are accessed through minio-go. The postgres backend stores bytes in the
// images.data column and is handled directly by the image service, so it has
// no adapter here.
type ImageStorage interface {
	Put(ctx context.Context, key string, data []byte, contentType string) error
	Open(ctx context.Context, key string) (io.ReadCloser, string, error)
	Delete(ctx context.Context, key string) error
	URL(ctx context.Context, key string) string
}
