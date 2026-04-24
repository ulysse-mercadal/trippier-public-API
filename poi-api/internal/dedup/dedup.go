// Package dedup merges duplicate POIs collected from multiple providers into
// unified EnrichedPoi records, using Wikidata ID matching and proximity + name similarity.
package dedup

import (
	"strings"

	"github.com/trippier/poi-api/internal/mathutil"
	"github.com/trippier/poi-api/pkg/types"
)

const (
	proximityThresholdMeters = 50.0
	nameSimilarityThreshold  = 0.80
)

var providerPriority = map[types.Provider]int{
	types.ProviderOverpass:   4,
	types.ProviderWikivoyage: 3,
	types.ProviderWikipedia:  2,
	types.ProviderGeoNames:   1,
}

// Merge groups raw POIs from all providers into deduplicated EnrichedPoi records.
func Merge(pois []types.RawPoi) []types.EnrichedPoi {
	groups := group(pois)
	result := make([]types.EnrichedPoi, 0, len(groups))
	for _, g := range groups {
		result = append(result, toEnriched(g))
	}
	return result
}

func group(pois []types.RawPoi) [][]types.RawPoi {
	used := make([]bool, len(pois))
	groups := make([][]types.RawPoi, 0, len(pois))

	for i := range pois {
		if used[i] {
			continue
		}
		g := []types.RawPoi{pois[i]}
		for j := i + 1; j < len(pois); j++ {
			if !used[j] && areDuplicates(pois[i], pois[j]) {
				g = append(g, pois[j])
				used[j] = true
			}
		}
		used[i] = true
		groups = append(groups, g)
	}
	return groups
}

func areDuplicates(a, b types.RawPoi) bool {
	if a.WikidataID != "" && a.WikidataID == b.WikidataID {
		return true
	}
	aApprox := a.Coords == nil || a.Coords.Approximate
	bApprox := b.Coords == nil || b.Coords.Approximate

	// When both POIs have precise coordinates, check proximity first — it is
	// cheaper than JaroWinkler and quickly eliminates distant duplicates.
	if !aApprox && !bApprox {
		dist := mathutil.Haversine(a.Coords.Lat, a.Coords.Lng, b.Coords.Lat, b.Coords.Lng)
		if dist >= proximityThresholdMeters {
			return false
		}
	}

	// Name similarity decides the final outcome.
	similarity := mathutil.JaroWinkler(
		strings.ToLower(strings.TrimSpace(a.Name)),
		strings.ToLower(strings.TrimSpace(b.Name)),
	)
	return similarity >= nameSimilarityThreshold
}

func toEnriched(group []types.RawPoi) types.EnrichedPoi {
	primary := primaryPoi(group)
	sources := make([]types.Provider, 0, len(group))
	data := make(map[types.Provider]types.RawPoi, len(group))

	for _, p := range group {
		sources = append(sources, p.Provider)
		data[p.Provider] = p
	}

	return types.EnrichedPoi{
		ID:            primary.ID,
		Name:          primary.Name,
		Type:          primary.Type,
		Coords:        bestCoords(group),
		Zone:          primary.Zone,
		Distance:      primary.Distance,
		Description:   firstNonEmpty(group, func(p types.RawPoi) string { return p.Description }),
		Thumbnail:     firstNonEmpty(group, func(p types.RawPoi) string { return p.Thumbnail }),
		Contact:       mergeContact(group),
		Sources:       sources,
		ProvidersData: data,
	}
}

func primaryPoi(group []types.RawPoi) types.RawPoi {
	best := group[0]
	for _, p := range group[1:] {
		if providerPriority[p.Provider] > providerPriority[best.Provider] {
			best = p
		}
	}
	return best
}

func bestCoords(group []types.RawPoi) *types.Coordinates {
	var best *types.Coordinates
	bestPrio := -1
	for _, p := range group {
		if p.Coords != nil && providerPriority[p.Provider] > bestPrio {
			best = p.Coords
			bestPrio = providerPriority[p.Provider]
		}
	}
	return best
}

func mergeContact(group []types.RawPoi) types.Contact {
	var c types.Contact
	for _, p := range group {
		if c.Website == "" {
			c.Website = p.Contact.Website
		}
		if c.Phone == "" {
			c.Phone = p.Contact.Phone
		}
		if c.Hours == "" {
			c.Hours = p.Contact.Hours
		}
		if c.Email == "" {
			c.Email = p.Contact.Email
		}
	}
	return c
}

func firstNonEmpty(group []types.RawPoi, fn func(types.RawPoi) string) string {
	for _, p := range group {
		if v := fn(p); v != "" {
			return v
		}
	}
	return ""
}
