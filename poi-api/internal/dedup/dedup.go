// Package dedup merges POIs from multiple providers into deduplicated EnrichedPoi records.
package dedup

import (
	"strings"

	"github.com/trippier/poi-api/internal/mathutil"
	"github.com/trippier/poi-api/pkg/types"
)

const (
	proximityThresholdMeters = 150.0
	nameSimilarityThreshold  = 0.80
)

var providerPriority = map[types.Provider]int{
	types.ProviderOverpass:        4,
	types.ProviderWikivoyage:      3,
	types.ProviderWikipedia:       2,
	types.ProviderGeoNames:        1,
	types.ProviderTicketmaster:    3,
	types.ProviderEventbrite:      3,
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

// group clusters raw POIs into duplicate groups using a greedy pairwise match.
func group(pois []types.RawPoi) [][]types.RawPoi {
	used := make([]bool, len(pois))
	groups := make([][]types.RawPoi, 0, len(pois))

	for i := range pois {
		if used[i] {
			continue
		}
		g := []types.RawPoi{pois[i]}
		used[i] = true
		for j := i + 1; j < len(pois); j++ {
			if used[j] {
				continue
			}
			for _, member := range g {
				if member.Coords == nil || member.Coords.Approximate {
					continue
				}
				if areDuplicates(member, pois[j]) {
					g = append(g, pois[j])
					used[j] = true
					break
				}
			}
		}
		groups = append(groups, g)
	}
	return groups
}

// areDuplicates returns true when two POIs refer to the same place or event.
// For events (those with a DateStart), two POIs are only duplicates when they
// also share the same start date — different dates mean different occurrences.
func areDuplicates(a, b types.RawPoi) bool {
	if a.WikidataID != "" && a.WikidataID == b.WikidataID {
		return true
	}
	if a.Coords == nil || a.Coords.Approximate || b.Coords == nil || b.Coords.Approximate {
		return false
	}
	if mathutil.Haversine(a.Coords.Lat, a.Coords.Lng, b.Coords.Lat, b.Coords.Lng) >= proximityThresholdMeters {
		return false
	}

	// Events at the same venue are only duplicates if they start on the same day.
	if a.DateStart != nil && b.DateStart != nil {
		ay, am, ad := a.DateStart.Date()
		by, bm, bd := b.DateStart.Date()
		if ay != by || am != bm || ad != bd {
			return false
		}
	}

	an, bn := normalizeName(a.Name), normalizeName(b.Name)

	if mathutil.JaroWinkler(an, bn) >= nameSimilarityThreshold {
		return tokenOverlapOK(an, bn)
	}

	short, long := an, bn
	if len(short) > len(long) {
		short, long = long, short
	}
	return len(short) >= 8 && strings.Contains(long, short)
}

// tokenOverlapOK guards against JW prefix-bonus false positives (e.g. "Hotel A"
// vs "Hotel B"). Requires shared_words/min(|a|,|b|) > 0.5 when both names
// have ≥ 2 words; single-word names rely on JW alone.
func tokenOverlapOK(a, b string) bool {
	wa, wb := strings.Fields(a), strings.Fields(b)
	if len(wa) < 2 || len(wb) < 2 {
		return true
	}
	set := make(map[string]bool, len(wb))
	for _, w := range wb {
		set[w] = true
	}
	shared := 0
	for _, w := range wa {
		if set[w] {
			shared++
		}
	}
	return float64(shared)/float64(min(len(wa), len(wb))) > 0.5
}

var diacriticReplacer = strings.NewReplacer(
	"é", "e", "è", "e", "ê", "e", "ë", "e",
	"à", "a", "â", "a", "ä", "a",
	"ô", "o", "ö", "o", "œ", "oe",
	"û", "u", "ù", "u", "ü", "u",
	"î", "i", "ï", "i",
	"ç", "c",
	"-", " ",
)

// normalizeName lowercases, trims, and replaces diacritics and hyphens for comparison.
func normalizeName(s string) string {
	return diacriticReplacer.Replace(strings.ToLower(strings.TrimSpace(s)))
}

// toEnriched builds an EnrichedPoi from a group by picking the highest-priority provider as primary.
func toEnriched(group []types.RawPoi) types.EnrichedPoi {
	primary := primaryPoi(group)
	sources := make([]types.Provider, 0, len(group))
	data := make(map[types.Provider]types.RawPoi, len(group))
	seen := make(map[types.Provider]bool, len(group))

	for _, p := range group {
		if !seen[p.Provider] {
			sources = append(sources, p.Provider)
			seen[p.Provider] = true
		}
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
		DateStart:     primary.DateStart,
		DateEnd:       primary.DateEnd,
		Recurring:     primary.Recurring,
	}
}

// primaryPoi returns the group member from the highest-priority provider.
func primaryPoi(group []types.RawPoi) types.RawPoi {
	best := group[0]
	for _, p := range group[1:] {
		if providerPriority[p.Provider] > providerPriority[best.Provider] {
			best = p
		}
	}
	return best
}

// bestCoords returns the coordinates from the highest-priority provider that has them.
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

// mergeContact fills each Contact field with the first non-empty value across the group.
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

// firstNonEmpty returns the first non-empty string extracted from the group by fn.
func firstNonEmpty(group []types.RawPoi, fn func(types.RawPoi) string) string {
	for _, p := range group {
		if v := fn(p); v != "" {
			return v
		}
	}
	return ""
}
