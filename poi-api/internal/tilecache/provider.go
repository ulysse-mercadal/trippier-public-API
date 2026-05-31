package tilecache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/trippier/poi-api/internal/providers"
	"github.com/trippier/poi-api/pkg/types"
)

// entry is the JSON-serialised cache value for one (provider, tile, type, lang) slot.
// An empty Pois slice with a non-zero BestRadiusM is the "fetched-but-empty" sentinel.
type entry struct {
	Pois        []types.RawPoi `json:"pois"`
	BestRadiusM int            `json:"best_radius"`
	FetchedAt   int64          `json:"fetched_at"`
}

// defaultCacheTypes is the type universe used when the caller doesn't specify
// q.Types. We need a finite set so cache keys are deterministic.
var defaultCacheTypes = []types.PoiType{
	types.TypeSee, types.TypeEat, types.TypeDrink,
	types.TypeDo, types.TypeBuy, types.TypeSleep, types.TypeGeneric,
}

// CachedProvider wraps a providers.Provider with an H3-tile Redis cache.
//
// Strategy: for each radius-mode query the wrapper quantizes the radius to a
// canonical tier, computes the tile cover, looks up every (tile, type) slot
// in Redis, and only calls the inner provider for the tiles it cannot serve
// from cache. The upstream fetch is centered on those missing tiles via an
// enclosing-circle approximation, so a small zoom or pan does not trigger a
// full re-fetch of the original query area.
//
// Tiles already in cache are never overwritten — if a fresh fetch happens to
// return POIs in a cached tile, those POIs are discarded for the cache write
// (the cached entry is presumed at least as precise). The fresh POIs returned
// to the caller are filtered the same way to avoid duplicates with hits.
type CachedProvider struct {
	inner providers.Provider
	rdb   *redis.Client
	ttl   time.Duration
	log   *zap.Logger
}

// NewCachedProvider returns a wrapper backed by the given Redis client.
// ttl is the per-entry expiry; entries also become stale as soon as the
// inner provider's underlying data changes (no invalidation hook).
func NewCachedProvider(inner providers.Provider, rdb *redis.Client, ttl time.Duration, log *zap.Logger) *CachedProvider {
	return &CachedProvider{inner: inner, rdb: rdb, ttl: ttl, log: log}
}

// Name implements providers.Provider — delegates to the inner provider so the
// wrapper is transparent to upstream routing and BYOK detection.
func (c *CachedProvider) Name() types.Provider { return c.inner.Name() }

// SupportsMode implements providers.Provider — delegates to the inner provider.
func (c *CachedProvider) SupportsMode(mode types.SearchMode) bool { return c.inner.SupportsMode(mode) }

// IsByok forwards the inner provider's BYOK status when applicable, so the
// service layer's ByokProvider type-assertion keeps working through the wrapper.
func (c *CachedProvider) IsByok() bool {
	if bp, ok := c.inner.(providers.ByokProvider); ok {
		return bp.IsByok()
	}
	return false
}

// Search implements providers.Provider with the tile-cache flow described on
// the type doc. Non-radius queries bypass the cache entirely.
func (c *CachedProvider) Search(ctx context.Context, q types.SearchQuery) ([]types.RawPoi, error) {
	if q.Mode != types.ModeRadius || (q.Lat == 0 && q.Lng == 0) {
		return c.inner.Search(ctx, q)
	}

	effectiveR := Quantize(q.Radius)
	tiles, err := TileCover(q.Lat, q.Lng, effectiveR)
	if err != nil {
		c.log.Warn("tilecache: cover failed, bypassing cache", zap.Error(err))
		return c.inner.Search(ctx, q)
	}

	poiTypes := q.Types
	if len(poiTypes) == 0 {
		poiTypes = defaultCacheTypes
	}

	providerName := string(c.inner.Name())
	keys, meta := c.buildKeys(providerName, tiles, poiTypes, q.Lang)

	hitPois, missingTiles := c.readCache(ctx, keys, meta, effectiveR)

	if len(missingTiles) == 0 {
		return hitPois, nil
	}

	freshPois, _, err := c.fetchMissing(ctx, q, missingTiles, poiTypes)
	if err != nil {
		return nil, err
	}

	// best_radius stored = the user's quantized radius, not the upstream fetch
	// radius. The fetch radius can be slightly larger (enclosing-circle margin)
	// or smaller (a single far-off missing tile), but the contract we want to
	// expose to future queries is: "this tile is trustworthy for any query at
	// precision >= effectiveR". Storing effectiveR preserves that contract and
	// avoids a pathological MISS when the next identical query is issued.
	c.writeCache(ctx, providerName, missingTiles, poiTypes, freshPois, effectiveR, q.Lang)

	// Drop POIs that landed in cached tiles to avoid double-counting with hits.
	keptFresh := freshPois[:0]
	for _, p := range freshPois {
		if p.Coords == nil {
			continue
		}
		t := TileOf(p.Coords.Lat, p.Coords.Lng)
		if _, ok := missingTiles[t]; ok {
			keptFresh = append(keptFresh, p)
		}
	}

	return append(hitPois, keptFresh...), nil
}

// keyMeta records which (tile, type) pair a given Redis key corresponds to,
// so the partition step can attribute MGet results back to their tile.
type keyMeta struct {
	Tile Tile
	Type types.PoiType
}

