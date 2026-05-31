package eventbrite_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/trippier/poi-api/internal/byok"
	"github.com/trippier/poi-api/internal/providers/eventbrite"
	"github.com/trippier/poi-api/pkg/types"
)

const sampleResponse = `{
  "events": [
    {
      "id": "eb1",
      "url": "https://example.com/eb1",
      "name":        {"text": "Tech Meetup"},
      "description": {"text": "A tech event"},
      "start": {"utc": "2026-06-20T18:00:00Z"},
      "end":   {"utc": "2026-06-20T21:00:00Z"},
      "logo":  {"url": "https://img.example.com/eb1.jpg"},
      "venue": {"latitude": "48.8566", "longitude": "2.3522"}
    },
    {
      "id": "eb2",
      "name":        {"text": ""},
      "description": {"text": ""},
      "start": {"utc": ""},
      "end":   {"utc": ""},
      "venue": {"latitude": "not-a-float", "longitude": "bad"}
    },
    {
      "id": "eb3",
      "name":        {"text": "Startup Fair"},
      "description": {"text": "Networking"},
      "start": {"utc": "2026-07-01T09:00:00Z"},
      "end":   {"utc": "2026-07-01T18:00:00Z"},
      "venue": null
    }
  ]
}`

func ctxWithEBToken(token string) context.Context {
	return byok.WithProviderKey(context.Background(), types.ProviderEventbrite, token)
}

// ── No token → silently skipped ───────────────────────────────────────────────

func TestSearch_NoToken_ReturnsNil(t *testing.T) {
	p := eventbrite.New()
	pois, err := p.Search(context.Background(), types.SearchQuery{
		Mode: types.ModeRadius, Lat: 48.86, Lng: 2.35, Radius: 50_000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pois != nil {
		t.Errorf("expected nil pois without token, got %d", len(pois))
	}
}

// ── Name / IsByok / SupportsMode ─────────────────────────────────────────────

func TestName(t *testing.T) {
	p := eventbrite.New()
	if p.Name() != types.ProviderEventbrite {
		t.Errorf("Name() = %q, want %q", p.Name(), types.ProviderEventbrite)
	}
}

func TestIsByok(t *testing.T) {
	p := eventbrite.New()
	if !p.IsByok() {
		t.Error("IsByok() = false, want true")
	}
}

func TestSupportsMode(t *testing.T) {
	p := eventbrite.New()
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

// ── Search with token + mock server ───────────────────────────────────────────

func TestSearch_WithToken_ParsesEvents(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sampleResponse))
	}))
	defer srv.Close()

	p := eventbrite.NewWithURL(srv.URL)
	pois, err := p.Search(ctxWithEBToken("test-token"), types.SearchQuery{
		Mode: types.ModeRadius, Lat: 48.86, Lng: 2.35, Radius: 50_000,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	// eb1 valid. eb2 has empty name → filtered. eb3 has nil venue → no coords but still included.
	if len(pois) == 0 {
		t.Fatal("expected at least one event")
	}
	if pois[0].Name != "Tech Meetup" {
		t.Errorf("first event name = %q, want %q", pois[0].Name, "Tech Meetup")
	}
	if pois[0].Type != types.TypeEvent {
		t.Errorf("first event type = %q, want %q", pois[0].Type, types.TypeEvent)
	}
	if pois[0].SourceURL != "https://example.com/eb1" {
		t.Errorf("first event SourceURL = %q, want %q", pois[0].SourceURL, "https://example.com/eb1")
	}
}

func TestSearch_ChecksAuthHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"events":[]}`))
	}))
	defer srv.Close()

	p := eventbrite.NewWithURL(srv.URL)
	p.Search(ctxWithEBToken("my-secret-token"), types.SearchQuery{ //nolint:errcheck
		Mode: types.ModeRadius, Lat: 48.86, Lng: 2.35, Radius: 50_000,
	})
	if gotAuth != "Bearer my-secret-token" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer my-secret-token")
	}
}

func TestSearch_NonOKStatus_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	p := eventbrite.NewWithURL(srv.URL)
	_, err := p.Search(ctxWithEBToken("bad"), types.SearchQuery{
		Mode: types.ModeRadius, Lat: 48.86, Lng: 2.35, Radius: 50_000,
	})
	if err == nil {
		t.Error("expected error for 401 response")
	}
}

func TestSearch_InvalidJSON_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`not-json`))
	}))
	defer srv.Close()

	p := eventbrite.NewWithURL(srv.URL)
	_, err := p.Search(ctxWithEBToken("key"), types.SearchQuery{
		Mode: types.ModeRadius, Lat: 48.86, Lng: 2.35, Radius: 50_000,
	})
	if err == nil {
		t.Error("expected error for invalid JSON response")
	}
}

func TestSearch_EmptyEvents_ReturnsEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"events":[]}`))
	}))
	defer srv.Close()

	p := eventbrite.NewWithURL(srv.URL)
	pois, err := p.Search(ctxWithEBToken("key"), types.SearchQuery{
		Mode: types.ModeRadius, Lat: 48.86, Lng: 2.35, Radius: 50_000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pois) != 0 {
		t.Errorf("expected 0 events, got %d", len(pois))
	}
}

func TestSearch_RadiusClamped(t *testing.T) {
	var capturedQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"events":[]}`))
	}))
	defer srv.Close()

	p := eventbrite.NewWithURL(srv.URL)
	// radius = 1000 m → should be clamped to minRadiusKm (50 km).
	p.Search(ctxWithEBToken("key"), types.SearchQuery{ //nolint:errcheck
		Mode: types.ModeRadius, Lat: 48.86, Lng: 2.35, Radius: 1000,
	})
	if !containsStr(capturedQuery, "50km") {
		t.Errorf("radius not clamped to 50km, got query: %q", capturedQuery)
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
