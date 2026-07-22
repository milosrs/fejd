package store

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
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
		testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp").WithStartupTimeout(60*time.Second)),
		testcontainers.CustomizeRequestOption(func(req *testcontainers.GenericContainerRequest) error {
			req.AutoRemove = true
			return nil
		}),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("failed to ping database: %v", err)
	}

	migrationSQL := readMigrationFile(t)

	if _, err := pool.Exec(ctx, migrationSQL); err != nil {
		pool.Close()
		t.Fatalf("failed to run migration: %v", err)
	}

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

func readMigrationFile(t *testing.T) string {
	t.Helper()

	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)

	candidates := []string{
		filepath.Join(dir, "..", "..", "migrations", "000001_initial_schema.up.sql"),
		filepath.Join(dir, "..", "..", "..", "backend", "migrations", "000001_initial_schema.up.sql"),
	}

	for _, p := range candidates {
		abs, err := filepath.Abs(p)
		if err != nil {
			continue
		}
		data, err := os.ReadFile(abs)
		if err == nil {
			return string(data)
		}
	}

	t.Fatalf("could not find migration file, tried: %v", candidates)
	return ""
}
