package overpass_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/trippier/poi-api/internal/providers/overpass"
	"github.com/trippier/poi-api/pkg/types"
)

func newCtx() context.Context { return context.Background() }

func newTestServerCapture(out *string, body string, status int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		vals, _ := url.ParseQuery(string(b))
		*out = vals.Get("data")
		w.WriteHeader(status)
		w.Write([]byte(body)) //nolint:errcheck
	}))
}

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

func TestSearch_SourceURL_NodeAndWay(t *testing.T) {
	srv := newTestServer(sampleOverpassResponse, http.StatusOK)
	defer srv.Close()

	p := overpass.NewWithURL(srv.URL)
	pois, err := p.Search(context.Background(), types.SearchQuery{
		Mode: types.ModeRadius, Lat: 48.8566, Lng: 2.3522, Radius: 5000,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	want := map[string]string{
		"overpass:12345": "https://www.openstreetmap.org/node/12345",
		"overpass:99999": "https://www.openstreetmap.org/way/99999",
	}
	for _, poi := range pois {
		if expected, ok := want[poi.ID]; ok && poi.SourceURL != expected {
			t.Errorf("%s SourceURL = %q, want %q", poi.ID, poi.SourceURL, expected)
		}
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

// TestSearch_FallsThroughOn5xx verifies that a server-error response on the
// first mirror is treated as retryable and the next mirror is tried.
func TestSearch_FallsThroughOn5xx(t *testing.T) {
	primaryHits, secondaryHits := 0, 0
	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		primaryHits++
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer primary.Close()
	secondary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		secondaryHits++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(sampleOverpassResponse))
	}))
	defer secondary.Close()

	p := overpass.NewWithURLs([]string{primary.URL, secondary.URL})
	q := types.SearchQuery{Mode: types.ModeRadius, Lat: 48.85, Lng: 2.35, Radius: 1000}
	pois, err := p.Search(context.Background(), q)
	if err != nil {
		t.Fatalf("expected fallback success, got: %v", err)
	}
	if primaryHits != 1 {
		t.Errorf("primary should be hit once, got %d", primaryHits)
	}
	if secondaryHits != 1 {
		t.Errorf("secondary should be hit once, got %d", secondaryHits)
	}
	if len(pois) == 0 {
		t.Error("expected POIs from the secondary mirror")
	}
}

// TestSearch_FallsThroughOn429 verifies the rate-limit case is retryable.
func TestSearch_FallsThroughOn429(t *testing.T) {
	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer primary.Close()
	secondary := newTestServer(sampleOverpassResponse, http.StatusOK)
	defer secondary.Close()

	p := overpass.NewWithURLs([]string{primary.URL, secondary.URL})
	q := types.SearchQuery{Mode: types.ModeRadius, Lat: 48.85, Lng: 2.35, Radius: 1000}
	if _, err := p.Search(context.Background(), q); err != nil {
		t.Fatalf("expected fallback success on 429, got: %v", err)
	}
}

// TestSearch_DoesNotRetryOn4xx verifies that a client error short-circuits
// the mirror loop — the same request would fail everywhere.
func TestSearch_DoesNotRetryOn4xx(t *testing.T) {
	primaryHits, secondaryHits := 0, 0
	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		primaryHits++
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer primary.Close()
	secondary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		secondaryHits++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(sampleOverpassResponse))
	}))
	defer secondary.Close()

	p := overpass.NewWithURLs([]string{primary.URL, secondary.URL})
	q := types.SearchQuery{Mode: types.ModeRadius, Lat: 48.85, Lng: 2.35, Radius: 1000}
	if _, err := p.Search(context.Background(), q); err == nil {
		t.Error("expected hard failure on HTTP 400, got nil")
	}
	if primaryHits != 1 || secondaryHits != 0 {
		t.Errorf("4xx should short-circuit: primary=%d secondary=%d", primaryHits, secondaryHits)
	}
}

// TestSearch_ImagesFromOSMTags exercises image/image:url/wikimedia_commons
// tag handling: direct URLs kept, non-http values dropped, Commons file refs
// turned into Special:FilePath URLs, duplicates removed and the result capped
// at 3 entries.
func TestSearch_ImagesFromOSMTags(t *testing.T) {
	body := `{"elements":[
		{"type":"node","id":1,"lat":48.86,"lon":2.29,"tags":{
			"name":"Direct URL","tourism":"attraction",
			"image":"https://example.org/a.jpg"
		}},
		{"type":"node","id":2,"lat":48.86,"lon":2.29,"tags":{
			"name":"Both tags","tourism":"attraction",
			"image":"https://example.org/b.jpg",
			"image:url":"https://example.org/b-alt.jpg"
		}},
		{"type":"node","id":3,"lat":48.86,"lon":2.29,"tags":{
			"name":"Commons file","tourism":"attraction",
			"wikimedia_commons":"File:Eiffel Tower.jpg"
		}},
		{"type":"node","id":4,"lat":48.86,"lon":2.29,"tags":{
			"name":"Commons category dropped","tourism":"attraction",
			"wikimedia_commons":"Category:Paris"
		}},
		{"type":"node","id":5,"lat":48.86,"lon":2.29,"tags":{
			"name":"Non-http image dropped","tourism":"attraction",
			"image":"file:///local/foo.jpg"
		}},
		{"type":"node","id":6,"lat":48.86,"lon":2.29,"tags":{
			"name":"Duplicate dropped","tourism":"attraction",
			"image":"https://example.org/x.jpg",
			"image:url":"https://example.org/x.jpg"
		}}
	]}`
	var got string
	srv := newTestServerCapture(&got, body, http.StatusOK)
	defer srv.Close()

	p := overpass.NewWithURLs([]string{srv.URL})
	pois, err := p.Search(newCtx(), types.SearchQuery{Mode: types.ModeRadius, Lat: 48.86, Lng: 2.29, Radius: 1000})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	byName := map[string][]string{}
	for _, poi := range pois {
		byName[poi.Name] = poi.Images
	}

	cases := []struct {
		name string
		want []string
	}{
		{"Direct URL", []string{"https://example.org/a.jpg"}},
		{"Both tags", []string{"https://example.org/b.jpg", "https://example.org/b-alt.jpg"}},
		{"Commons file", []string{"https://commons.wikimedia.org/wiki/Special:FilePath/Eiffel%20Tower.jpg"}},
		{"Commons category dropped", nil},
		{"Non-http image dropped", nil},
		{"Duplicate dropped", []string{"https://example.org/x.jpg"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := byName[c.name]
			if len(got) != len(c.want) {
				t.Fatalf("len = %d, want %d (got=%v)", len(got), len(c.want), got)
			}
			for i := range got {
				if got[i] != c.want[i] {
					t.Errorf("[%d] = %q, want %q", i, got[i], c.want[i])
				}
			}
		})
	}
}

// TestSearch_AllMirrorsFail reports the last error when every mirror fails.
func TestSearch_AllMirrorsFail(t *testing.T) {
	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer primary.Close()
	secondary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusGatewayTimeout)
	}))
	defer secondary.Close()

	p := overpass.NewWithURLs([]string{primary.URL, secondary.URL})
	q := types.SearchQuery{Mode: types.ModeRadius, Lat: 48.85, Lng: 2.35, Radius: 1000}
	if _, err := p.Search(context.Background(), q); err == nil {
		t.Error("expected failure when all mirrors return 5xx")
	}
}
