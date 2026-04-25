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

// Score returns a relevance score in [0, 100].
// Weights: source count 50%, type 30%, distance 10%, coord precision 10%.
func Score(poi types.EnrichedPoi, q types.SearchQuery) float64 {
	s := sourceScore(len(poi.Sources))*50 +
		typeScore(poi.Type, q.Weights)*30 +
		distanceScore(poi.Distance, float64(q.Radius))*10 +
		coordScore(poi)*10
	return math.Min(s, 100)
}

// sourceScore uses a stepped function so multi-provider POIs always outrank
// single-provider ones regardless of distance or type bonuses.
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
