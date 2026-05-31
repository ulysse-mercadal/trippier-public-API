package ticketmaster_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/trippier/poi-api/internal/byok"
	"github.com/trippier/poi-api/internal/geo"
	"github.com/trippier/poi-api/internal/providers/ticketmaster"
	"github.com/trippier/poi-api/pkg/types"
)

const sampleResponse = `{
  "_embedded": {
    "events": [
      {
        "id": "tm1",
        "name": "Rock Concert",
        "url": "https://example.com/tm1",
        "info": "Great show",
        "images": [{"url":"https://img.example.com/1.jpg","width":1024,"height":576,"ratio":"16_9"}],
        "dates": {
          "start": {"dateTime": "2026-06-15T19:00:00Z"},
          "end":   {"dateTime": "2026-06-15T23:00:00Z"}
        },
        "_embedded": {
          "venues": [{"name":"Paris Arena","location":{"latitude":"48.8750","longitude":"2.3200"}}]
        }
      },
      {
        "id": "tm2",
        "name": "Jazz Night",
        "dates": {"start":{"dateTime":""},"end":{"dateTime":""}},
        "_embedded": {"venues": [{"name":"Unknown Venue","location":{"latitude":"","longitude":""}}]}
      }
    ]
  }
}`

func mockNominatim(t *testing.T, body string) func() {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	old := geo.NominatimReverseURL
	geo.NominatimReverseURL = srv.URL
	return func() {
		geo.NominatimReverseURL = old
		srv.Close()
	}
}

func ctxWithTMKey(key string) context.Context {
	return context.WithValue(context.Background(), byok.TicketmasterKey, key)
}

// ── No key → silently skipped ─────────────────────────────────────────────────

func TestSearch_NoKey_ReturnsNil(t *testing.T) {
	p := ticketmaster.New()
	pois, err := p.Search(context.Background(), types.SearchQuery{
		Mode: types.ModeRadius, Lat: 48.86, Lng: 2.35, Radius: 50_000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pois != nil {
		t.Errorf("expected nil pois without key, got %d", len(pois))
	}
}

// ── Name / IsByok / SupportsMode ─────────────────────────────────────────────

func TestName(t *testing.T) {
	p := ticketmaster.New()
	if p.Name() != types.ProviderTicketmaster {
		t.Errorf("Name() = %q, want %q", p.Name(), types.ProviderTicketmaster)
	}
}

func TestIsByok(t *testing.T) {
	p := ticketmaster.New()
	if !p.IsByok() {
		t.Error("IsByok() = false, want true")
	}
}

func TestSupportsMode(t *testing.T) {
	p := ticketmaster.New()
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

// ── Search with key + mock server ─────────────────────────────────────────────

func TestSearch_WithKey_ParsesEvents(t *testing.T) {
	// Mock both Nominatim (reverse geocode) and the TM API.
	restore := mockNominatim(t, `{"address":{"city":"Paris","country_code":"fr"}}`)
	defer restore()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sampleResponse))
	}))
	defer srv.Close()

	p := ticketmaster.NewWithURL(srv.URL)
	pois, err := p.Search(ctxWithTMKey("test-api-key"), types.SearchQuery{
		Mode: types.ModeRadius, Lat: 48.86, Lng: 2.35, Radius: 50_000,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	// 2 events — even the one with empty venue coords should be included (uses center).
	if len(pois) != 2 {
		t.Errorf("expected 2 events, got %d", len(pois))
	}
	if pois[0].Name != "Rock Concert" {
		t.Errorf("first event name = %q, want %q", pois[0].Name, "Rock Concert")
	}
	if pois[0].Type != types.TypeEvent {
		t.Errorf("first event type = %q, want %q", pois[0].Type, types.TypeEvent)
	}
	if pois[0].SourceURL != "https://example.com/tm1" {
		t.Errorf("first event SourceURL = %q, want %q", pois[0].SourceURL, "https://example.com/tm1")
	}
}

func TestSearch_EmptyEmbedded_ReturnsNil(t *testing.T) {
	restore := mockNominatim(t, `{"address":{"city":"Paris","country_code":"fr"}}`)
	defer restore()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	p := ticketmaster.NewWithURL(srv.URL)
	pois, err := p.Search(ctxWithTMKey("key"), types.SearchQuery{
		Mode: types.ModeRadius, Lat: 48.86, Lng: 2.35, Radius: 50_000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pois != nil {
		t.Errorf("expected nil for empty _embedded, got %d", len(pois))
	}
}

func TestSearch_NonOKStatus_ReturnsError(t *testing.T) {
	restore := mockNominatim(t, `{"address":{"city":"Paris","country_code":"fr"}}`)
	defer restore()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	p := ticketmaster.NewWithURL(srv.URL)
	_, err := p.Search(ctxWithTMKey("bad-key"), types.SearchQuery{
		Mode: types.ModeRadius, Lat: 48.86, Lng: 2.35, Radius: 50_000,
	})
	if err == nil {
		t.Error("expected error for 401 response")
	}
}

func TestSearch_InvalidJSON_ReturnsError(t *testing.T) {
	restore := mockNominatim(t, `{"address":{"city":"Paris","country_code":"fr"}}`)
	defer restore()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`not-json`))
	}))
	defer srv.Close()

	p := ticketmaster.NewWithURL(srv.URL)
	_, err := p.Search(ctxWithTMKey("key"), types.SearchQuery{
		Mode: types.ModeRadius, Lat: 48.86, Lng: 2.35, Radius: 50_000,
	})
	if err == nil {
		t.Error("expected error for invalid JSON response")
	}
}

