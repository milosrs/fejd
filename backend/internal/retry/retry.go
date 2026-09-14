package retry

import (
	"context"
	"fmt"
	"log"
	"time"

	avastretry "github.com/avast/retry-go/v4"
)

const (
	maxAttempts = 5
	initialWait = time.Second
	maxWait     = 8 * time.Second
)

func Do(ctx context.Context, name string, operation func() error) error {
	attempt := uint(0)
	err := avastretry.Do(
		func() error {
			attempt++
			return operation()
		},
		avastretry.Attempts(maxAttempts),
		avastretry.Context(ctx),
		avastretry.Delay(initialWait),
		avastretry.MaxDelay(maxWait),
		avastretry.DelayType(avastretry.BackOffDelay),
		avastretry.OnRetry(func(_ uint, err error) {
			log.Printf("[startup] %s connection attempt %d/%d failed: %v; retrying with exponential backoff", name, attempt, maxAttempts, err)
		}),
	)
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("%s connection cancelled: %w", name, ctx.Err())
		}
		return fmt.Errorf("%s connection failed after %d attempts: %w", name, attempt, err)
	}
	if attempt > 1 {
		log.Printf("[startup] %s connection established on attempt %d", name, attempt)
	}
	return nil
}
