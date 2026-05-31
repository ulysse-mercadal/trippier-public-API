// Package byok defines context keys used to thread per-request BYOK credentials
// (Bring Your Own Key) through the search pipeline without polluting SearchQuery.
//
// Two APIs coexist:
//   - Legacy specific keys (TicketmasterKey, EventbriteKey constants + typed accessors) —
//     kept for the existing Ticketmaster and Eventbrite provider implementations.
//   - Generic provider key (WithProviderKey / GetProviderKey) — used for all new providers;
//     keyed by provider ID so the set of supported headers is open-ended.
package byok

import (
	"context"

	"github.com/trippier/poi-api/pkg/types"
)

// legacyKey is an unexported type for the original specific-provider context keys.
type legacyKey int

const (
	// TicketmasterKey stores the caller's Ticketmaster API key in the context.
	TicketmasterKey legacyKey = iota
	// EventbriteKey stores the caller's Eventbrite private token in the context.
	EventbriteKey
)

// TicketmasterAPIKey returns the Ticketmaster API key stored in ctx, or "".
func TicketmasterAPIKey(ctx context.Context) string {
	v, _ := ctx.Value(TicketmasterKey).(string)
	return v
}

// EventbriteToken returns the Eventbrite private token stored in ctx, or "".
func EventbriteToken(ctx context.Context) string {
	v, _ := ctx.Value(EventbriteKey).(string)
	return v
}

// genericKey is the context key type for the open-ended provider key store.
type genericKey struct{ id types.Provider }

// WithProviderKey stores a BYOK API key for the given provider in the context.
// Use GetProviderKey to retrieve it inside a provider implementation.
func WithProviderKey(ctx context.Context, id types.Provider, key string) context.Context {
	return context.WithValue(ctx, genericKey{id}, key)
}

// GetProviderKey returns the BYOK key stored for the given provider, or "".
func GetProviderKey(ctx context.Context, id types.Provider) string {
	v, _ := ctx.Value(genericKey{id}).(string)
	return v
}
