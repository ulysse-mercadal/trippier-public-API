package mathutil_test

import (
	"math"
	"testing"

	"github.com/trippier/poi-api/internal/mathutil"
)

func TestHaversine(t *testing.T) {
	tests := []struct {
		name    string
		lat1    float64
		lng1    float64
		lat2    float64
		lng2    float64
		wantApx float64 // expected distance in metres (±1 %)
	}{
		{
			name: "same point",
			lat1: 48.8566, lng1: 2.3522,
			lat2: 48.8566, lng2: 2.3522,
			wantApx: 0,
		},
		{
			name: "Paris → London (~340 km)",
			lat1: 48.8566, lng1: 2.3522,
			lat2: 51.5074, lng2: -0.1278,
			wantApx: 340_600,
		},
		{
			name: "short distance 50 m",
			lat1: 48.85680, lng1: 2.35200,
			lat2: 48.85725, lng2: 2.35200,
			wantApx: 50,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := mathutil.Haversine(tc.lat1, tc.lng1, tc.lat2, tc.lng2)
			if tc.wantApx == 0 {
				if got != 0 {
					t.Errorf("want 0, got %f", got)
				}
				return
			}
			tol := tc.wantApx * 0.01 // 1 %
			if math.Abs(got-tc.wantApx) > tol {
				t.Errorf("Haversine = %f, want ~%f (±1%%)", got, tc.wantApx)
			}
		})
	}
}
