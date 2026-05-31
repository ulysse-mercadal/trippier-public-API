package types

// SearchMode defines the geographic search strategy used for a query.
type SearchMode string

const (
	ModeRadius   SearchMode = "radius"
	ModePolygon  SearchMode = "polygon"
	ModeDistrict SearchMode = "district"
)

// SearchQuery holds all parameters for a POI search request.
// Weights maps each PoiType to a relative importance factor (e.g. {"see":2,"eat":1}).
//
// Custom-route-only fields (CountryHint, ProviderWeights, ExcludeProviders) are parsed
// by the custom handler and ignored on the standard auto route.
type SearchQuery struct {
	Mode      SearchMode          `form:"mode"      json:"mode"      binding:"omitempty,oneof=radius polygon district"`
	Lat       float64             `form:"lat"       json:"lat"`
	Lng       float64             `form:"lng"       json:"lng"`
	Radius    int                 `form:"radius"    json:"radius"`
	Polygon   string              `form:"polygon"   json:"polygon,omitempty"`
	District  string              `form:"district"  json:"district,omitempty"`
	Providers []Provider          `form:"providers" json:"providers"`
	Types     []PoiType           `form:"types"     json:"types,omitempty"`
	Weights   map[PoiType]float64 `form:"-"         json:"weights,omitempty"`
	Lang      string              `form:"lang"      json:"lang"`
	Limit     int                 `form:"limit"     json:"limit"`
	Offset    int                 `form:"offset"    json:"offset"`
	MinScore  float64             `form:"min_score" json:"min_score,omitempty"`
	Date      string              `form:"date"      json:"date,omitempty"`

	// ── Custom-route fields ────────────────────────────────────────────────────
	// CountryHint overrides geo-detected country code (ISO 3166-1 alpha-2, e.g. "CN").
	// When set, auto-selection and recommend use this code instead of calling Nominatim.
	CountryHint string `form:"country_hint" json:"country_hint,omitempty"`

	// ExcludeProviders is a blacklist applied on top of auto-selection or the explicit
	// providers list. Comma-separated on the wire, same as Providers.
	ExcludeProviders []Provider `form:"exclude_providers" json:"exclude_providers,omitempty"`

	// ProviderWeights overrides registry confidence scores for the auto-selection step.
	// Parsed from the "provider_weights" query param as a JSON object, e.g.
	// {"overpass":0.2,"foursquare":0.9}. Values must be in [0, 1].
	// Providers with an effective score below 0.1 are excluded from auto-selection.
	ProviderWeights map[Provider]float64 `form:"-" json:"provider_weights,omitempty"`
}

// SearchResult is the top-level API response body for GET /pois/search.
type SearchResult struct {
	Query   SearchQuery   `json:"query"`
	Total   int           `json:"total"`
	Results []EnrichedPoi `json:"results"`
}
