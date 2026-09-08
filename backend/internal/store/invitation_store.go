package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"fejd-backend/internal/models"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InvitationStore struct {
	pool *pgxpool.Pool
}

func NewInvitationStore(pool *pgxpool.Pool) *InvitationStore {
	return &InvitationStore{pool: pool}
}

func invitationColumns() []string {
	return []string{"id", "business_id", "token_hash", "role", "created_by", "max_uses", "use_count", "expires_at", "created_at"}
}

func scanInvitation(row pgx.Row) (*models.Invitation, error) {
	var inv models.Invitation
	err := row.Scan(&inv.ID, &inv.BusinessID, &inv.TokenHash, &inv.Role, &inv.CreatedBy, &inv.MaxUses, &inv.UseCount, &inv.ExpiresAt, &inv.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to scan invitation: %w", err)
	}
	return &inv, nil
}

func (s *InvitationStore) Create(ctx context.Context, q Querier, inv *models.Invitation) error {
	if inv.ID == uuid.Nil {
		inv.ID = uuid.New()
	}
	if inv.Role == "" {
		inv.Role = "employee"
	}
	if inv.MaxUses == 0 {
		inv.MaxUses = 1
	}
	sql, args, err := psql.
		Insert("invitations").
		Columns("id", "business_id", "token_hash", "role", "created_by", "max_uses", "use_count", "expires_at").
		Values(inv.ID, inv.BusinessID, inv.TokenHash, inv.Role, inv.CreatedBy, inv.MaxUses, inv.UseCount, inv.ExpiresAt).
		Suffix("RETURNING created_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	return q.QueryRow(ctx, sql, args...).Scan(&inv.CreatedAt)
}

func (s *InvitationStore) GetByTokenHash(ctx context.Context, tokenHash string) (*models.Invitation, error) {
	sql, args, err := psql.
		Select(invitationColumns()...).
		From("invitations").
		Where(sq.Eq{"token_hash": tokenHash}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	return scanInvitation(s.pool.QueryRow(ctx, sql, args...))
}

func (s *InvitationStore) GetByID(ctx context.Context, id uuid.UUID) (*models.Invitation, error) {
	sql, args, err := psql.
		Select(invitationColumns()...).
		From("invitations").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	return scanInvitation(s.pool.QueryRow(ctx, sql, args...))
}

func (s *InvitationStore) ListByBusiness(ctx context.Context, businessID uuid.UUID) ([]models.Invitation, error) {
	sql, args, err := psql.
		Select(invitationColumns()...).
		From("invitations").
		Where(sq.Eq{"business_id": businessID}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list invitations: %w", err)
	}
	defer rows.Close()

	var invitations []models.Invitation
	for rows.Next() {
		var inv models.Invitation
		if err := rows.Scan(&inv.ID, &inv.BusinessID, &inv.TokenHash, &inv.Role, &inv.CreatedBy, &inv.MaxUses, &inv.UseCount, &inv.ExpiresAt, &inv.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan invitation: %w", err)
		}
		invitations = append(invitations, inv)
	}
	return invitations, nil
}

// TryConsume atomically increments use_count when the invitation is still
// valid (unexpired and under its use limit). It returns true when the caller
// won the consumption; false when the invite was expired, already used up, or
// missing.
func (s *InvitationStore) TryConsume(ctx context.Context, q Querier, id uuid.UUID) (bool, error) {
	sql, args, err := psql.
		Update("invitations").
		Set("use_count", sq.Expr("use_count + 1")).
		Where(sq.Eq{"id": id}).
		Where(sq.Expr("use_count < max_uses")).
		Where(sq.Expr("expires_at > ?", time.Now().UTC())).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return false, fmt.Errorf("failed to build query: %w", err)
	}

	var consumedID uuid.UUID
	err = q.QueryRow(ctx, sql, args...).Scan(&consumedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to consume invitation: %w", err)
	}
	return true, nil
}

func (s *InvitationStore) Delete(ctx context.Context, q Querier, id uuid.UUID) error {
	sql, args, err := psql.
		Delete("invitations").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = q.Exec(ctx, sql, args...)
	return err
}
