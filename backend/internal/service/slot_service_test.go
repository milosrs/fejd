package service

import (
	"fejd-backend/internal/models"
	"testing"
	"time"
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

	expectedSlotCount := 16
	if len(slots) != expectedSlotCount {
		t.Errorf("expected %d slots, got %d", expectedSlotCount, len(slots))
	}

	if !slots[0].StartTime.Equal(dayStart) {
		t.Errorf("first slot should start at %v, got %v", dayStart, slots[0].StartTime)
	}
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
		if slot.StartTime.Equal(busyStart) {
			t.Error("slot at busy time should not be available")
		}
		if slot.StartTime.Equal(busyStart.Add(30 * time.Minute)) {
			t.Error("slot overlapping busy period should not be available")
		}
	}
}

func TestComputeSlotsWithHourDuration(t *testing.T) {
	dayStart := futureDay().Add(9 * time.Hour)
	dayEnd := futureDay().Add(17 * time.Hour)
	duration := 60 * time.Minute

	slots := computeSlots(dayStart, dayEnd, duration, nil)

	expectedSlotCount := 8
	if len(slots) != expectedSlotCount {
		t.Errorf("expected %d hour-long slots, got %d", expectedSlotCount, len(slots))
	}
}

func TestComputeSlotsEmptyRange(t *testing.T) {
	dayStart := futureDay().Add(9 * time.Hour)
	dayEnd := dayStart
	duration := 30 * time.Minute

	slots := computeSlots(dayStart, dayEnd, duration, nil)

	if len(slots) != 0 {
		t.Errorf("expected 0 slots for empty range, got %d", len(slots))
	}
}
