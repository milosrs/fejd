package concurrency

import (
	"context"
	"fmt"
	"hash/fnv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Key derives a stable bigint advisory-lock key from a primitive value.
// All lock acquisitions in the codebase must go through this function so
// different write paths serialize on the same key.
func Key(v any) (int64, error) {
	switch x := v.(type) {
	case int:
		return int64(x), nil
	case int32:
		return int64(x), nil
	case int64:
		return x, nil
	case uint64:
		return int64(x), nil
	case string:
		h := fnv.New64a()
		_, _ = h.Write([]byte(x))
		return int64(h.Sum64()), nil
	case fmt.Stringer:
		return Key(x.String())
	default:
		return 0, fmt.Errorf("concurrency: unsupported lock key type %T", v)
	}
}

// XactLock acquires a transaction-scoped advisory lock, released automatically
// on commit or rollback.
func XactLock(ctx context.Context, tx pgx.Tx, v any) error {
	key, err := Key(v)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", key)
	return err
}

// SessionLock acquires a session-scoped advisory lock on a dedicated
// connection. The returned release function must be called to unlock and
// release the connection.
func SessionLock(ctx context.Context, pool *pgxpool.Pool, v any) (release func(), err error) {
	key, err := Key(v)
	if err != nil {
		return nil, err
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1)", key); err != nil {
		conn.Release()
		return nil, err
	}

	return func() {
		_, _ = conn.Exec(context.Background(), "SELECT pg_advisory_unlock($1)", key)
		conn.Release()
	}, nil
}
