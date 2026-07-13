// Package mathutil provides small geometric helper functions.
package mathutil

// PointInPolygon reports whether the point at (lat, lng) lies inside polygon,
// a list of vertices given as [lat, lng] pairs, using the ray-casting
// algorithm. It returns true if the point is inside the polygon.
func PointInPolygon(lat, lng float64, polygon [][2]float64) bool {
	n := len(polygon)
	inside := false
	j := n - 1
	for i := 0; i < n; i++ {
		yi, xi := polygon[i][0], polygon[i][1]
		yj, xj := polygon[j][0], polygon[j][1]
		if ((yi > lat) != (yj > lat)) && (lng < (xj-xi)*(lat-yi)/(yj-yi)+xi) {
			inside = !inside
		}
		j = i
	}
	return inside
}
