package store

import (
	"context"
	"fejd-backend/internal/models"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BusinessStore struct {
	pool *pgxpool.Pool
}

func NewBusinessStore(pool *pgxpool.Pool) *BusinessStore {
	return &BusinessStore{pool: pool}
}

func (s *BusinessStore) GetBySlug(ctx context.Context, slug string) (*models.Business, error) {
	sql, args, err := psql.
		Select("id", "name", "slug", "created_at", "updated_at", "cancellation_lead_hours", "no_show_after_minutes", "slot_interval_minutes", "auto_approve", "appointment_reminder_enabled", "appointment_reminder_lead_minutes", "staff_notifications_enabled", "reminder_title", "reminder_body", "COALESCE(address_line, '')", "COALESCE(city, '')", "COALESCE(postal_code, '')", "COALESCE(country, '')", "latitude", "longitude", "COALESCE(phone, '')", "realm_admin_created").
		From("businesses").
		Where(sq.Eq{"slug": slug}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var b models.Business
	err = s.pool.QueryRow(ctx, sql, args...).Scan(&b.ID, &b.Name, &b.Slug, &b.CreatedAt, &b.UpdatedAt, &b.CancellationLeadHours, &b.NoShowAfterMinutes, &b.SlotIntervalMinutes, &b.AutoApprove, &b.AppointmentReminderEnabled, &b.AppointmentReminderLeadMinutes, &b.StaffNotificationsEnabled, &b.ReminderTitle, &b.ReminderBody, &b.AddressLine, &b.City, &b.PostalCode, &b.Country, &b.Latitude, &b.Longitude, &b.Phone, &b.RealmAdminCreated)
	if err != nil {
		return nil, fmt.Errorf("business not found: %w", err)
	}
	return &b, nil
}

func (s *BusinessStore) GetByID(ctx context.Context, id uuid.UUID) (*models.Business, error) {
	sql, args, err := psql.
		Select("id", "name", "slug", "created_at", "updated_at", "cancellation_lead_hours", "no_show_after_minutes", "slot_interval_minutes", "auto_approve", "appointment_reminder_enabled", "appointment_reminder_lead_minutes", "staff_notifications_enabled", "reminder_title", "reminder_body", "COALESCE(address_line, '')", "COALESCE(city, '')", "COALESCE(postal_code, '')", "COALESCE(country, '')", "latitude", "longitude", "COALESCE(phone, '')", "realm_admin_created").
		From("businesses").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var b models.Business
	err = s.pool.QueryRow(ctx, sql, args...).Scan(&b.ID, &b.Name, &b.Slug, &b.CreatedAt, &b.UpdatedAt, &b.CancellationLeadHours, &b.NoShowAfterMinutes, &b.SlotIntervalMinutes, &b.AutoApprove, &b.AppointmentReminderEnabled, &b.AppointmentReminderLeadMinutes, &b.StaffNotificationsEnabled, &b.ReminderTitle, &b.ReminderBody, &b.AddressLine, &b.City, &b.PostalCode, &b.Country, &b.Latitude, &b.Longitude, &b.Phone, &b.RealmAdminCreated)
	if err != nil {
		return nil, fmt.Errorf("business not found: %w", err)
	}
	return &b, nil
}

func (s *BusinessStore) List(ctx context.Context) ([]models.Business, error) {
	sql, args, err := psql.
		Select("id", "name", "slug", "created_at", "updated_at", "cancellation_lead_hours", "no_show_after_minutes", "slot_interval_minutes", "auto_approve", "appointment_reminder_enabled", "appointment_reminder_lead_minutes", "staff_notifications_enabled", "reminder_title", "reminder_body", "COALESCE(address_line, '')", "COALESCE(city, '')", "COALESCE(postal_code, '')", "COALESCE(country, '')", "latitude", "longitude", "COALESCE(phone, '')", "realm_admin_created").
		From("businesses").
		OrderBy("created_at").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list businesses: %w", err)
	}
	defer rows.Close()

	var businesses []models.Business
	for rows.Next() {
		var b models.Business
		if err := rows.Scan(&b.ID, &b.Name, &b.Slug, &b.CreatedAt, &b.UpdatedAt, &b.CancellationLeadHours, &b.NoShowAfterMinutes, &b.SlotIntervalMinutes, &b.AutoApprove, &b.AppointmentReminderEnabled, &b.AppointmentReminderLeadMinutes, &b.StaffNotificationsEnabled, &b.ReminderTitle, &b.ReminderBody, &b.AddressLine, &b.City, &b.PostalCode, &b.Country, &b.Latitude, &b.Longitude, &b.Phone, &b.RealmAdminCreated); err != nil {
			return nil, fmt.Errorf("failed to scan business: %w", err)
		}
		businesses = append(businesses, b)
	}
	return businesses, nil
}

// ListNeedingGeocode returns businesses that have an address but are missing a
// city or geo coordinates, for the one-time location geocoding backfill.
func (s *BusinessStore) ListNeedingGeocode(ctx context.Context) ([]models.Business, error) {
	sql, args, err := psql.
		Select("id", "name", "slug", "created_at", "updated_at", "cancellation_lead_hours", "no_show_after_minutes", "slot_interval_minutes", "auto_approve", "appointment_reminder_enabled", "appointment_reminder_lead_minutes", "staff_notifications_enabled", "reminder_title", "reminder_body", "COALESCE(address_line, '')", "COALESCE(city, '')", "COALESCE(postal_code, '')", "COALESCE(country, '')", "latitude", "longitude", "COALESCE(phone, '')", "realm_admin_created").
		From("businesses").
		Where(sq.Expr("COALESCE(address_line, '') <> ''")).
		Where(sq.Expr("city IS NULL OR latitude IS NULL OR longitude IS NULL")).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list businesses needing geocode: %w", err)
	}
	defer rows.Close()

	var businesses []models.Business
	for rows.Next() {
		var b models.Business
		if err := rows.Scan(&b.ID, &b.Name, &b.Slug, &b.CreatedAt, &b.UpdatedAt, &b.CancellationLeadHours, &b.NoShowAfterMinutes, &b.SlotIntervalMinutes, &b.AutoApprove, &b.AppointmentReminderEnabled, &b.AppointmentReminderLeadMinutes, &b.StaffNotificationsEnabled, &b.ReminderTitle, &b.ReminderBody, &b.AddressLine, &b.City, &b.PostalCode, &b.Country, &b.Latitude, &b.Longitude, &b.Phone, &b.RealmAdminCreated); err != nil {
			return nil, fmt.Errorf("failed to scan business: %w", err)
		}
		businesses = append(businesses, b)
	}
	return businesses, nil
}

