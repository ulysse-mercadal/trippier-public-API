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

// PointKind classifies whether a result is a place or a time-bound event, orthogonal to PoiType.
type PointKind string

const (
	KindPOI   PointKind = "poi"
	KindEvent PointKind = "event"
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
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Kind        PointKind    `json:"kind"`
	Type        PoiType      `json:"type"`
	Coords      *Coordinates `json:"coords,omitempty"`
	Zone        *Zone        `json:"zone,omitempty"`
	Description string       `json:"description,omitempty"`
	Contact     Contact      `json:"contact,omitempty"`
	Thumbnail   string       `json:"thumbnail,omitempty"`
	Images      []string     `json:"images,omitempty"`
	Provider    Provider     `json:"provider"`
	WikidataID  string       `json:"wikidata_id,omitempty"`
	// SourceURL is the canonical link to this POI's page on the originating
	// provider (e.g. https://www.openstreetmap.org/node/12345). Empty when
	// the provider does not expose a stable browse URL.
	SourceURL string  `json:"source_url,omitempty"`
	Distance  float64 `json:"distance,omitempty"`
	// ExtraSources lets a provider declare cross-references it knows about
	// (e.g. a Wikivoyage listing carrying a wikipedia= or wikidata= field).
	// They surface as additional entries in EnrichedPoi.Sources after the
	// dedup pass but never produce standalone RawPoi records of their own,
	// so the pipeline never has to filter phantom POIs.
	ExtraSources []SourceLink `json:"-"`
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

// SlimEvent is the lightweight projection returned by GET /events/slim.
type SlimEvent struct {
	Name      string       `json:"name"`
	Coords    *Coordinates `json:"coords,omitempty"`
	DateStart *time.Time   `json:"date_start,omitempty"`
	DateEnd   *time.Time   `json:"date_end,omitempty"`
	Recurring bool         `json:"recurring,omitempty"`
}

// SlimEventResult is the top-level response body for GET /events/slim.
type SlimEventResult struct {
	Total   int         `json:"total"`
	Results []SlimEvent `json:"results"`
}

// SourceLink names one provider that contributed to a merged POI and the
// canonical URL clients can follow for richer detail on that source.
type SourceLink struct {
	Provider Provider `json:"provider"`
	URL      string   `json:"url,omitempty"`
}

// EnrichedPoi is the final merged, scored result returned to the caller;
// per-provider detail is folded into the top-level fields — follow Sources[i].URL for more.
type EnrichedPoi struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Kind        PointKind    `json:"kind"`
	Type        PoiType      `json:"type"`
	Score       float64      `json:"score"`
	Coords      *Coordinates `json:"coords,omitempty"`
	Zone        *Zone        `json:"zone,omitempty"`
	Distance    float64      `json:"distance"`
	Description string       `json:"description,omitempty"`
	Contact     Contact      `json:"contact,omitempty"`
	Thumbnail   string       `json:"thumbnail,omitempty"`
	Images      []string     `json:"images,omitempty"`
	Sources     []SourceLink `json:"sources"`
	// Event-specific fields — nil/zero for non-event POIs.
	DateStart *time.Time `json:"date_start,omitempty"`
	DateEnd   *time.Time `json:"date_end,omitempty"`
	Recurring bool       `json:"recurring,omitempty"`
}
