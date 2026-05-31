package search

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/trippier/poi-api/internal/providers"
	"github.com/trippier/poi-api/pkg/types"
)

// stubEnricher implements providers.Provider + providers.Enricher for the
// generic enrichment tests. The borrowing rule is intentionally simple: fill
// WikidataID on every target that lacks one.
type stubEnricher struct {
	name   types.Provider
	pois   []types.RawPoi
	radius float64
}

func (s *stubEnricher) Name() types.Provider                 { return s.name }
func (s *stubEnricher) SupportsMode(_ types.SearchMode) bool { return true }
func (s *stubEnricher) Search(_ context.Context, _ types.SearchQuery) ([]types.RawPoi, error) {
	return s.pois, nil
}
func (s *stubEnricher) EnrichmentRadius() float64 { return s.radius }
func (s *stubEnricher) EnrichTarget(target *types.RawPoi, source types.RawPoi) {
	if target.WikidataID == "" && source.WikidataID != "" {
		target.WikidataID = source.WikidataID
	}
}

var (
	_ providers.Provider = (*stubEnricher)(nil)
	_ providers.Enricher = (*stubEnricher)(nil)
)

func ec(lat, lng float64) *types.Coordinates {
	return &types.Coordinates{Lat: lat, Lng: lng}
}

func newServiceWithEnricher(e *stubEnricher) *Service {
	return NewService([]providers.Provider{e}, 1*time.Second, zap.NewNop())
}

func TestEnrichRaw_AppliesEnricherWithinRadius(t *testing.T) {
	enricher := &stubEnricher{
		name:   types.ProviderWikipedia,
		radius: 50,
		pois: []types.RawPoi{
			{ID: "wikipedia:1", Provider: types.ProviderWikipedia,
				Coords: ec(48.85837, 2.29450), WikidataID: "Q243"},
		},
	}
	svc := newServiceWithEnricher(enricher)
	raw := []types.RawPoi{
		// 2m from enricher's POI — should be enriched.
		{ID: "overpass:1", Provider: types.ProviderOverpass, Coords: ec(48.85838, 2.29451)},
		// >100m away — out of radius, not enriched.
		{ID: "overpass:2", Provider: types.ProviderOverpass, Coords: ec(48.86200, 2.30000)},
	}
	out := svc.enrichRaw(context.Background(), raw, types.SearchQuery{Mode: types.ModeRadius})
	if out[0].WikidataID != "Q243" {
		t.Errorf("nearby POI WikidataID = %q, want Q243", out[0].WikidataID)
	}
	if out[1].WikidataID != "" {
		t.Errorf("far POI WikidataID = %q, want empty", out[1].WikidataID)
	}
}

func TestEnrichRaw_SkipsSelfProvider(t *testing.T) {
	enricher := &stubEnricher{
		name:   types.ProviderWikipedia,
		radius: 50,
		pois: []types.RawPoi{
			{ID: "wikipedia:1", Provider: types.ProviderWikipedia,
				Coords: ec(48.85837, 2.29450), WikidataID: "Q243"},
		},
	}
	svc := newServiceWithEnricher(enricher)
	raw := []types.RawPoi{
		// Wikipedia target — the enricher must NOT enrich its own provider.
		{ID: "wikipedia:other", Provider: types.ProviderWikipedia, Coords: ec(48.85838, 2.29451)},
	}
	out := svc.enrichRaw(context.Background(), raw, types.SearchQuery{Mode: types.ModeRadius})
	if out[0].WikidataID != "" {
		t.Errorf("self-provider target should not be enriched, got %q", out[0].WikidataID)
	}
}

func TestEnrichRaw_SkipsApproximateCoords(t *testing.T) {
	enricher := &stubEnricher{
		name:   types.ProviderWikipedia,
		radius: 50,
		pois: []types.RawPoi{
			{Provider: types.ProviderWikipedia, Coords: ec(48.85837, 2.29450), WikidataID: "Q243"},
		},
	}
	svc := newServiceWithEnricher(enricher)
	raw := []types.RawPoi{
		{Provider: types.ProviderOverpass, Coords: &types.Coordinates{Lat: 48.85837, Lng: 2.29450, Approximate: true}},
		{Provider: types.ProviderOverpass},
	}
	out := svc.enrichRaw(context.Background(), raw, types.SearchQuery{Mode: types.ModeRadius})
	if out[0].WikidataID != "" || out[1].WikidataID != "" {
		t.Errorf("targets with approximate/nil coords must be skipped, got %+v", out)
	}
}

func TestEnrichRaw_NoEnrichers_LeavesRawUnchanged(t *testing.T) {
	plain := &stubProvider{name: types.ProviderOverpass}
	svc := NewService([]providers.Provider{plain}, 1*time.Second, zap.NewNop())
	raw := []types.RawPoi{
		{Provider: types.ProviderGeoNames, Coords: ec(48.85837, 2.29450)},
	}
	out := svc.enrichRaw(context.Background(), raw, types.SearchQuery{Mode: types.ModeRadius})
	if out[0].WikidataID != "" {
		t.Errorf("no enrichers registered, target must stay untouched")
	}
}

// stubProvider implements only providers.Provider (no Enricher) so we can
// verify the generic loop ignores plain providers.
type stubProvider struct{ name types.Provider }

func (s *stubProvider) Name() types.Provider                 { return s.name }
func (s *stubProvider) SupportsMode(_ types.SearchMode) bool { return true }
func (s *stubProvider) Search(_ context.Context, _ types.SearchQuery) ([]types.RawPoi, error) {
	return nil, nil
}

func TestClosestNeighbour(t *testing.T) {
	target := types.RawPoi{Coords: ec(48.8584, 2.2945)}
	sources := []types.RawPoi{
		{ID: "far", Coords: ec(48.8584, 2.2960)},
		{ID: "close", Coords: ec(48.8585, 2.2946)},
		{ID: "mid", Coords: ec(48.8586, 2.2945)},
	}
	if got := closestNeighbour(target, sources, 50); got == nil || got.ID != "close" {
		t.Errorf("closestNeighbour = %+v, want close", got)
	}
}

func TestClosestNeighbour_OutsideRadius(t *testing.T) {
	target := types.RawPoi{Coords: ec(48.8584, 2.2945)}
	sources := []types.RawPoi{
		{Coords: ec(48.8600, 2.2960)}, // > 50m
	}
	if got := closestNeighbour(target, sources, 50); got != nil {
		t.Errorf("closestNeighbour should be nil for out-of-radius input, got %+v", got)
	}
}

func TestClosestNeighbour_SkipsMalformed(t *testing.T) {
	target := types.RawPoi{Coords: ec(48.8584, 2.2945)}
	sources := []types.RawPoi{
		{}, // nil coords
		{Coords: &types.Coordinates{Lat: 48.8585, Lng: 2.2946, Approximate: true}},
		{ID: "ok", Coords: ec(48.8584, 2.2946)},
	}
	if got := closestNeighbour(target, sources, 50); got == nil || got.ID != "ok" {
		t.Errorf("closestNeighbour = %+v, want ok", got)
	}
}
