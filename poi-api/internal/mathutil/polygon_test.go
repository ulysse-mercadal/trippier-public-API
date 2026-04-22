package mathutil_test

import (
	"testing"

	"github.com/trippier/poi-api/internal/mathutil"
)

func TestPointInPolygon(t *testing.T) {
	// Simple square: (0,0) → (0,1) → (1,1) → (1,0)
	square := [][2]float64{
		{0, 0}, {0, 1}, {1, 1}, {1, 0},
	}

	tests := []struct {
		name string
		lat  float64
		lng  float64
		want bool
	}{
		{"centre", 0.5, 0.5, true},
		{"outside right", 0.5, 2.0, false},
		{"outside top", 2.0, 0.5, false},
		{"outside bottom-left", -0.5, -0.5, false},
		{"near edge", 0.99, 0.99, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := mathutil.PointInPolygon(tc.lat, tc.lng, square)
			if got != tc.want {
				t.Errorf("PointInPolygon(%f, %f) = %v, want %v", tc.lat, tc.lng, got, tc.want)
			}
		})
	}
}
