package mathutil_test

import (
	"math"
	"testing"

	"github.com/trippier/poi-api/internal/mathutil"
)

func TestHaversine(t *testing.T) {
	tests := []struct {
		name     string
		lat1     float64
		lng1     float64
		lat2     float64
		lng2     float64
		wantApx  float64 // expected distance in metres (±1 %)
	}{
		{
			name:    "same point",
			lat1:    48.8566, lng1: 2.3522,
			lat2:    48.8566, lng2: 2.3522,
			wantApx: 0,
		},
		{
			name:    "Paris → London (~340 km)",
			lat1:    48.8566, lng1: 2.3522,
			lat2:    51.5074, lng2: -0.1278,
			wantApx: 340_600,
		},
		{
			name:    "short distance 50 m",
			lat1:    48.85680, lng1: 2.35200,
			lat2:    48.85725, lng2: 2.35200,
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

func TestBoundingBox(t *testing.T) {
	minLat, minLng, maxLat, maxLng := mathutil.BoundingBox(48.8566, 2.3522, 1000)

	if minLat >= 48.8566 || maxLat <= 48.8566 {
		t.Errorf("lat %f not inside [%f, %f]", 48.8566, minLat, maxLat)
	}
	if minLng >= 2.3522 || maxLng <= 2.3522 {
		t.Errorf("lng %f not inside [%f, %f]", 2.3522, minLng, maxLng)
	}

	// Width should be symmetric around the centre.
	latSpan := maxLat - minLat
	lngSpan := maxLng - minLng
	if latSpan <= 0 || lngSpan <= 0 {
		t.Errorf("non-positive span: lat=%f lng=%f", latSpan, lngSpan)
	}
}
