package types

import "time"

// PoiType represents the category of a point of interest,
// aligned with Wikivoyage listing sections.
type PoiType string

const (
	TypeSee     PoiType = "see"
	TypeEat     PoiType = "eat"
	TypeDrink   PoiType = "drink"
	TypeDo      PoiType = "do"
	TypeBuy     PoiType = "buy"
	TypeSleep   PoiType = "sleep"
	TypeGeneric PoiType = "generic"
	TypeEvent   PoiType = "event"
)

// Contact groups reachability information for a POI.
type Contact struct {
	Website string `json:"website,omitempty"`
	Phone   string `json:"phone,omitempty"`
	Email   string `json:"email,omitempty"`
	Hours   string `json:"opening_hours,omitempty"`
}

// RawPoi is the normalised output of a single provider before merging.
// The ID field is namespaced as "{provider}:{native_id}".
type RawPoi struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        PoiType           `json:"type"`
	Coords      *Coordinates      `json:"coords,omitempty"`
	Zone        *Zone             `json:"zone,omitempty"`
	Description string            `json:"description,omitempty"`
	Contact     Contact           `json:"contact,omitempty"`
	Thumbnail   string            `json:"thumbnail,omitempty"`
	Provider    Provider          `json:"provider"`
	WikidataID  string            `json:"wikidata_id,omitempty"`
	Distance    float64           `json:"distance,omitempty"`
	// Event-specific fields — nil/zero for non-event POIs.
	DateStart *time.Time `json:"date_start,omitempty"`
	DateEnd   *time.Time `json:"date_end,omitempty"`
	Recurring bool       `json:"recurring,omitempty"`
}

// SlimPoi is the lightweight projection returned by GET /pois/search/slim.
type SlimPoi struct {
	Name   string       `json:"name"`
	Type   PoiType      `json:"type"`
	Coords *Coordinates `json:"coords,omitempty"`
}

// SlimResult is the top-level response body for GET /pois/search/slim.
type SlimResult struct {
	Total   int       `json:"total"`
	Results []SlimPoi `json:"results"`
}

// EnrichedPoi is the final merged and scored result returned to the caller.
type EnrichedPoi struct {
	ID            string              `json:"id"`
	Name          string              `json:"name"`
	Type          PoiType             `json:"type"`
	Score         float64             `json:"score"`
	Coords        *Coordinates        `json:"coords,omitempty"`
	Zone          *Zone               `json:"zone,omitempty"`
	Distance      float64             `json:"distance"`
	Description   string              `json:"description,omitempty"`
	Contact       Contact             `json:"contact,omitempty"`
	Thumbnail     string              `json:"thumbnail,omitempty"`
	Sources       []Provider          `json:"sources"`
	ProvidersData map[Provider]RawPoi `json:"providers_data,omitempty"`
	// Event-specific fields — nil/zero for non-event POIs.
	DateStart *time.Time `json:"date_start,omitempty"`
	DateEnd   *time.Time `json:"date_end,omitempty"`
	Recurring bool       `json:"recurring,omitempty"`
}
