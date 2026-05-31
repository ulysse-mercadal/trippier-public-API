package tilecache_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/trippier/poi-api/internal/tilecache"
	"github.com/trippier/poi-api/pkg/types"
)

// mockProvider records every Search call and returns a fixed set of POIs
// scattered around the query centre. Used to assert when the cache wrapper
// hits the upstream vs serves from Redis.
type mockProvider struct {
	name     types.Provider
	callCnt  atomic.Int32
	lastQ    types.SearchQuery
	response []types.RawPoi
}

func (m *mockProvider) Name() types.Provider               { return m.name }
func (m *mockProvider) SupportsMode(types.SearchMode) bool { return true }
func (m *mockProvider) Search(_ context.Context, q types.SearchQuery) ([]types.RawPoi, error) {
	m.callCnt.Add(1)
	m.lastQ = q
	return m.response, nil
}

func newMock(pois []types.RawPoi) *mockProvider {
	return &mockProvider{name: "mock", response: pois}
}

// newCacheHarness wires miniredis + the wrapper for one test.
func newCacheHarness(t *testing.T, m *mockProvider) (*tilecache.CachedProvider, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { rdb.Close() })
	cp := tilecache.NewCachedProvider(m, rdb, time.Hour, zap.NewNop())
	return cp, mr
}

// makePoi returns a RawPoi at the given coordinates with the given type.
func makePoi(id string, lat, lng float64, pt types.PoiType) types.RawPoi {
	return types.RawPoi{
		ID:       id,
		Name:     id,
		Type:     pt,
		Coords:   &types.Coordinates{Lat: lat, Lng: lng},
		Provider: "mock",
	}
}

// ── 1. First call: cold cache → one upstream call.

func TestCachedProvider_ColdCache_HitsUpstreamOnce(t *testing.T) {
	pois := []types.RawPoi{
		makePoi("p1", 48.8566, 2.3522, types.TypeSee),
		makePoi("p2", 48.8576, 2.3532, types.TypeEat),
	}
	m := newMock(pois)
	cp, _ := newCacheHarness(t, m)

	q := types.SearchQuery{
		Mode: types.ModeRadius, Lat: 48.8566, Lng: 2.3522, Radius: 1000,
		Types: []types.PoiType{types.TypeSee, types.TypeEat},
	}
	got, err := cp.Search(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 pois, got %d", len(got))
	}
	if m.callCnt.Load() != 1 {
		t.Errorf("expected 1 upstream call, got %d", m.callCnt.Load())
	}
}

// ── 2. Identical second call: warm cache → no extra upstream call.

func TestCachedProvider_WarmCache_NoUpstream(t *testing.T) {
	pois := []types.RawPoi{makePoi("p1", 48.8566, 2.3522, types.TypeSee)}
	m := newMock(pois)
	cp, _ := newCacheHarness(t, m)

	q := types.SearchQuery{
		Mode: types.ModeRadius, Lat: 48.8566, Lng: 2.3522, Radius: 1000,
		Types: []types.PoiType{types.TypeSee},
	}
	if _, err := cp.Search(context.Background(), q); err != nil {
		t.Fatal(err)
	}
	if _, err := cp.Search(context.Background(), q); err != nil {
		t.Fatal(err)
	}
	if m.callCnt.Load() != 1 {
		t.Errorf("expected 1 upstream call (second served from cache), got %d", m.callCnt.Load())
	}
}

// ── 3. Tiny radius variation inside the same tier: no refetch.

func TestCachedProvider_RadiusEpsilonSameTier_NoRefetch(t *testing.T) {
	pois := []types.RawPoi{makePoi("p1", 48.8566, 2.3522, types.TypeSee)}
	m := newMock(pois)
	cp, _ := newCacheHarness(t, m)

	base := types.SearchQuery{
		Mode: types.ModeRadius, Lat: 48.8566, Lng: 2.3522, Radius: 900,
		Types: []types.PoiType{types.TypeSee},
	}
	if _, err := cp.Search(context.Background(), base); err != nil {
		t.Fatal(err)
	}
	// 950m and 1000m both quantize to the 1000m tier — must hit cache.
	for _, r := range []int{950, 1000, 800} {
		q := base
		q.Radius = r
		if _, err := cp.Search(context.Background(), q); err != nil {
			t.Fatal(err)
		}
	}
	if m.callCnt.Load() != 1 {
		t.Errorf("expected 1 upstream call across same-tier variations, got %d", m.callCnt.Load())
	}
}

// ── 4. Zoom into a higher precision (smaller tier): refetch.

