// Package providers defines the Provider interface that all data source adapters must implement.
package providers

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/trippier/poi-api/pkg/types"
)

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

// Pingable is an optional interface for providers that offer a lightweight
// health-check endpoint. ProvidersStatus uses Ping instead of Search when available,
// avoiding quota consumption on quota-constrained APIs (Ticketmaster, Eventbrite).
type Pingable interface {
	Ping(ctx context.Context) error
}

// ByokProvider is an optional interface for providers that require a user-supplied key.
type ByokProvider interface {
	IsByok() bool
}
