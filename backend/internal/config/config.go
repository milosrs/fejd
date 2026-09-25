package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Backend selects where image bytes are stored.
type Backend string

const (
	BackendPostgres  Backend = "postgres"
	BackendSeaweedfs Backend = "seaweedfs"
	BackendS3        Backend = "s3"
)

// ObjectStoreConfig holds connection settings for an S3-compatible object
// store (SeaweedFS or a custom endpoint).
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
	Region         string
	Endpoint       string
	AccessKey      string
	SecretKey      string
	Bucket         string
	UseSSL         bool
	ForcePathStyle bool
}

// StorageConfig selects the active image backend and its settings.
type StorageConfig struct {
	Backend        Backend
	Seaweedfs      ObjectStoreConfig
	S3             S3Config
	MaxUploadBytes int64
}

// KeycloakConfig holds connection and admin settings for Keycloak.
type KeycloakConfig struct {
	AdminURL string
	Realm    string
	// IssuerURL is the realm issuer used to validate the JWT "iss" claim. It
	// must match the public hostname Keycloak uses to mint tokens (e.g.
	// https://auth.fejd.fyi/realms/fejd), which can differ from AdminURL when
	// Keycloak sits behind a proxy. Empty falls back to AdminURL-derived URL.
	IssuerURL         string
	Audiences         []string
	AdminClientID     string
	AdminClientSecret string
}

// CORSConfig controls cross-origin access to the API.
type CORSConfig struct {
	// AllowedOrigins are exact origins permitted to call the API (local dev
	// servers and the native app webview origins). Salon subdomains are covered
	// separately by AllowedSuffix.
	AllowedOrigins []string
	// AllowedSuffix, when non-empty, permits any origin whose host is equal to
	// or a subdomain of this suffix (e.g. "fejd.com" allows
	// "https://dragicevic.fejd.com").
	AllowedSuffix string
}

// JobsConfig holds scheduled-job settings.
type JobsConfig struct {
	PendingScanInterval time.Duration
	CleanupInterval     time.Duration
	RunJobs             bool
	InviteExpiryHours   int
	InviteRedirectURI   string
	InviteBaseURL       string
}

// EmailConfig holds SMTP settings and the superadmin notification target.
type EmailConfig struct {
	SMTPHost              string
	SMTPPort              int
	SMTPUser              string
	SMTPPass              string
	SMTPFrom              string
	SuperadminNotifyEmail string
}

// Config is the typed view of the service environment configuration.
type Config struct {
	ImageStorage StorageConfig
	Keycloak     KeycloakConfig
	Jobs         JobsConfig
	Email        EmailConfig
	CORS         CORSConfig
	AppDomain    string
}

