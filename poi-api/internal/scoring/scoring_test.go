package scoring_test

import (
	"testing"

	"github.com/trippier/poi-api/internal/scoring"
	"github.com/trippier/poi-api/pkg/types"
)

func TestScoreRange(t *testing.T) {
	poi := types.EnrichedPoi{
		Type:     types.TypeSee,
		Distance: 200,
		Sources:  []types.Provider{types.ProviderOverpass, types.ProviderWikipedia},
		Coords:   &types.Coordinates{Lat: 48.85, Lng: 2.35, Approximate: false},
	}
	q := types.SearchQuery{
		Mode:   types.ModeRadius,
		Radius: 5000,
		Weights: map[types.PoiType]float64{
			types.TypeSee: 2,
			types.TypeEat: 1,
		},
	}

	s := scoring.Score(poi, q)
	if s < 0 || s > 100 {
		t.Errorf("score %f outside [0, 100]", s)
	}
}

func TestScoreWeightedTypeBetter(t *testing.T) {
	base := types.EnrichedPoi{
		Distance: 100,
		Sources:  []types.Provider{types.ProviderOverpass},
		Coords:   &types.Coordinates{},
	}
	q := types.SearchQuery{
		Radius: 5000,
		Weights: map[types.PoiType]float64{
			types.TypeSee: 2,
			types.TypeEat: 1,
		},
	}

	see := base
	see.Type = types.TypeSee
	eat := base
	eat.Type = types.TypeEat

	if scoring.Score(see, q) <= scoring.Score(eat, q) {
		t.Error("TypeSee (weight 2) should score higher than TypeEat (weight 1)")
	}
}

func TestScoreMoreSourcesBetter(t *testing.T) {
	q := types.SearchQuery{Radius: 5000}
	one := types.EnrichedPoi{
		Distance: 100,
		Sources:  []types.Provider{types.ProviderOverpass},
		Coords:   &types.Coordinates{},
	}
	many := types.EnrichedPoi{
		Distance: 100,
		Sources: []types.Provider{
			types.ProviderOverpass,
			types.ProviderWikivoyage,
			types.ProviderWikipedia,
		},
		Coords: &types.Coordinates{},
	}

	if scoring.Score(many, q) <= scoring.Score(one, q) {
		t.Error("more sources should yield a higher score")
	}
}

func TestScoreCloserIsBetter(t *testing.T) {
	q := types.SearchQuery{Radius: 5000}
	near := types.EnrichedPoi{
		Distance: 100,
		Sources:  []types.Provider{types.ProviderOverpass},
		Coords:   &types.Coordinates{},
	}
	far := types.EnrichedPoi{
		Distance: 4900,
		Sources:  []types.Provider{types.ProviderOverpass},
		Coords:   &types.Coordinates{},
	}

	if scoring.Score(near, q) <= scoring.Score(far, q) {
		t.Error("closer POI should score higher")
	}
}

func TestScoreApproximateCoordsLower(t *testing.T) {
	q := types.SearchQuery{Radius: 5000}
	exact := types.EnrichedPoi{
		Distance: 100,
		Sources:  []types.Provider{types.ProviderOverpass},
		Coords:   &types.Coordinates{Approximate: false},
	}
	approx := types.EnrichedPoi{
		Distance: 100,
		Sources:  []types.Provider{types.ProviderOverpass},
		Coords:   &types.Coordinates{Approximate: true},
	}

	if scoring.Score(exact, q) <= scoring.Score(approx, q) {
		t.Error("exact coords should score higher than approximate")
	}
}
