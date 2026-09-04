package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("IMAGE_STORAGE_BACKEND", "minio")
	t.Setenv("MINIO_ENDPOINT", "localhost:9000")
	t.Setenv("MINIO_BUCKET", "bucket")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, BackendMinio, cfg.ImageStorage.Backend)
	assert.Equal(t, int64(10*1024*1024), cfg.ImageStorage.MaxUploadBytes)
	assert.Equal(t, "localhost:9000", cfg.ImageStorage.Minio.Endpoint)
}

func TestLoadRejectsUnknownBackend(t *testing.T) {
	t.Setenv("IMAGE_STORAGE_BACKEND", "bogus")
	_, err := Load()
	require.Error(t, err)
}

func TestLoadRejectsIncompleteS3(t *testing.T) {
	t.Setenv("IMAGE_STORAGE_BACKEND", "s3")
	t.Setenv("S3_REGION", "eu-west-1")
	_, err := Load()
	require.Error(t, err)
}

func TestLoadS3(t *testing.T) {
	t.Setenv("IMAGE_STORAGE_BACKEND", "s3")
	t.Setenv("S3_REGION", "eu-west-1")
	t.Setenv("S3_BUCKET", "bucket")
	t.Setenv("S3_ACCESS_KEY", "key")
	t.Setenv("S3_SECRET_KEY", "secret")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "eu-west-1", cfg.ImageStorage.S3.Region)
}

func TestLoadRejectsNonPositiveMaxUpload(t *testing.T) {
	t.Setenv("IMAGE_STORAGE_BACKEND", "minio")
	t.Setenv("MINIO_ENDPOINT", "localhost:9000")
	t.Setenv("MINIO_BUCKET", "bucket")
	t.Setenv("IMAGE_MAX_UPLOAD_MB", "0")

	_, err := Load()
	require.Error(t, err)
}

func TestLoadKeycloakDefaults(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:9090", cfg.Keycloak.AdminURL)
	assert.Equal(t, "fejd", cfg.Keycloak.Realm)
	assert.Equal(t, []string{"salon-mobile", "fejd-frontend"}, cfg.Keycloak.Audiences)
	assert.Equal(t, "fejd-admin", cfg.Keycloak.AdminClientID)
}

func TestLoadKeycloakAudiencesParsing(t *testing.T) {
	t.Setenv("KEYCLOAK_AUDIENCES", " a, b , c ")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, cfg.Keycloak.Audiences)
}

func TestLoadJobsDefaults(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	assert.True(t, cfg.Jobs.RunJobs)
	assert.Equal(t, 48, cfg.Jobs.InviteExpiryHours)
	assert.Equal(t, 15*time.Minute, cfg.Jobs.PendingScanInterval)
	assert.Equal(t, time.Hour, cfg.Jobs.CleanupInterval)
}

func TestLoadEmailDefaults(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, 587, cfg.Email.SMTPPort)
	assert.Empty(t, cfg.Email.SMTPHost)
	assert.Empty(t, cfg.Email.SMTPUser)
	assert.Empty(t, cfg.Email.SMTPPass)
	assert.Empty(t, cfg.Email.SMTPFrom)
	assert.Empty(t, cfg.Email.SuperadminNotifyEmail)
}

func TestLoadRejectsEmptyAudiences(t *testing.T) {
	t.Setenv("KEYCLOAK_AUDIENCES", " , ")

	_, err := Load()
	require.Error(t, err)
}

func TestLoadRejectsNonPositiveInviteExpiry(t *testing.T) {
	t.Setenv("INVITE_EXPIRY_HOURS", "0")

	_, err := Load()
	require.Error(t, err)
}
