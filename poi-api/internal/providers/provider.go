// Package providers defines the Provider interface that all data source adapters must implement.
package providers

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/trippier/poi-api/pkg/types"
)

// langCodeRe matches a syntactically valid Wikimedia language code such as
// "en", "fr", "zh-yue" or "bat-smg": a lowercase letter followed by up to
// eleven more lowercase alphanumerics or hyphens.
var langCodeRe = regexp.MustCompile(`^[a-z][a-z0-9-]{0,11}$`)

// NormalizeLang lowercases and validates a user-supplied language code so it
// is safe to interpolate into a Wikimedia subdomain (guarding against SSRF via
// dots, slashes, or an over-long value). It returns the normalized code, or ""
// when lang is empty or malformed, in which case callers fall back to their
// default edition.
func NormalizeLang(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if !langCodeRe.MatchString(lang) {
		return ""
	}
	return lang
}

// BuildConfig carries every input a provider's Factory may need.
type BuildConfig struct {
	Lang             string
	GeoNamesUsername string
	Redis            *redis.Client
	CacheTTL         time.Duration
	Log              *zap.Logger
}

// Factory builds a concrete Provider from BuildConfig; returning (nil, nil) is a clean opt-out (e.g. a missing env var).
type Factory func(BuildConfig) (Provider, error)

var factories = map[types.Provider]Factory{}

// Register associates the Factory f with the provider identifier id. It is
// called from each provider package's init().
func Register(id types.Provider, f Factory) {
	factories[id] = f
}

// BuildAll instantiates every registered Factory with cfg, dropping opt-outs
// (nil) silently and stopping on the first error. It returns the
// successfully built providers, or an error from the first factory that
// failed.
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

// CommonsFileURL builds the direct file URL for a Wikimedia Commons
// reference, such as "File:Eiffel_Tower.jpg", "Eiffel_Tower.jpg" or
// "Category:Eiffel". It returns the canonical Special:FilePath URL for the
// file, or an empty string when ref is empty, a category, or otherwise
// unusable.
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

// SafeURL filters s, a URL string drawn from untrusted upstream data (OSM
// tags, MediaWiki listings), down to allow-listed schemes before it is
// exposed to API consumers. It returns s if it starts with http://,
// https://, or mailto:, otherwise the empty string — defanging javascript:
// and other XSS-prone schemes before they reach API consumers.
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

// SetUserAgent stamps the shared User-Agent on req, the outgoing HTTP
// request.
func SetUserAgent(req *http.Request) {
	req.Header.Set("User-Agent", userAgent)
}

// Provider is the interface implemented by every data source adapter.
type Provider interface {
	// Name returns the unique identifier of this provider.
	Name() types.Provider

	// SupportsMode reports whether this provider can handle the given search
	// mode.
	SupportsMode(mode types.SearchMode) bool

	// Search fetches raw POIs matching q, the search query to fulfil.
	// Providers must respect context cancellation and deadline via ctx. It
	// returns the matching raw POIs, or an error.
	Search(ctx context.Context, q types.SearchQuery) ([]types.RawPoi, error)
}

// ByokProvider is an optional interface for providers that require a user-supplied key.
type ByokProvider interface {
	// IsByok reports whether this provider requires a user-supplied key.
	IsByok() bool
}

// Enricher lets a provider lend attributes (WikidataID, SourceURL, Description, …) to nearby POIs found by other providers within its EnrichmentRadius.
type Enricher interface {
	// EnrichmentRadius returns the max distance in metres for treating one of
	// this provider's POIs as the same place as a target.
	EnrichmentRadius() float64

	// EnrichTarget applies this provider's contribution to target (mutated in
	// place) using source, the nearest matching POI from this provider.
	EnrichTarget(target *types.RawPoi, source types.RawPoi)
}
