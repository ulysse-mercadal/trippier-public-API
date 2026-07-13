// Package types defines shared data types for POI search requests and results.
package types

// SearchMode defines the geographic search strategy used for a query.
type SearchMode string

const (
	ModeRadius   SearchMode = "radius"
	ModePolygon  SearchMode = "polygon"
	ModeDistrict SearchMode = "district"
)

// SearchQuery holds all parameters for a POI search request, including per-type
// weights and custom-route-only overrides (CountryHint, ProviderWeights, ExcludeProviders).
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

	// CountryHint overrides the geo-detected country code (ISO 3166-1 alpha-2, e.g. "CN"),
	// used by auto-selection and recommend instead of calling Nominatim.
	CountryHint string `form:"country_hint" json:"country_hint,omitempty"`

	// ExcludeProviders is a blacklist applied on top of auto-selection or Providers.
	ExcludeProviders []Provider `form:"exclude_providers" json:"exclude_providers,omitempty"`

	// ProviderWeights overrides registry confidence scores for auto-selection (values in [0,1]);
	// providers with an effective score below 0.1 are excluded.
	ProviderWeights map[Provider]float64 `form:"-" json:"provider_weights,omitempty"`
}

// SearchResult is the top-level API response body for GET /pois/search.
type SearchResult struct {
	Query   SearchQuery   `json:"query"`
	Total   int           `json:"total"`
	Results []EnrichedPoi `json:"results"`
}
