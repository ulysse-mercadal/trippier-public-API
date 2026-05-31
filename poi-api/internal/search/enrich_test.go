package search

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/trippier/poi-api/internal/providers"
	"github.com/trippier/poi-api/pkg/types"
)

type stubWikipediaProvider struct {
	pois []types.RawPoi
	err  error
}

func (s *stubWikipediaProvider) Name() types.Provider                 { return types.ProviderWikipedia }
func (s *stubWikipediaProvider) SupportsMode(_ types.SearchMode) bool { return true }
func (s *stubWikipediaProvider) Search(_ context.Context, _ types.SearchQuery) ([]types.RawPoi, error) {
	return s.pois, s.err
}

var _ providers.Provider = (*stubWikipediaProvider)(nil)

func newServiceWithStub(stub *stubWikipediaProvider) *Service {
	return NewService([]providers.Provider{stub}, 1*time.Second, zap.NewNop())
}

func ec(lat, lng float64) *types.Coordinates {
	return &types.Coordinates{Lat: lat, Lng: lng}
}

func TestNeedsWikipediaEnrichment(t *testing.T) {
	cases := []struct {
		name string
		p    types.RawPoi
		want bool
	}{
		{
			name: "overpass without wikidata id",
			p: types.RawPoi{
				Provider: types.ProviderOverpass,
				Coords:   ec(48.86, 2.29),
			},
			want: true,
		},
		{
			name: "overpass with wikidata id",
			p: types.RawPoi{
				Provider:   types.ProviderOverpass,
				Coords:     ec(48.86, 2.29),
				WikidataID: "Q243",
			},
			want: false,
		},
		{
			name: "geonames without wikidata id",
			p: types.RawPoi{
				Provider: types.ProviderGeoNames,
				Coords:   ec(48.86, 2.29),
			},
			want: true,
		},
		{
			name: "geonames with wikidata id",
			p: types.RawPoi{
				Provider:   types.ProviderGeoNames,
				Coords:     ec(48.86, 2.29),
				WikidataID: "Q243",
			},
			want: true,
		},
		{
			name: "wikipedia POI",
			p: types.RawPoi{
				Provider: types.ProviderWikipedia,
				Coords:   ec(48.86, 2.29),
			},
			want: false,
		},
		{
			name: "wikipedia events POI",
			p: types.RawPoi{
				Provider: types.ProviderWikipediaEvents,
				Coords:   ec(48.86, 2.29),
			},
			want: false,
		},
		{
			name: "nil coords",
			p: types.RawPoi{
				Provider: types.ProviderGeoNames,
			},
			want: false,
		},
		{
			name: "approximate coords",
			p: types.RawPoi{
				Provider: types.ProviderGeoNames,
				Coords:   &types.Coordinates{Lat: 48.86, Lng: 2.29, Approximate: true},
			},
			want: false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := needsWikipediaEnrichment(c.p); got != c.want {
				t.Errorf("needsWikipediaEnrichment = %v, want %v", got, c.want)
			}
		})
	}
}

func TestHasPoiNeedingEnrichment(t *testing.T) {
	overpassEnriched := types.RawPoi{Provider: types.ProviderOverpass, Coords: ec(48.86, 2.29), WikidataID: "Q1"}
	overpassMissing := types.RawPoi{Provider: types.ProviderOverpass, Coords: ec(48.86, 2.29)}
	geonamesEnriched := types.RawPoi{Provider: types.ProviderGeoNames, Coords: ec(48.86, 2.29), WikidataID: "Q1"}

	if hasPoiNeedingEnrichment(nil) {
		t.Error("empty slice: expected false")
	}
	if hasPoiNeedingEnrichment([]types.RawPoi{overpassEnriched}) {
		t.Error("only enriched non-geonames: expected false")
	}
	if !hasPoiNeedingEnrichment([]types.RawPoi{overpassEnriched, overpassMissing}) {
		t.Error("one missing wikidata id: expected true")
	}
	if !hasPoiNeedingEnrichment([]types.RawPoi{geonamesEnriched}) {
		t.Error("geonames always needs URL upgrade check: expected true")
	}
}

