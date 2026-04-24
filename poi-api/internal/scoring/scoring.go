// Package scoring computes relevance scores for merged POI candidates.
package scoring

import (
	"math"

	"github.com/trippier/poi-api/pkg/types"
)

// defaultTypeWeights ranks POI types by tourist relevance when the caller
// has not provided explicit weights. "see" and "do" score highest because
// travellers prioritise landmarks and activities over generic services.
var defaultTypeWeights = map[types.PoiType]float64{
	types.TypeSee:   1.0,
	types.TypeDo:    0.8,
	types.TypeEat:   0.6,
	types.TypeDrink: 0.5,
	types.TypeSleep: 0.4,
	types.TypeBuy:   0.4,
	// TypeGeneric falls through to the 0.5 fallback in typeScore.
}

// Score computes a relevance score in [0, 100] for an enriched POI.
// Weights:
//   - type match      40 — what kind of POI it is
//   - source count    30 — confirmed by multiple independent sources
//   - distance        20 — closer is better, but doesn't dominate
//   - coord precision 10 — precise location preferred over approximate
func Score(poi types.EnrichedPoi, q types.SearchQuery) float64 {
	s := typeScore(poi.Type, q.Weights)*40 +
		sourceScore(len(poi.Sources))*30 +
		distanceScore(poi.Distance, float64(q.Radius))*20 +
		coordScore(poi)*10
	return math.Min(s, 100)
}

func typeScore(t types.PoiType, weights map[types.PoiType]float64) float64 {
	if len(weights) == 0 {
		w := defaultTypeWeights[t]
		if w == 0 {
			return 0.5
		}
		return w
	}
	w, ok := weights[t]
	if !ok {
		return 0.2
	}
	max := maxWeight(weights)
	if max == 0 {
		return 0.5
	}
	return w / max
}

func sourceScore(count int) float64 {
	return math.Min(float64(count)/float64(len(types.AllProviders)), 1.0)
}

func distanceScore(dist, radius float64) float64 {
	if radius <= 0 || dist >= radius {
		return 0
	}
	return 1 - (dist / radius)
}

func coordScore(poi types.EnrichedPoi) float64 {
	if poi.Coords == nil {
		return 0
	}
	if poi.Coords.Approximate {
		return 0.5
	}
	return 1.0
}

func maxWeight(weights map[types.PoiType]float64) float64 {
	var m float64
	for _, v := range weights {
		if v > m {
			m = v
		}
	}
	return m
}
