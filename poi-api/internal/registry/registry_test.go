package registry_test

import (
	"testing"

	"github.com/trippier/poi-api/internal/registry"
	"github.com/trippier/poi-api/pkg/types"
)

func TestCountryScore_ExactMatch(t *testing.T) {
	m := registry.All[types.ProviderOverpass]
	score := m.CountryScore("DE")
	if score < 0.9 {
		t.Errorf("Overpass DE score = %.2f, want >= 0.9 (OSM is exceptional in Germany)", score)
	}
}

func TestCountryScore_GlobalFallback(t *testing.T) {
	m := registry.All[types.ProviderOverpass]
	// XX is not in the map, should fall back to "*"
	score := m.CountryScore("XX")
	global := m.CountryScores["*"]
	if score != global {
		t.Errorf("CountryScore(XX) = %.2f, want global default %.2f", score, global)
	}
}

func TestCountryScore_NoGlobalFallback(t *testing.T) {
	// Build a meta with no "*" key and no matching country.
	m := registry.Meta{
		CountryScores: map[string]float64{"FR": 0.9},
	}
	score := m.CountryScore("JP")
	if score != 0.5 {
		t.Errorf("CountryScore fallback = %.2f, want 0.5", score)
	}
}

func TestCategoryScore_NoTypes(t *testing.T) {
	m := registry.All[types.ProviderFoursquare]
	score := m.CategoryScore(nil)
	if score != 1.0 {
		t.Errorf("CategoryScore(nil) = %.2f, want 1.0 (no filter)", score)
	}
}

func TestCategoryScore_NoScoresDefined(t *testing.T) {
	m := registry.Meta{} // no CategoryScores
	score := m.CategoryScore([]types.PoiType{types.TypeSee})
	if score != 1.0 {
		t.Errorf("CategoryScore with no scores defined = %.2f, want 1.0", score)
	}
}

func TestCategoryScore_PartialMatch(t *testing.T) {
	m := registry.Meta{
		CategoryScores: map[types.PoiType]float64{
			types.TypeEat: 0.8,
			// TypeSee not present
		},
	}
	// Only TypeEat matches; TypeSee gets no contribution.
	score := m.CategoryScore([]types.PoiType{types.TypeEat, types.TypeSee})
	if score != 0.8 {
		t.Errorf("CategoryScore partial match = %.2f, want 0.8", score)
	}
}

func TestCategoryScore_NoMatch(t *testing.T) {
	m := registry.Meta{
		CategoryScores: map[types.PoiType]float64{types.TypeEat: 0.9},
	}
	// None of the requested types are in CategoryScores.
	score := m.CategoryScore([]types.PoiType{types.TypeSee})
	if score != 0.5 {
		t.Errorf("CategoryScore no match = %.2f, want 0.5", score)
	}
}

func TestScore_CompositeMultiplication(t *testing.T) {
	m := registry.Meta{
		CountryScores:  map[string]float64{"*": 0.8},
		CategoryScores: map[types.PoiType]float64{types.TypeEat: 0.5},
	}
	score := m.Score("US", []types.PoiType{types.TypeEat})
	want := 0.8 * 0.5
	if abs(score-want) > 1e-9 {
		t.Errorf("Score = %.4f, want %.4f", score, want)
	}
}

func TestScore_China_Baidu_High(t *testing.T) {
	m := registry.All[types.ProviderBaidu]
	score := m.Score("CN", nil)
	if score < 0.9 {
		t.Errorf("Baidu CN score = %.2f, want >= 0.9", score)
	}
}

func TestScore_China_Baidu_LowGlobal(t *testing.T) {
	m := registry.All[types.ProviderBaidu]
	score := m.Score("FR", nil)
	if score >= 0.2 {
		t.Errorf("Baidu FR score = %.2f, want < 0.2 (not relevant outside China)", score)
	}
}

func TestScore_GrabMaps_SEA(t *testing.T) {
	m := registry.All[types.ProviderGrabMaps]
	if m.Score("SG", nil) < 0.9 {
		t.Errorf("GrabMaps SG score < 0.9")
	}
	if m.Score("FR", nil) >= 0.2 {
		t.Errorf("GrabMaps FR score should be low")
	}
}

func TestAllProvidersHaveLabel(t *testing.T) {
	for id, m := range registry.All {
		if m.Label == "" {
			t.Errorf("provider %q has empty Label", id)
		}
		if m.ID != id {
			t.Errorf("provider %q: Meta.ID = %q, want %q", id, m.ID, id)
		}
	}
}

func TestByokProvidersHaveHeader(t *testing.T) {
	for id, m := range registry.All {
		if m.Byok && m.ByokHeader == "" {
			t.Errorf("BYOK provider %q has empty ByokHeader", id)
		}
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
