package middleware

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"fejd-backend/auth"
	"fejd-backend/internal/db"
	"fejd-backend/internal/models"
	"fejd-backend/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupMiddlewareDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("fejd"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp").WithStartupTimeout(60*time.Second)),
		testcontainers.CustomizeRequestOption(func(req *testcontainers.GenericContainerRequest) error {
			req.AutoRemove = true
			return nil
		}),
	)
	require.NoError(t, err)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)
	require.NoError(t, pool.Ping(ctx))

	sqlDB, err := sql.Open("pgx", connStr)
	require.NoError(t, err)
	require.NoError(t, db.Migrate(sqlDB, "file://"+locateMigrations(t)))
	require.NoError(t, sqlDB.Close())

	t.Cleanup(func() {
		pool.Close()
		if err := pgContainer.Terminate(context.Background()); err != nil {
			fmt.Fprintf(os.Stderr, "failed to terminate container: %v\n", err)
		}
	})

	return pool
}

func locateMigrations(t *testing.T) string {
	t.Helper()
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	for _, p := range []string{
		filepath.Join(dir, "..", "..", "migrations"),
		filepath.Join(dir, "..", "..", "..", "backend", "migrations"),
	} {
		if abs, err := filepath.Abs(p); err == nil {
			if info, err := os.Stat(abs); err == nil && info.IsDir() {
				return abs
			}
		}
	}
	require.FailNow(t, "could not find migrations directory")
	return ""
}

func memberRequest(t *testing.T, userID, businessID string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/business/"+businessID+"/invitations", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("businessID", businessID)
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, auth.ContextKeyUserID, userID)
	return req.WithContext(ctx)
}

func TestRequireBusinessMember(t *testing.T) {
	pool := setupMiddlewareDB(t)
	ctx := context.Background()
	buStore := store.NewBusinessUserStore(pool)
	businessStore := store.NewBusinessStore(pool)

	b := &models.Business{Name: "Salon", Slug: "salon"}
	require.NoError(t, businessStore.Create(ctx, pool, b))

	admin := &models.BusinessUser{BusinessID: b.ID, UserID: "owner", Role: "admin"}
	require.NoError(t, buStore.Create(ctx, pool, admin))
	employee := &models.BusinessUser{BusinessID: b.ID, UserID: "emp", Role: "employee"}
	require.NoError(t, buStore.Create(ctx, pool, employee))
	inactive := &models.BusinessUser{BusinessID: b.ID, UserID: "gone", Role: "employee"}
	require.NoError(t, buStore.Create(ctx, pool, inactive))
	require.NoError(t, buStore.SetActive(ctx, pool, inactive.ID, false))

	mw := RequireBusinessMember(buStore)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	allowed := []string{"owner", "emp"}
	for _, uid := range allowed {
		rr := httptest.NewRecorder()
		mw(next).ServeHTTP(rr, memberRequest(t, uid, b.ID.String()))
		require.Equal(t, http.StatusNoContent, rr.Code, "user %q should be allowed", uid)
	}

	forbidden := []string{"gone", "stranger"}
	for _, uid := range forbidden {
		rr := httptest.NewRecorder()
		mw(next).ServeHTTP(rr, memberRequest(t, uid, b.ID.String()))
		require.Equal(t, http.StatusForbidden, rr.Code, "user %q should be forbidden", uid)
	}
}
