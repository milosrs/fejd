package service

import (
	"context"
	"fejd-backend/internal/models"
	"fejd-backend/internal/store"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type WorkingHoursService struct {
	workingHours *store.WorkingHoursStore
	overrides    *store.WorkingHoursOverrideStore
	businessUser *store.BusinessUserStore
	slotService  *SlotService
	business     *store.BusinessStore
}

func NewWorkingHoursService(
	workingHours *store.WorkingHoursStore,
	overrides *store.WorkingHoursOverrideStore,
	businessUser *store.BusinessUserStore,
	slotService *SlotService,
	business *store.BusinessStore,
) *WorkingHoursService {
	return &WorkingHoursService{
		workingHours: workingHours,
		overrides:    overrides,
		businessUser: businessUser,
		slotService:  slotService,
		business:     business,
	}
}

func (s *WorkingHoursService) SetWeeklyHours(ctx context.Context, businessID uuid.UUID, targetUserID string, hours []models.WorkingHours) error {
	isAdmin, _ := s.businessUser.IsAdmin(ctx, businessID, targetUserID)
	_ = isAdmin

	targetBU, err := s.businessUser.GetByBusinessAndUser(ctx, businessID, targetUserID)
	if err != nil {
		return fmt.Errorf("target user not found in business: %w", err)
	}

	if err := s.workingHours.DeleteByBusinessUser(ctx, targetBU.ID); err != nil {
		return fmt.Errorf("failed to clear existing hours: %w", err)
	}

	for i := range hours {
		hours[i].BusinessUserID = targetBU.ID
		if err := s.workingHours.Upsert(ctx, &hours[i]); err != nil {
			return fmt.Errorf("failed to set hours for day %d: %w", hours[i].DayOfWeek, err)
		}
	}

	today := time.Now()
	for i := 0; i < 7; i++ {
		date := today.AddDate(0, 0, i)
		s.slotService.PublishSlotUpdate(businessID, targetBU.ID, date)
	}

	return nil
}

func (s *WorkingHoursService) GetWeeklyHours(ctx context.Context, businessID uuid.UUID, targetUserID string) ([]models.WorkingHours, error) {
	targetBU, err := s.businessUser.GetByBusinessAndUser(ctx, businessID, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("target user not found in business: %w", err)
	}
	return s.workingHours.GetByBusinessUser(ctx, targetBU.ID)
}

func (s *WorkingHoursService) AddOverride(ctx context.Context, businessID uuid.UUID, targetUserID string, override *models.WorkingHoursOverride) error {
	targetBU, err := s.businessUser.GetByBusinessAndUser(ctx, businessID, targetUserID)
	if err != nil {
		return fmt.Errorf("target user not found in business: %w", err)
	}

	override.BusinessUserID = targetBU.ID
	if err := s.overrides.Create(ctx, override); err != nil {
		return fmt.Errorf("failed to create override: %w", err)
	}

	date, err := time.Parse("2006-01-02", override.OverrideDate)
	if err == nil {
		s.slotService.PublishSlotUpdate(businessID, targetBU.ID, date)
	}

	return nil
}

func (s *WorkingHoursService) DeleteOverride(ctx context.Context, businessID uuid.UUID, overrideID uuid.UUID) error {
	override, err := s.overrides.GetByBusinessUserAndDate(ctx, uuid.Nil, time.Now())
	if err != nil {
		return fmt.Errorf("override not found: %w", err)
	}
	_ = override

	if err := s.overrides.Delete(ctx, overrideID); err != nil {
		return fmt.Errorf("failed to delete override: %w", err)
	}

	return nil
}

func (s *WorkingHoursService) GetOverrides(ctx context.Context, businessID uuid.UUID, targetUserID string, from, to time.Time) ([]models.WorkingHoursOverride, error) {
	targetBU, err := s.businessUser.GetByBusinessAndUser(ctx, businessID, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("target user not found in business: %w", err)
	}
	return s.overrides.ListByBusinessUser(ctx, targetBU.ID, from, to)
}
