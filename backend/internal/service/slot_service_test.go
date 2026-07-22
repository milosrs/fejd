package service

import (
	"fejd-backend/internal/models"
	"testing"
	"time"
)

func TestComputeSlots(t *testing.T) {
	dayStart := time.Date(2026, 7, 22, 9, 0, 0, 0, time.UTC)
	dayEnd := time.Date(2026, 7, 22, 17, 0, 0, 0, time.UTC)
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
	dayStart := time.Date(2026, 7, 22, 9, 0, 0, 0, time.UTC)
	dayEnd := time.Date(2026, 7, 22, 17, 0, 0, 0, time.UTC)
	duration := 30 * time.Minute

	busySlots := []models.TimeSlot{
		{
			StartTime: time.Date(2026, 7, 22, 10, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2026, 7, 22, 11, 0, 0, 0, time.UTC),
		},
	}

	slots := computeSlots(dayStart, dayEnd, duration, busySlots)

	for _, slot := range slots {
		if slot.StartTime.Equal(busySlots[0].StartTime) {
			t.Error("slot at busy time should not be available")
		}
		if slot.StartTime.Equal(time.Date(2026, 7, 22, 10, 30, 0, 0, time.UTC)) {
			t.Error("slot overlapping busy period should not be available")
		}
	}
}

func TestComputeSlotsWithHourDuration(t *testing.T) {
	dayStart := time.Date(2026, 7, 22, 9, 0, 0, 0, time.UTC)
	dayEnd := time.Date(2026, 7, 22, 17, 0, 0, 0, time.UTC)
	duration := 60 * time.Minute

	slots := computeSlots(dayStart, dayEnd, duration, nil)

	expectedSlotCount := 8
	if len(slots) != expectedSlotCount {
		t.Errorf("expected %d hour-long slots, got %d", expectedSlotCount, len(slots))
	}
}

func TestComputeSlotsEmptyRange(t *testing.T) {
	dayStart := time.Date(2026, 7, 22, 9, 0, 0, 0, time.UTC)
	dayEnd := dayStart
	duration := 30 * time.Minute

	slots := computeSlots(dayStart, dayEnd, duration, nil)

	if len(slots) != 0 {
		t.Errorf("expected 0 slots for empty range, got %d", len(slots))
	}
}
