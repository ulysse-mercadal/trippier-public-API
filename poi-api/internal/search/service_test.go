package search_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/trippier/poi-api/internal/providers"
	"github.com/trippier/poi-api/internal/search"
	"github.com/trippier/poi-api/pkg/types"
)

// mockProvider is a test double for providers.Provider.
type mockProvider struct {
	name      types.Provider
	modes     []types.SearchMode
	returnErr error
	pois      []types.RawPoi
}

func (m *mockProvider) Name() types.Provider { return m.name }

func (m *mockProvider) SupportsMode(mode types.SearchMode) bool {
	for _, mo := range m.modes {
		if mo == mode {
			return true
		}
	}
	return false
}

func (m *mockProvider) Search(_ context.Context, _ types.SearchQuery) ([]types.RawPoi, error) {
	return m.pois, m.returnErr
}

// Ensure mockProvider satisfies the interface at compile time.
var _ providers.Provider = (*mockProvider)(nil)

func newCoords(lat, lng float64) *types.Coordinates {
	return &types.Coordinates{Lat: lat, Lng: lng}
}

func TestServiceSearch_BasicRadius(t *testing.T) {
	pois := []types.RawPoi{
		{ID: "overpass:1", Name: "Louvre", Type: types.TypeSee,
			Provider: types.ProviderOverpass, Coords: newCoords(48.8606, 2.3376)},
		{ID: "overpass:2", Name: "Notre-Dame", Type: types.TypeSee,
			Provider: types.ProviderOverpass, Coords: newCoords(48.8530, 2.3499)},
	}

	p := &mockProvider{
		name:  types.ProviderOverpass,
		modes: []types.SearchMode{types.ModeRadius},
		pois:  pois,
	}

	svc := search.NewService([]providers.Provider{p}, 5*time.Second)

	q := types.SearchQuery{
		Mode:      types.ModeRadius,
		Lat:       48.8566,
		Lng:       2.3522,
		Radius:    5000,
		Providers: []types.Provider{types.ProviderOverpass},
		Limit:     20,
		Lang:      "en",
	}

	result, err := svc.Search(context.Background(), q)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("expected 2 results, got %d", result.Total)
	}
}

func TestServiceSearch_Pagination(t *testing.T) {
	var pois []types.RawPoi
	for i := 0; i < 10; i++ {
		// Spread POIs across Paris so the deduplication does not collapse them.
		// Each point is ~1.1 km apart in latitude.
		pois = append(pois, types.RawPoi{
			ID:       fmt.Sprintf("overpass:%d", i),
			Name:     fmt.Sprintf("Place %d", i),
			Provider: types.ProviderOverpass,
			Coords:   newCoords(48.80+float64(i)*0.01, 2.35),
		})
	}

	p := &mockProvider{
		name:  types.ProviderOverpass,
		modes: []types.SearchMode{types.ModeRadius},
		pois:  pois,
	}

	svc := search.NewService([]providers.Provider{p}, 5*time.Second)

	q := types.SearchQuery{
		Mode:      types.ModeRadius,
		Lat:       48.85,
		Lng:       2.35,
		Radius:    15_000,
		Providers: []types.Provider{types.ProviderOverpass},
		Limit:     3,
		Offset:    0,
		Lang:      "en",
	}

	result, err := svc.Search(context.Background(), q)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Results) != 3 {
		t.Errorf("expected 3 paginated results, got %d", len(result.Results))
	}
	if result.Total != 10 {
		t.Errorf("total should be 10, got %d", result.Total)
	}
}

func TestServiceSearch_UnsupportedModeSkipped(t *testing.T) {
	p := &mockProvider{
		name:  types.ProviderOverpass,
		modes: []types.SearchMode{types.ModeRadius}, // polygon not supported
		pois: []types.RawPoi{
			{ID: "overpass:1", Name: "Musée", Provider: types.ProviderOverpass,
				Coords: newCoords(48.85, 2.35)},
		},
	}

	svc := search.NewService([]providers.Provider{p}, 5*time.Second)

	q := types.SearchQuery{
		Mode:      types.ModePolygon,
		Polygon:   "48.84 2.34 48.86 2.34 48.86 2.36 48.84 2.36",
		Providers: []types.Provider{types.ProviderOverpass},
		Limit:     20,
		Lang:      "en",
	}

	result, err := svc.Search(context.Background(), q)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("expected 0 results (provider skipped), got %d", result.Total)
	}
}

func TestServiceSearch_MinScore(t *testing.T) {
	p := &mockProvider{
		name:  types.ProviderOverpass,
		modes: []types.SearchMode{types.ModeRadius},
		pois: []types.RawPoi{
			{ID: "overpass:1", Name: "POI", Provider: types.ProviderOverpass,
				Coords: newCoords(48.85, 2.35)},
		},
	}

	svc := search.NewService([]providers.Provider{p}, 5*time.Second)

	q := types.SearchQuery{
		Mode:      types.ModeRadius,
		Lat:       48.85,
		Lng:       2.35,
		Radius:    5000,
		Providers: []types.Provider{types.ProviderOverpass},
		Limit:     20,
		MinScore:  99, // impossibly high
		Lang:      "en",
	}

	result, err := svc.Search(context.Background(), q)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("min_score=99 should filter everything out, got %d", result.Total)
	}
}

func TestParseWeights(t *testing.T) {
	tests := []struct {
		raw     string
		wantLen int
		wantErr bool
	}{
		{"", 0, false},
		{`{"see":2,"eat":1}`, 2, false},
		{`not-json`, 0, true},
	}

	for _, tc := range tests {
		weights, err := search.ParseWeights(tc.raw)
		if tc.wantErr && err == nil {
			t.Errorf("ParseWeights(%q) expected error", tc.raw)
		}
		if !tc.wantErr && err != nil {
			t.Errorf("ParseWeights(%q) unexpected error: %v", tc.raw, err)
		}
		if !tc.wantErr && len(weights) != tc.wantLen {
			t.Errorf("ParseWeights(%q) len = %d, want %d", tc.raw, len(weights), tc.wantLen)
		}
	}
}
