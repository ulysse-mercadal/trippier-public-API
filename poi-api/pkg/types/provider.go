package types

// Provider identifies a data source adapter.
type Provider string

const (
	ProviderOverpass        Provider = "overpass"
	ProviderWikivoyage      Provider = "wikivoyage"
	ProviderGeoNames        Provider = "geonames"
	ProviderWikipedia       Provider = "wikipedia"
	ProviderWikipediaEvents Provider = "wikipedia_events"
)

// AllProviders is the default set used when no providers are specified on a places search.
// Wikipedia is intentionally excluded: its geosearch returns non-physical articles
// (historical events, meta-articles, organisations) that cannot be filtered reliably
// at query time without prohibitive latency. It is used for enrichment only.
var AllProviders = []Provider{
	ProviderOverpass,
	ProviderWikivoyage,
	ProviderGeoNames,
}

// AllEventProviders is the default set used when no providers are specified on an events search.
var AllEventProviders = []Provider{
	ProviderWikipediaEvents,
}

// ProviderStatus is returned by GET /pois/providers.
type ProviderStatus struct {
	Name      Provider `json:"name"`
	Available bool     `json:"available"`
	LatencyMs int64    `json:"latency_ms,omitempty"`
	Error     string   `json:"error,omitempty"`
}
