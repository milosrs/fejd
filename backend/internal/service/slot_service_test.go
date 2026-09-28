package service

import (
	"testing"
	"time"

	"fejd-backend/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestComputeSlotsScalesAfterAppointment(t *testing.T) {
	day := futureDay()
	dayStart := day.Add(10 * time.Hour)
	dayEnd := day.Add(13 * time.Hour)
	serviceDuration := 30 * time.Minute
	busy := []models.TimeSlot{
		{
			StartTime: day.Add(10*time.Hour + 30*time.Minute),
			EndTime:   day.Add(10*time.Hour + 50*time.Minute),
		},
	}

	slots := computeSlots(dayStart, dayEnd, serviceDuration, busy)

	want := []time.Time{
		day.Add(10 * time.Hour),
		day.Add(10*time.Hour + 50*time.Minute),
		day.Add(11*time.Hour + 20*time.Minute),
		day.Add(11*time.Hour + 50*time.Minute),
	}
	require.GreaterOrEqual(t, len(slots), len(want))
	for i, s := range want {
		assert.True(t, slots[i].StartTime.Equal(s))
	}
}

func TestComputeSlotsConflictsWithLongService(t *testing.T) {
	day := futureDay()
	dayStart := day.Add(10 * time.Hour)
	dayEnd := day.Add(12 * time.Hour)
	serviceDuration := 40 * time.Minute
	busy := []models.TimeSlot{
		{
			StartTime: day.Add(10*time.Hour + 30*time.Minute),
			EndTime:   day.Add(10*time.Hour + 50*time.Minute),
		},
	}

	slots := computeSlots(dayStart, dayEnd, serviceDuration, busy)

	require.NotEmpty(t, slots)
	assert.True(t, slots[0].StartTime.Equal(day.Add(10*time.Hour+50*time.Minute)))
	for _, s := range slots {
		assert.False(t, s.StartTime.Equal(day.Add(10*time.Hour)), "10:00 must not be offered for a 40m service overlapping a 10:30 booking")
	}
}

func TestComputeSlotsPacksShortService(t *testing.T) {
	day := futureDay()
	dayStart := day.Add(10 * time.Hour)
	dayEnd := day.Add(11 * time.Hour)
	serviceDuration := 20 * time.Minute

	slots := computeSlots(dayStart, dayEnd, serviceDuration, nil)

	require.Len(t, slots, 3)
	assert.True(t, slots[0].StartTime.Equal(day.Add(10 * time.Hour)))
	assert.True(t, slots[1].StartTime.Equal(day.Add(10*time.Hour + 20*time.Minute)))
	assert.True(t, slots[2].StartTime.Equal(day.Add(10*time.Hour + 40*time.Minute)))
}

func TestComputeSlotsSqueezesBetweenMixedReservations(t *testing.T) {
	day := futureDay()
	dayStart := day.Add(9 * time.Hour)
	dayEnd := day.Add(20 * time.Hour)
	serviceDuration := 30 * time.Minute

	busy := []models.TimeSlot{
		{StartTime: day.Add(9 * time.Hour), EndTime: day.Add(9*time.Hour + 40*time.Minute)},
		{StartTime: day.Add(10 * time.Hour), EndTime: day.Add(10*time.Hour + 20*time.Minute)},
		{StartTime: day.Add(11 * time.Hour), EndTime: day.Add(11*time.Hour + 50*time.Minute)},
	}

	slots := computeSlots(dayStart, dayEnd, serviceDuration, busy)

	require.NotEmpty(t, slots)

	// 09:40-10:00 (20m) is too short for a 30m service; the first bookable start
	// is 10:20 (after the 10:00-10:20 reservation), then 11:50 (after the
	// 11:00-11:50 reservation).
	assert.True(t, slots[0].StartTime.Equal(day.Add(10*time.Hour+20*time.Minute)))
	assert.True(t, slots[1].StartTime.Equal(day.Add(11*time.Hour+50*time.Minute)))

	for _, s := range slots {
		for _, b := range busy {
			assert.False(t,
				s.StartTime.Before(b.EndTime) && s.EndTime.After(b.StartTime),
				"slot %v overlaps reservation %v-%v", s.StartTime, b.StartTime, b.EndTime,
			)
		}
	}
}
