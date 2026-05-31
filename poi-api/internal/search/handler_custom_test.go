package search_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/trippier/poi-api/pkg/types"
)

// ── /pois/search/custom ───────────────────────────────────────────────────────

func TestHandlerSearchCustom_OK(t *testing.T) {
	p := &mockProvider{
		name:  types.ProviderOverpass,
		modes: []types.SearchMode{types.ModeRadius},
		pois: []types.RawPoi{
			{ID: "o:1", Name: "Custom POI", Provider: types.ProviderOverpass, Coords: newCoords(48.86, 2.35)},
		},
	}
	w := get(newRouter(p), "/pois/search/custom?mode=radius&lat=48.8566&lng=2.3522&radius=5000&providers=overpass")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body.String())
	}
	var body struct {
		Total int `json:"total"`
	}
	json.NewDecoder(w.Body).Decode(&body)
	if body.Total != 1 {
		t.Errorf("total = %d, want 1", body.Total)
	}
}

func TestHandlerSearchCustom_InvalidParams(t *testing.T) {
	w := get(newRouter(), "/pois/search/custom?mode=radius")
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestHandlerSearchCustom_InvalidProviderWeights(t *testing.T) {
	w := get(newRouter(), `/pois/search/custom?mode=radius&lat=48.8566&lng=2.3522&provider_weights={"overpass":2.0}`)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 for out-of-range provider_weight", w.Code)
	}
}

func TestHandlerSearchCustom_ExcludeProviders(t *testing.T) {
	p1 := &mockProvider{
		name:  types.ProviderOverpass,
		modes: []types.SearchMode{types.ModeRadius},
		pois: []types.RawPoi{
			{ID: "o:1", Name: "Overpass", Provider: types.ProviderOverpass, Coords: newCoords(48.86, 2.35)},
		},
	}
	p2 := &mockProvider{
		name:  types.ProviderWikivoyage,
		modes: []types.SearchMode{types.ModeRadius},
		pois: []types.RawPoi{
			{ID: "wv:1", Name: "Wikivoyage", Provider: types.ProviderWikivoyage, Coords: newCoords(48.87, 2.36)},
		},
	}
	w := get(newRouter(p1, p2),
		"/pois/search/custom?mode=radius&lat=48.8566&lng=2.3522&radius=10000"+
			"&providers=overpass&providers=wikivoyage&exclude_providers=wikivoyage")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var body struct {
		Total int `json:"total"`
	}
	json.NewDecoder(w.Body).Decode(&body)
	if body.Total != 1 {
		t.Errorf("total = %d, want 1 (wikivoyage excluded)", body.Total)
	}
}

