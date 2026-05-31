// Package registry holds static metadata for every known provider: geographic confidence
// scores per ISO 3166-1 alpha-2 country and per POI category. This data drives the
// geo-aware auto-selection logic and powers the /pois/providers/catalog endpoint.
//
// Adding a new provider: append an entry to All. The backend implementation is independent;
// providers with Implemented=false appear in the catalog but are never auto-selected.
package registry

import "github.com/trippier/poi-api/pkg/types"

// Meta holds static metadata for one provider.
type Meta struct {
	ID             types.Provider
	Label          string
	Byok           bool
	ByokHeader     string // HTTP request header carrying the user key, e.g. "X-Foursquare-Key"
	ForEvents      bool   // true = event provider, false = POI provider
	Categories     []types.PoiType
	CountryScores  map[string]float64        // ISO 3166-1 alpha-2 (uppercase) → [0,1]; "*" = global default
	CategoryScores map[types.PoiType]float64 // category → [0,1]; empty = neutral (1.0)
	// Priority drives dedup tie-breaks: when several providers contribute to the
	// same merged POI, the one with the highest Priority is "primary" (canonical
	// ID, name, image ordering, etc.). Higher = preferred. Zero is acceptable for
	// providers that don't care which way the tie goes.
	Priority int
	// AccuracyMeters is the provider's typical coordinate precision. Used by the
	// dedup proximity threshold and Wikipedia-enrichment radius — a coarse
	// provider should not be discarded just because its coords are 80m off when
	// it says so up front. Zero means "use the global default".
	AccuracyMeters float64
	// MinRadius caps how small the event radius can shrink before this provider
	// returns useful results. Most event APIs return nothing under 50 km; map
	// providers don't care (leave 0).
	MinRadius int
}

// CountryScore returns the provider's confidence for a country code.
// Falls back to "*" global default, then 0.5 if neither is defined.
func (m Meta) CountryScore(cc string) float64 {
	if s, ok := m.CountryScores[cc]; ok {
		return s
	}
	if s, ok := m.CountryScores["*"]; ok {
		return s
	}
	return 0.5
}

// CategoryScore returns the average specialisation score for the requested types.
// Returns 1.0 when no types are requested or no category scores are defined.
func (m Meta) CategoryScore(requested []types.PoiType) float64 {
	if len(requested) == 0 || len(m.CategoryScores) == 0 {
		return 1.0
	}
	var total float64
	var n int
	for _, t := range requested {
		if s, ok := m.CategoryScores[t]; ok {
			total += s
			n++
		}
	}
	if n == 0 {
		return 0.5
	}
	return total / float64(n)
}

// Score returns the composite confidence for a country + category combination.
func (m Meta) Score(cc string, requested []types.PoiType) float64 {
	return m.CountryScore(cc) * m.CategoryScore(requested)
}

