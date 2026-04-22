package wikipedia_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/trippier/poi-api/internal/providers/wikipedia"
	"github.com/trippier/poi-api/pkg/types"
)

// geosearchResp returns a minimal Wikimedia geosearch API response.
const geosearchResp = `{
  "query": {
    "geosearch": [
      {"pageid": 1, "title": "Eiffel Tower", "lat": 48.8584, "lon": 2.2945, "dist": 50.5},
      {"pageid": 2, "title": "Champ de Mars",  "lat": 48.8555, "lon": 2.2988, "dist": 300.0}
    ]
  }
}`

// enrichResp returns a minimal pages enrichment response.
const enrichResp = `{
  "query": {
    "pages": {
      "1": {
        "pageid": 1,
        "title": "Eiffel Tower",
        "extract": "The Eiffel Tower is a wrought-iron lattice tower.",
        "thumbnail": {"source": "https://upload.wikimedia.org/thumb/eiffel.jpg"},
        "pageprops": {"wikibase_item": "Q243"}
      },
      "2": {
        "pageid": 2,
        "title": "Champ de Mars",
        "extract": "A large public greenspace.",
        "pageprops": {}
      }
    }
  }
}`

func newMultiHandler(geosearch, enrich string) http.Handler {
	call := 0
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call++
		if call == 1 {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(geosearch))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(enrich))
	})
}

func TestSearch_ReturnsEnrichedPois(t *testing.T) {
	srv := httptest.NewServer(newMultiHandler(geosearchResp, enrichResp))
	defer srv.Close()

	p := wikipedia.NewWithBaseURL(srv.URL)

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
	if len(pois) != 2 {
		t.Fatalf("expected 2 POIs, got %d", len(pois))
	}

	var eiffel *types.RawPoi
	for i := range pois {
		if pois[i].ID == "wikipedia:1" {
			eiffel = &pois[i]
			break
		}
	}
	if eiffel == nil {
		t.Fatal("Eiffel Tower POI not found")
	}
	if eiffel.WikidataID != "Q243" {
		t.Errorf("WikidataID = %q, want Q243", eiffel.WikidataID)
	}
	if eiffel.Description == "" {
		t.Error("description should be set from extract")
	}
	if eiffel.Thumbnail == "" {
		t.Error("thumbnail should be set")
	}
}

func TestSearch_EmptyGeosearch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"query":{"geosearch":[]}}`))
	}))
	defer srv.Close()

	p := wikipedia.NewWithBaseURL(srv.URL)

	q := types.SearchQuery{Mode: types.ModeRadius, Lat: 0, Lng: 0, Radius: 100}

	pois, err := p.Search(context.Background(), q)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pois) != 0 {
		t.Errorf("expected 0 pois, got %d", len(pois))
	}
}

func TestSupportsMode(t *testing.T) {
	p := wikipedia.NewWithBaseURL("http://example.com")
	if !p.SupportsMode(types.ModeRadius) {
		t.Error("should support radius mode")
	}
	if p.SupportsMode(types.ModePolygon) {
		t.Error("should NOT support polygon mode")
	}
}
