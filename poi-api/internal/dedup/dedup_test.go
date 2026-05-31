package dedup_test

import (
	"testing"

	"github.com/trippier/poi-api/internal/dedup"
	"github.com/trippier/poi-api/pkg/types"
)

func coords(lat, lng float64) *types.Coordinates {
	return &types.Coordinates{Lat: lat, Lng: lng}
}

// TestMergeByWikidataID confirms that two POIs sharing a Wikidata ID are collapsed.
func TestMergeByWikidataID(t *testing.T) {
	pois := []types.RawPoi{
		{
			ID: "overpass:1", Name: "Louvre Museum",
			Provider:   types.ProviderOverpass,
			Coords:     coords(48.8606, 2.3376),
			WikidataID: "Q19675",
		},
		{
			ID: "wikipedia:123", Name: "Louvre",
			Provider:   types.ProviderWikipedia,
			Coords:     coords(48.8606, 2.3376),
			WikidataID: "Q19675",
		},
	}

	merged := dedup.Merge(pois)

	if len(merged) != 1 {
		t.Fatalf("expected 1 merged POI, got %d", len(merged))
	}
	if len(merged[0].Sources) != 2 {
		t.Errorf("expected 2 sources, got %d", len(merged[0].Sources))
	}
}

// TestMergeByProximityAndName confirms that near-identical POIs are merged via
// spatial proximity + Jaro-Winkler name similarity.
func TestMergeByProximityAndName(t *testing.T) {
	pois := []types.RawPoi{
		{
			ID: "overpass:2", Name: "Tour Eiffel",
			Provider: types.ProviderOverpass,
			Coords:   coords(48.85837, 2.29450),
		},
		{
			ID: "geonames:999", Name: "Tour Eiffel",
			Provider: types.ProviderGeoNames,
			Coords:   coords(48.85840, 2.29452), // ~3 m away
		},
	}

	merged := dedup.Merge(pois)

	if len(merged) != 1 {
		t.Fatalf("expected 1 merged POI, got %d", len(merged))
	}
}

// TestNoDuplicateDifferentPlaces ensures distant POIs are not merged.
func TestNoDuplicateDifferentPlaces(t *testing.T) {
	pois := []types.RawPoi{
		{
			ID: "overpass:3", Name: "Café de Flore",
			Provider: types.ProviderOverpass,
			Coords:   coords(48.8540, 2.3330),
		},
		{
			ID: "overpass:4", Name: "Café de Flore",
			Provider: types.ProviderOverpass,
			Coords:   coords(48.8800, 2.3500), // ~3 km away
		},
	}

	merged := dedup.Merge(pois)

	if len(merged) != 2 {
		t.Errorf("expected 2 distinct POIs, got %d", len(merged))
	}
}

// TestPrimaryProviderPriority confirms Overpass is chosen as primary over Wikipedia.
func TestPrimaryProviderPriority(t *testing.T) {
	pois := []types.RawPoi{
		{
			ID: "wikipedia:1", Name: "Panthéon",
			Provider:   types.ProviderWikipedia,
			Coords:     coords(48.8462, 2.3461),
			WikidataID: "Q188715",
		},
		{
			ID: "overpass:10", Name: "Panthéon",
			Provider:   types.ProviderOverpass,
			Coords:     coords(48.8462, 2.3461),
			WikidataID: "Q188715",
		},
	}

	merged := dedup.Merge(pois)

	if len(merged) != 1 {
		t.Fatalf("expected 1 merged POI, got %d", len(merged))
	}
	// Overpass has higher priority — its ID should be the canonical one.
	if merged[0].ID != "overpass:10" {
		t.Errorf("primary ID = %q, want overpass:10", merged[0].ID)
	}
}

func TestTourEiffelMergedViaWikidataBridge(t *testing.T) {
	pois := []types.RawPoi{
		{
			ID: "overpass:5013364", Name: "Tour Eiffel",
			Provider:   types.ProviderOverpass,
			Coords:     coords(48.85837, 2.29450),
			WikidataID: "Q243",
		},
		{
			ID: "wikipedia:Eiffel_Tower", Name: "Eiffel Tower",
			Provider:   types.ProviderWikipedia,
			Coords:     coords(48.85840, 2.29470),
			WikidataID: "Q243",
		},
		{
			ID: "wikivoyage:tour_eiffel", Name: "Tour Eiffel",
			Provider:   types.ProviderWikivoyage,
			Coords:     coords(48.85839, 2.29452),
			WikidataID: "Q243",
		},
		{
			ID: "geonames:6254976", Name: "Eiffel Tower",
			Provider:   types.ProviderGeoNames,
			Coords:     coords(48.85838, 2.29453),
			WikidataID: "Q243",
		},
	}

	merged := dedup.Merge(pois)

	if len(merged) != 1 {
		t.Fatalf("expected 1 merged Eiffel Tower POI, got %d", len(merged))
	}
	got := merged[0]
	if got.ID != "overpass:5013364" {
		t.Errorf("primary ID = %q, want overpass:5013364", got.ID)
	}
	if len(got.Sources) != 4 {
		t.Errorf("expected 4 sources, got %d: %v", len(got.Sources), got.Sources)
	}
}