func TestClosestWikipediaNeighbour(t *testing.T) {
	target := types.RawPoi{Coords: ec(48.8584, 2.2945)}
	wikiPois := []types.RawPoi{
		{ID: "wikipedia:1", Coords: ec(48.8584, 2.2960), WikidataID: "Q_FAR", SourceURL: "https://en.wikipedia.org/?curid=1"},
		{ID: "wikipedia:2", Coords: ec(48.8585, 2.2946), WikidataID: "Q_CLOSE", SourceURL: "https://en.wikipedia.org/?curid=2"},
		{ID: "wikipedia:3", Coords: ec(48.8586, 2.2945), WikidataID: "Q_MID", SourceURL: "https://en.wikipedia.org/?curid=3"},
	}
	got := closestWikipediaNeighbour(target, wikiPois)
	if got == nil {
		t.Fatal("got nil, want wikipedia:2")
	}
	if got.ID != "wikipedia:2" {
		t.Errorf("got %q, want wikipedia:2", got.ID)
	}
}

func TestClosestWikipediaNeighbourOutsideRadius(t *testing.T) {
	target := types.RawPoi{Coords: ec(48.8584, 2.2945)}
	wikiPois := []types.RawPoi{
		{Coords: ec(48.8600, 2.2960), WikidataID: "Q_FAR"},
	}
	if got := closestWikipediaNeighbour(target, wikiPois); got != nil {
		t.Errorf("got %+v, want nil", got)
	}
}

func TestClosestWikipediaNeighbourSkipsMalformedCoords(t *testing.T) {
	target := types.RawPoi{Coords: ec(48.8584, 2.2945)}
	wikiPois := []types.RawPoi{
		{WikidataID: "Q_NIL_COORDS"},
		{Coords: &types.Coordinates{Lat: 48.8585, Lng: 2.2946, Approximate: true}, WikidataID: "Q_APPROX"},
		{ID: "wikipedia:ok", Coords: ec(48.8584, 2.2946), WikidataID: "Q_OK"},
	}
	got := closestWikipediaNeighbour(target, wikiPois)
	if got == nil {
		t.Fatal("got nil, want wikipedia:ok")
	}
	if got.ID != "wikipedia:ok" {
		t.Errorf("got %q, want wikipedia:ok", got.ID)
	}
}

func TestEnrichWithWikidata_SwapsGeoNamesSourceURL(t *testing.T) {
	wikiPois := []types.RawPoi{
		{
			ID: "wikipedia:5013364", Name: "Eiffel Tower",
			Provider:    types.ProviderWikipedia,
			Coords:      ec(48.85837, 2.29450),
			WikidataID:  "Q243",
			SourceURL:   "https://en.wikipedia.org/?curid=5013364",
			Description: "The Eiffel Tower is a wrought-iron lattice tower on the Champ de Mars in Paris.",
		},
	}
	svc := newServiceWithStub(&stubWikipediaProvider{pois: wikiPois})

	raw := []types.RawPoi{
		{
			ID: "geonames:6254976", Name: "Eiffel Tower",
			Provider:  types.ProviderGeoNames,
			Coords:    ec(48.85838, 2.29452),
			SourceURL: "https://www.geonames.org/6254976",
		},
	}
	out := svc.enrichWithWikidata(context.Background(), raw, types.SearchQuery{Mode: types.ModeRadius})

	if len(out) != 1 {
		t.Fatalf("expected 1 POI, got %d", len(out))
	}
	if out[0].SourceURL != "https://en.wikipedia.org/?curid=5013364" {
		t.Errorf("SourceURL = %q, want Wikipedia article URL", out[0].SourceURL)
	}
	if out[0].WikidataID != "Q243" {
		t.Errorf("WikidataID = %q, want Q243", out[0].WikidataID)
	}
	if out[0].Description != "The Eiffel Tower is a wrought-iron lattice tower on the Champ de Mars in Paris." {
		t.Errorf("Description = %q, want Wikipedia extract", out[0].Description)
	}
	if out[0].ID != "geonames:6254976" {
		t.Errorf("ID should stay geonames:6254976, got %q", out[0].ID)
	}
}

