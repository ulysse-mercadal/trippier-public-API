package geonames_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/trippier/poi-api/internal/providers/geonames"
	"github.com/trippier/poi-api/pkg/types"
)

// sampleResponse has three items:
//   - Aiguille du Midi (MUS) — in fcodeTypeMap, valid coords → kept
//   - Mont Blanc (MT, mountain) — not in fcodeTypeMap → filtered
//   - Bad Coords (MUS but unparseable lat/lng) → filtered
const sampleResponse = `{
  "geonames": [
    {"geonameId": 6255148, "name": "Aiguille du Midi", "lat": "45.8791", "lng": "6.8874", "fcode": "MUS", "fcodeName": "museum",   "countryCode": "FR", "distance": "5.2"},
    {"geonameId": 6269131, "name": "Mont Blanc",        "lat": "45.8326", "lng": "6.8652", "fcode": "MT",  "fcodeName": "mountain", "countryCode": "FR", "distance": "0"},
    {"geonameId": 9999999, "name": "Bad Coords",        "lat": "not-a-float", "lng": "also-bad", "fcode": "MUS", "fcodeName": "", "countryCode": "FR", "distance": "1"}
  ]
}`

func newTestServer(t *testing.T, body string, status int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
}

// ── SupportsMode ──────────────────────────────────────────────────────────────

func TestSupportsMode(t *testing.T) {
	p := geonames.New("testuser")
	cases := []struct {
		mode types.SearchMode
		want bool
	}{
		{types.ModeRadius, true},
		{types.ModeDistrict, true},
		{types.ModePolygon, false},
	}
	for _, tc := range cases {
		if got := p.SupportsMode(tc.mode); got != tc.want {
			t.Errorf("SupportsMode(%s) = %v, want %v", tc.mode, got, tc.want)
		}
	}
}

func TestName(t *testing.T) {
	p := geonames.New("testuser")
	if p.Name() != types.ProviderGeoNames {
		t.Errorf("Name() = %q, want %q", p.Name(), types.ProviderGeoNames)
	}
}

// ── Search — radius mode ──────────────────────────────────────────────────────

