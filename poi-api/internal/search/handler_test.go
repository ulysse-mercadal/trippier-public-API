package search_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trippier/poi-api/internal/providers"
	"github.com/trippier/poi-api/internal/search"
	"github.com/trippier/poi-api/pkg/types"
	"go.uber.org/zap"
)

func newRouter(pp ...providers.Provider) *gin.Engine {
	svc := search.NewService(pp, 5*time.Second, zap.NewNop())
	h := search.NewHandler(svc)
	r := gin.New()
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	pois := r.Group("/pois")
	h.RegisterRoutes(pois)
	events := pois.Group("/events")
	h.RegisterEventRoutes(events)
	return r
}

func get(r *gin.Engine, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	return w
}

// ── /pois/search ─────────────────────────────────────────────────────────────

func TestHandlerSearch_OK(t *testing.T) {
	p := &mockProvider{
		name:  types.ProviderOverpass,
		modes: []types.SearchMode{types.ModeRadius},
		pois: []types.RawPoi{
			{ID: "o:1", Name: "Louvre", Type: types.TypeSee,
				Provider: types.ProviderOverpass, Coords: newCoords(48.86, 2.33)},
		},
	}
	w := get(newRouter(p), "/pois/search?mode=radius&lat=48.8566&lng=2.3522&radius=5000")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var body struct {
		Total int `json:"total"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Total != 1 {
		t.Errorf("total = %d, want 1", body.Total)
	}
}

func TestHandlerSearch_MissingMode(t *testing.T) {
	w := get(newRouter(), "/pois/search?lat=48.8566&lng=2.3522")
	// mode defaults to radius — lat+lng present so 200.
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200 (mode defaults to radius)", w.Code)
	}
}

func TestHandlerSearch_InvalidMode(t *testing.T) {
	w := get(newRouter(), "/pois/search?mode=invalid&lat=48.8566&lng=2.3522")
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestHandlerSearch_MissingLatLng(t *testing.T) {
	w := get(newRouter(), "/pois/search?mode=radius")
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestHandlerSearch_InvalidWeights(t *testing.T) {
	w := get(newRouter(), `/pois/search?mode=radius&lat=48.8566&lng=2.3522&weights={"see":2}`)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 for out-of-range weight", w.Code)
	}
}

func TestHandlerSearch_TypesAndWeightsMutuallyExclusive(t *testing.T) {
	w := get(newRouter(), `/pois/search?mode=radius&lat=48.8566&lng=2.3522&types=see&weights={"see":1}`)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 for types+weights", w.Code)
	}
}

func TestHandlerSearch_TypeFilter(t *testing.T) {
	p := &mockProvider{
		name:  types.ProviderOverpass,
		modes: []types.SearchMode{types.ModeRadius},
		pois: []types.RawPoi{
			{ID: "o:1", Name: "Louvre", Type: types.TypeSee, Provider: types.ProviderOverpass, Coords: newCoords(48.86, 2.33)},
			{ID: "o:2", Name: "Brasserie", Type: types.TypeEat, Provider: types.ProviderOverpass, Coords: newCoords(48.86, 2.34)},
		},
	}
	w := get(newRouter(p), "/pois/search?mode=radius&lat=48.8566&lng=2.3522&radius=5000&types=see")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var body struct {
		Total int `json:"total"`
	}
	json.NewDecoder(w.Body).Decode(&body)
	if body.Total != 1 {
		t.Errorf("type filter: total = %d, want 1 (only see)", body.Total)
	}
}

// ── /pois/search/slim ────────────────────────────────────────────────────────

func TestHandlerSearchSlim_OK(t *testing.T) {
	p := &mockProvider{
		name:  types.ProviderOverpass,
		modes: []types.SearchMode{types.ModeRadius},
		pois: []types.RawPoi{
			{ID: "o:1", Name: "Eiffel", Type: types.TypeSee,
				Provider: types.ProviderOverpass, Coords: newCoords(48.858, 2.294)},
		},
	}
	w := get(newRouter(p), "/pois/search/slim?mode=radius&lat=48.8566&lng=2.3522&radius=5000")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var body struct {
		Results []struct {
			Name string `json:"name"`
		} `json:"results"`
	}
	json.NewDecoder(w.Body).Decode(&body)
	if len(body.Results) != 1 || body.Results[0].Name != "Eiffel" {
		t.Errorf("slim results = %+v", body.Results)
	}
}

func TestHandlerSearchSlim_InvalidParams(t *testing.T) {
	w := get(newRouter(), "/pois/search/slim?mode=radius")
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// ── /pois/events ─────────────────────────────────────────────────────────────

func TestHandlerEvents_OK(t *testing.T) {
	p := &mockProvider{
		name:  types.ProviderWikipedia,
		modes: []types.SearchMode{types.ModeRadius},
		pois: []types.RawPoi{
			{ID: "wp:1", Name: "Jazz Fest", Type: types.TypeDo,
				Provider: types.ProviderWikipedia, Coords: newCoords(48.86, 2.35)},
		},
	}
	w := get(newRouter(p), "/pois/events?mode=radius&lat=48.8566&lng=2.3522&radius=5000")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestHandlerEvents_InvalidParams(t *testing.T) {
	w := get(newRouter(), "/pois/events?mode=radius")
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// ── /pois/providers ──────────────────────────────────────────────────────────

func TestHandlerProviders_OK(t *testing.T) {
	p := &mockProvider{name: types.ProviderOverpass, modes: []types.SearchMode{types.ModeRadius}}
	w := get(newRouter(p), "/pois/providers")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var statuses []types.ProviderStatus
	if err := json.NewDecoder(w.Body).Decode(&statuses); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(statuses) != 1 {
		t.Errorf("len(statuses) = %d, want 1", len(statuses))
	}
}

// ── SearchEvents ─────────────────────────────────────────────────────────────

func TestServiceSearchEvents(t *testing.T) {
	p := &mockProvider{
		name:  types.ProviderWikipedia,
		modes: []types.SearchMode{types.ModeRadius},
		pois: []types.RawPoi{
			{ID: "wp:1", Name: "Festival", Type: types.TypeDo,
				Provider: types.ProviderWikipedia, Coords: newCoords(48.86, 2.35)},
		},
	}
	svc := search.NewService([]providers.Provider{p}, 5*time.Second, zap.NewNop())
	result, err := svc.SearchEvents(context.Background(), types.SearchQuery{
		Mode: types.ModeRadius, Lat: 48.8566, Lng: 2.3522, Radius: 5000,
		Providers: []types.Provider{types.ProviderWikipedia},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("total = %d, want 1", result.Total)
	}
}

// ── ProvidersStatus ───────────────────────────────────────────────────────────

func TestServiceProvidersStatus(t *testing.T) {
	p := &mockProvider{name: types.ProviderOverpass, modes: []types.SearchMode{types.ModeRadius}}
	svc := search.NewService([]providers.Provider{p}, 5*time.Second, zap.NewNop())
	statuses := svc.ProvidersStatus(context.Background())
	if len(statuses) != 1 {
		t.Fatalf("len = %d, want 1", len(statuses))
	}
	if statuses[0].Name != types.ProviderOverpass {
		t.Errorf("name = %s, want overpass", statuses[0].Name)
	}
}