func TestCachedProvider_ZoomToFinerTier_Refetches(t *testing.T) {
	pois := []types.RawPoi{makePoi("p1", 48.8566, 2.3522, types.TypeSee)}
	m := newMock(pois)
	cp, _ := newCacheHarness(t, m)

	q1 := types.SearchQuery{
		Mode: types.ModeRadius, Lat: 48.8566, Lng: 2.3522, Radius: 2000,
		Types: []types.PoiType{types.TypeSee},
	}
	if _, err := cp.Search(context.Background(), q1); err != nil {
		t.Fatal(err)
	}
	// Same centre, zoom in to 500m — must MISS because best_radius=2000 > 500.
	q2 := q1
	q2.Radius = 500
	if _, err := cp.Search(context.Background(), q2); err != nil {
		t.Fatal(err)
	}
	if m.callCnt.Load() != 2 {
		t.Errorf("expected 2 upstream calls (zoom-in is a miss), got %d", m.callCnt.Load())
	}
}

// ── 5. Zoom out within an already-covered area: cache hit.

func TestCachedProvider_ZoomOutWithinCovered_NoRefetch(t *testing.T) {
	pois := []types.RawPoi{makePoi("p1", 48.8566, 2.3522, types.TypeSee)}
	m := newMock(pois)
	cp, _ := newCacheHarness(t, m)

	// Fetch with small radius first → best_radius=500.
	q1 := types.SearchQuery{
		Mode: types.ModeRadius, Lat: 48.8566, Lng: 2.3522, Radius: 500,
		Types: []types.PoiType{types.TypeSee},
	}
	if _, err := cp.Search(context.Background(), q1); err != nil {
		t.Fatal(err)
	}
	callsAfter := m.callCnt.Load()
	// Zoom out to 1km — tiles inside the 500m cover already have best_radius=500
	// (≤ 1000), so those tiles HIT. The outer ring is missing → 1 extra fetch.
	q2 := q1
	q2.Radius = 1000
	if _, err := cp.Search(context.Background(), q2); err != nil {
		t.Fatal(err)
	}
	if m.callCnt.Load()-callsAfter > 1 {
		t.Errorf("expected at most 1 extra upstream call for the ring, got %d", m.callCnt.Load()-callsAfter)
	}
}

// ── 6. Empty-set sentinel: a subsequent identical query does not refetch.

func TestCachedProvider_EmptyResult_SentinelPreventsRefetch(t *testing.T) {
	m := newMock(nil) // upstream returns zero POIs
	cp, _ := newCacheHarness(t, m)

	q := types.SearchQuery{
		Mode: types.ModeRadius, Lat: 48.8566, Lng: 2.3522, Radius: 500,
		Types: []types.PoiType{types.TypeSee},
	}
	if _, err := cp.Search(context.Background(), q); err != nil {
		t.Fatal(err)
	}
	if _, err := cp.Search(context.Background(), q); err != nil {
		t.Fatal(err)
	}
	if m.callCnt.Load() != 1 {
		t.Errorf("expected 1 upstream call (empty sentinel served), got %d", m.callCnt.Load())
	}
}

// ── 7. Non-radius mode: bypass cache entirely.

func TestCachedProvider_NonRadiusMode_Bypasses(t *testing.T) {
	pois := []types.RawPoi{makePoi("p1", 48.8566, 2.3522, types.TypeSee)}
	m := newMock(pois)
	cp, _ := newCacheHarness(t, m)

	q := types.SearchQuery{Mode: types.ModeDistrict, District: "Paris"}
	for i := 0; i < 3; i++ {
		if _, err := cp.Search(context.Background(), q); err != nil {
			t.Fatal(err)
		}
	}
	if m.callCnt.Load() != 3 {
		t.Errorf("expected 3 upstream calls (cache bypassed for district mode), got %d", m.callCnt.Load())
	}
}

// ── 8. Pan to overlapping area: only the missing strip is refetched, and
//      the upstream is centred on the missing tiles, not on the user's query.

func TestCachedProvider_OverlappingPan_FetchCenterShifts(t *testing.T) {
	pois := []types.RawPoi{makePoi("p1", 48.8566, 2.3522, types.TypeSee)}
	m := newMock(pois)
	cp, _ := newCacheHarness(t, m)

	// First fetch at Notre-Dame, radius 1km.
	q1 := types.SearchQuery{
		Mode: types.ModeRadius, Lat: 48.8566, Lng: 2.3522, Radius: 1000,
		Types: []types.PoiType{types.TypeSee},
	}
	if _, err := cp.Search(context.Background(), q1); err != nil {
		t.Fatal(err)
	}
	firstFetchCenterLat := m.lastQ.Lat
	firstFetchCenterLng := m.lastQ.Lng

	// Pan ~500m east, same radius. Some tiles overlap → upstream fetch should
	// be centred east of the original Notre-Dame coordinates.
	q2 := q1
	q2.Lng = 2.3590
	if _, err := cp.Search(context.Background(), q2); err != nil {
		t.Fatal(err)
	}
	if m.callCnt.Load() != 2 {
		t.Errorf("expected 2 upstream calls, got %d", m.callCnt.Load())
	}
	if m.lastQ.Lng <= firstFetchCenterLng {
		t.Errorf("expected fetch centre to shift east after pan, got %f (first %f)", m.lastQ.Lng, firstFetchCenterLng)
	}
	_ = firstFetchCenterLat
}