// All is the complete provider registry, keyed by provider ID.
// It contains both implemented providers (active in the runtime) and future ones
// (documented in the catalog, auto-selected when the backend is added).
var All = map[types.Provider]Meta{

	// ── Free / always-on providers ─────────────────────────────────────────────

	types.ProviderOverpass: {
		ID:             types.ProviderOverpass,
		Label:          "OpenStreetMap / Overpass",
		ForEvents:      false,
		Priority:       4,
		AccuracyMeters: 15,
		Categories:     []types.PoiType{types.TypeSee, types.TypeEat, types.TypeDrink, types.TypeDo, types.TypeBuy, types.TypeSleep},
		CountryScores: map[string]float64{
			"*": 0.75,
			// Europe — OSM is exceptional
			"DE": 0.97, "AT": 0.96, "CH": 0.95, "NL": 0.95, "BE": 0.93,
			"FR": 0.92, "GB": 0.92, "SE": 0.93, "NO": 0.92, "DK": 0.91,
			"FI": 0.91, "IT": 0.90, "ES": 0.90, "PT": 0.89, "PL": 0.88,
			"CZ": 0.88, "HU": 0.87, "RO": 0.82, "HR": 0.85, "SK": 0.85,
			"SI": 0.86, "GR": 0.83, "BG": 0.80, "RS": 0.78, "UA": 0.76,
			// Americas
			"US": 0.82, "CA": 0.83, "BR": 0.72, "MX": 0.70, "AR": 0.68,
			"CL": 0.67, "CO": 0.65, "PE": 0.62, "VE": 0.55, "BO": 0.55,
			// Asia-Pacific
			"AU": 0.84, "NZ": 0.83, "JP": 0.77, "KR": 0.65, "TW": 0.70,
			"SG": 0.80, "HK": 0.78, "MY": 0.70, "TH": 0.68, "ID": 0.65,
			"PH": 0.63, "VN": 0.62, "MM": 0.52, "KH": 0.50,
			"IN": 0.70, "PK": 0.55, "BD": 0.50, "LK": 0.58,
			"CN": 0.60, // OSM has decent China data but less dense than Baidu/Amap
			// Middle East
			"AE": 0.78, "SA": 0.70, "IL": 0.82, "TR": 0.75, "EG": 0.65,
			"JO": 0.68, "LB": 0.65, "KW": 0.70, "QA": 0.72, "BH": 0.72,
			// Africa
			"ZA": 0.72, "KE": 0.60, "NG": 0.52, "GH": 0.55, "ET": 0.45,
			"TZ": 0.55, "UG": 0.50, "MA": 0.65, "TN": 0.62, "DZ": 0.58,
		},
		CategoryScores: map[types.PoiType]float64{
			types.TypeSee: 0.88, types.TypeDo: 0.82, types.TypeEat: 0.72,
			types.TypeDrink: 0.70, types.TypeBuy: 0.68, types.TypeSleep: 0.62,
		},
	},

	types.ProviderWikivoyage: {
		ID:             types.ProviderWikivoyage,
		Label:          "Wikivoyage",
		ForEvents:      false,
		Priority:       3,
		AccuracyMeters: 40,
		Categories:     []types.PoiType{types.TypeSee, types.TypeDo, types.TypeEat, types.TypeDrink, types.TypeBuy, types.TypeSleep},
		CountryScores: map[string]float64{
			"*":  0.60,
			"FR": 0.88, "DE": 0.87, "IT": 0.87, "ES": 0.85, "GB": 0.85,
			"US": 0.83, "JP": 0.80, "CN": 0.65, "AU": 0.78, "CA": 0.78,
			"IN": 0.68, "BR": 0.65, "MX": 0.65, "RU": 0.72, "TR": 0.70,
			"TH": 0.68, "VN": 0.62, "ID": 0.60,
		},
		CategoryScores: map[types.PoiType]float64{
			types.TypeSee: 0.92, types.TypeDo: 0.85, types.TypeSleep: 0.78,
			types.TypeEat: 0.68, types.TypeDrink: 0.62, types.TypeBuy: 0.55,
		},
	},

	types.ProviderGeoNames: {
		ID:             types.ProviderGeoNames,
		Label:          "GeoNames",
		ForEvents:      false,
		Priority:       1,
		AccuracyMeters: 80,
		Categories:     []types.PoiType{types.TypeSee, types.TypeDo},
		CountryScores: map[string]float64{
			"*": 0.60,
		},
		CategoryScores: map[types.PoiType]float64{
			types.TypeSee: 0.70, types.TypeDo: 0.65,
		},
	},

	// ── Global BYOK providers ──────────────────────────────────────────────────

	types.ProviderFoursquare: {
		ID:         types.ProviderFoursquare,
		Label:      "Foursquare FSQ",
		Byok:       true,
		ByokHeader: "X-Foursquare-Key",
		ForEvents:  false,
		Categories: []types.PoiType{types.TypeSee, types.TypeEat, types.TypeDrink, types.TypeDo, types.TypeBuy, types.TypeSleep},
		CountryScores: map[string]float64{
			"*":  0.72,
			"US": 0.96, "CA": 0.92, "GB": 0.90, "AU": 0.90, "NZ": 0.86,
			"FR": 0.87, "DE": 0.87, "IT": 0.86, "ES": 0.85, "NL": 0.86,
			"BE": 0.84, "SE": 0.83, "NO": 0.82, "DK": 0.82, "PT": 0.80,
			"PL": 0.78, "CH": 0.86,
			"JP": 0.80, "KR": 0.72, "TW": 0.76, "HK": 0.82, "SG": 0.83,
			"MY": 0.74, "TH": 0.72, "ID": 0.70, "PH": 0.70, "VN": 0.67,
			"CN": 0.52, // Ctrip partnership but limited access
			"IN": 0.67, "LK": 0.58,
			"BR": 0.74, "MX": 0.72, "AR": 0.70, "CL": 0.68, "CO": 0.65,
			"AE": 0.78, "SA": 0.68, "IL": 0.75, "TR": 0.74, "EG": 0.58,
			"QA": 0.72, "KW": 0.70, "BH": 0.70,
			"ZA": 0.62, "KE": 0.45, "NG": 0.42, "GH": 0.45, "MA": 0.55,
		},
		CategoryScores: map[types.PoiType]float64{
			types.TypeEat: 0.93, types.TypeDrink: 0.91, types.TypeBuy: 0.82,
			types.TypeSleep: 0.78, types.TypeSee: 0.80, types.TypeDo: 0.74,
		},
	},

	types.ProviderHere: {
		ID:         types.ProviderHere,
		Label:      "HERE Places",
		Byok:       true,
		ByokHeader: "X-Here-Key",
		ForEvents:  false,
		Categories: []types.PoiType{types.TypeSee, types.TypeEat, types.TypeDrink, types.TypeDo, types.TypeBuy, types.TypeSleep},
		CountryScores: map[string]float64{
			"*":  0.70,
			"DE": 0.95, "AT": 0.93, "CH": 0.92, "NL": 0.92, "BE": 0.90,
			"FR": 0.91, "GB": 0.91, "IT": 0.89, "ES": 0.88, "SE": 0.88,
			"NO": 0.87, "DK": 0.87, "FI": 0.86, "PL": 0.84,
			"US": 0.90, "CA": 0.89, "AU": 0.86, "NZ": 0.82,
			"AE": 0.90, "SA": 0.85, "IL": 0.87, "TR": 0.78, "EG": 0.72,
			"QA": 0.82, "KW": 0.80, "BH": 0.80, "JO": 0.70, "MA": 0.68,
			"JP": 0.74, "KR": 0.70, "SG": 0.80, "MY": 0.72, "HK": 0.78,
			"TW": 0.72, "TH": 0.68, "ID": 0.65, "PH": 0.62, "VN": 0.60,
			"IN": 0.70, "BR": 0.72, "MX": 0.70, "AR": 0.65,
			"CN": 0.42, "ZA": 0.65, "KE": 0.48, "NG": 0.42,
		},
		CategoryScores: map[types.PoiType]float64{
			types.TypeSleep: 0.85, types.TypeSee: 0.82, types.TypeEat: 0.80,
			types.TypeDo: 0.77, types.TypeBuy: 0.76, types.TypeDrink: 0.74,
		},
	},

	// ── China BYOK providers ───────────────────────────────────────────────────

	types.ProviderBaidu: {
		ID:         types.ProviderBaidu,
		Label:      "Baidu Maps",
		Byok:       true,
		ByokHeader: "X-Baidu-Key",
		ForEvents:  false,
		Categories: []types.PoiType{types.TypeSee, types.TypeEat, types.TypeDrink, types.TypeDo, types.TypeBuy},
		CountryScores: map[string]float64{
			"CN": 0.95,
			"*":  0.05,
		},
		CategoryScores: map[types.PoiType]float64{
			types.TypeEat: 0.92, types.TypeDrink: 0.88, types.TypeBuy: 0.86,
			types.TypeSee: 0.85, types.TypeDo: 0.80,
		},
	},

	types.ProviderAmap: {
		ID:         types.ProviderAmap,
		Label:      "Amap / AutoNavi",
		Byok:       true,
		ByokHeader: "X-Amap-Key",
		ForEvents:  false,
		Categories: []types.PoiType{types.TypeSee, types.TypeEat, types.TypeDrink, types.TypeDo, types.TypeBuy},
		CountryScores: map[string]float64{
			"CN": 0.97,
			"*":  0.05,
		},
		CategoryScores: map[types.PoiType]float64{
			types.TypeEat: 0.95, types.TypeDrink: 0.90, types.TypeBuy: 0.88,
			types.TypeSee: 0.87, types.TypeDo: 0.82,
		},
	},

	// ── Korea BYOK providers ───────────────────────────────────────────────────

	types.ProviderKakao: {
		ID:         types.ProviderKakao,
		Label:      "Kakao Maps",
		Byok:       true,
		ByokHeader: "X-Kakao-Key",
		ForEvents:  false,
		Categories: []types.PoiType{types.TypeSee, types.TypeEat, types.TypeDrink, types.TypeDo, types.TypeBuy},
		CountryScores: map[string]float64{
			"KR": 0.95,
			"*":  0.05,
		},
		CategoryScores: map[types.PoiType]float64{
			types.TypeEat: 0.92, types.TypeDrink: 0.88, types.TypeBuy: 0.84,
			types.TypeSee: 0.80, types.TypeDo: 0.78,
		},
	},

	// ── Japan BYOK providers ───────────────────────────────────────────────────

	types.ProviderNavitime: {
		ID:         types.ProviderNavitime,
		Label:      "NAVITIME",
		Byok:       true,
		ByokHeader: "X-Navitime-Key",
		ForEvents:  false,
		Categories: []types.PoiType{types.TypeSee, types.TypeDo, types.TypeEat},
		CountryScores: map[string]float64{
			"JP": 0.95,
			"*":  0.05,
		},
		CategoryScores: map[types.PoiType]float64{
			types.TypeSee: 0.92, types.TypeDo: 0.87, types.TypeEat: 0.76,
		},
	},

	// ── India BYOK providers ───────────────────────────────────────────────────

	types.ProviderMappls: {
		ID:         types.ProviderMappls,
		Label:      "Mappls / MapmyIndia",
		Byok:       true,
		ByokHeader: "X-Mappls-Key",
		ForEvents:  false,
		Categories: []types.PoiType{types.TypeSee, types.TypeEat, types.TypeDrink, types.TypeDo, types.TypeBuy, types.TypeSleep},
		CountryScores: map[string]float64{
			"IN": 0.97,
			"*":  0.05,
		},
		CategoryScores: map[types.PoiType]float64{
			types.TypeSee: 0.88, types.TypeSleep: 0.86, types.TypeEat: 0.85,
			types.TypeDo: 0.80, types.TypeBuy: 0.80, types.TypeDrink: 0.75,
		},
	},

	// ── Southeast Asia BYOK providers ─────────────────────────────────────────

	types.ProviderGrabMaps: {
		ID:         types.ProviderGrabMaps,
		Label:      "GrabMaps",
		Byok:       true,
		ByokHeader: "X-Grabmaps-Key",
		ForEvents:  false,
		Categories: []types.PoiType{types.TypeSee, types.TypeEat, types.TypeDrink, types.TypeDo, types.TypeBuy, types.TypeSleep},
		CountryScores: map[string]float64{
			"SG": 0.96, "ID": 0.93, "MY": 0.92, "TH": 0.91,
			"PH": 0.89, "VN": 0.89, "MM": 0.83, "KH": 0.81,
			"*": 0.05,
		},
		CategoryScores: map[types.PoiType]float64{
			types.TypeEat: 0.91, types.TypeDrink: 0.86, types.TypeSleep: 0.82,
			types.TypeBuy: 0.79, types.TypeSee: 0.76, types.TypeDo: 0.73,
		},
	},

	// ── Event providers ────────────────────────────────────────────────────────

	types.ProviderWikipedia: {
		ID:             types.ProviderWikipedia,
		Label:          "Wikipedia",
		ForEvents:      false,
		Priority:       2,
		AccuracyMeters: 35,
		Categories:     []types.PoiType{types.TypeSee},
		// Intentionally absent from default-providers selection: Wikipedia
		// geosearch returns too many non-physical articles. It is used only
		// for enrichment (filling WikidataID / SourceURL / Description on
		// the POIs returned by other providers).
		CountryScores: map[string]float64{
			"*": 0.0,
		},
	},

	types.ProviderWikipediaEvents: {
		ID:             types.ProviderWikipediaEvents,
		Label:          "Wikipedia Events",
		ForEvents:      true,
		Priority:       2,
		AccuracyMeters: 35,
		Categories:     []types.PoiType{types.TypeEvent},
		CountryScores: map[string]float64{
			"*":  0.65,
			"US": 0.85, "GB": 0.85, "FR": 0.82, "DE": 0.82, "JP": 0.78,
		},
	},

	types.ProviderTicketmaster: {
		ID:         types.ProviderTicketmaster,
		Label:      "Ticketmaster",
		Byok:       true,
		ByokHeader: "X-Ticketmaster-Key",
		ForEvents:  true,
		Priority:   3,
		MinRadius:  50_000,
		Categories: []types.PoiType{types.TypeEvent},
		CountryScores: map[string]float64{
			"*":  0.38,
			"US": 0.97, "CA": 0.93, "GB": 0.91, "AU": 0.89, "NZ": 0.86,
			"IE": 0.89, "DE": 0.84, "FR": 0.83, "ES": 0.81, "IT": 0.79,
			"MX": 0.76, "BR": 0.72, "AR": 0.68, "CL": 0.65,
		},
	},

	types.ProviderEventbrite: {
		ID:         types.ProviderEventbrite,
		Label:      "Eventbrite",
		Byok:       true,
		ByokHeader: "X-Eventbrite-Token",
		ForEvents:  true,
		Priority:   3,
		MinRadius:  50_000,
		Categories: []types.PoiType{types.TypeEvent},
		CountryScores: map[string]float64{
			"*":  0.42,
			"US": 0.93, "GB": 0.89, "CA": 0.86, "AU": 0.83, "NZ": 0.80,
			"FR": 0.79, "DE": 0.79, "ES": 0.76, "IT": 0.73, "NL": 0.75,
			"BR": 0.72, "MX": 0.70, "IN": 0.67, "SG": 0.68,
		},
	},

	types.ProviderMeetup: {
		ID:         types.ProviderMeetup,
		Label:      "Meetup",
		Byok:       true,
		ByokHeader: "X-Meetup-Token",
		ForEvents:  true,
		Categories: []types.PoiType{types.TypeEvent},
		CountryScores: map[string]float64{
			"*":  0.48,
			"US": 0.92, "GB": 0.87, "AU": 0.83, "CA": 0.83, "IN": 0.74,
			"DE": 0.79, "FR": 0.76, "NL": 0.78, "SG": 0.72, "JP": 0.65,
			"BR": 0.65, "ES": 0.70, "SE": 0.72, "CH": 0.74,
		},
	},

	types.ProviderOpenAgenda: {
		ID:         types.ProviderOpenAgenda,
		Label:      "OpenAgenda",
		Byok:       true,
		ByokHeader: "X-OpenAgenda-Key",
		ForEvents:  true,
		Categories: []types.PoiType{types.TypeEvent},
		CountryScores: map[string]float64{
			"FR": 0.96, "BE": 0.77, "CH": 0.72, "LU": 0.72,
			"*": 0.12,
		},
	},
}
