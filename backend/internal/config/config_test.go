package config

import (
	"reflect"
	"testing"
	"time"
)

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

func TestLoadKeycloakDefaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Keycloak.AdminURL != "http://localhost:9090" {
		t.Fatalf("unexpected admin url: %q", cfg.Keycloak.AdminURL)
	}
	if cfg.Keycloak.Realm != "fejd" {
		t.Fatalf("unexpected realm: %q", cfg.Keycloak.Realm)
	}
	want := []string{"salon-mobile", "fejd-frontend"}
	if !reflect.DeepEqual(cfg.Keycloak.Audiences, want) {
		t.Fatalf("unexpected audiences: %v", cfg.Keycloak.Audiences)
	}
	if cfg.Keycloak.AdminClientID != "fejd-admin" {
		t.Fatalf("unexpected admin client id: %q", cfg.Keycloak.AdminClientID)
	}
}

func TestLoadKeycloakAudiencesParsing(t *testing.T) {
	t.Setenv("KEYCLOAK_AUDIENCES", " a, b , c ")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"a", "b", "c"}
	if !reflect.DeepEqual(cfg.Keycloak.Audiences, want) {
		t.Fatalf("unexpected audiences: %v", cfg.Keycloak.Audiences)
	}
}

func TestLoadJobsDefaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Jobs.RunJobs {
		t.Fatal("expected RunJobs default true")
	}
	if cfg.Jobs.InviteExpiryHours != 48 {
		t.Fatalf("unexpected invite expiry hours: %d", cfg.Jobs.InviteExpiryHours)
	}
	if cfg.Jobs.PendingScanInterval != 15*time.Minute {
		t.Fatalf("unexpected pending scan interval: %v", cfg.Jobs.PendingScanInterval)
	}
	if cfg.Jobs.CleanupInterval != time.Hour {
		t.Fatalf("unexpected cleanup interval: %v", cfg.Jobs.CleanupInterval)
	}
}

func TestLoadEmailDefaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Email.SMTPPort != 587 {
		t.Fatalf("unexpected smtp port: %d", cfg.Email.SMTPPort)
	}
	if cfg.Email.SMTPHost != "" || cfg.Email.SMTPUser != "" || cfg.Email.SMTPPass != "" || cfg.Email.SMTPFrom != "" || cfg.Email.SuperadminNotifyEmail != "" {
		t.Fatal("expected empty smtp defaults")
	}
}

func TestLoadRejectsEmptyAudiences(t *testing.T) {
	t.Setenv("KEYCLOAK_AUDIENCES", " , ")

	if _, err := Load(); err == nil {
		t.Fatal("expected error for empty audiences")
	}
}

func TestLoadRejectsNonPositiveInviteExpiry(t *testing.T) {
	t.Setenv("INVITE_EXPIRY_HOURS", "0")

	if _, err := Load(); err == nil {
		t.Fatal("expected error for non-positive invite expiry")
	}
}
