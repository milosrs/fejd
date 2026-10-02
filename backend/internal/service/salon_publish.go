package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"fejd-backend/internal/indexing"
	"fejd-backend/internal/models"
	"fejd-backend/internal/store"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	// ErrBusinessNotFound is returned when a publish targets a missing salon.
	ErrBusinessNotFound = errors.New("business not found")
	// ErrIndexingUnavailable is returned when the search engine cannot be
	// reached; the publish is not persisted so the owner can retry.
	ErrIndexingUnavailable = errors.New("indexing service unavailable")
)

// PublishStatus reports the outcome of a publish request.
type PublishStatus string

const (
	PublishStatusPublished PublishStatus = "published"
	PublishStatusNoChanges PublishStatus = "no_changes"
)

// SalonPublishService detects whether a salon's public content changed since it
// was last published and, when it has, submits it to a search engine and
// refreshes the sitemap cache.
type SalonPublishService struct {
	businessStore        *store.BusinessStore
	serviceStore         *store.ServiceStore
	buStore              *store.BusinessUserStore
	employeeServiceStore *store.EmployeeServiceStore
	pageStore            *store.PageStore
	sectionStore         *store.SectionStore
	userStore            *store.UserStore
	imageLinkStore       *store.ImageLinkStore
	publishStore         *store.SalonPublishStore
	indexer              indexing.Indexer
	cache                *SitemapCache
	pool                 *pgxpool.Pool
	appDomain            string
}

func NewSalonPublishService(
	businessStore *store.BusinessStore,
	serviceStore *store.ServiceStore,
	buStore *store.BusinessUserStore,
	employeeServiceStore *store.EmployeeServiceStore,
	pageStore *store.PageStore,
	sectionStore *store.SectionStore,
	userStore *store.UserStore,
	imageLinkStore *store.ImageLinkStore,
	publishStore *store.SalonPublishStore,
	indexer indexing.Indexer,
	cache *SitemapCache,
	pool *pgxpool.Pool,
	appDomain string,
) *SalonPublishService {
	return &SalonPublishService{
		businessStore:        businessStore,
		serviceStore:         serviceStore,
		buStore:              buStore,
		employeeServiceStore: employeeServiceStore,
		pageStore:            pageStore,
		sectionStore:         sectionStore,
		userStore:            userStore,
		imageLinkStore:       imageLinkStore,
		publishStore:         publishStore,
		indexer:              indexer,
		cache:                cache,
		pool:                 pool,
		appDomain:            appDomain,
	}
}

// Publish compares the salon's current public snapshot against its last
// published hash. When unchanged it returns PublishStatusNoChanges without
// contacting the search engine. Otherwise it submits the salon's URLs, persists
// the new hash and invalidates the sitemap cache.
func (s *SalonPublishService) Publish(ctx context.Context, businessID uuid.UUID) (PublishStatus, error) {
	b, err := s.businessStore.GetByID(ctx, businessID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrBusinessNotFound
		}
		return "", err
	}

	hash, urls, err := s.snapshot(ctx, b)
	if err != nil {
		return "", err
	}

	stored, err := s.publishStore.GetByBusiness(ctx, businessID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}
	if stored != nil && stored.ContentHash == hash {
		return PublishStatusNoChanges, nil
	}

	if err := s.indexer.PublishURLs(ctx, urls); err != nil {
		return "", ErrIndexingUnavailable
	}

	if err := s.publishStore.Upsert(ctx, s.pool, &models.SalonPublish{
		BusinessID:  businessID,
		ContentHash: hash,
		PublishedAt: time.Now().UTC(),
	}); err != nil {
		return "", err
	}

	if s.cache != nil {
		s.cache.Invalidate()
	}
	return PublishStatusPublished, nil
}

// snapshot computes the SHA-256 hash of the salon's public content and the
// absolute URLs a search engine should (re)index.
func (s *SalonPublishService) snapshot(ctx context.Context, b *models.Business) (string, []string, error) {
	services, err := s.serviceStore.ListByBusiness(ctx, b.ID)
	if err != nil {
		return "", nil, err
	}
	active := make([]models.Service, 0, len(services))
	for _, svc := range services {
		if svc.Active {
			active = append(active, svc)
		}
	}

	barbers, err := s.publicBarbers(ctx, b.ID)
	if err != nil {
		return "", nil, err
	}

	sections, err := s.publicSections(ctx, b.ID)
	if err != nil {
		return "", nil, err
	}

	sort.Slice(active, func(i, j int) bool { return active[i].Slug < active[j].Slug })
	sort.Slice(barbers, func(i, j int) bool { return barbers[i].ID.String() < barbers[j].ID.String() })
	sort.Slice(sections, func(i, j int) bool { return sections[i].Position < sections[j].Position })

	snap := publishSnapshot{
		Business: publishBusiness{
			Name:        b.Name,
			Slug:        b.Slug,
			AddressLine: b.AddressLine,
			City:        b.City,
			PostalCode:  b.PostalCode,
			Country:     b.Country,
			Phone:       b.Phone,
			Latitude:    b.Latitude,
			Longitude:   b.Longitude,
		},
		Services: make([]publishService, len(active)),
		Barbers:  barbers,
		Sections: sections,
	}
	for i, svc := range active {
		snap.Services[i] = publishService{
			Slug:            svc.Slug,
			Name:            svc.Name,
			Description:     svc.Description,
			Price:           svc.Price,
			DurationMinutes: svc.DurationMinutes,
			PictureID:       svc.PictureID,
		}
	}

	raw, err := json.Marshal(snap)
	if err != nil {
		return "", nil, fmt.Errorf("failed to serialize salon snapshot: %w", err)
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), s.salonURLs(b, active), nil
}

