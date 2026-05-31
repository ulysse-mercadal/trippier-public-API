// Package byok defines context keys used to thread per-request BYOK credentials
// (Bring Your Own Key) through the search pipeline without polluting SearchQuery.
package byok

import (
	"context"

	"github.com/trippier/poi-api/pkg/types"
)

// genericKey is the context key type for the open-ended provider key store.
type genericKey struct{ id types.Provider }

// @param ctx parent context.
// @param id provider identifier the key belongs to.
// @param key the user-supplied BYOK credential.
// @return a derived context carrying the key under a provider-specific slot.
func WithProviderKey(ctx context.Context, id types.Provider, key string) context.Context {
	return context.WithValue(ctx, genericKey{id}, key)
}

// @param ctx the request context.
// @param id provider identifier to look up.
// @return the stored BYOK key, or "" when none was set.
func GetProviderKey(ctx context.Context, id types.Provider) string {
	v, _ := ctx.Value(genericKey{id}).(string)
	return v
}
