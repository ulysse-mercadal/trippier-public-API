// Package byok defines context keys used to thread per-request BYOK credentials
// (Bring Your Own Key) through the search pipeline without polluting SearchQuery.
//
// Callers extract credentials from request headers and inject them with
// context.WithValue; providers read them back and skip the call when absent.
package byok

import "context"

// key is an unexported type to prevent collisions with other packages.
type key int

const (
	// TicketmasterKey stores the caller's Ticketmaster API key in the context.
	TicketmasterKey key = iota
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
