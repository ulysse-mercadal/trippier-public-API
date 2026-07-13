// Package scoring computes relevance scores for merged POI candidates.
package scoring

import (
	"math"

	"github.com/trippier/poi-api/pkg/types"
)

var defaultTypeWeights = map[types.PoiType]float64{
	types.TypeSee:   1.0,
	types.TypeDo:    0.8,
	types.TypeEat:   0.6,
	types.TypeDrink: 0.5,
	types.TypeSleep: 0.4,
	types.TypeBuy:   0.4,
}

// Score computes an overall relevance score for poi given the search query
// q, combining source count, type weight, distance, and coordinate
// precision. It returns a value in [0, 100].
func Score(poi types.EnrichedPoi, q types.SearchQuery) float64 {
	s := sourceScore(len(poi.Sources))*50 +
		typeScore(poi.Type, q.Weights)*30 +
		distanceScore(poi.Distance, float64(q.Radius))*10 +
		coordScore(poi)*10
	return math.Min(s, 100)
}

// sourceScore ranks POIs by count, the number of sources reporting the POI,
// favoring multi-provider results. It returns a score in [0, 1].
func sourceScore(count int) float64 {
	switch {
	case count >= 3:
		return 1.0
	case count == 2:
		return 0.70
	default:
		return 0.25
	}
}

// typeScore rates POI type t against the caller-supplied weights map, or the
// package defaults when weights is empty. It returns a normalized score in
// [0, 1].
func typeScore(t types.PoiType, weights map[types.PoiType]float64) float64 {
	if len(weights) == 0 {
		if w := defaultTypeWeights[t]; w != 0 {
			return w
		}
		return 0.5
	}
	w, ok := weights[t]
	if !ok {
		return 0.2
	}
	if m := maxWeight(weights); m != 0 {
		return w / m
	}
	return 0.5
}

// distanceScore linearly rewards proximity, scoring dist, the POI's
// distance from the search center, against radius, the search radius. It
// returns a score in [0, 1], or 0 if the POI is outside the radius.
func distanceScore(dist, radius float64) float64 {
	if radius <= 0 || dist >= radius {
		return 0
	}
	return 1 - (dist / radius)
}

// coordScore rates the coordinate precision of poi, favoring exact over
// approximate locations. It returns a score in [0, 1].
func coordScore(poi types.EnrichedPoi) float64 {
	if poi.Coords == nil {
		return 0
	}
	if poi.Coords.Approximate {
		return 0.5
	}
	return 1.0
}

// maxWeight finds the largest weight value in weights, a map of POI type
// weights. It returns the maximum weight, or 0 if the map is empty.
func maxWeight(weights map[types.PoiType]float64) float64 {
	var m float64
	for _, v := range weights {
		if v > m {
			m = v
		}
	}
	return m
}