func TestHandlerSearchCustom_WithCountryHint(t *testing.T) {
	w := get(newRouter(), "/pois/search/custom?mode=radius&lat=48.8566&lng=2.3522&radius=5000&country_hint=FR")
	// Should succeed even with no providers registered (returns empty results).
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

// ── /pois/search/custom/slim ──────────────────────────────────────────────────

func TestHandlerSearchCustomSlim_OK(t *testing.T) {
	p := &mockProvider{
		name:  types.ProviderOverpass,
		modes: []types.SearchMode{types.ModeRadius},
		pois: []types.RawPoi{
			{ID: "o:1", Name: "Slim POI", Type: types.TypeSee, Provider: types.ProviderOverpass, Coords: newCoords(48.86, 2.35)},
		},
	}
	w := get(newRouter(p), "/pois/search/custom/slim?mode=radius&lat=48.8566&lng=2.3522&radius=5000&providers=overpass")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var body struct {
		Results []struct {
			Name string `json:"name"`
		} `json:"results"`
	}
	json.NewDecoder(w.Body).Decode(&body)
	if len(body.Results) != 1 {
		t.Errorf("slim results count = %d, want 1", len(body.Results))
	}
}

func TestHandlerSearchCustomSlim_InvalidParams(t *testing.T) {
	w := get(newRouter(), "/pois/search/custom/slim?mode=radius")
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// ── /pois/events/slim ─────────────────────────────────────────────────────────

func TestHandlerEventsSlim_OK(t *testing.T) {
	p := &mockProvider{
		name:  types.ProviderWikipediaEvents,
		modes: []types.SearchMode{types.ModeRadius},
		pois: []types.RawPoi{
			{ID: "wp:1", Name: "Festival", Type: types.TypeEvent,
				Provider: types.ProviderWikipediaEvents, Coords: newCoords(48.86, 2.35)},
		},
	}
	w := get(newRouter(p), "/pois/events/slim?mode=radius&lat=48.8566&lng=2.3522&radius=5000")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var body struct {
		Total   int `json:"total"`
		Results []struct {
			Name string `json:"name"`
		} `json:"results"`
	}
	json.NewDecoder(w.Body).Decode(&body)
	if body.Total == 0 {
		t.Errorf("expected at least one event in slim response")
	}
}

func TestHandlerEventsSlim_InvalidParams(t *testing.T) {
	w := get(newRouter(), "/pois/events/slim?mode=radius")
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// ── /pois/events/custom ───────────────────────────────────────────────────────

func TestHandlerEventsCustom_OK(t *testing.T) {
	p := &mockProvider{
		name:  types.ProviderWikipediaEvents,
		modes: []types.SearchMode{types.ModeRadius},
		pois: []types.RawPoi{
			{ID: "wp:1", Name: "Event Custom", Type: types.TypeEvent,
				Provider: types.ProviderWikipediaEvents, Coords: newCoords(48.86, 2.35)},
		},
	}
	w := get(newRouter(p), "/pois/events/custom?mode=radius&lat=48.8566&lng=2.3522&radius=5000&providers=wikipedia_events")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body.String())
	}
}

func TestHandlerEventsCustom_InvalidParams(t *testing.T) {
	w := get(newRouter(), "/pois/events/custom?mode=radius")
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestHandlerEventsCustom_InvalidProviderWeights(t *testing.T) {
	w := get(newRouter(), `/pois/events/custom?mode=radius&lat=48.8566&lng=2.3522&provider_weights=bad-json`)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// ── /pois/events/custom/slim ──────────────────────────────────────────────────

func TestHandlerEventsCustomSlim_OK(t *testing.T) {
	p := &mockProvider{
		name:  types.ProviderWikipediaEvents,
		modes: []types.SearchMode{types.ModeRadius},
		pois: []types.RawPoi{
			{ID: "wp:1", Name: "Slim Event", Type: types.TypeEvent,
				Provider: types.ProviderWikipediaEvents, Coords: newCoords(48.86, 2.35)},
		},
	}
	w := get(newRouter(p), "/pois/events/custom/slim?mode=radius&lat=48.8566&lng=2.3522&radius=5000&providers=wikipedia_events")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var body struct {
		Results []struct {
			Name string `json:"name"`
		} `json:"results"`
	}
	json.NewDecoder(w.Body).Decode(&body)
	if len(body.Results) == 0 {
		t.Error("expected at least one slim event")
	}
}

func TestHandlerEventsCustomSlim_InvalidParams(t *testing.T) {
	w := get(newRouter(), "/pois/events/custom/slim?mode=radius")
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// ── /pois/providers/catalog ───────────────────────────────────────────────────

func TestHandlerProvidersCatalog_OK(t *testing.T) {
	p := &mockProvider{name: types.ProviderOverpass, modes: []types.SearchMode{types.ModeRadius}}
	w := get(newRouter(p), "/pois/providers/catalog")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var catalog []types.ProviderCatalogEntry
	if err := json.NewDecoder(w.Body).Decode(&catalog); err != nil {
		t.Fatalf("decode catalog: %v", err)
	}
	if len(catalog) == 0 {
		t.Error("catalog should not be empty")
	}
	// Find overpass and verify it's implemented.
	for _, entry := range catalog {
		if entry.ID == types.ProviderOverpass && !entry.Implemented {
			t.Error("overpass should be implemented=true")
		}
	}
}

// ── /pois/providers/recommend ─────────────────────────────────────────────────

func TestHandlerProvidersRecommend_OK(t *testing.T) {
	w := get(newRouter(), "/pois/providers/recommend?lat=48.8566&lng=2.3522")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body.String())
	}
	var result types.RecommendResult
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
}

func TestHandlerProvidersRecommend_MissingLatLng(t *testing.T) {
	w := get(newRouter(), "/pois/providers/recommend")
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestHandlerProvidersRecommend_ForEvents(t *testing.T) {
	w := get(newRouter(), "/pois/providers/recommend?lat=48.8566&lng=2.3522&for_events=true")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var result types.RecommendResult
	json.NewDecoder(w.Body).Decode(&result)
	for _, p := range result.Providers {
		if !p.ForEvents {
			t.Errorf("provider %q should be for_events", p.ID)
		}
	}
}

func TestHandlerProvidersRecommend_WithLimit(t *testing.T) {
	w := get(newRouter(), "/pois/providers/recommend?lat=48.8566&lng=2.3522&limit=2")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var result types.RecommendResult
	json.NewDecoder(w.Body).Decode(&result)
	if len(result.Providers) > 2 {
		t.Errorf("expected at most 2 providers with limit=2, got %d", len(result.Providers))
	}
}

func TestHandlerProvidersRecommend_WithTypes(t *testing.T) {
	w := get(newRouter(), "/pois/providers/recommend?lat=48.8566&lng=2.3522&types=eat,drink")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}
