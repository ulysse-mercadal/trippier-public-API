package mathutil

// PointInPolygon reports whether the point (lat, lng) lies inside the given polygon
// using the ray-casting algorithm. The polygon is expressed as a slice of [lat, lng] pairs.
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
