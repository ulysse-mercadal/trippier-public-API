package wikipedia_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/trippier/poi-api/internal/providers/wikipedia"
	"github.com/trippier/poi-api/pkg/types"
)

// ── Shared fixtures ───────────────────────────────────────────────────────────

const geosearchResp = `{
  "query": {
    "geosearch": [
      {"pageid": 1, "title": "Eiffel Tower",  "lat": 48.8584, "lon": 2.2945, "dist": 50.5},
      {"pageid": 2, "title": "Champ de Mars", "lat": 48.8555, "lon": 2.2988, "dist": 300.0}
    ]
  }
}`

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

// sparqlNone is a SPARQL result with no matching entities.
const sparqlNone = `{"results":{"bindings":[]}}`

// sparqlMatchQ243 simulates SPARQL classifying Q243 as matching the queried class.
const sparqlMatchQ243 = `{
  "results": {
    "bindings": [
      {"item": {"type": "uri", "value": "http://www.wikidata.org/entity/Q243"}}
    ]
  }
}`

// newWikipediaHandler sequences responses: first call → geosearch, second → enrich.
func newWikipediaHandler(geosearch, enrich string) http.Handler {
	call := 0
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		call++
		w.WriteHeader(http.StatusOK)
		if call == 1 {
			_, _ = w.Write([]byte(geosearch))
		} else {
			_, _ = w.Write([]byte(enrich))
		}
	})
}

func newSPARQLHandler(body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	})
}

var testQuery = types.SearchQuery{Mode: types.ModeRadius, Lat: 48.8566, Lng: 2.3522, Radius: 5000}

// ── Provider (places) tests ───────────────────────────────────────────────────

func TestProvider_ReturnsEnrichedPois(t *testing.T) {
	wikiSrv := httptest.NewServer(newWikipediaHandler(geosearchResp, enrichResp))
	defer wikiSrv.Close()
	sparqlSrv := httptest.NewServer(newSPARQLHandler(sparqlNone))
	defer sparqlSrv.Close()

	p := wikipedia.NewWithURLs(wikiSrv.URL, sparqlSrv.URL)

	pois, err := p.Search(context.Background(), testQuery)
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
	if eiffel.Type != types.TypeSee {
		t.Errorf("type = %q, want %q", eiffel.Type, types.TypeSee)
	}
}

func TestProvider_ReturnsAllGeosearchResults(t *testing.T) {
	// The places Provider no longer filters by Wikidata — it is excluded from
	// AllProviders and used only for enrichment. All geosearch results are returned.
	wikiSrv := httptest.NewServer(newWikipediaHandler(geosearchResp, enrichResp))
	defer wikiSrv.Close()
	sparqlSrv := httptest.NewServer(newSPARQLHandler(sparqlMatchQ243))
	defer sparqlSrv.Close()

	p := wikipedia.NewWithURLs(wikiSrv.URL, sparqlSrv.URL)

	pois, err := p.Search(context.Background(), testQuery)
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(pois) != 2 {
		t.Fatalf("expected 2 POIs (no filtering), got %d", len(pois))
	}
}

func TestProvider_EmptyGeosearch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"query":{"geosearch":[]}}`))
	}))
	defer srv.Close()

	p := wikipedia.NewWithURLs(srv.URL, srv.URL)

	pois, err := p.Search(context.Background(), types.SearchQuery{Mode: types.ModeRadius, Lat: 0, Lng: 0, Radius: 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pois) != 0 {
		t.Errorf("expected 0 pois, got %d", len(pois))
	}
}

func TestProvider_SupportsMode(t *testing.T) {
	p := wikipedia.NewWithURLs("http://example.com", "http://example.com")
	if !p.SupportsMode(types.ModeRadius) {
		t.Error("should support radius mode")
	}
	if p.SupportsMode(types.ModePolygon) {
		t.Error("should NOT support polygon mode")
	}
}

// ── EventProvider (festivals) tests ──────────────────────────────────────────

func TestEventProvider_ReturnsFestivals(t *testing.T) {
	wikiSrv := httptest.NewServer(newWikipediaHandler(geosearchResp, enrichResp))
	defer wikiSrv.Close()
	// SPARQL reports Q243 as a festival → Eiffel Tower is kept as an event POI.
	sparqlSrv := httptest.NewServer(newSPARQLHandler(sparqlMatchQ243))
	defer sparqlSrv.Close()

	p := wikipedia.NewEventProviderWithURLs(wikiSrv.URL, sparqlSrv.URL)

	pois, err := p.Search(context.Background(), testQuery)
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(pois) != 1 {
		t.Fatalf("expected 1 festival POI, got %d", len(pois))
	}
	if pois[0].Type != types.TypeEvent {
		t.Errorf("type = %q, want %q", pois[0].Type, types.TypeEvent)
	}
	if pois[0].Provider != types.ProviderWikipediaEvents {
		t.Errorf("provider = %q, want %q", pois[0].Provider, types.ProviderWikipediaEvents)
	}
	if !pois[0].Recurring {
		t.Error("recurring should be true for festival POIs")
	}
}

func TestEventProvider_NonFestivalsAreExcluded(t *testing.T) {
	wikiSrv := httptest.NewServer(newWikipediaHandler(geosearchResp, enrichResp))
	defer wikiSrv.Close()
	// SPARQL finds no festivals → nothing is returned.
	sparqlSrv := httptest.NewServer(newSPARQLHandler(sparqlNone))
	defer sparqlSrv.Close()

	p := wikipedia.NewEventProviderWithURLs(wikiSrv.URL, sparqlSrv.URL)

	pois, err := p.Search(context.Background(), testQuery)
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(pois) != 0 {
		t.Errorf("expected 0 non-festival POIs, got %d", len(pois))
	}
}

func TestEventProvider_SupportsMode(t *testing.T) {
	p := wikipedia.NewEventProviderWithURLs("http://example.com", "http://example.com")
	if !p.SupportsMode(types.ModeRadius) {
		t.Error("should support radius mode")
	}
	if p.SupportsMode(types.ModePolygon) {
		t.Error("should NOT support polygon mode")
	}
}
