package service

import (
	"testing"
	"time"

	"fejd-backend/internal/models"

	"github.com/stretchr/testify/assert"
)

func futureDay() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
}

func TestComputeSlots(t *testing.T) {
	dayStart := futureDay().Add(9 * time.Hour)
	dayEnd := futureDay().Add(17 * time.Hour)
	duration := 30 * time.Minute

	slots := computeSlots(dayStart, dayEnd, duration, nil)

	assert.Len(t, slots, 16)
	assert.True(t, slots[0].StartTime.Equal(dayStart))
}

func TestComputeSlotsWithBusy(t *testing.T) {
	dayStart := futureDay().Add(9 * time.Hour)
	dayEnd := futureDay().Add(17 * time.Hour)
	duration := 30 * time.Minute

	busyStart := futureDay().Add(10 * time.Hour)
	busyEnd := futureDay().Add(11 * time.Hour)
	busySlots := []models.TimeSlot{
		{
			StartTime: busyStart,
			EndTime:   busyEnd,
		},
	}

	slots := computeSlots(dayStart, dayEnd, duration, busySlots)

	for _, slot := range slots {
		assert.False(t, slot.StartTime.Equal(busyStart), "slot at busy time should not be available")
		assert.False(t, slot.StartTime.Equal(busyStart.Add(30*time.Minute)), "slot overlapping busy period should not be available")
	}
}

func TestComputeSlotsWithHourDuration(t *testing.T) {
	dayStart := futureDay().Add(9 * time.Hour)
	dayEnd := futureDay().Add(17 * time.Hour)
	duration := 60 * time.Minute

	slots := computeSlots(dayStart, dayEnd, duration, nil)

	assert.Len(t, slots, 8)
}

func TestComputeSlotsEmptyRange(t *testing.T) {
	dayStart := futureDay().Add(9 * time.Hour)
	dayEnd := dayStart
	duration := 30 * time.Minute

	slots := computeSlots(dayStart, dayEnd, duration, nil)

	assert.Empty(t, slots)
}
