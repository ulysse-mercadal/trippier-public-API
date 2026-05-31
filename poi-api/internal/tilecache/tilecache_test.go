package tilecache_test

import (
	"math"
	"testing"

	"github.com/trippier/poi-api/internal/tilecache"
)

func TestQuantize_RoundsUpToTier(t *testing.T) {
	cases := []struct {
		in   int
		want int
	}{
		{1, 500},
		{499, 500},
		{500, 500},
		{501, 1000},
		{900, 1000},
		{1000, 1000},
		{1500, 2000},
		{4999, 5000},
		{5001, 10000},
		{25000, 25000},
		{30000, 50000},
		{49999, 50000},
		{50000, 50000},
		// Above the max tier: clamped to max.
		{100000, 50000},
	}
	for _, c := range cases {
		if got := tilecache.Quantize(c.in); got != c.want {
			t.Errorf("Quantize(%d) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestTileCover_OriginInCover(t *testing.T) {
	lat, lng := 48.8566, 2.3522 // Notre-Dame
	tiles, err := tilecache.TileCover(lat, lng, 500)
	if err != nil {
		t.Fatalf("TileCover: %v", err)
	}
	if len(tiles) == 0 {
		t.Fatal("expected at least one tile")
	}
	origin, err := tilecache.LatLngToTile(lat, lng)
	if err != nil {
		t.Fatalf("LatLngToTile: %v", err)
	}
	for _, tile := range tiles {
		if tile == origin {
			return
		}
	}
	t.Errorf("origin tile not in cover")
}

func TestTileCover_RadiusGrowsCoverage(t *testing.T) {
	lat, lng := 48.8566, 2.3522
	small, err := tilecache.TileCover(lat, lng, 500)
	if err != nil {
		t.Fatal(err)
	}
	large, err := tilecache.TileCover(lat, lng, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if len(large) <= len(small) {
		t.Errorf("expected larger radius to cover more tiles, got %d vs %d", len(large), len(small))
	}
}

func TestEnclosingCircle_ContainsAllTiles(t *testing.T) {
	lat, lng := 48.8566, 2.3522
	tiles, err := tilecache.TileCover(lat, lng, 1000)
	if err != nil {
		t.Fatal(err)
	}
	cLat, cLng, r, err := tilecache.EnclosingCircle(tiles)
	if err != nil {
		t.Fatal(err)
	}
	if r <= 0 {
		t.Fatalf("expected positive radius, got %d", r)
	}
	// Every tile centre must be within r of (cLat, cLng).
	for _, tile := range tiles {
		tlat, tlng, err := tilecache.TileCenter(tile)
		if err != nil {
			t.Fatal(err)
		}
		d := haversine(cLat, cLng, tlat, tlng)
		if d > float64(r) {
			t.Errorf("tile centre at distance %.1f > radius %d", d, r)
		}
	}
}

func TestEnclosingCircle_SingleTile(t *testing.T) {
	tile, err := tilecache.LatLngToTile(48.8566, 2.3522)
	if err != nil {
		t.Fatal(err)
	}
	cLat, cLng, r, err := tilecache.EnclosingCircle([]tilecache.Tile{tile})
	if err != nil {
		t.Fatal(err)
	}
	tlat, tlng, _ := tilecache.TileCenter(tile)
	if cLat != tlat || cLng != tlng {
		t.Errorf("single-tile centre should equal tile centre")
	}
	if r <= 0 {
		t.Errorf("single-tile radius should be > 0")
	}
}

func TestEnclosingCircle_EmptyFails(t *testing.T) {
	if _, _, _, err := tilecache.EnclosingCircle(nil); err == nil {
		t.Error("expected error on empty tile set")
	}
}

func TestTileOf_RoundTrip(t *testing.T) {
	lat, lng := 48.8566, 2.3522
	t1 := tilecache.TileOf(lat, lng)
	t2, err := tilecache.LatLngToTile(lat, lng)
	if err != nil {
		t.Fatal(err)
	}
	if t1 != t2 {
		t.Errorf("TileOf and LatLngToTile disagree: %v vs %v", t1, t2)
	}
}

// haversine duplicates the formula used inside tilecache for verifying that
// EnclosingCircle's radius really covers every tile centre.
func haversine(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6_371_000.0
	rad := math.Pi / 180.0
	dLat := (lat2 - lat1) * rad
	dLng := (lng2 - lng1) * rad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*rad)*math.Cos(lat2*rad)*math.Sin(dLng/2)*math.Sin(dLng/2)
	return 2 * R * math.Asin(math.Sqrt(a))
}
