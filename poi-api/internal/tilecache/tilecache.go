// Package tilecache implements an H3-tile-based POI cache used by CachedProvider.
//
// The cache partitions space into fixed-size H3 r8 hexagons (~460m edge,
// ~0.7km²) and stores one Redis entry per (provider, tile, type, lang) slot.
// Each entry records the smallest fetch radius that has ever populated it
// ("best_radius") so the wrapper can decide whether the cached data is precise
// enough to satisfy an incoming query — a tile fetched as part of a 25km
// query is less trustworthy for a 500m zoom-in than a tile fetched as part of
// a 500m query.
package tilecache

import (
	"fmt"
	"math"

	"github.com/uber/h3-go/v4"
)

// Resolution is the fixed H3 resolution used for all cache tiles.
// r8 hexagons have an edge of ~461m and an area of ~0.74km² — small enough
// that a tile fetched as part of a wide query rarely contains the full set
// of POIs, large enough that nation-scale traffic stays under a few million
// cells in Redis.
const Resolution = 8

// edgeMeters is the average H3 r8 hexagon edge length, used to convert a
// fetch radius into a grid-disk ring count k. Hard-coded so callers don't
// pay a cgo round-trip per request.
const edgeMeters = 461.354684

// RadiusTiers lists the canonical fetch radii (meters). Any incoming radius
// is rounded up to the smallest tier ≥ r by Quantize, which means
// near-identical queries (radius 4900 vs 5100) share the same cache slot and
// near-identical fetches go upstream with the same parameters.
//
// The geometric ~×2 spacing keeps the number of distinct cache versions
// bounded: a tile can only ever hold one of seven best_radius values.
var RadiusTiers = []int{500, 1000, 2000, 5000, 10000, 25000, 50000}

// Tile is an opaque identifier for one H3 r8 hexagon.
type Tile = h3.Cell

// Quantize rounds r up to the smallest RadiusTiers entry ≥ r.
// Values above the maximum tier are capped at the maximum tier.
func Quantize(r int) int {
	for _, tier := range RadiusTiers {
		if r <= tier {
			return tier
		}
	}
	return RadiusTiers[len(RadiusTiers)-1]
}

// LatLngToTile returns the H3 r8 cell containing (lat, lng).
func LatLngToTile(lat, lng float64) (Tile, error) {
	return h3.LatLngToCell(h3.NewLatLng(lat, lng), Resolution)
}

// TileCenter returns the geographic centre of the given tile.
func TileCenter(t Tile) (lat, lng float64, err error) {
	ll, err := h3.CellToLatLng(t)
	if err != nil {
		return 0, 0, err
	}
	return ll.Lat, ll.Lng, nil
}

// TileCover returns the set of H3 r8 tiles intersecting the disk of radius
// radiusM around (lat, lng). The cover is conservative — it never misses an
// intersecting tile, but may include extras at the boundary (callers should
// post-filter against the exact disk if precise membership matters).
//
// k is derived as ceil(radiusM / edgeMeters) + 1; the +1 absorbs the fact
// that hexagon corners stick out further than edges and that the origin tile
// itself is at k=0.
func TileCover(lat, lng float64, radiusM int) ([]Tile, error) {
	origin, err := LatLngToTile(lat, lng)
	if err != nil {
		return nil, fmt.Errorf("tilecache: origin cell: %w", err)
	}
	k := int(math.Ceil(float64(radiusM)/edgeMeters)) + 1
	if k < 1 {
		k = 1
	}
	cells, err := h3.GridDisk(origin, k)
	if err != nil {
		return nil, fmt.Errorf("tilecache: grid disk: %w", err)
	}
	return cells, nil
}

// EnclosingCircle returns a (centre, radius) pair such that every tile in
// `tiles` is fully contained in the resulting disk. The centre is the
// arithmetic mean of tile centres and the radius is the farthest tile-centre
// distance plus one edge-length margin so the disk swallows the whole
// hexagon (not just its centre).
//
// This is the approximate centroid method — not the optimal Welzl
// smallest-enclosing-circle. For compact clusters of cache-miss tiles the
// extra fetch surface is usually under ~10%, which is a fair trade for an
// O(n) algorithm with no external dependency.
func EnclosingCircle(tiles []Tile) (centerLat, centerLng float64, radiusM int, err error) {
	if len(tiles) == 0 {
		return 0, 0, 0, fmt.Errorf("tilecache: enclosing circle on empty tile set")
	}
	if len(tiles) == 1 {
		lat, lng, err := TileCenter(tiles[0])
		if err != nil {
			return 0, 0, 0, err
		}
		return lat, lng, int(math.Ceil(edgeMeters)), nil
	}

	var sumLat, sumLng float64
	centers := make([]h3.LatLng, len(tiles))
	for i, t := range tiles {
		lat, lng, err := TileCenter(t)
		if err != nil {
			return 0, 0, 0, err
		}
		centers[i] = h3.NewLatLng(lat, lng)
		sumLat += lat
		sumLng += lng
	}
	cLat := sumLat / float64(len(tiles))
	cLng := sumLng / float64(len(tiles))
	cLL := h3.NewLatLng(cLat, cLng)

	maxDist := 0.0
	for _, ll := range centers {
		if d := h3.GreatCircleDistanceM(cLL, ll); d > maxDist {
			maxDist = d
		}
	}
	return cLat, cLng, int(math.Ceil(maxDist + edgeMeters)), nil
}

// TileOf returns the H3 r8 cell at the given coordinates, suitable for
// attributing a POI to its tile bucket. Returns the zero Tile on error so
// callers can short-circuit invalid POIs without an extra check.
func TileOf(lat, lng float64) Tile {
	c, err := LatLngToTile(lat, lng)
	if err != nil {
		return 0
	}
	return c
}

// Key builds the Redis key for one (provider, tile, type, lang) slot.
// Tile is serialised as its hex H3 index so keys are inspectable in redis-cli.
func Key(provider, tileHex, poiType, lang string) string {
	return fmt.Sprintf("poi:tile:%s:%s:%s:%s", provider, tileHex, poiType, lang)
}

// TileHex returns the hex-string representation of a tile, used in cache keys.
func TileHex(t Tile) string {
	return h3.CellToString(t)
}
