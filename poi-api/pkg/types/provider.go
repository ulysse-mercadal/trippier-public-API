package types

// Provider identifies a data source adapter.
type Provider string

const (
	// ── Free / always-on ──────────────────────────────────────────────────────.
	ProviderOverpass        Provider = "overpass"
	ProviderWikivoyage      Provider = "wikivoyage"
	ProviderGeoNames        Provider = "geonames"
	ProviderWikipedia       Provider = "wikipedia"
	ProviderWikipediaEvents Provider = "wikipedia_events"

	// ── Global BYOK ───────────────────────────────────────────────────────────.
	ProviderFoursquare Provider = "foursquare"
	ProviderHere       Provider = "here"

	// ── China BYOK ────────────────────────────────────────────────────────────.
	ProviderBaidu Provider = "baidu"
	ProviderAmap  Provider = "amap"

	// ── Korea BYOK ────────────────────────────────────────────────────────────.
	ProviderKakao Provider = "kakao"

	// ── Japan BYOK ────────────────────────────────────────────────────────────.
	ProviderNavitime Provider = "navitime"

	// ── India BYOK ────────────────────────────────────────────────────────────.
	ProviderMappls Provider = "mappls"

	// ── Southeast Asia BYOK ───────────────────────────────────────────────────.
	ProviderGrabMaps Provider = "grabmaps"

	// ── Event providers ───────────────────────────────────────────────────────.
	ProviderTicketmaster Provider = "ticketmaster"
	ProviderEventbrite   Provider = "eventbrite"
	ProviderMeetup       Provider = "meetup"
	ProviderOpenAgenda   Provider = "openagenda"
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
// Includes live event providers (Ticketmaster, Eventbrite) alongside Wikipedia festivals.
var AllEventProviders = []Provider{
	ProviderWikipediaEvents,
	ProviderTicketmaster,
	ProviderEventbrite,
}

// ProviderStatus is returned by GET /pois/providers.
type ProviderStatus struct {
	Name      Provider `json:"name"`
	Available bool     `json:"available"`
	LatencyMs int64    `json:"latency_ms"`
	Byok      bool     `json:"byok,omitempty"`
	Error     string   `json:"error,omitempty"`
}

// ProviderCatalogEntry is one item in the GET /pois/providers/catalog response.
// It combines static registry metadata with the runtime "implemented" flag.
type ProviderCatalogEntry struct {
	ID             Provider            `json:"id"`
	Label          string              `json:"label"`
	Byok           bool                `json:"byok"`
	ByokHeader     string              `json:"byok_header,omitempty"`
	ForEvents      bool                `json:"for_events"`
	Categories     []PoiType           `json:"categories"`
	CountryScores  map[string]float64  `json:"country_scores"`
	CategoryScores map[PoiType]float64 `json:"category_scores,omitempty"`
	Implemented    bool                `json:"implemented"` // backend exists and is registered
}

// RecommendedProvider is one ranked entry in GET /pois/providers/recommend.
type RecommendedProvider struct {
	ID          Provider `json:"id"`
	Label       string   `json:"label"`
	Score       float64  `json:"score"`
	Byok        bool     `json:"byok"`
	ByokHeader  string   `json:"byok_header,omitempty"`
	ForEvents   bool     `json:"for_events"`
	Implemented bool     `json:"implemented"`
}

// RecommendResult is the top-level response for GET /pois/providers/recommend.
type RecommendResult struct {
	CountryCode string                `json:"country_code"`
	Providers   []RecommendedProvider `json:"providers"`
}
