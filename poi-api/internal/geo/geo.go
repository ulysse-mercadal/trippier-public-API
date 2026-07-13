// Package geo provides geographic filtering operations on POI collections.
// Pure mathematical primitives are delegated to the mathutil package.
package geo

import (
	"github.com/trippier/poi-api/internal/mathutil"
	"github.com/trippier/poi-api/pkg/types"
)

// SetDistances annotates each RawPoi in pois with its distance in meters
// from (lat, lng); POIs with approximate or missing coordinates are
// skipped. It returns the same slice with distances set.
func SetDistances(pois []types.RawPoi, lat, lng float64) []types.RawPoi {
	for i, p := range pois {
		if p.Coords != nil && !p.Coords.Approximate {
			pois[i].Distance = mathutil.Haversine(lat, lng, p.Coords.Lat, p.Coords.Lng)
		}
	}
	return pois
}

// FilterByRadius keeps only the POIs in pois within radiusMeters of
// (lat, lng); POIs without precise coordinates are kept (zone-based
// results). It returns the filtered slice.
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

// FilterByPolygon keeps only the POIs in pois whose coordinates fall
// within polygon; POIs without precise coordinates are kept. It returns
// the filtered slice.
func FilterByPolygon(pois []types.RawPoi, polygon [][2]float64) []types.RawPoi {
	result := pois[:0]
	for _, p := range pois {
		if p.Coords == nil || p.Coords.Approximate || mathutil.PointInPolygon(p.Coords.Lat, p.Coords.Lng, polygon) {
			result = append(result, p)
		}
	}
	return result
}
