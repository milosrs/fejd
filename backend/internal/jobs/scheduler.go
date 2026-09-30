package jobs

import (
	"context"
	"log"
	"time"
)

type Scheduler struct {
	notifier         *PendingNotifier
	cleanup          *InviteCleanup
	reminder         *ReminderNotifier
	scanInterval     time.Duration
	cleanupInterval  time.Duration
	reminderInterval time.Duration
}

func NewScheduler(notifier *PendingNotifier, cleanup *InviteCleanup, reminder *ReminderNotifier, scanInterval, cleanupInterval, reminderInterval time.Duration) *Scheduler {
	return &Scheduler{
		notifier:         notifier,
		cleanup:          cleanup,
		reminder:         reminder,
		scanInterval:     scanInterval,
		cleanupInterval:  cleanupInterval,
		reminderInterval: reminderInterval,
	}
}

// Run blocks running all jobs on their intervals until ctx is cancelled.
func (s *Scheduler) Run(ctx context.Context) {
	scan := time.NewTicker(s.scanInterval)
	cleanup := time.NewTicker(s.cleanupInterval)
	reminder := time.NewTicker(s.reminderInterval)
	defer scan.Stop()
	defer cleanup.Stop()
	defer reminder.Stop()

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
		case <-reminder.C:
			if s.reminder != nil {
				if err := s.reminder.Run(ctx); err != nil {
					log.Printf("[jobs] appointment reminder failed: %v", err)
				}
			}
		}
	}
}