// TestMergeImages_CapsAtThreeAndPrefersHigherPriority gathers images from
// every group member, keeps the highest-priority provider's URLs first,
// drops duplicates, and caps the result at maxImagesPerPoi (3).
func TestMergeImages_CapsAtThreeAndPrefersHigherPriority(t *testing.T) {
	pois := []types.RawPoi{
		// GeoNames (lowest priority) — listed first to verify ordering still
		// puts Overpass URLs ahead of it in the output.
		{
			ID: "geonames:1", Provider: types.ProviderGeoNames,
			Coords:     coords(48.86, 2.29),
			WikidataID: "Q1",
			Images:     []string{"https://img/geonames.jpg"},
		},
		// Overpass (highest priority) — its two URLs should occupy slots 0+1.
		{
			ID: "overpass:1", Provider: types.ProviderOverpass,
			Coords:     coords(48.86, 2.29),
			WikidataID: "Q1",
			Images:     []string{"https://img/overpass-a.jpg", "https://img/overpass-b.jpg"},
		},
		// Wikipedia — contributes one new URL and one duplicate (overpass-a).
		// The duplicate must be dropped; the new URL fills slot 2.
		{
			ID: "wikipedia:1", Provider: types.ProviderWikipedia,
			Coords:     coords(48.86, 2.29),
			WikidataID: "Q1",
			Images:     []string{"https://img/overpass-a.jpg", "https://img/wikipedia.jpg"},
		},
	}

	merged := dedup.Merge(pois)
	if len(merged) != 1 {
		t.Fatalf("expected 1 merged POI, got %d", len(merged))
	}
	want := []string{
		"https://img/overpass-a.jpg",
		"https://img/overpass-b.jpg",
		"https://img/wikipedia.jpg",
	}
	got := merged[0].Images
	if len(got) != len(want) {
		t.Fatalf("Images = %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// TestMergeImages_EmptyWhenNoSourceHasAny returns a nil/empty slice when no
// group member carries any image URL, so the JSON output omits the field.
func TestMergeImages_EmptyWhenNoSourceHasAny(t *testing.T) {
	pois := []types.RawPoi{
		{ID: "overpass:1", Provider: types.ProviderOverpass, Coords: coords(48.86, 2.29), WikidataID: "Q1"},
		{ID: "geonames:1", Provider: types.ProviderGeoNames, Coords: coords(48.86, 2.29), WikidataID: "Q1"},
	}
	merged := dedup.Merge(pois)
	if len(merged[0].Images) != 0 {
		t.Errorf("Images = %v, want empty", merged[0].Images)
	}
}

// TestMergeContactMerge confirms that contact fields are combined across providers.
func TestMergeContactMerge(t *testing.T) {
	pois := []types.RawPoi{
		{
			ID: "overpass:20", Name: "Brasserie Lipp",
			Provider:   types.ProviderOverpass,
			Coords:     coords(48.8540, 2.3330),
			WikidataID: "Q123",
			Contact:    types.Contact{Website: "https://brasserie-lipp.fr"},
		},
		{
			ID: "wikivoyage:brasserie_lipp", Name: "Brasserie Lipp",
			Provider:   types.ProviderWikivoyage,
			Coords:     coords(48.8540, 2.3330),
			WikidataID: "Q123",
			Contact:    types.Contact{Phone: "+33 1 45 48 53 91"},
		},
	}

	merged := dedup.Merge(pois)

	if len(merged) != 1 {
		t.Fatalf("expected 1 merged POI, got %d", len(merged))
	}
	c := merged[0].Contact
	if c.Website == "" {
		t.Error("website should be merged from overpass")
	}
	if c.Phone == "" {
		t.Error("phone should be merged from wikivoyage")
	}
}
