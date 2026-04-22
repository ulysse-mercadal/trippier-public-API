package types

// Provider identifies a data source adapter.
type Provider string

const (
	ProviderOverpass   Provider = "overpass"
	ProviderWikivoyage Provider = "wikivoyage"
	ProviderGeoNames   Provider = "geonames"
	ProviderWikipedia  Provider = "wikipedia"
)

// AllProviders is the default set used when the caller specifies none.
var AllProviders = []Provider{
	ProviderOverpass,
	ProviderWikivoyage,
	ProviderGeoNames,
	ProviderWikipedia,
}

// ProviderStatus is returned by GET /pois/providers.
type ProviderStatus struct {
	Name      Provider `json:"name"`
	Available bool     `json:"available"`
	LatencyMs int64    `json:"latency_ms,omitempty"`
	Error     string   `json:"error,omitempty"`
}
