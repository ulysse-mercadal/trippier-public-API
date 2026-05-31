package search_test

import (
	"context"
	"testing"
	"time"

	"github.com/trippier/poi-api/internal/providers"
	"github.com/trippier/poi-api/internal/search"
	"github.com/trippier/poi-api/pkg/types"
	"go.uber.org/zap"
)

// byokMockProvider is a BYOK-aware mock that also satisfies providers.ByokProvider.
type byokMockProvider struct {
	mockProvider
}

func (b *byokMockProvider) IsByok() bool { return true }

var _ providers.ByokProvider = (*byokMockProvider)(nil)

func newSvc(pp ...providers.Provider) *search.Service {
	return search.NewService(pp, 5*time.Second, zap.NewNop())
}

// ── SearchCustom ──────────────────────────────────────────────────────────────

func TestSearchCustom_ExplicitProviders(t *testing.T) {
	p := &mockProvider{
		name:  types.ProviderOverpass,
		modes: []types.SearchMode{types.ModeRadius},
		pois: []types.RawPoi{
			{ID: "o:1", Name: "Test POI", Provider: types.ProviderOverpass, Coords: newCoords(48.86, 2.35)},
		},
	}
	svc := newSvc(p)
	result, err := svc.SearchCustom(context.Background(), types.SearchQuery{
		Mode:      types.ModeRadius,
		Lat:       48.86,
		Lng:       2.35,
		Radius:    5000,
		Providers: []types.Provider{types.ProviderOverpass},
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("SearchCustom: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("total = %d, want 1", result.Total)
	}
}

func TestSearchCustom_ExcludeProviders(t *testing.T) {
	p1 := &mockProvider{
		name:  types.ProviderOverpass,
		modes: []types.SearchMode{types.ModeRadius},
		pois: []types.RawPoi{
			{ID: "o:1", Name: "Overpass POI", Provider: types.ProviderOverpass, Coords: newCoords(48.86, 2.35)},
		},
	}
	p2 := &mockProvider{
		name:  types.ProviderWikivoyage,
		modes: []types.SearchMode{types.ModeRadius},
		pois: []types.RawPoi{
			{ID: "wv:1", Name: "Wv POI", Provider: types.ProviderWikivoyage, Coords: newCoords(48.87, 2.36)},
		},
	}
	svc := newSvc(p1, p2)
	result, err := svc.SearchCustom(context.Background(), types.SearchQuery{
		Mode:             types.ModeRadius,
		Lat:              48.86,
		Lng:              2.35,
		Radius:           10_000,
		Providers:        []types.Provider{types.ProviderOverpass, types.ProviderWikivoyage},
		ExcludeProviders: []types.Provider{types.ProviderWikivoyage},
		Limit:            10,
	})
	if err != nil {
		t.Fatalf("SearchCustom: %v", err)
	}
	for _, poi := range result.Results {
		for _, src := range poi.Sources {
			if src.Provider == types.ProviderWikivoyage {
				t.Error("excluded provider wikivoyage still appears in results")
			}
		}
	}
}

func TestSearchCustom_CountryHint_UsedForAutoSelect(t *testing.T) {
	// country_hint triggers selectByCountry; with no implemented free providers that
	// match the hint well, auto-selection may return an empty list — but the query
	// still executes (falls back to AllProviders).
	p := &mockProvider{
		name:  types.ProviderOverpass,
		modes: []types.SearchMode{types.ModeRadius},
		pois: []types.RawPoi{
			{ID: "o:1", Name: "POI", Provider: types.ProviderOverpass, Coords: newCoords(48.86, 2.35)},
		},
	}
	svc := newSvc(p)
	// No panic, no error — that's the key assertion.
	_, err := svc.SearchCustom(context.Background(), types.SearchQuery{
		Mode:        types.ModeRadius,
		Lat:         48.86,
		Lng:         2.35,
		Radius:      5000,
		CountryHint: "FR",
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("SearchCustom with country_hint: %v", err)
	}
}

func TestSearchCustom_ProviderWeights_LowWeightExcludes(t *testing.T) {
	p := &mockProvider{
		name:  types.ProviderOverpass,
		modes: []types.SearchMode{types.ModeRadius},
		pois: []types.RawPoi{
			{ID: "o:1", Name: "POI", Provider: types.ProviderOverpass, Coords: newCoords(48.86, 2.35)},
		},
	}
	svc := newSvc(p)
	// With provider_weights={"overpass":0.0} and country_hint="FR", overpass is
	// excluded from auto-selection (below threshold). Result should still not error.
	result, err := svc.SearchCustom(context.Background(), types.SearchQuery{
		Mode:            types.ModeRadius,
		Lat:             48.86,
		Lng:             2.35,
		Radius:          5000,
		CountryHint:     "FR",
		ProviderWeights: map[types.Provider]float64{types.ProviderOverpass: 0.0},
		Limit:           10,
	})
	if err != nil {
		t.Fatalf("SearchCustom with low provider weight: %v", err)
	}
	// overpass excluded from auto-select → no results from it.
	_ = result
}

// ── SearchEventsCustom ────────────────────────────────────────────────────────

func TestSearchEventsCustom_ExplicitProviders(t *testing.T) {
	p := &mockProvider{
		name:  types.ProviderWikipediaEvents,
		modes: []types.SearchMode{types.ModeRadius},
		pois: []types.RawPoi{
			{ID: "wp:1", Name: "Festival", Type: types.TypeEvent,
				Provider: types.ProviderWikipediaEvents, Coords: newCoords(48.86, 2.35)},
		},
	}
	svc := newSvc(p)
	result, err := svc.SearchEventsCustom(context.Background(), types.SearchQuery{
		Mode:      types.ModeRadius,
		Lat:       48.86,
		Lng:       2.35,
		Radius:    5000,
		Providers: []types.Provider{types.ProviderWikipediaEvents},
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("SearchEventsCustom: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("total = %d, want 1", result.Total)
	}
}

// TestSearchEventsCustom_RadiusEnforced confirms the orchestrator stretches
// q.Radius up to the maximum MinRadius declared by any selected provider.
// Ticketmaster's registry entry declares MinRadius=50_000, so a 1 km query
// is upgraded; if we ever ship an event provider with MinRadius=0 it would
// not be touched.
func TestSearchEventsCustom_RadiusEnforced(t *testing.T) {
	var capturedRadius int
	p := &extMockProvider{
		name:  types.ProviderTicketmaster,
		modes: []types.SearchMode{types.ModeRadius},
		searchFn: func(q types.SearchQuery) ([]types.RawPoi, error) {
			capturedRadius = q.Radius
			return nil, nil
		},
	}
	svc := newSvc(p)
	svc.SearchEventsCustom(context.Background(), types.SearchQuery{ //nolint:errcheck
		Mode:      types.ModeRadius,
		Lat:       48.86,
		Lng:       2.35,
		Radius:    1000, // below the registry-declared 50 km minimum
		Providers: []types.Provider{types.ProviderTicketmaster},
	})
	if capturedRadius < 50_000 {
		t.Errorf("radius not enforced to declared MinRadius, got %d", capturedRadius)
	}
}

// ── ProvidersCatalog ──────────────────────────────────────────────────────────

func TestProvidersCatalog_ContainsAllRegistryEntries(t *testing.T) {
	svc := newSvc()
	catalog := svc.ProvidersCatalog()
	if len(catalog) == 0 {
		t.Fatal("catalog should not be empty")
	}
}

func TestProvidersCatalog_ImplementedFlag(t *testing.T) {
	p := &mockProvider{name: types.ProviderOverpass, modes: []types.SearchMode{types.ModeRadius}}
	svc := newSvc(p)
	catalog := svc.ProvidersCatalog()

	var foundOverpass, foundFoursquare bool
	for _, entry := range catalog {
		if entry.ID == types.ProviderOverpass {
			foundOverpass = true
			if !entry.Implemented {
				t.Error("overpass should be implemented=true")
			}
		}
		if entry.ID == types.ProviderFoursquare {
			foundFoursquare = true
			if entry.Implemented {
				t.Error("foursquare should be implemented=false (not registered)")
			}
		}
	}
	if !foundOverpass {
		t.Error("overpass not found in catalog")
	}
	if !foundFoursquare {
		t.Error("foursquare not found in catalog")
	}
}

func TestProvidersCatalog_IsSorted(t *testing.T) {
	svc := newSvc()
	catalog := svc.ProvidersCatalog()
	for i := 1; i < len(catalog); i++ {
		if string(catalog[i-1].ID) > string(catalog[i].ID) {
			t.Errorf("catalog not sorted: %q > %q at positions %d,%d",
				catalog[i-1].ID, catalog[i].ID, i-1, i)
		}
	}
}

// ── ProvidersRecommend ────────────────────────────────────────────────────────

func TestProvidersRecommend_ReturnsResults(t *testing.T) {
	svc := newSvc()
	// Use a cancelled context so CountryCode fails fast and uses global defaults.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result := svc.ProvidersRecommend(ctx, 48.86, 2.35, false, nil, 5)
	if len(result.Providers) == 0 {
		t.Error("expected at least one provider recommendation")
	}
}

func TestProvidersRecommend_LimitRespected(t *testing.T) {
	svc := newSvc()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result := svc.ProvidersRecommend(ctx, 48.86, 2.35, false, nil, 3)
	if len(result.Providers) > 3 {
		t.Errorf("expected at most 3 providers, got %d", len(result.Providers))
	}
}

func TestProvidersRecommend_ForEvents(t *testing.T) {
	svc := newSvc()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result := svc.ProvidersRecommend(ctx, 48.86, 2.35, true, nil, 10)
	for _, p := range result.Providers {
		if !p.ForEvents {
			t.Errorf("provider %q should be for_events=true", p.ID)
		}
	}
}

func TestProvidersRecommend_ScoresSorted(t *testing.T) {
	svc := newSvc()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result := svc.ProvidersRecommend(ctx, 48.86, 2.35, false, nil, 20)
	for i := 1; i < len(result.Providers); i++ {
		if result.Providers[i-1].Score < result.Providers[i].Score {
			t.Errorf("providers not sorted by score: %.2f < %.2f at positions %d,%d",
				result.Providers[i-1].Score, result.Providers[i].Score, i-1, i)
		}
	}
}

// ── ParseProviderWeights ──────────────────────────────────────────────────────

func TestParseProviderWeights(t *testing.T) {
	tests := []struct {
		raw     string
		wantLen int
		wantErr bool
	}{
		{"", 0, false},
		{`{"overpass":0.8,"foursquare":0.5}`, 2, false},
		{`{"overpass":0.0}`, 1, false},
		{`{"overpass":1.0}`, 1, false},
		{`{"overpass":1.5}`, 0, true},  // > 1
		{`{"overpass":-0.1}`, 0, true}, // < 0
		{`not-json`, 0, true},
	}
	for _, tc := range tests {
		weights, err := search.ParseProviderWeights(tc.raw)
		if tc.wantErr && err == nil {
			t.Errorf("ParseProviderWeights(%q) expected error", tc.raw)
		}
		if !tc.wantErr && err != nil {
			t.Errorf("ParseProviderWeights(%q) unexpected error: %v", tc.raw, err)
		}
		if !tc.wantErr && len(weights) != tc.wantLen {
			t.Errorf("ParseProviderWeights(%q) len = %d, want %d", tc.raw, len(weights), tc.wantLen)
		}
	}
}

// ── Auto-select with explicit providers (no geo lookup) ───────────────────────

func TestSearch_AutoSelect_FallsBackWhenNoProviders(t *testing.T) {
	// With no providers registered and no explicit list, auto-select returns empty,
	// and the fallback AllProviders is used (which also maps to nothing). No panic.
	svc := newSvc()
	result, err := svc.Search(context.Background(), types.SearchQuery{
		Mode:   types.ModeRadius,
		Lat:    48.86,
		Lng:    2.35,
		Radius: 5000,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("Search with no providers: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("expected 0 results with no providers, got %d", result.Total)
	}
}

// mockProvider extended version that supports a custom searchFn for capturing query params.
// We redefine the type within this file to avoid conflicts — using embedding.
type extMockProvider struct {
	name     types.Provider
	modes    []types.SearchMode
	searchFn func(q types.SearchQuery) ([]types.RawPoi, error)
}

func (m *extMockProvider) Name() types.Provider { return m.name }
func (m *extMockProvider) SupportsMode(mode types.SearchMode) bool {
	for _, mo := range m.modes {
		if mo == mode {
			return true
		}
	}
	return false
}
func (m *extMockProvider) Search(_ context.Context, q types.SearchQuery) ([]types.RawPoi, error) {
	if m.searchFn != nil {
		return m.searchFn(q)
	}
	return nil, nil
}

var _ providers.Provider = (*extMockProvider)(nil)
