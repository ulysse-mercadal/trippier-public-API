// Package providers defines the Provider interface that all data source adapters must implement.
package providers

import (
	"context"

	"github.com/trippier/poi-api/pkg/types"
)

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
