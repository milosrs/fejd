package jobs

import (
	"context"
	"log"
	"time"
)

type Scheduler struct {
	notifier        *PendingNotifier
	cleanup         *InviteCleanup
	scanInterval    time.Duration
	cleanupInterval time.Duration
}

func NewScheduler(notifier *PendingNotifier, cleanup *InviteCleanup, scanInterval, cleanupInterval time.Duration) *Scheduler {
	return &Scheduler{
		notifier:        notifier,
		cleanup:         cleanup,
		scanInterval:    scanInterval,
		cleanupInterval: cleanupInterval,
	}
}

// Run blocks running both jobs on their intervals until ctx is cancelled.
func (s *Scheduler) Run(ctx context.Context) {
	scan := time.NewTicker(s.scanInterval)
	cleanup := time.NewTicker(s.cleanupInterval)
	defer scan.Stop()
	defer cleanup.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-scan.C:
			if err := s.notifier.Run(ctx); err != nil {
				log.Printf("[jobs] pending notifier failed: %v", err)
			}
		case <-cleanup.C:
			if err := s.cleanup.Run(ctx); err != nil {
				log.Printf("[jobs] invite cleanup failed: %v", err)
			}
		}
	}
}