func (s *BusinessStore) ListMembershipsByUser(ctx context.Context, userID string) ([]models.BusinessMembership, error) {
	sql, args, err := psql.
		Select("b.id", "b.name", "b.slug", "bu.role").
		From("businesses b").
		Join("business_users bu ON bu.business_id = b.id").
		Where(sq.Eq{"bu.user_id": userID}).
		OrderBy("b.created_at").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list memberships: %w", err)
	}
	defer rows.Close()

	var memberships []models.BusinessMembership
	for rows.Next() {
		var m models.BusinessMembership
		if err := rows.Scan(&m.BusinessID, &m.Name, &m.Slug, &m.Role); err != nil {
			return nil, fmt.Errorf("failed to scan membership: %w", err)
		}
		memberships = append(memberships, m)
	}
	return memberships, nil
}

func (s *BusinessStore) Create(ctx context.Context, q Querier, b *models.Business) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	sql, args, err := psql.
		Insert("businesses").
		Columns("id", "name", "slug", "realm_admin_created").
		Values(b.ID, b.Name, b.Slug, b.RealmAdminCreated).
		Suffix("RETURNING created_at, updated_at, cancellation_lead_hours, no_show_after_minutes, slot_interval_minutes, appointment_reminder_enabled, appointment_reminder_lead_minutes, staff_notifications_enabled, reminder_title, reminder_body, COALESCE(address_line, ''), COALESCE(city, ''), COALESCE(postal_code, ''), COALESCE(country, ''), latitude, longitude, COALESCE(phone, ''), realm_admin_created").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	return q.QueryRow(ctx, sql, args...).Scan(&b.CreatedAt, &b.UpdatedAt, &b.CancellationLeadHours, &b.NoShowAfterMinutes, &b.SlotIntervalMinutes, &b.AppointmentReminderEnabled, &b.AppointmentReminderLeadMinutes, &b.StaffNotificationsEnabled, &b.ReminderTitle, &b.ReminderBody, &b.AddressLine, &b.City, &b.PostalCode, &b.Country, &b.Latitude, &b.Longitude, &b.Phone, &b.RealmAdminCreated)
}

