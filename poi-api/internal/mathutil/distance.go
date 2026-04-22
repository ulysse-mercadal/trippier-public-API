// Package mathutil contains all pure mathematical functions used across the poi-api pipeline:
// geographic distance, polygon containment, and string similarity.
package mathutil

import "math"

// Haversine returns the great-circle distance in meters between two WGS-84 coordinates.
func Haversine(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadius = 6_371_000.0
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	return earthRadius * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

// BoundingBox returns the (minLat, minLng, maxLat, maxLng) bounding box
// for a center point and a radius in meters.
func BoundingBox(lat, lng, radiusMeters float64) (minLat, minLng, maxLat, maxLng float64) {
	const earthRadius = 6_371_000.0
	deltaLat := (radiusMeters / earthRadius) * (180 / math.Pi)
	deltaLng := deltaLat / math.Cos(lat*math.Pi/180)
	return lat - deltaLat, lng - deltaLng, lat + deltaLat, lng + deltaLng
}
