package storage

import (
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3ImageStorage struct {
	*objectStore
}

func NewS3ImageStorage(region, endpoint, accessKey, secretKey, bucket string, useSSL bool) (*S3ImageStorage, error) {
	if endpoint == "" {
		endpoint = fmt.Sprintf("s3.%s.amazonaws.com", region)
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create s3 client: %w", err)
	}

	return &S3ImageStorage{objectStore: &objectStore{client: client, bucket: bucket}}, nil
}
