package types

// Coordinates holds a geographic position.
// Approximate is true when the position is derived from a zone centroid
// rather than the exact POI location.
type Coordinates struct {
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	Approximate bool    `json:"approximate"`
}

// GeoJSONPolygon is a minimal representation of a GeoJSON Polygon geometry.
// Coordinates follows the GeoJSON spec: [ring][point][lng, lat].
type GeoJSONPolygon struct {
	Type        string         `json:"type"`        // always "Polygon"
	Coordinates [][][2]float64 `json:"coordinates"` // outer ring first, then holes
}

// Zone describes the approximate area of a POI whose precise location is unknown.
// Polygon is an optional GeoJSON Polygon geometry returned when the exact
// position of a POI is unknown.
type Zone struct {
	Name    string          `json:"name"`
	Polygon *GeoJSONPolygon `json:"polygon,omitempty"`
	Source  Provider        `json:"source"`
}
