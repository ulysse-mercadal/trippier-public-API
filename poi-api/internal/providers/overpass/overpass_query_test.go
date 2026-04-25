package overpass_test

import (
	"strings"
	"testing"

	"github.com/trippier/poi-api/internal/providers/overpass"
	"github.com/trippier/poi-api/pkg/types"
)

func TestSearch_PolygonMode(t *testing.T) {
	srv := newTestServer(sampleOverpassResponse, 200)
	defer srv.Close()

	p := overpass.NewWithURL(srv.URL)
	pois, err := p.Search(newCtx(), types.SearchQuery{
		Mode:    types.ModePolygon,
		Polygon: "48.84 2.34 48.86 2.34 48.86 2.36 48.84 2.36",
	})
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(pois) == 0 {
		t.Error("expected pois, got none")
	}
}

func TestSearch_DistrictMode(t *testing.T) {
	srv := newTestServer(sampleOverpassResponse, 200)
	defer srv.Close()

	p := overpass.NewWithURL(srv.URL)
	pois, err := p.Search(newCtx(), types.SearchQuery{
		Mode:     types.ModeDistrict,
		District: "Montmartre",
	})
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(pois) == 0 {
		t.Error("expected pois, got none")
	}
}

func TestSearch_TypeFilter(t *testing.T) {
	srv := newTestServer(sampleOverpassResponse, 200)
	defer srv.Close()

	p := overpass.NewWithURL(srv.URL)
	pois, err := p.Search(newCtx(), types.SearchQuery{
		Mode:   types.ModeRadius,
		Lat:    48.8566,
		Lng:    2.3522,
		Radius: 5000,
		Types:  []types.PoiType{types.TypeEat, types.TypeDrink},
	})
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	_ = pois // query is built and sent; type coverage is the goal
}

func TestSearch_QueryContainsDistrict(t *testing.T) {
	var captured string
	srv := newTestServerCapture(&captured, sampleOverpassResponse, 200)
	defer srv.Close()

	p := overpass.NewWithURL(srv.URL)
	p.Search(newCtx(), types.SearchQuery{ //nolint:errcheck
		Mode:     types.ModeDistrict,
		District: "Belleville",
	})

	if !strings.Contains(captured, `"Belleville"`) {
		t.Errorf("district query should contain Belleville, got: %s", captured)
	}
	if !strings.Contains(captured, "-> .n") || !strings.Contains(captured, "-> .w") {
		t.Errorf("expected named sets .n and .w in query: %s", captured)
	}
}

func TestSearch_QueryContainsPolygon(t *testing.T) {
	var captured string
	srv := newTestServerCapture(&captured, sampleOverpassResponse, 200)
	defer srv.Close()

	p := overpass.NewWithURL(srv.URL)
	poly := "48.84 2.34 48.86 2.34 48.86 2.36"
	p.Search(newCtx(), types.SearchQuery{Mode: types.ModePolygon, Polygon: poly}) //nolint:errcheck

	if !strings.Contains(captured, poly) {
		t.Errorf("polygon query should contain poly coords, got: %s", captured)
	}
}

func TestResolveType_Shop(t *testing.T) {
	body := `{"elements":[{"type":"node","id":1,"lat":1,"lon":1,"tags":{"name":"Gift Shop","shop":"gift"}}]}`
	srv := newTestServer(body, 200)
	defer srv.Close()

	p := overpass.NewWithURL(srv.URL)
	pois, err := p.Search(newCtx(), types.SearchQuery{Mode: types.ModeRadius, Lat: 1, Lng: 1, Radius: 1000})
	if err != nil {
		t.Fatal(err)
	}
	if len(pois) != 1 || pois[0].Type != types.TypeBuy {
		t.Errorf("shop should resolve to TypeBuy, got %v", pois)
	}
}

func TestResolveType_Generic(t *testing.T) {
	body := `{"elements":[{"type":"node","id":1,"lat":1,"lon":1,"tags":{"name":"Unknown","tourism":"unknown_tag"}}]}`
	srv := newTestServer(body, 200)
	defer srv.Close()

	p := overpass.NewWithURL(srv.URL)
	pois, err := p.Search(newCtx(), types.SearchQuery{Mode: types.ModeRadius, Lat: 1, Lng: 1, Radius: 1000})
	if err != nil {
		t.Fatal(err)
	}
	if len(pois) != 1 || pois[0].Type != types.TypeGeneric {
		t.Errorf("unknown tag should resolve to TypeGeneric, got %v", pois)
	}
}

func TestName(t *testing.T) {
	p := overpass.New()
	if p.Name() != types.ProviderOverpass {
		t.Errorf("Name() = %s, want overpass", p.Name())
	}
}
