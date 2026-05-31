// Package providers defines the Provider interface that all data source adapters must implement.
package providers

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/trippier/poi-api/pkg/types"
)

// BuildConfig carries every input a provider's Factory may need. New
// providers can declare new dependencies by adding fields — existing
// factories ignore what they don't use.
type BuildConfig struct {
	Lang             string
	GeoNamesUsername string
	Redis            *redis.Client
	CacheTTL         time.Duration
	Log              *zap.Logger
}

// Factory builds a concrete Provider from BuildConfig. Returning (nil, nil)
// is a clean opt-out — used when a required env var is absent (e.g. GeoNames
// with no username) so the rest of the boot is unaffected.
type Factory func(BuildConfig) (Provider, error)

var factories = map[types.Provider]Factory{}

// Register associates a Factory with a provider id. Called from each provider
// package's init() so the binary only needs `_ "…/providers/<name>"` blank
// imports to enrol every backend — no central switch statement to maintain.
func Register(id types.Provider, f Factory) {
	factories[id] = f
}

// BuildAll instantiates every registered Factory with cfg. Providers that
// opted out (returned nil) are dropped silently; the first hard error stops
// the boot.
func BuildAll(cfg BuildConfig) ([]Provider, error) {
	out := make([]Provider, 0, len(factories))
	for _, f := range factories {
		p, err := f(cfg)
		if err != nil {
			return nil, err
		}
		if p == nil {
			continue
		}
		out = append(out, p)
	}
	return out, nil
}

const userAgent = "Trippier/1.0 (+https://github.com/trippier)"

// @param ref a Wikimedia file reference such as "File:Eiffel_Tower.jpg", "Eiffel_Tower.jpg" or "Category:Eiffel".
// @return the canonical Special:FilePath URL for the file, or empty string when ref is empty, a category, or otherwise unusable.
func CommonsFileURL(ref string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(ref), "category:") {
		return ""
	}
	if strings.HasPrefix(ref, "File:") || strings.HasPrefix(ref, "file:") {
		ref = ref[len("File:"):]
	}
	if ref == "" {
		return ""
	}
	return "https://commons.wikimedia.org/wiki/Special:FilePath/" + url.PathEscape(ref)
}

// @param s a URL string drawn from untrusted upstream data (OSM tags, MediaWiki listings).
// @return s if it starts with http:// / https:// / mailto:, otherwise the empty string — defangs javascript: and other XSS-prone schemes before they reach API consumers.
func SafeURL(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	lower := strings.ToLower(s)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") || strings.HasPrefix(lower, "mailto:") {
		return s
	}
	return ""
}

// SetUserAgent stamps the shared User-Agent on an outgoing request.
// All provider HTTP calls must use this so external APIs (Overpass, Wikimedia)
// can identify the application.
func SetUserAgent(req *http.Request) {
	req.Header.Set("User-Agent", userAgent)
}

// Provider is the interface implemented by every data source adapter.
type Provider interface {
	// Name returns the unique identifier of this provider.
	Name() types.Provider

	// SupportsMode reports whether this provider can handle the given search mode.
	SupportsMode(mode types.SearchMode) bool

	// Search fetches raw POIs matching the given query.
	// Providers must respect context cancellation and deadline.
	Search(ctx context.Context, q types.SearchQuery) ([]types.RawPoi, error)
}

// ByokProvider is an optional interface for providers that require a user-supplied key.
type ByokProvider interface {
	IsByok() bool
}

// Enricher is an optional interface for providers that lend attributes
// (WikidataID, SourceURL, Description, …) to nearby POIs returned by other
// providers. The core enrichment loop is provider-agnostic: it runs every
// registered Enricher and lets each one decide what to copy onto target
// POIs found within its declared EnrichmentRadius. Provider-specific
// borrowing rules live inside the implementing package, never in the core.
type Enricher interface {
	// EnrichmentRadius returns the maximum distance, in metres, at which one
	// of this provider's POIs is considered "the same place" as a target POI
	// for the purpose of borrowing attributes.
	EnrichmentRadius() float64

	// EnrichTarget applies this provider's contribution to target, given that
	// source is the nearest matching POI returned by this provider during the
	// current request. Implementations decide which fields to fill or replace.
	EnrichTarget(target *types.RawPoi, source types.RawPoi)
}
