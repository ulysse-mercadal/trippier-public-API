package overpass_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/trippier/poi-api/internal/providers/overpass"
	"github.com/trippier/poi-api/pkg/types"
)

const sampleOverpassResponse = `{
  "elements": [
    {
      "type": "node",
      "id": 12345,
      "lat": 48.8606,
      "lon": 2.3376,
      "tags": {
        "name": "Musée du Louvre",
        "tourism": "museum",
        "website": "https://louvre.fr",
        "opening_hours": "Mo-Su 09:00-18:00",
        "wikidata": "Q19675"
      }
    },
    {
      "type": "way",
      "id": 99999,
      "center": {"lat": 48.8530, "lon": 2.3499},
      "tags": {
        "name": "Notre-Dame de Paris",
        "tourism": "attraction"
      }
    },
    {
      "type": "node",
      "id": 11111,
      "lat": 48.8700,
      "lon": 2.3600,
      "tags": {
        "amenity": "restaurant",
        "name": "Le Relais"
      }
    },
    {
      "type": "node",
      "id": 22222,
      "lat": 1.0,
      "lon": 1.0,
      "tags": {}
    }
  ]
}`

func newTestServer(body string, status int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
}

func TestSearch_RadiusMode(t *testing.T) {
	srv := newTestServer(sampleOverpassResponse, http.StatusOK)
	defer srv.Close()

	p := overpass.NewWithURL(srv.URL)

	q := types.SearchQuery{
		Mode:   types.ModeRadius,
		Lat:    48.8566,
		Lng:    2.3522,
		Radius: 5000,
	}

	pois, err := p.Search(context.Background(), q)
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}

	// Element 22222 has no name — must be skipped.
	if len(pois) != 3 {
		t.Fatalf("expected 3 POIs (unnamed element skipped), got %d", len(pois))
	}
}

func TestSearch_WayUsesCenter(t *testing.T) {
	srv := newTestServer(sampleOverpassResponse, http.StatusOK)
	defer srv.Close()

	p := overpass.NewWithURL(srv.URL)

	q := types.SearchQuery{Mode: types.ModeRadius, Lat: 48.8566, Lng: 2.3522, Radius: 5000}
	pois, err := p.Search(context.Background(), q)
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}

	var notredame *types.RawPoi
	for i := range pois {
		if pois[i].ID == "overpass:99999" {
			notredame = &pois[i]
			break
		}
	}
	if notredame == nil {
		t.Fatal("Notre-Dame way not found")
	}
	if notredame.Coords.Lat != 48.8530 || notredame.Coords.Lng != 2.3499 {
		t.Errorf("way coords should come from center, got %+v", notredame.Coords)
	}
}

func TestSearch_TypeResolution(t *testing.T) {
	srv := newTestServer(sampleOverpassResponse, http.StatusOK)
	defer srv.Close()

	p := overpass.NewWithURL(srv.URL)

	q := types.SearchQuery{Mode: types.ModeRadius, Lat: 48.8566, Lng: 2.3522, Radius: 5000}
	pois, err := p.Search(context.Background(), q)
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}

	for _, poi := range pois {
		switch poi.ID {
		case "overpass:12345":
			if poi.Type != types.TypeSee {
				t.Errorf("Louvre: want TypeSee, got %s", poi.Type)
			}
			if poi.WikidataID != "Q19675" {
				t.Errorf("Louvre: want wikidata Q19675, got %s", poi.WikidataID)
			}
		case "overpass:11111":
			if poi.Type != types.TypeEat {
				t.Errorf("Le Relais: want TypeEat, got %s", poi.Type)
			}
		}
	}
}

func TestSearch_HTTPError(t *testing.T) {
	srv := newTestServer("", http.StatusInternalServerError)
	defer srv.Close()

	p := overpass.NewWithURL(srv.URL)

	q := types.SearchQuery{Mode: types.ModeRadius, Lat: 48.85, Lng: 2.35, Radius: 1000}
	_, err := p.Search(context.Background(), q)
	if err == nil {
		t.Error("expected error on HTTP 500, got nil")
	}
}

func TestSearch_InvalidJSON(t *testing.T) {
	srv := newTestServer("not-json", http.StatusOK)
	defer srv.Close()

	p := overpass.NewWithURL(srv.URL)

	q := types.SearchQuery{Mode: types.ModeRadius, Lat: 48.85, Lng: 2.35, Radius: 1000}
	_, err := p.Search(context.Background(), q)
	if err == nil {
		t.Error("expected error on invalid JSON, got nil")
	}
}

func TestSupportsMode(t *testing.T) {
	p := overpass.New()
	for _, mode := range []types.SearchMode{types.ModeRadius, types.ModePolygon, types.ModeDistrict} {
		if !p.SupportsMode(mode) {
			t.Errorf("Overpass should support mode %s", mode)
		}
	}
}
