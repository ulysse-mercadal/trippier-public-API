// Package dedup merges POIs from multiple providers into deduplicated EnrichedPoi records.
package dedup

import (
	"math"
	"sort"
	"strings"

	"github.com/trippier/poi-api/internal/mathutil"
	"github.com/trippier/poi-api/internal/registry"
	"github.com/trippier/poi-api/pkg/types"
)

// dedupCellSizeDeg is the side of the spatial bucket used to skip POI pairs that are obviously too far apart.
const dedupCellSizeDeg = 0.003

// cellKey identifies a bucket in the dedup spatial index.
type cellKey struct{ lat, lng int32 }

// cellOf maps coordinates c (which must be non-nil) to their bucket key.
func cellOf(c *types.Coordinates) cellKey {
	return cellKey{
		lat: int32(math.Floor(c.Lat / dedupCellSizeDeg)),
		lng: int32(math.Floor(c.Lng / dedupCellSizeDeg)),
	}
}

const (
	// proximityThresholdMeters is the base distance in metres below which two POIs may merge, scaled per pair by providerAccuracy.
	proximityThresholdMeters = 150.0
	nameSimilarityThreshold  = 0.80
)

// providerPriority reads provider p's priority from the static registry,
// returning 0 if p is unknown.
func providerPriority(p types.Provider) int {
	return registry.All[p].Priority
}

// providerAccuracy returns provider p's declared coordinate accuracy in
// metres, or 0 if undeclared.
func providerAccuracy(p types.Provider) float64 {
	return registry.All[p].AccuracyMeters
}

// Merge groups raw POIs from all providers into deduplicated EnrichedPoi
// records, returning the resulting slice of merged POIs.
func Merge(pois []types.RawPoi) []types.EnrichedPoi {
	groups := group(pois)
	result := make([]types.EnrichedPoi, 0, len(groups))
	for _, g := range groups {
		result = append(result, toEnriched(g))
	}
	return result
}

