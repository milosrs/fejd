package config

import (
	"fmt"
	"os"
	"strconv"
)

// Backend selects where image bytes are stored.
type Backend string

const (
	BackendPostgres Backend = "postgres"
	BackendMinio    Backend = "minio"
	BackendS3       Backend = "s3"
)

// ObjectStoreConfig holds connection settings for an S3-compatible object
// store (MinIO or a custom endpoint).
type ObjectStoreConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

// S3Config holds connection settings for AWS S3 (or an S3-compatible
// endpoint via Endpoint).
type S3Config struct {
	Region    string
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

// StorageConfig selects the active image backend and its settings.
type StorageConfig struct {
	Backend        Backend
	Minio          ObjectStoreConfig
	S3             S3Config
	MaxUploadBytes int64
}

// Config is the typed view of the service environment configuration.
type Config struct {
	ImageStorage StorageConfig
}

// Load reads configuration from environment variables, applies defaults, and
// validates the result.
func Load() (*Config, error) {
	backend := Backend(getEnv("IMAGE_STORAGE_BACKEND", "minio"))

	cfg := &Config{
		ImageStorage: StorageConfig{
			Backend: backend,
			Minio: ObjectStoreConfig{
				Endpoint:  getEnv("MINIO_ENDPOINT", "minio:9000"),
				AccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
				SecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
				Bucket:    getEnv("MINIO_BUCKET", "fejd-images"),
				UseSSL:    getEnvBool("MINIO_USE_SSL", false),
			},
			S3: S3Config{
				Region:    getEnv("S3_REGION", ""),
				Endpoint:  getEnv("S3_ENDPOINT", ""),
				AccessKey: getEnv("S3_ACCESS_KEY", ""),
				SecretKey: getEnv("S3_SECRET_KEY", ""),
				Bucket:    getEnv("S3_BUCKET", ""),
				UseSSL:    getEnvBool("S3_USE_SSL", true),
			},
			MaxUploadBytes: getEnvInt("IMAGE_MAX_UPLOAD_MB", 10) * 1024 * 1024,
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	switch c.ImageStorage.Backend {
	case BackendPostgres, BackendMinio, BackendS3:
	default:
		return fmt.Errorf("invalid IMAGE_STORAGE_BACKEND %q: must be postgres, minio or s3", c.ImageStorage.Backend)
	}

	if c.ImageStorage.MaxUploadBytes <= 0 {
		return fmt.Errorf("IMAGE_MAX_UPLOAD_MB must be greater than zero")
	}

	switch c.ImageStorage.Backend {
	case BackendMinio:
		if c.ImageStorage.Minio.Endpoint == "" || c.ImageStorage.Minio.Bucket == "" {
			return fmt.Errorf("minio backend requires MINIO_ENDPOINT and MINIO_BUCKET")
		}
	case BackendS3:
		s3 := c.ImageStorage.S3
		if s3.Region == "" || s3.Bucket == "" || s3.AccessKey == "" || s3.SecretKey == "" {
			return fmt.Errorf("s3 backend requires S3_REGION, S3_BUCKET, S3_ACCESS_KEY and S3_SECRET_KEY")
		}
	}

	return nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

func getEnvInt(key string, def int64) int64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return def
}
