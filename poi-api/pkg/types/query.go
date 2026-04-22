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
	Mode      SearchMode          `form:"mode"      binding:"required,oneof=radius polygon district"`
	Lat       float64             `form:"lat"`
	Lng       float64             `form:"lng"`
	Radius    int                 `form:"radius"`
	Polygon   string              `form:"polygon"`
	District  string              `form:"district"`
	Providers []Provider          `form:"providers"`
	Types     []PoiType           `form:"types"`
	Weights   map[PoiType]float64 `form:"-"`
	Lang      string              `form:"lang"`
	Limit     int                 `form:"limit"`
	Offset    int                 `form:"offset"`
	MinScore  float64             `form:"min_score"`
}

// SearchResult is the top-level API response body for GET /pois/search.
type SearchResult struct {
	Query   SearchQuery   `json:"query"`
	Total   int           `json:"total"`
	Results []EnrichedPoi `json:"results"`
}