func TestEnrichWithWikidata_KeepsGeoNamesDescriptionWhenAlreadySet(t *testing.T) {
	wikiPois := []types.RawPoi{
		{
			ID: "wikipedia:1", Provider: types.ProviderWikipedia,
			Coords:      ec(48.85837, 2.29450),
			WikidataID:  "Q243",
			Description: "Wikipedia copy",
		},
	}
	svc := newServiceWithStub(&stubWikipediaProvider{pois: wikiPois})

	raw := []types.RawPoi{
		{
			ID: "geonames:1", Provider: types.ProviderGeoNames,
			Coords:      ec(48.85838, 2.29452),
			Description: "Pre-existing copy",
		},
	}
	out := svc.enrichWithWikidata(context.Background(), raw, types.SearchQuery{Mode: types.ModeRadius})

	if out[0].Description != "Pre-existing copy" {
		t.Errorf("Description = %q, want pre-existing value preserved", out[0].Description)
	}
}

func TestEnrichWithWikidata_KeepsGeoNamesURLWhenNoNeighbour(t *testing.T) {
	wikiPois := []types.RawPoi{
		{
			ID: "wikipedia:far", Name: "Somewhere Else",
			Provider:   types.ProviderWikipedia,
			Coords:     ec(48.8700, 2.3500),
			WikidataID: "Q_FAR",
			SourceURL:  "https://en.wikipedia.org/?curid=999",
		},
	}
	svc := newServiceWithStub(&stubWikipediaProvider{pois: wikiPois})

	raw := []types.RawPoi{
		{
			ID: "geonames:1", Name: "Local Place",
			Provider:  types.ProviderGeoNames,
			Coords:    ec(48.85837, 2.29450),
			SourceURL: "https://www.geonames.org/1",
		},
	}
	out := svc.enrichWithWikidata(context.Background(), raw, types.SearchQuery{Mode: types.ModeRadius})

	if out[0].SourceURL != "https://www.geonames.org/1" {
		t.Errorf("SourceURL = %q, want original geonames.org URL", out[0].SourceURL)
	}
}

func TestEnrichWithWikidata_LeavesOverpassFieldsUntouched(t *testing.T) {
	wikiPois := []types.RawPoi{
		{
			ID: "wikipedia:1", Name: "Eiffel Tower",
			Provider:    types.ProviderWikipedia,
			Coords:      ec(48.85837, 2.29450),
			WikidataID:  "Q243",
			SourceURL:   "https://en.wikipedia.org/?curid=5013364",
			Description: "Wikipedia copy",
		},
	}
	svc := newServiceWithStub(&stubWikipediaProvider{pois: wikiPois})

	raw := []types.RawPoi{
		{
			ID: "overpass:5013364", Name: "Tour Eiffel",
			Provider:  types.ProviderOverpass,
			Coords:    ec(48.85838, 2.29452),
			SourceURL: "https://www.openstreetmap.org/node/5013364",
		},
	}
	out := svc.enrichWithWikidata(context.Background(), raw, types.SearchQuery{Mode: types.ModeRadius})

	if out[0].WikidataID != "Q243" {
		t.Errorf("WikidataID = %q, want Q243", out[0].WikidataID)
	}
	if out[0].SourceURL != "https://www.openstreetmap.org/node/5013364" {
		t.Errorf("Overpass SourceURL should be preserved, got %q", out[0].SourceURL)
	}
	if out[0].Description != "" {
		t.Errorf("Overpass Description should not be filled by Wikipedia, got %q", out[0].Description)
	}
}
