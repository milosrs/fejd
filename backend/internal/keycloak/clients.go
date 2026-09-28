package keycloak

import (
	"context"
)

// EnsureClientRedirect registers the given redirect URI and web origin on the
// named client, preserving every other field of the client representation. It
// is idempotent: values already present are not duplicated.
func (c *Client) EnsureClientRedirect(ctx context.Context, clientID, redirectURI, webOrigin string) error {
	clientUUID, err := c.clientUUID(ctx, clientID)
	if err != nil {
		return err
	}

	path := "/admin/realms/" + c.realm + "/clients/" + clientUUID

	var rep map[string]any
	if err := c.do(ctx, "GET", path, nil, &rep); err != nil {
		return err
	}

	redirects, _ := rep["redirectUris"].([]any)
	origins, _ := rep["webOrigins"].([]any)

	if !anyContains(redirects, redirectURI) {
		redirects = append(redirects, redirectURI)
		rep["redirectUris"] = redirects
	}
	if !anyContains(origins, webOrigin) {
		origins = append(origins, webOrigin)
		rep["webOrigins"] = origins
	}

	return c.do(ctx, "PUT", path, rep, nil)
}

// anyContains reports whether items contains the string s.
func anyContains(items []any, s string) bool {
	for _, item := range items {
		if v, ok := item.(string); ok && v == s {
			return true
		}
	}
	return false
}
