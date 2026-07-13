// Package byok defines context keys used to thread per-request BYOK credentials
// (Bring Your Own Key) through the search pipeline without polluting SearchQuery.
package byok

import (
	"context"

	"github.com/trippier/poi-api/pkg/types"
)

// genericKey is the context key type for the open-ended provider key store.
type genericKey struct{ id types.Provider }

// WithProviderKey returns a copy of ctx that carries key as the BYOK
// credential for provider id.
func WithProviderKey(ctx context.Context, id types.Provider, key string) context.Context {
	return context.WithValue(ctx, genericKey{id}, key)
}

// GetProviderKey returns the BYOK credential stored in ctx for provider id,
// or "" if none was set.
func GetProviderKey(ctx context.Context, id types.Provider) string {
	v, _ := ctx.Value(genericKey{id}).(string)
	return v
}