// buildKeys returns the ordered list of Redis keys to probe along with their
// (tile, type) metadata. Order is preserved between the two slices so the
// MGet response can be zipped back with meta[i].
func (c *CachedProvider) buildKeys(provider string, tiles []Tile, poiTypes []types.PoiType, lang string) ([]string, []keyMeta) {
	n := len(tiles) * len(poiTypes)
	keys := make([]string, 0, n)
	meta := make([]keyMeta, 0, n)
	for _, t := range tiles {
		hex := TileHex(t)
		for _, pt := range poiTypes {
			keys = append(keys, Key(provider, hex, string(pt), lang))
			meta = append(meta, keyMeta{Tile: t, Type: pt})
		}
	}
	return keys, meta
}

// readCache fans the keys through MGet and partitions the result into the
// already-cached POIs and the set of tiles needing an upstream fetch.
//
// A tile is considered "missing" if any of its (tile, type) slots is absent
// or has a BestRadiusM strictly greater than effectiveR. A single weak slot
// promotes the whole tile to missing so the next upstream fetch is centred
// correctly — partial-coverage tiles would otherwise distort the enclosing
// circle.
func (c *CachedProvider) readCache(ctx context.Context, keys []string, meta []keyMeta, effectiveR int) ([]types.RawPoi, map[Tile]struct{}) {
	missing := make(map[Tile]struct{})
	if len(keys) == 0 {
		return nil, missing
	}

	vals, err := c.rdb.MGet(ctx, keys...).Result()
	if err != nil {
		c.log.Warn("tilecache: MGet failed, treating as miss", zap.Error(err))
		for _, m := range meta {
			missing[m.Tile] = struct{}{}
		}
		return nil, missing
	}

	var hits []types.RawPoi
	for i, v := range vals {
		if v == nil {
			missing[meta[i].Tile] = struct{}{}
			continue
		}
		raw, ok := v.(string)
		if !ok {
			missing[meta[i].Tile] = struct{}{}
			continue
		}
		var e entry
		if err := json.Unmarshal([]byte(raw), &e); err != nil {
			missing[meta[i].Tile] = struct{}{}
			continue
		}
		if e.BestRadiusM > effectiveR {
			missing[meta[i].Tile] = struct{}{}
			continue
		}
		hits = append(hits, e.Pois...)
	}
	return hits, missing
}

// fetchMissing computes the enclosing-circle fetch parameters for the missing
// tile set and runs one inner.Search call against the upstream provider.
// The fetch radius is quantized to the same tier ladder as the request radius
// so future requests can hit this fetch's results when their tier matches.
func (c *CachedProvider) fetchMissing(ctx context.Context, q types.SearchQuery, missingTiles map[Tile]struct{}, poiTypes []types.PoiType) ([]types.RawPoi, int, error) {
	missingList := make([]Tile, 0, len(missingTiles))
	for t := range missingTiles {
		missingList = append(missingList, t)
	}

	fetchLat, fetchLng, rawR, err := EnclosingCircle(missingList)
	if err != nil {
		return nil, 0, fmt.Errorf("tilecache: enclosing circle: %w", err)
	}
	fetchR := Quantize(rawR)

	fetchQuery := q
	fetchQuery.Lat = fetchLat
	fetchQuery.Lng = fetchLng
	fetchQuery.Radius = fetchR
	fetchQuery.Types = poiTypes

	pois, err := c.inner.Search(ctx, fetchQuery)
	if err != nil {
		return nil, 0, err
	}
	return pois, fetchR, nil
}

// writeCache pipelines one SET per (missing_tile, type) slot. Empty buckets
// get an empty-POI sentinel so a subsequent identical query short-circuits
// without re-hitting the upstream API.
//
// Important: only tiles in missingTiles are written. POIs that the upstream
// returned for tiles already in cache are silently dropped here — the
// pre-existing cache entry is presumed at least as precise (its best_radius
// is ≤ effectiveR by the readCache contract).
func (c *CachedProvider) writeCache(ctx context.Context, provider string, missingTiles map[Tile]struct{}, poiTypes []types.PoiType, freshPois []types.RawPoi, bestRadius int, lang string) {
	buckets := make(map[Tile]map[types.PoiType][]types.RawPoi, len(missingTiles))
	for _, p := range freshPois {
		if p.Coords == nil {
			continue
		}
		t := TileOf(p.Coords.Lat, p.Coords.Lng)
		if t == 0 {
			continue
		}
		if _, ok := missingTiles[t]; !ok {
			continue
		}
		if buckets[t] == nil {
			buckets[t] = make(map[types.PoiType][]types.RawPoi)
		}
		buckets[t][p.Type] = append(buckets[t][p.Type], p)
	}

	pipe := c.rdb.Pipeline()
	now := time.Now().Unix()
	for t := range missingTiles {
		hex := TileHex(t)
		for _, pt := range poiTypes {
			pois := buckets[t][pt]
			if pois == nil {
				pois = []types.RawPoi{}
			}
			e := entry{Pois: pois, BestRadiusM: bestRadius, FetchedAt: now}
			data, err := json.Marshal(e)
			if err != nil {
				continue
			}
			pipe.Set(ctx, Key(provider, hex, string(pt), lang), data, c.ttl)
		}
	}
	if _, err := pipe.Exec(ctx); err != nil {
		c.log.Warn("tilecache: pipeline exec failed", zap.Error(err))
	}
}
