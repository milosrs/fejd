package service

import (
	"context"
	"fmt"
)

// webClientID is the Keycloak client that serves the salon web app. Salon
// subdomains must be registered as valid redirect URIs and web origins on it so
// users can log in from <slug>.<app-domain>.
const webClientID = "fejd-frontend"

// ClientRedirectManager is the Keycloak admin surface needed to register a
// salon subdomain on the web client. *keycloak.Client satisfies it; tests
// provide a fake.
type ClientRedirectManager interface {
	EnsureClientRedirect(ctx context.Context, clientID, redirectURI, webOrigin string) error
}

// SalonDomainService registers salon subdomains on the Keycloak web client so
// each salon's <slug>.<app-domain> URL is an accepted redirect URI and web
// origin.
type SalonDomainService struct {
	clients   ClientRedirectManager
	appDomain string
}

func NewSalonDomainService(clients ClientRedirectManager, appDomain string) *SalonDomainService {
	return &SalonDomainService{clients: clients, appDomain: appDomain}
}

// RegisterSalon ensures the salon's subdomain is registered on the web client.
// It is a no-op when subdomain mode is disabled (empty app domain, e.g. local
// dev path-based routing).
func (s *SalonDomainService) RegisterSalon(ctx context.Context, slug string) error {
	if s.appDomain == "" {
		return nil
	}
	redirectURI := fmt.Sprintf("https://%s.%s/*", slug, s.appDomain)
	webOrigin := fmt.Sprintf("https://%s.%s", slug, s.appDomain)
	return s.clients.EnsureClientRedirect(ctx, webClientID, redirectURI, webOrigin)
}
