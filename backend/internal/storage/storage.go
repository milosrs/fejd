package storage

import (
	"context"
	"errors"
)

// ErrNotImplemented is returned by backends that are declared but not yet
// wired up (the Postgres adapter is a stub for now).
var ErrNotImplemented = errors.New("storage backend not implemented")

// ImageStorage abstracts where image bytes live. MinIO is the current backend;
// Postgres is the planned switch target. The images table records which backend
// owns each image (storage column) together with the reference (object_key /
// data / url).
type ImageStorage interface {
	Put(ctx context.Context, key string, data []byte, contentType string) error
	Get(ctx context.Context, key string) ([]byte, string, error)
	Delete(ctx context.Context, key string) error
	URL(ctx context.Context, key string) string
}
