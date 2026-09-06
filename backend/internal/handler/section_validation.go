package handler

import (
	"context"
	"errors"

	"fejd-backend/internal/models"
	"fejd-backend/internal/store"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ownedSection returns the section only when it belongs to the given business's
// landing page, so a caller cannot address another business's sections.
func (h *AdminHandler) ownedSection(ctx context.Context, businessID, sectionID uuid.UUID) (*models.Section, error) {
	page, err := h.pageStore.GetByBusinessAndName(ctx, businessID, store.LandingPageName)
	if err != nil {
		return nil, err
	}

	section, err := h.sectionStore.GetByID(ctx, sectionID)
	if err != nil {
		return nil, err
	}
	if section.PageID != page.ID {
		return nil, errors.New("section does not belong to this business")
	}
	return section, nil
}

// ensureLandingPage returns the business's landing page, creating it if it does
// not exist yet.
func (h *AdminHandler) ensureLandingPage(ctx context.Context, businessID uuid.UUID) (*models.Page, error) {
	page, err := h.pageStore.GetByBusinessAndName(ctx, businessID, store.LandingPageName)
	if err == nil {
		return page, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	page = &models.Page{BusinessID: businessID, Name: store.LandingPageName}
	if err := h.pageStore.Create(ctx, h.pool, page); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return h.pageStore.GetByBusinessAndName(ctx, businessID, store.LandingPageName)
		}
		return nil, err
	}
	return page, nil
}
