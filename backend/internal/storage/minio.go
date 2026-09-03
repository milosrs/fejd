package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// objectStore implements the ImageStorage interface against an S3-compatible
// object store (MinIO, AWS S3, etc.) via minio-go.
type objectStore struct {
	client *minio.Client
	bucket string
}

func (s *objectStore) Put(ctx context.Context, key string, data []byte, contentType string) error {
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	_, err := s.client.PutObject(ctx, s.bucket, key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to put object: %w", err)
	}
	return nil
}

func (s *objectStore) Open(ctx context.Context, key string) (io.ReadCloser, string, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", fmt.Errorf("failed to get object: %w", err)
	}

	stat, err := obj.Stat()
	if err != nil {
		obj.Close()
		return nil, "", fmt.Errorf("failed to stat object: %w", err)
	}

	return obj, stat.ContentType, nil
}

func (s *objectStore) Delete(ctx context.Context, key string) error {
	if err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("failed to remove object: %w", err)
	}
	return nil
}

func (s *objectStore) URL(ctx context.Context, key string) string {
	u, err := s.client.PresignedGetObject(ctx, s.bucket, key, time.Hour, nil)
	if err != nil {
		return ""
	}
	return u.String()
}

type MinioImageStorage struct {
	*objectStore
}

func NewMinioImageStorage(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*MinioImageStorage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	return &MinioImageStorage{objectStore: &objectStore{client: client, bucket: bucket}}, nil
}