// Load reads configuration from environment variables, applies defaults, and
// validates the result.
func Load() (*Config, error) {
	backend := Backend(getEnv("IMAGE_STORAGE_BACKEND", "seaweedfs"))

	cfg := &Config{
		ImageStorage: StorageConfig{
			Backend: backend,
			Seaweedfs: ObjectStoreConfig{
				Endpoint:  getEnv("SEAWEEDFS_ENDPOINT", "localhost:8333"),
				AccessKey: getEnv("SEAWEEDFS_ACCESS_KEY", "fejd_dev"),
				SecretKey: getEnv("SEAWEEDFS_SECRET_KEY", "fejd_dev_secret"),
				Bucket:    getEnv("SEAWEEDFS_BUCKET", "fejd-images"),
				UseSSL:    getEnvBool("SEAWEEDFS_USE_SSL", false),
			},
			S3: S3Config{
				Region:         getEnv("S3_REGION", ""),
				Endpoint:       getEnv("S3_ENDPOINT", ""),
				AccessKey:      getEnv("S3_ACCESS_KEY", ""),
				SecretKey:      getEnv("S3_SECRET_KEY", ""),
				Bucket:         getEnv("S3_BUCKET", ""),
				UseSSL:         getEnvBool("S3_USE_SSL", true),
				ForcePathStyle: getEnvBool("S3_FORCE_PATH_STYLE", false),
			},
			MaxUploadBytes: getEnvInt("IMAGE_MAX_UPLOAD_MB", 10) * 1024 * 1024,
		},
		Keycloak: KeycloakConfig{
			AdminURL:          getEnv("KEYCLOAK_URL", "http://localhost:9090"),
			Realm:             getEnv("KEYCLOAK_REALM", "fejd"),
			IssuerURL:         getEnv("KEYCLOAK_ISSUER_URL", ""),
			Audiences:         splitCSV(getEnv("KEYCLOAK_AUDIENCES", "salon-mobile,fejd-frontend")),
			AdminClientID:     getEnv("KEYCLOAK_ADMIN_CLIENT_ID", "fejd-admin"),
			AdminClientSecret: getEnv("KEYCLOAK_ADMIN_CLIENT_SECRET", ""),
		},
		Jobs: JobsConfig{
			PendingScanInterval: getEnvDuration("PENDING_SCAN_INTERVAL", 15*time.Minute),
			CleanupInterval:     getEnvDuration("CLEANUP_INTERVAL", time.Hour),
			RunJobs:             getEnvBool("RUN_JOBS", true),
			InviteExpiryHours:   int(getEnvInt("INVITE_EXPIRY_HOURS", 48)),
			InviteRedirectURI:   getEnv("INVITE_REDIRECT_URI", ""),
			InviteBaseURL:       getEnv("INVITE_BASE_URL", "http://localhost:5173"),
		},
		Email: EmailConfig{
			SMTPHost:              getEnv("SMTP_HOST", ""),
			SMTPPort:              int(getEnvInt("SMTP_PORT", 587)),
			SMTPUser:              getEnv("SMTP_USER", ""),
			SMTPPass:              getEnv("SMTP_PASS", ""),
			SMTPFrom:              getEnv("SMTP_FROM", ""),
			SuperadminNotifyEmail: getEnv("SUPERADMIN_NOTIFY_EMAIL", ""),
		},
		CORS: CORSConfig{
			AllowedOrigins: splitCSV(getEnv("CORS_ALLOWED_ORIGINS",
				"http://localhost:5173,http://localhost:8080,http://localhost,https://localhost,capacitor://localhost")),
			AllowedSuffix: getEnv("CORS_ALLOWED_SUFFIX", ""),
		},
		AppDomain: getEnv("FEJD_DOMAIN", ""),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	switch c.ImageStorage.Backend {
	case BackendPostgres, BackendSeaweedfs, BackendS3:
	default:
		return fmt.Errorf("invalid IMAGE_STORAGE_BACKEND %q: must be postgres, seaweedfs or s3", c.ImageStorage.Backend)
	}

	if c.ImageStorage.MaxUploadBytes <= 0 {
		return fmt.Errorf("IMAGE_MAX_UPLOAD_MB must be greater than zero")
	}

	switch c.ImageStorage.Backend {
	case BackendSeaweedfs:
		if c.ImageStorage.Seaweedfs.Endpoint == "" || c.ImageStorage.Seaweedfs.Bucket == "" {
			return fmt.Errorf("seaweedfs backend requires SEAWEEDFS_ENDPOINT and SEAWEEDFS_BUCKET")
		}
	case BackendS3:
		s3 := c.ImageStorage.S3
		if s3.Region == "" || s3.Bucket == "" || s3.AccessKey == "" || s3.SecretKey == "" {
			return fmt.Errorf("s3 backend requires S3_REGION, S3_BUCKET, S3_ACCESS_KEY and S3_SECRET_KEY")
		}
	}

	if len(c.Keycloak.Audiences) == 0 {
		return fmt.Errorf("KEYCLOAK_AUDIENCES must contain at least one audience")
	}

	if c.Jobs.InviteExpiryHours <= 0 {
		return fmt.Errorf("INVITE_EXPIRY_HOURS must be greater than zero")
	}

	if c.Jobs.InviteBaseURL != "" {
		u, err := url.Parse(c.Jobs.InviteBaseURL)
		if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
			return fmt.Errorf("INVITE_BASE_URL must be an absolute http(s) URL")
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

func getEnvDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func splitCSV(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}