// publicBarbers returns the active staff who offer at least one service, with
// their resolved display name and avatar image ID — the representation shown on
// the public barbers page.
func (s *SalonPublishService) publicBarbers(ctx context.Context, businessID uuid.UUID) ([]publishBarber, error) {
	members, err := s.buStore.ListByBusiness(ctx, businessID)
	if err != nil {
		return nil, err
	}
	providerIDs, err := s.employeeServiceStore.ListServiceProviderIDs(ctx, businessID)
	if err != nil {
		return nil, err
	}
	provides := make(map[uuid.UUID]bool, len(providerIDs))
	for _, id := range providerIDs {
		provides[id] = true
	}

	barbers := make([]publishBarber, 0, len(members))
	for _, m := range members {
		if !m.Active || !provides[m.ID] {
			continue
		}
		displayName := m.DisplayName
		var avatarID *uuid.UUID
		if u, err := s.userStore.GetByID(ctx, m.UserID); err == nil {
			if displayName == "" {
				displayName = u.DisplayName
			}
			avatarID = u.AvatarID
		}
		if id := s.businessUserAvatarID(ctx, m.ID); id != nil {
			avatarID = id
		}
		barbers = append(barbers, publishBarber{
			ID:          m.ID,
			DisplayName: displayName,
			AvatarID:    avatarID,
		})
	}
	return barbers, nil
}

func (s *SalonPublishService) businessUserAvatarID(ctx context.Context, businessUserID uuid.UUID) *uuid.UUID {
	if s.imageLinkStore == nil {
		return nil
	}
	links, err := s.imageLinkStore.ListByEntity(ctx, "business_user", businessUserID)
	if err != nil {
		return nil
	}
	for _, l := range links {
		if l.Purpose == "avatar" {
			id := l.ImageID
			return &id
		}
	}
	return nil
}

func (s *SalonPublishService) publicSections(ctx context.Context, businessID uuid.UUID) ([]publishSection, error) {
	page, err := s.pageStore.GetByBusinessAndName(ctx, businessID, store.LandingPageName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []publishSection{}, nil
		}
		return nil, err
	}
	sections, err := s.sectionStore.ListByPage(ctx, page.ID)
	if err != nil {
		return nil, err
	}
	out := make([]publishSection, len(sections))
	for i, sec := range sections {
		out[i] = publishSection{
			Type:     sec.Type,
			Position: sec.Position,
			Content:  string(sec.Content),
		}
	}
	return out, nil
}

// salonURLs builds the absolute public URLs for a salon (home, services,
// barbers, and each active service detail page).
func (s *SalonPublishService) salonURLs(b *models.Business, services []models.Service) []string {
	base := strings.TrimSuffix(strings.TrimSpace(s.appDomain), "/")
	if base == "" {
		base = "fejd.fyi"
	}
	host := "https://" + b.Slug + "." + base
	urls := []string{host + "/", host + "/services", host + "/barbers"}
	for _, svc := range services {
		if svc.Slug == "" {
			continue
		}
		urls = append(urls, host+"/services/"+svc.Slug)
	}
	return urls
}

type publishSnapshot struct {
	Business publishBusiness  `json:"business"`
	Services []publishService `json:"services"`
	Barbers  []publishBarber  `json:"barbers"`
	Sections []publishSection `json:"sections"`
}

type publishBusiness struct {
	Name        string   `json:"name"`
	Slug        string   `json:"slug"`
	AddressLine string   `json:"address_line"`
	City        string   `json:"city"`
	PostalCode  string   `json:"postal_code"`
	Country     string   `json:"country"`
	Phone       string   `json:"phone"`
	Latitude    *float64 `json:"latitude"`
	Longitude   *float64 `json:"longitude"`
}

type publishService struct {
	Slug            string     `json:"slug"`
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	Price           float64    `json:"price"`
	DurationMinutes int        `json:"duration_minutes"`
	PictureID       *uuid.UUID `json:"picture_id"`
}

type publishBarber struct {
	ID          uuid.UUID  `json:"id"`
	DisplayName string     `json:"display_name"`
	AvatarID    *uuid.UUID `json:"avatar_id"`
}

type publishSection struct {
	Type     string `json:"type"`
	Position int    `json:"position"`
	Content  string `json:"content"`
}
