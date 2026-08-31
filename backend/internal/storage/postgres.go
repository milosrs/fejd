package storage

import (
	"context"
)

// PostgresImageStorage is the planned storage backend for switching image bytes
// into the database (images.data). It is intentionally a stub for now; wire it
// up when the switch from MinIO is required. The images table already carries
// the storage='postgres' discriminator and a BYTEA data column for it.
type PostgresImageStorage struct{}

func NewPostgresImageStorage() *PostgresImageStorage {
	return &PostgresImageStorage{}
}

func (s *PostgresImageStorage) Put(ctx context.Context, key string, data []byte, contentType string) error {
	return ErrNotImplemented
}

func (s *PostgresImageStorage) Get(ctx context.Context, key string) ([]byte, string, error) {
	return nil, "", ErrNotImplemented
}

func (s *PostgresImageStorage) Delete(ctx context.Context, key string) error {
	return ErrNotImplemented
}

func (s *PostgresImageStorage) URL(ctx context.Context, key string) string {
	return ""
}
