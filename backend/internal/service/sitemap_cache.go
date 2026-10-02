package service

import (
	"sync"
	"time"
)

// SitemapCache is a short-lived in-memory cache for the generated sitemap.xml.
// It is invalidated when a salon is published so the change shows up
// immediately, with a TTL backstop for anything changed outside the publish
// flow.
type SitemapCache struct {
	mu   sync.Mutex
	body []byte
	at   time.Time
	ttl  time.Duration
}

func NewSitemapCache(ttl time.Duration) *SitemapCache {
	return &SitemapCache{ttl: ttl}
}

// Get returns the cached sitemap body and true when it is still fresh.
func (c *SitemapCache) Get() ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.body == nil || time.Since(c.at) >= c.ttl {
		return nil, false
	}
	return c.body, true
}

// Set stores the sitemap body and stamps the current time.
func (c *SitemapCache) Set(body []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.body = body
	c.at = time.Now()
}

// Invalidate clears the cache so the next request regenerates the sitemap.
func (c *SitemapCache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.body = nil
}
