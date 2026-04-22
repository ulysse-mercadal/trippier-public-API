package search_test

import (
	"testing"

	"github.com/trippier/poi-api/internal/search"
	"github.com/trippier/poi-api/pkg/types"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		q       types.SearchQuery
		wantErr bool
	}{
		{
			name:    "valid radius",
			q:       types.SearchQuery{Mode: types.ModeRadius, Lat: 48.85, Lng: 2.35, Radius: 5000},
			wantErr: false,
		},
		{
			name:    "radius missing lat/lng",
			q:       types.SearchQuery{Mode: types.ModeRadius},
			wantErr: true,
		},
		{
			name:    "radius lat out of range",
			q:       types.SearchQuery{Mode: types.ModeRadius, Lat: 200, Lng: 2.35, Radius: 1000},
			wantErr: true,
		},
		{
			name:    "radius too large",
			q:       types.SearchQuery{Mode: types.ModeRadius, Lat: 48.85, Lng: 2.35, Radius: 100_000},
			wantErr: true,
		},
		{
			name:    "valid polygon",
			q:       types.SearchQuery{Mode: types.ModePolygon, Polygon: "48.84 2.34 48.86 2.36"},
			wantErr: false,
		},
		{
			name:    "polygon missing polygon",
			q:       types.SearchQuery{Mode: types.ModePolygon},
			wantErr: true,
		},
		{
			name:    "valid district",
			q:       types.SearchQuery{Mode: types.ModeDistrict, District: "Montmartre"},
			wantErr: false,
		},
		{
			name:    "district missing district",
			q:       types.SearchQuery{Mode: types.ModeDistrict},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := search.Validate(tc.q)
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
