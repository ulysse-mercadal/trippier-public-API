package types

// Coordinates holds a geographic position.
// Approximate is true when the position is derived from a zone centroid
// rather than the exact POI location.
type Coordinates struct {
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	Approximate bool    `json:"approximate"`
}

// Zone describes the approximate area of a POI whose precise location is unknown.
type Zone struct {
	Name   string   `json:"name"`
	Source Provider `json:"source"`
}