func TestSearch_WithDate_UsesDateFilter(t *testing.T) {
	restore := mockNominatim(t, `{"address":{"city":"Paris","country_code":"fr"}}`)
	defer restore()

	var capturedURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.RawQuery
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	p := ticketmaster.NewWithURL(srv.URL)
	p.Search(ctxWithTMKey("key"), types.SearchQuery{ //nolint:errcheck
		Mode: types.ModeRadius, Lat: 48.86, Lng: 2.35, Radius: 50_000,
		Date: "2026-07-14",
	})
	if capturedURL == "" || !containsStr(capturedURL, "2026-07-14") {
		t.Errorf("expected date in query params, got: %q", capturedURL)
	}
}

// TestSearch_ImagesUpToThreePreferLandscape confirms pickImages returns at
// most 3 URLs, prioritising 16:9 entries at width ≥ 640 before falling back
// to other shapes, and dropping duplicates.
func TestSearch_ImagesUpToThreePreferLandscape(t *testing.T) {
	restore := mockNominatim(t, `{"address":{"city":"Paris","country_code":"fr"}}`)
	defer restore()

	body := `{"_embedded":{"events":[{
		"id":"tm-images","name":"Show","url":"https://example.com/tm",
		"images":[
			{"url":"https://img/portrait.jpg","width":640,"height":1024,"ratio":"3_4"},
			{"url":"https://img/large.jpg","width":1024,"height":576,"ratio":"16_9"},
			{"url":"https://img/medium.jpg","width":800,"height":450,"ratio":"16_9"},
			{"url":"https://img/large.jpg","width":1024,"height":576,"ratio":"16_9"},
			{"url":"https://img/wide.jpg","width":1280,"height":720,"ratio":"16_9"},
			{"url":"https://img/extra.jpg","width":2048,"height":1152,"ratio":"16_9"}
		],
		"dates":{"start":{"dateTime":"2026-06-15T19:00:00Z"},"end":{"dateTime":"2026-06-15T23:00:00Z"}},
		"_embedded":{"venues":[{"location":{"latitude":"48.87","longitude":"2.32"}}]}
	}]}}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	p := ticketmaster.NewWithURL(srv.URL)
	pois, err := p.Search(ctxWithTMKey("key"), types.SearchQuery{
		Mode: types.ModeRadius, Lat: 48.86, Lng: 2.35, Radius: 50_000,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(pois) != 1 {
		t.Fatalf("len pois = %d, want 1", len(pois))
	}
	want := []string{
		"https://img/large.jpg",
		"https://img/medium.jpg",
		"https://img/wide.jpg",
	}
	got := pois[0].Images
	if len(got) != len(want) {
		t.Fatalf("Images len = %d (%v), want %d (%v)", len(got), got, len(want), want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func containsStr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