// UpdatePolicy updates a business's cancellation, no-show, slot, automatic
// approval, appointment-reminder and staff-notification policies.
func (s *BusinessStore) UpdatePolicy(ctx context.Context, businessID uuid.UUID, cancellationLeadHours, noShowAfterMinutes, slotIntervalMinutes int, autoApprove bool, appointmentReminderEnabled bool, appointmentReminderLeadMinutes int, staffNotificationsEnabled bool, reminderTitle, reminderBody string) error {
	sql, args, err := psql.
		Update("businesses").
		Set("cancellation_lead_hours", cancellationLeadHours).
		Set("no_show_after_minutes", noShowAfterMinutes).
		Set("slot_interval_minutes", slotIntervalMinutes).
		Set("auto_approve", autoApprove).
		Set("appointment_reminder_enabled", appointmentReminderEnabled).
		Set("appointment_reminder_lead_minutes", appointmentReminderLeadMinutes).
		Set("staff_notifications_enabled", staffNotificationsEnabled).
		Set("reminder_title", reminderTitle).
		Set("reminder_body", reminderBody).
		Where(sq.Eq{"id": businessID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = s.pool.Exec(ctx, sql, args...)
	return err
}

// UpdateLocation persists the salon's structured location. City is required by
// the caller; the geo coordinates are nullable and set by geocoding.
func (s *BusinessStore) UpdateLocation(ctx context.Context, businessID uuid.UUID, addressLine, city, postalCode, country, phone string, lat, lon *float64) error {
	sql, args, err := psql.
		Update("businesses").
		Set("address_line", nullIfEmpty(addressLine)).
		Set("city", nullIfEmpty(city)).
		Set("postal_code", nullIfEmpty(postalCode)).
		Set("country", nullIfEmpty(country)).
		Set("phone", nullIfEmpty(phone)).
		Set("latitude", lat).
		Set("longitude", lon).
		Set("updated_at", sq.Expr("now()")).
		Where(sq.Eq{"id": businessID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = s.pool.Exec(ctx, sql, args...)
	return err
}

// nullIfEmpty maps an empty string to NULL so optional location fields stay
// NULL rather than empty strings.
func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// Rename updates a business's display name and slug.
func (s *BusinessStore) Rename(ctx context.Context, q Querier, businessID uuid.UUID, name, slug string) error {
	sql, args, err := psql.
		Update("businesses").
		Set("name", name).
		Set("slug", slug).
		Set("updated_at", sq.Expr("now()")).
		Where(sq.Eq{"id": businessID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = q.Exec(ctx, sql, args...)
	return err
}

// SlugTakenByOther reports whether a business other than excludeID already
// uses the given slug.
func (s *BusinessStore) SlugTakenByOther(ctx context.Context, q Querier, slug string, excludeID uuid.UUID) (bool, error) {
	sql, args, err := psql.
		Select("1").
		From("businesses").
		Where(sq.Eq{"slug": slug}).
		Where(sq.NotEq{"id": excludeID}).
		Prefix("SELECT EXISTS (").
		Suffix(")").
		ToSql()
	if err != nil {
		return false, fmt.Errorf("failed to build query: %w", err)
	}

	var exists bool
	if err := q.QueryRow(ctx, sql, args...).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check slug: %w", err)
	}
	return exists, nil
}

func (s *BusinessStore) SlugExists(ctx context.Context, q Querier, slug string) (bool, error) {
	sql, args, err := psql.
		Select("1").
		From("businesses").
		Where(sq.Eq{"slug": slug}).
		Prefix("SELECT EXISTS (").
		Suffix(")").
		ToSql()
	if err != nil {
		return false, fmt.Errorf("failed to build query: %w", err)
	}

	var exists bool
	if err := q.QueryRow(ctx, sql, args...).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check slug: %w", err)
	}
	return exists, nil
}

// Delete removes a business row. Callers must delete dependent rows (images and
// appointments) first because those reference businesses without ON DELETE
// CASCADE; every other dependent table cascades automatically.
func (s *BusinessStore) Delete(ctx context.Context, q Querier, id uuid.UUID) error {
	sql, args, err := psql.
		Delete("businesses").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = q.Exec(ctx, sql, args...)
	return err
}
