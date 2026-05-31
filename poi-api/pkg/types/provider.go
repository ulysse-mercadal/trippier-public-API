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

// Default provider sets are now derived from the registry at startup — see
// search.DefaultProviders / search.DefaultEventProviders.

// ProviderStatus is returned by GET /pois/providers. It carries only the
// static metadata we can produce without calling the upstream — historical
// availability/latency fields were dropped because the probe consumed paid
// quota without delivering reliable signal.
type ProviderStatus struct {
	Name Provider `json:"name"`
	Byok bool     `json:"byok,omitempty"`
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
