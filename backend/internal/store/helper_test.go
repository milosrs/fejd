package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"fejd-backend/internal/db"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

type testDB struct {
	pool     *pgxpool.Pool
	teardown func()
}

func setupTestDB(t *testing.T) *testDB {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("fejd"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		tcpostgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		require.FailNow(t, "failed to ping database: %v", err)
	}

	sqlDB, err := sql.Open("pgx", connStr)
	if err != nil {
		pool.Close()
		require.FailNow(t, "failed to open database for migrations: %v", err)
	}

	if err := db.Migrate(sqlDB, "file://"+locateMigrationsDir(t)); err != nil {
		sqlDB.Close()
		pool.Close()
		require.FailNow(t, "failed to run migrations: %v", err)
	}
	sqlDB.Close()

	return &testDB{
		pool: pool,
		teardown: func() {
			pool.Close()
			if err := pgContainer.Terminate(context.Background()); err != nil {
				fmt.Fprintf(os.Stderr, "failed to terminate container: %v\n", err)
			}
		},
	}
}

func locateMigrationsDir(t *testing.T) string {
	t.Helper()

	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)

	candidates := []string{
		filepath.Join(dir, "..", "..", "migrations"),
		filepath.Join(dir, "..", "..", "..", "backend", "migrations"),
	}

	for _, p := range candidates {
		abs, err := filepath.Abs(p)
		if err != nil {
			continue
		}
		if info, err := os.Stat(abs); err == nil && info.IsDir() {
			return abs
		}
	}

	require.FailNow(t, "could not find migrations directory", "tried: %v", candidates)
	return ""
}
