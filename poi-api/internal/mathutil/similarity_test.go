package mathutil_test

import (
	"testing"

	"github.com/trippier/poi-api/internal/mathutil"
)

func TestJaroWinkler(t *testing.T) {
	tests := []struct {
		s1, s2  string
		wantMin float64
		wantMax float64
	}{
		{"", "", 1.0, 1.0},
		{"abc", "", 0.0, 0.0},
		{"", "abc", 0.0, 0.0},
		{"louvre", "louvre", 1.0, 1.0},
		{"louvre", "louvr", 0.95, 1.0},  // very close
		{"cafe", "restaurant", 0.0, 0.7}, // clearly different
		{"musée", "musee", 0.85, 1.0},    // accent difference (byte-level comparison)
	}

	for _, tc := range tests {
		got := mathutil.JaroWinkler(tc.s1, tc.s2)
		if got < tc.wantMin || got > tc.wantMax {
			t.Errorf("JaroWinkler(%q, %q) = %f, want [%f, %f]",
				tc.s1, tc.s2, got, tc.wantMin, tc.wantMax)
		}
	}
}
