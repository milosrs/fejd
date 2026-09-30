package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("IMAGE_STORAGE_BACKEND", "seaweedfs")
	t.Setenv("SEAWEEDFS_ENDPOINT", "localhost:8333")
	t.Setenv("SEAWEEDFS_BUCKET", "bucket")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, BackendSeaweedfs, cfg.ImageStorage.Backend)
	assert.Equal(t, int64(10*1024*1024), cfg.ImageStorage.MaxUploadBytes)
	assert.Equal(t, "localhost:8333", cfg.ImageStorage.Seaweedfs.Endpoint)
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
	t.Setenv("S3_FORCE_PATH_STYLE", "true")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "eu-west-1", cfg.ImageStorage.S3.Region)
	assert.True(t, cfg.ImageStorage.S3.ForcePathStyle)
}

func TestLoadRejectsNonPositiveMaxUpload(t *testing.T) {
	t.Setenv("IMAGE_STORAGE_BACKEND", "seaweedfs")
	t.Setenv("SEAWEEDFS_ENDPOINT", "localhost:8333")
	t.Setenv("SEAWEEDFS_BUCKET", "bucket")
	t.Setenv("IMAGE_MAX_UPLOAD_MB", "0")

	_, err := Load()
	require.Error(t, err)
}

func TestLoadKeycloakDefaults(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:9090", cfg.Keycloak.AdminURL)
	assert.Equal(t, "fejd", cfg.Keycloak.Realm)
	assert.Empty(t, cfg.Keycloak.IssuerURL)
	assert.Equal(t, []string{"salon-mobile", "fejd-frontend"}, cfg.Keycloak.Audiences)
	assert.Equal(t, "fejd-admin", cfg.Keycloak.AdminClientID)
}

func TestLoadKeycloakIssuerURL(t *testing.T) {
	t.Setenv("KEYCLOAK_ISSUER_URL", "https://auth.fejd.fyi/realms/fejd")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "https://auth.fejd.fyi/realms/fejd", cfg.Keycloak.IssuerURL)
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
	assert.Equal(t, time.Minute, cfg.Jobs.ReminderScanInterval)
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

func TestLoadCORSParsing(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", " http://localhost:5173, https://www.fejd.com ")
	t.Setenv("CORS_ALLOWED_SUFFIX", "fejd.com")
	t.Setenv("FEJD_DOMAIN", "fejd.com")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, []string{"http://localhost:5173", "https://www.fejd.com"}, cfg.CORS.AllowedOrigins)
	assert.Equal(t, "fejd.com", cfg.CORS.AllowedSuffix)
	assert.Equal(t, "fejd.com", cfg.AppDomain)
}

func TestLoadCORSDefaults(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	assert.Contains(t, cfg.CORS.AllowedOrigins, "http://localhost:5173")
	assert.Contains(t, cfg.CORS.AllowedOrigins, "capacitor://localhost")
	assert.Empty(t, cfg.CORS.AllowedSuffix)
	assert.Empty(t, cfg.AppDomain)
}

func TestLoadAcceptsValidInviteBaseURL(t *testing.T) {
	for _, base := range []string{"https://fejd.fyi", "http://localhost:5173"} {
		t.Run(base, func(t *testing.T) {
			t.Setenv("INVITE_BASE_URL", base)

			cfg, err := Load()
			require.NoError(t, err)
			assert.Equal(t, base, cfg.Jobs.InviteBaseURL)
		})
	}
}

func TestLoadRejectsInvalidInviteBaseURL(t *testing.T) {
	for _, base := range []string{"fejd.fyi", "localhost:5173", "/invite", "ftp://fejd.fyi", "https://"} {
		t.Run(base, func(t *testing.T) {
			t.Setenv("INVITE_BASE_URL", base)

			_, err := Load()
			require.Error(t, err)
		})
	}
}

func TestLoadFirebaseDefaults(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	assert.Empty(t, cfg.Firebase.ServiceAccountJSON)
	assert.Empty(t, cfg.Firebase.ServiceAccountPath)
	assert.True(t, cfg.Firebase.Enabled)
	assert.False(t, cfg.Firebase.Configured())
}

func TestLoadFirebaseInlineServiceAccount(t *testing.T) {
	t.Setenv("FIREBASE_SERVICE_ACCOUNT_JSON", `{"project_id":"fejd-c8cab"}`)

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, `{"project_id":"fejd-c8cab"}`, cfg.Firebase.ServiceAccountJSON)
	assert.True(t, cfg.Firebase.Configured())

	data, err := cfg.Firebase.ServiceAccount()
	require.NoError(t, err)
	assert.JSONEq(t, `{"project_id":"fejd-c8cab"}`, string(data))
}

func TestLoadFirebaseServiceAccountPath(t *testing.T) {
	file := filepath.Join(t.TempDir(), "sa.json")
	require.NoError(t, os.WriteFile(file, []byte(`{"project_id":"fejd-c8cab"}`), 0o600))
	t.Setenv("FIREBASE_SERVICE_ACCOUNT_PATH", file)

	cfg, err := Load()
	require.NoError(t, err)
	assert.True(t, cfg.Firebase.Configured())

	data, err := cfg.Firebase.ServiceAccount()
	require.NoError(t, err)
	assert.JSONEq(t, `{"project_id":"fejd-c8cab"}`, string(data))
}

func TestLoadFirebaseCanBeDisabled(t *testing.T) {
	t.Setenv("FIREBASE_ENABLED", "false")

	cfg, err := Load()
	require.NoError(t, err)
	assert.False(t, cfg.Firebase.Enabled)
}

func TestLoadPublicURLDefaultsToLocalhost(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:8080", cfg.PublicURL)
}

func TestLoadPublicURLFromDomain(t *testing.T) {
	t.Setenv("FEJD_DOMAIN", "fejd.fyi")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "https://fejd.fyi", cfg.PublicURL)
}

func TestLoadPublicURLExplicit(t *testing.T) {
	t.Setenv("PUBLIC_URL", "http://localhost:9090/")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:9090", cfg.PublicURL)
}
