package geo_test

import (
	"testing"

	"github.com/trippier/poi-api/internal/geo"
	"github.com/trippier/poi-api/pkg/types"
)

func coords(lat, lng float64, approx bool) *types.Coordinates {
	return &types.Coordinates{Lat: lat, Lng: lng, Approximate: approx}
}

func TestSetDistances(t *testing.T) {
	pois := []types.RawPoi{
		{ID: "a", Coords: coords(48.8566, 2.3522, false)}, // exact
		{ID: "b", Coords: coords(0, 0, true)},             // approximate — skipped
		{ID: "c"},                                         // nil coords — skipped
	}

	result := geo.SetDistances(pois, 48.8566, 2.3522)

	if result[0].Distance != 0 {
		t.Errorf("same-point distance = %f, want 0", result[0].Distance)
	}
	if result[1].Distance != 0 {
		t.Errorf("approximate coords should not set distance, got %f", result[1].Distance)
	}
	if result[2].Distance != 0 {
		t.Errorf("nil coords should not set distance, got %f", result[2].Distance)
	}
}

func TestFilterByRadius(t *testing.T) {
	// poi at 0 m, poi at ~10 km, zone-only poi (approximate)
	pois := []types.RawPoi{
		{ID: "near", Coords: coords(48.8566, 2.3522, false), Distance: 100},
		{ID: "far", Coords: coords(48.9566, 2.3522, false), Distance: 15_000},
		{ID: "zone", Coords: coords(0, 0, true)},
	}

	got := geo.FilterByRadius(pois, 48.8566, 2.3522, 5_000)

	if len(got) != 2 {
		t.Fatalf("expected 2 results (near + zone), got %d", len(got))
	}
	for _, p := range got {
		if p.ID == "far" {
			t.Errorf("far POI should have been filtered out")
		}
	}
}

func TestFilterByPolygon(t *testing.T) {
	square := [][2]float64{
		{0, 0}, {0, 1}, {1, 1}, {1, 0},
	}

	pois := []types.RawPoi{
		{ID: "inside", Coords: coords(0.5, 0.5, false)},
		{ID: "outside", Coords: coords(2.0, 2.0, false)},
		{ID: "zone", Coords: coords(0, 0, true)}, // kept regardless
	}

	got := geo.FilterByPolygon(pois, square)

	if len(got) != 2 {
		t.Fatalf("expected 2 (inside + zone), got %d", len(got))
	}
	for _, p := range got {
		if p.ID == "outside" {
			t.Errorf("outside POI should have been filtered out")
		}
	}
}
