// Package indexing submits public salon URLs to search engines for
// (re)indexing. The real implementation targets the Google Indexing API; the
// NoopIndexer is a placeholder wired until then.
package indexing

import "context"

// Indexer submits absolute URLs to a search engine so they are crawled or
// re-crawled promptly.
type Indexer interface {
	PublishURLs(ctx context.Context, urls []string) error
}

// NoopIndexer drops every submission. It keeps the publish flow operational
// until the Google Indexing API client is implemented.
type NoopIndexer struct{}

func (NoopIndexer) PublishURLs(context.Context, []string) error { return nil }
