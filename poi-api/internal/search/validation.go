package search

import (
	"errors"
	"fmt"

	"github.com/trippier/poi-api/pkg/types"
)

// Validate checks that a SearchQuery is semantically valid for its mode.
// It is called after query binding so all fields are already populated.
func Validate(q types.SearchQuery) error {
	switch q.Mode {
	case types.ModeRadius:
		return validateRadius(q)
	case types.ModePolygon:
		return validatePolygon(q)
	case types.ModeDistrict:
		return validateDistrict(q)
	default:
		return fmt.Errorf("unknown mode %q", q.Mode)
	}
}

func validateRadius(q types.SearchQuery) error {
	if q.Lat == 0 && q.Lng == 0 {
		return errors.New("mode=radius requires lat and lng")
	}
	if q.Lat < -90 || q.Lat > 90 {
		return fmt.Errorf("lat %f out of range [-90, 90]", q.Lat)
	}
	if q.Lng < -180 || q.Lng > 180 {
		return fmt.Errorf("lng %f out of range [-180, 180]", q.Lng)
	}
	if q.Radius < 0 {
		return errors.New("radius must be non-negative")
	}
	if q.Radius > 50_000 {
		return fmt.Errorf("radius %d exceeds maximum of 50 000 m", q.Radius)
	}
	return nil
}

func validatePolygon(q types.SearchQuery) error {
	if q.Polygon == "" {
		return errors.New("mode=polygon requires a polygon parameter")
	}
	return nil
}

func validateDistrict(q types.SearchQuery) error {
	if q.District == "" {
		return errors.New("mode=district requires a district parameter")
	}
	return nil
}
