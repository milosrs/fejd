package store

import (
	"context"
	"fejd-backend/internal/models"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WorkingHoursOverrideStore struct {
	pool *pgxpool.Pool
}

func NewWorkingHoursOverrideStore(pool *pgxpool.Pool) *WorkingHoursOverrideStore {
	return &WorkingHoursOverrideStore{pool: pool}
}

func (s *WorkingHoursOverrideStore) ListByBusinessUser(ctx context.Context, businessUserID uuid.UUID, from, to time.Time) ([]models.WorkingHoursOverride, error) {
	sql, args, err := psql.
		Select("id", "business_user_id", "override_date::text", "start_time::text", "end_time::text", "is_off", "COALESCE(reason, '')").
		From("working_hours_overrides").
		Where(sq.Eq{"business_user_id": businessUserID}).
		Where(sq.GtOrEq{"override_date": from.Format("2006-01-02")}).
		Where(sq.LtOrEq{"override_date": to.Format("2006-01-02")}).
		OrderBy("override_date").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list overrides: %w", err)
	}
	defer rows.Close()

	var overrides []models.WorkingHoursOverride
	for rows.Next() {
		var o models.WorkingHoursOverride
		var startTime, endTime *string
		if err := rows.Scan(&o.ID, &o.BusinessUserID, &o.OverrideDate, &startTime, &endTime, &o.IsOff, &o.Reason); err != nil {
			return nil, fmt.Errorf("failed to scan override: %w", err)
		}
		o.StartTime = startTime
		o.EndTime = endTime
		overrides = append(overrides, o)
	}
	return overrides, nil
}

func (s *WorkingHoursOverrideStore) GetByBusinessUserAndDate(ctx context.Context, businessUserID uuid.UUID, date time.Time) (*models.WorkingHoursOverride, error) {
	sql, args, err := psql.
		Select("id", "business_user_id", "override_date::text", "start_time::text", "end_time::text", "is_off", "COALESCE(reason, '')").
		From("working_hours_overrides").
		Where(sq.Eq{"business_user_id": businessUserID, "override_date": date.Format("2006-01-02")}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var o models.WorkingHoursOverride
	var startTime, endTime *string
	err = s.pool.QueryRow(ctx, sql, args...).Scan(&o.ID, &o.BusinessUserID, &o.OverrideDate, &startTime, &endTime, &o.IsOff, &o.Reason)
	if err != nil {
		return nil, fmt.Errorf("override not found: %w", err)
	}
	o.StartTime = startTime
	o.EndTime = endTime
	return &o, nil
}

func (s *WorkingHoursOverrideStore) Create(ctx context.Context, o *models.WorkingHoursOverride) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	sql, args, err := psql.
		Insert("working_hours_overrides").
		Columns("id", "business_user_id", "override_date", "start_time", "end_time", "is_off", "reason").
		Values(o.ID, o.BusinessUserID, o.OverrideDate, o.StartTime, o.EndTime, o.IsOff, o.Reason).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = s.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to create override: %w", err)
	}
	return nil
}

func (s *WorkingHoursOverrideStore) Delete(ctx context.Context, id uuid.UUID) error {
	sql, args, err := psql.
		Delete("working_hours_overrides").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = s.pool.Exec(ctx, sql, args...)
	return err
}
