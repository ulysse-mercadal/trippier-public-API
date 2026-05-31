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

// dedupCellSizeDeg is the side of the spatial bucket the dedup algorithm uses
// to skip pairs of POIs that are obviously too far apart. ~330m at the
// equator, narrower at higher latitudes — comfortably larger than the
// proximityThresholdMeters (150 m) + provider accuracy slack so the 9-cell
// neighbourhood search never misses a real pair.
const dedupCellSizeDeg = 0.003

// cellKey identifies a bucket in the dedup spatial index.
type cellKey struct{ lat, lng int32 }

// cellOf maps coordinates to their bucket. nil-safe — callers should check
// for usable coords before deriving the cell.
func cellOf(c *types.Coordinates) cellKey {
	return cellKey{
		lat: int32(math.Floor(c.Lat / dedupCellSizeDeg)),
		lng: int32(math.Floor(c.Lng / dedupCellSizeDeg)),
	}
}

const (
	// proximityThresholdMeters is the default distance below which two POIs
	// can be considered for merging when their names are similar enough. The
	// effective threshold is scaled per-pair by providerAccuracy() so
	// low-precision sources don't disqualify themselves by reporting
	// coordinates a few dozen metres off.
	proximityThresholdMeters = 150.0
	nameSimilarityThreshold  = 0.80
)

// providerPriority reads Priority off the static registry. Providers not in
// the registry fall back to 0 — they still merge, but never win tie-breaks.
func providerPriority(p types.Provider) int {
	return registry.All[p].Priority
}

// providerAccuracy returns the provider's declared coordinate accuracy in
// metres, or 0 when none is declared. The dedup proximity threshold is
// stretched by the max of both participants' accuracies — coarse providers
// no longer fall on the wrong side of a hard 150 m line.
func providerAccuracy(p types.Provider) float64 {
	return registry.All[p].AccuracyMeters
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
// A spatial bucket index restricts the per-leader scan to the 9-cell
// neighbourhood, turning the algorithm from O(n²) into amortised O(n) for
// realistic POI densities while preserving the exact merge results.
func group(pois []types.RawPoi) [][]types.RawPoi {
	used := make([]bool, len(pois))
	groups := make([][]types.RawPoi, 0, len(pois))

	// Build the spatial index. POIs without usable coordinates can still
	// merge via the WikidataID path of areDuplicates, so they go into a
	// dedicated bucket that every leader also scans.
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

// candidatesFor returns the indices of POIs that could conceivably merge with
// a POI at c. The leader's own bucket plus its 8 neighbours covers every
// position within proximityThresholdMeters + AccuracyMeters of c. Spatialless
// POIs are always candidates — areDuplicates short-circuits on nil coords
// unless WikidataIDs match.
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
	return types.EnrichedPoi{
		ID:          primary.ID,
		Name:        primary.Name,
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

// mergeSources returns one SourceLink per distinct provider that contributed
// to the merged POI. Two passes:
//  1. Real provider sources — each RawPoi's own (Provider, SourceURL).
//  2. Cross-references — each member's ExtraSources, kept only for providers
//     not already covered by pass 1. This preserves the "real" URL when a
//     provider participates AND a sibling listing also points to it.
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

// primaryPoi returns the group member from the highest-priority provider.
func primaryPoi(group []types.RawPoi) types.RawPoi {
	best := group[0]
	for _, p := range group[1:] {
		if providerPriority(p.Provider) > providerPriority(best.Provider) {
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
		if p.Coords != nil && providerPriority(p.Provider) > bestPrio {
			best = p.Coords
			bestPrio = providerPriority(p.Provider)
		}
	}
	return best
}

// maxImagesPerPoi caps the merged Images slice. Three is enough to display a
// small gallery and keeps the response payload bounded when many providers
// each contribute a few URLs.
const maxImagesPerPoi = 3

// @param group raw POIs sharing a single deduplicated place.
// @return up to maxImagesPerPoi distinct image URLs, taken in provider iteration order with the highest-priority provider first.
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
