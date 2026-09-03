package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("IMAGE_STORAGE_BACKEND", "minio")
	t.Setenv("MINIO_ENDPOINT", "localhost:9000")
	t.Setenv("MINIO_BUCKET", "bucket")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ImageStorage.Backend != BackendMinio {
		t.Fatalf("expected minio backend, got %q", cfg.ImageStorage.Backend)
	}
	if cfg.ImageStorage.MaxUploadBytes != 10*1024*1024 {
		t.Fatalf("expected default max upload of 10MiB, got %d", cfg.ImageStorage.MaxUploadBytes)
	}
	if cfg.ImageStorage.Minio.Endpoint != "localhost:9000" {
		t.Fatalf("unexpected endpoint: %q", cfg.ImageStorage.Minio.Endpoint)
	}
}

func TestLoadRejectsUnknownBackend(t *testing.T) {
	t.Setenv("IMAGE_STORAGE_BACKEND", "bogus")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for unknown backend")
	}
}

func TestLoadRejectsIncompleteS3(t *testing.T) {
	t.Setenv("IMAGE_STORAGE_BACKEND", "s3")
	t.Setenv("S3_REGION", "eu-west-1")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for incomplete s3 config")
	}
}

func TestLoadS3(t *testing.T) {
	t.Setenv("IMAGE_STORAGE_BACKEND", "s3")
	t.Setenv("S3_REGION", "eu-west-1")
	t.Setenv("S3_BUCKET", "bucket")
	t.Setenv("S3_ACCESS_KEY", "key")
	t.Setenv("S3_SECRET_KEY", "secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ImageStorage.S3.Region != "eu-west-1" {
		t.Fatalf("unexpected region: %q", cfg.ImageStorage.S3.Region)
	}
}

func TestLoadRejectsNonPositiveMaxUpload(t *testing.T) {
	t.Setenv("IMAGE_STORAGE_BACKEND", "minio")
	t.Setenv("MINIO_ENDPOINT", "localhost:9000")
	t.Setenv("MINIO_BUCKET", "bucket")
	t.Setenv("IMAGE_MAX_UPLOAD_MB", "0")

	if _, err := Load(); err == nil {
		t.Fatal("expected error for non-positive max upload size")
	}
}