func TestSearch_RadiusMode_ParsesResponse(t *testing.T) {
	srv := newTestServer(t, sampleResponse, http.StatusOK)
	defer srv.Close()

	p := geonames.NewWithURL(srv.URL, "testuser")
	pois, err := p.Search(context.Background(), types.SearchQuery{
		Mode:   types.ModeRadius,
		Lat:    45.83,
		Lng:    6.86,
		Radius: 10000,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	// 3 items: Mont Blanc (unknown fcode) filtered, Bad Coords (bad lat/lng) filtered.
	if len(pois) != 1 {
		t.Fatalf("expected 1 POI (unknown fcode and bad coords filtered), got %d", len(pois))
	}
}

func TestSearch_RadiusMode_IDFormat(t *testing.T) {
	srv := newTestServer(t, sampleResponse, http.StatusOK)
	defer srv.Close()

	p := geonames.NewWithURL(srv.URL, "testuser")
	pois, err := p.Search(context.Background(), types.SearchQuery{
		Mode: types.ModeRadius, Lat: 45.83, Lng: 6.86, Radius: 10000,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	ids := map[string]bool{}
	for _, poi := range pois {
		ids[poi.ID] = true
	}
	if !ids["geonames:6255148"] {
		t.Error("expected ID geonames:6255148 for Aiguille du Midi")
	}
}

func TestSearch_RadiusMode_TypeResolution(t *testing.T) {
	srv := newTestServer(t, sampleResponse, http.StatusOK)
	defer srv.Close()

	p := geonames.NewWithURL(srv.URL, "testuser")
	pois, err := p.Search(context.Background(), types.SearchQuery{
		Mode: types.ModeRadius, Lat: 45.83, Lng: 6.86, Radius: 10000,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	byName := map[string]types.RawPoi{}
	for _, poi := range pois {
		byName[poi.Name] = poi
	}

	// fcode "MUS" → TypeSee
	if byName["Aiguille du Midi"].Type != types.TypeSee {
		t.Errorf("Aiguille du Midi: want TypeSee, got %s", byName["Aiguille du Midi"].Type)
	}
	// fcode "MT" (mountain) is not in fcodeTypeMap → filtered out entirely
	if _, present := byName["Mont Blanc"]; present {
		t.Error("Mont Blanc (fcode MT) should be filtered out — not in fcodeTypeMap")
	}
}

func TestSearch_RadiusMode_Coordinates(t *testing.T) {
	srv := newTestServer(t, sampleResponse, http.StatusOK)
	defer srv.Close()

	p := geonames.NewWithURL(srv.URL, "testuser")
	pois, err := p.Search(context.Background(), types.SearchQuery{
		Mode: types.ModeRadius, Lat: 45.83, Lng: 6.86, Radius: 10000,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	byName := map[string]types.RawPoi{}
	for _, poi := range pois {
		byName[poi.Name] = poi
	}

	if byName["Aiguille du Midi"].Coords == nil {
		t.Fatal("Aiguille du Midi should have coordinates")
	}
	if byName["Aiguille du Midi"].Coords.Lat != 45.8791 {
		t.Errorf("Aiguille du Midi lat = %v, want 45.8791", byName["Aiguille du Midi"].Coords.Lat)
	}
}

// ── Search — district mode ────────────────────────────────────────────────────

func TestSearch_DistrictMode_CallsSearchEndpoint(t *testing.T) {
	var capturedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"geonames": []any{}}) //nolint:errcheck
	}))
	defer srv.Close()

	p := geonames.NewWithURL(srv.URL, "testuser")
	_, err := p.Search(context.Background(), types.SearchQuery{
		Mode:     types.ModeDistrict,
		District: "Chamonix",
	})
	if err != nil {
		t.Fatalf("Search district: %v", err)
	}

	if capturedPath != "/searchJSON" {
		t.Errorf("district mode should call /searchJSON, got %s", capturedPath)
	}
}

// ── Error handling ────────────────────────────────────────────────────────────

func TestSearch_HTTPError(t *testing.T) {
	srv := newTestServer(t, "", http.StatusInternalServerError)
	defer srv.Close()

	p := geonames.NewWithURL(srv.URL, "testuser")
	_, err := p.Search(context.Background(), types.SearchQuery{
		Mode: types.ModeRadius, Lat: 45.83, Lng: 6.86, Radius: 1000,
	})
	if err == nil {
		t.Error("expected error on HTTP 500, got nil")
	}
}

func TestSearch_InvalidJSON(t *testing.T) {
	srv := newTestServer(t, "not-json", http.StatusOK)
	defer srv.Close()

	p := geonames.NewWithURL(srv.URL, "testuser")
	_, err := p.Search(context.Background(), types.SearchQuery{
		Mode: types.ModeRadius, Lat: 45.83, Lng: 6.86, Radius: 1000,
	})
	if err == nil {
		t.Error("expected error on invalid JSON, got nil")
	}
}

func TestSearch_UnsupportedMode_ReturnsNil(t *testing.T) {
	p := geonames.New("testuser")
	pois, err := p.Search(context.Background(), types.SearchQuery{
		Mode:    types.ModePolygon,
		Polygon: "48.84 2.34 48.86 2.34",
	})
	if err != nil {
		t.Fatalf("unexpected error for unsupported mode: %v", err)
	}
	if len(pois) != 0 {
		t.Errorf("expected nil for unsupported mode, got %d POIs", len(pois))
	}
}

func TestSearch_EmptyResponse(t *testing.T) {
	srv := newTestServer(t, `{"geonames":[]}`, http.StatusOK)
	defer srv.Close()

	p := geonames.NewWithURL(srv.URL, "testuser")
	pois, err := p.Search(context.Background(), types.SearchQuery{
		Mode: types.ModeRadius, Lat: 45.83, Lng: 6.86, Radius: 1000,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(pois) != 0 {
		t.Errorf("expected 0 POIs for empty response, got %d", len(pois))
	}
}