// group clusters pois into duplicate groups using a greedy pairwise match,
// restricted to a 9-cell spatial neighbourhood for near-O(n) performance. It
// returns the resulting groups of duplicate POIs.
func group(pois []types.RawPoi) [][]types.RawPoi {
	used := make([]bool, len(pois))
	groups := make([][]types.RawPoi, 0, len(pois))

	buckets := make(map[cellKey][]int, len(pois))
	var spatialless []int
	for i, p := range pois {
		if p.Coords == nil || p.Coords.Approximate {
			spatialless = append(spatialless, i)
			continue
		}
		k := cellOf(p.Coords)
		buckets[k] = append(buckets[k], i)
	}

	for i := range pois {
		if used[i] {
			continue
		}
		g := []types.RawPoi{pois[i]}
		used[i] = true

		for _, j := range candidatesFor(pois[i].Coords, buckets, spatialless) {
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

// candidatesFor returns the indices of POIs that could plausibly merge with
// a POI at coordinates c, using the 9-cell neighbourhood of buckets plus the
// spatialless indices (POIs without usable coordinates). It returns the
// resulting candidate POI indices to check for duplication.
func candidatesFor(c *types.Coordinates, buckets map[cellKey][]int, spatialless []int) []int {
	out := append([]int(nil), spatialless...)
	if c == nil || c.Approximate {
		return out
	}
	k := cellOf(c)
	for dLat := int32(-1); dLat <= 1; dLat++ {
		for dLng := int32(-1); dLng <= 1; dLng++ {
			out = append(out, buckets[cellKey{lat: k.lat + dLat, lng: k.lng + dLng}]...)
		}
	}
	return out
}

// areDuplicates reports whether POIs a and b are the same place, matching
// WikidataID, proximity, name similarity, and (for events) start date.
func areDuplicates(a, b types.RawPoi) bool {
	if a.WikidataID != "" && a.WikidataID == b.WikidataID {
		return true
	}
	if a.Coords == nil || a.Coords.Approximate || b.Coords == nil || b.Coords.Approximate {
		return false
	}
	threshold := proximityThresholdMeters
	if acc := providerAccuracy(a.Provider); acc > 0 {
		threshold += acc
	}
	if acc := providerAccuracy(b.Provider); acc > 0 {
		threshold += acc
	}
	if mathutil.Haversine(a.Coords.Lat, a.Coords.Lng, b.Coords.Lat, b.Coords.Lng) >= threshold {
		return false
	}

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

// tokenOverlapOK guards against Jaro-Winkler prefix-bonus false positives by
// requiring >50% shared-word overlap between the normalized names a and b
// for multi-word names. It returns true if the names share enough words, or
// if either isn't multi-word.
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

// normalizeName lowercases, trims, and replaces diacritics and hyphens in s
// for comparison, returning the normalized name.
func normalizeName(s string) string {
	return diacriticReplacer.Replace(strings.ToLower(strings.TrimSpace(s)))
}

// toEnriched builds an EnrichedPoi from group by picking the
// highest-priority provider as primary, returning the merged enriched POI.
func toEnriched(group []types.RawPoi) types.EnrichedPoi {
	primary := primaryPoi(group)
	return types.EnrichedPoi{
		ID:          primary.ID,
		Name:        primary.Name,
		Kind:        primary.Kind,
		Type:        primary.Type,
		Coords:      bestCoords(group),
		Zone:        primary.Zone,
		Distance:    primary.Distance,
		Description: firstNonEmpty(group, func(p types.RawPoi) string { return p.Description }),
		Thumbnail:   firstNonEmpty(group, func(p types.RawPoi) string { return p.Thumbnail }),
		Images:      mergeImages(group),
		Contact:     mergeContact(group),
		Sources:     mergeSources(group),
		DateStart:   primary.DateStart,
		DateEnd:     primary.DateEnd,
		Recurring:   primary.Recurring,
	}
}

// mergeSources returns one SourceLink per distinct provider that contributed to
// the merged POI: first each member's own (Provider, SourceURL), then any
// ExtraSources cross-references for providers not already covered by the first
// pass.
func mergeSources(group []types.RawPoi) []types.SourceLink {
	out := make([]types.SourceLink, 0, len(group)*2)
	seen := make(map[types.Provider]bool, len(group))
	for _, p := range group {
		if seen[p.Provider] {
			continue
		}
		seen[p.Provider] = true
		out = append(out, types.SourceLink{Provider: p.Provider, URL: p.SourceURL})
	}
	for _, p := range group {
		for _, extra := range p.ExtraSources {
			if extra.Provider == "" || seen[extra.Provider] {
				continue
			}
			seen[extra.Provider] = true
			out = append(out, extra)
		}
	}
	return out
}

// primaryPoi returns the member of group from the highest-priority provider.
func primaryPoi(group []types.RawPoi) types.RawPoi {
	best := group[0]
	for _, p := range group[1:] {
		if providerPriority(p.Provider) > providerPriority(best.Provider) {
			best = p
		}
	}
	return best
}

// bestCoords returns the coordinates from group's highest-priority provider
// that has them, or nil if none do.
func bestCoords(group []types.RawPoi) *types.Coordinates {
	var best *types.Coordinates
	bestPrio := -1
	for _, p := range group {
		if p.Coords != nil && providerPriority(p.Provider) > bestPrio {
			best = p.Coords
			bestPrio = providerPriority(p.Provider)
		}
	}
	return best
}

// maxImagesPerPoi caps the number of image URLs returned per merged POI.
const maxImagesPerPoi = 3

// mergeImages returns up to maxImagesPerPoi distinct image URLs from group,
// highest-priority provider first.
func mergeImages(group []types.RawPoi) []string {
	ordered := make([]types.RawPoi, len(group))
	copy(ordered, group)
	sort.SliceStable(ordered, func(i, j int) bool {
		return providerPriority(ordered[i].Provider) > providerPriority(ordered[j].Provider)
	})
	out := make([]string, 0, maxImagesPerPoi)
	seen := make(map[string]bool, maxImagesPerPoi)
	for _, p := range ordered {
		for _, u := range p.Images {
			if len(out) >= maxImagesPerPoi {
				return out
			}
			if u == "" || seen[u] {
				continue
			}
			out = append(out, u)
			seen[u] = true
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// mergeContact fills each Contact field with the first non-empty value
// across group, returning the merged contact info.
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

// firstNonEmpty returns the first non-empty string extracted from group by
// the field extractor fn, or "" if none is found.
func firstNonEmpty(group []types.RawPoi, fn func(types.RawPoi) string) string {
	for _, p := range group {
		if v := fn(p); v != "" {
			return v
		}
	}
	return ""
}
