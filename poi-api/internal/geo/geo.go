// Package geo provides geographic filtering operations on POI collections.
// Pure mathematical primitives are delegated to the mathutil package.
package geo

import (
	"github.com/trippier/poi-api/internal/mathutil"
	"github.com/trippier/poi-api/pkg/types"
)

// SetDistances annotates each RawPoi with its distance in meters from (lat, lng).
// POIs with approximate or missing coordinates are skipped.
func SetDistances(pois []types.RawPoi, lat, lng float64) []types.RawPoi {
	for i, p := range pois {
		if p.Coords != nil && !p.Coords.Approximate {
			pois[i].Distance = mathutil.Haversine(lat, lng, p.Coords.Lat, p.Coords.Lng)
		}
	}
	return pois
}

// FilterByRadius returns only the POIs within radiusMeters of (lat, lng).
// POIs without precise coordinates are kept (they are zone-based results).
func FilterByRadius(pois []types.RawPoi, lat, lng, radiusMeters float64) []types.RawPoi {
	result := pois[:0]
	for _, p := range pois {
		if p.Coords == nil || p.Coords.Approximate {
			result = append(result, p)
			continue
		}
		if p.Distance <= radiusMeters {
			result = append(result, p)
		}
	}
	return result
}

// FilterByPolygon returns only the POIs whose coordinates fall within the given polygon.
// POIs without precise coordinates are kept.
func FilterByPolygon(pois []types.RawPoi, polygon [][2]float64) []types.RawPoi {
	result := pois[:0]
	for _, p := range pois {
		if p.Coords == nil || p.Coords.Approximate || mathutil.PointInPolygon(p.Coords.Lat, p.Coords.Lng, polygon) {
			result = append(result, p)
		}
	}
	return result
}
