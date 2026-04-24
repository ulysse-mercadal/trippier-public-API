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
}

// SearchResult is the top-level API response body for GET /pois/search.
type SearchResult struct {
	Query   SearchQuery   `json:"query"`
	Total   int           `json:"total"`
	Results []EnrichedPoi `json:"results"`
}
