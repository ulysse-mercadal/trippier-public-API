// Package search validates and executes point-of-interest search queries.
package search

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/trippier/poi-api/pkg/types"
)

// districtRe whitelists characters valid in a place name and blocks Overpass QL injection.
// Single-quote is excluded: it is not needed in place names and could break the QL string delimiter.
var districtRe = regexp.MustCompile(`^[\p{L}\p{N}\s\-,\.]+$`)

// polygonCoordRe accepts a single decimal coordinate token, optionally negative.
var polygonCoordRe = regexp.MustCompile(`^-?\d{1,3}(?:\.\d{1,10})?$`)

// Validate dispatches semantic validation of q to its mode-specific
// validator based on q.Mode; an empty mode defaults to ModeRadius. It
// returns an error describing why the query is invalid, or nil.
func Validate(q types.SearchQuery) error {
	switch q.Mode {
	case "", types.ModeRadius:
		return validateRadius(q)
	case types.ModePolygon:
		return validatePolygon(q)
	case types.ModeDistrict:
		return validateDistrict(q)
	default:
		return fmt.Errorf("unknown mode %q", q.Mode)
	}
}

// validateRadius checks that q's lat/lng are present and within valid
// ranges, and that q.Radius does not exceed 50 000 m. It returns an error
// describing the invalid field, or nil.
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

// validatePolygon checks that q.Polygon is a string of 3-100
// whitespace-separated lat/lon pairs, blocking QL injection. It returns an
// error describing the invalid field, or nil.
func validatePolygon(q types.SearchQuery) error {
	if q.Polygon == "" {
		return errors.New("mode=polygon requires a polygon parameter")
	}
	parts := strings.Fields(q.Polygon)
	if len(parts) < 6 {
		return errors.New("polygon requires at least 3 coordinate pairs (lat lon ...)")
	}
	if len(parts)%2 != 0 {
		return errors.New("polygon must contain an even number of values (lat lon pairs)")
	}
	if len(parts) > 200 {
		return errors.New("polygon exceeds maximum of 100 coordinate pairs")
	}
	for _, p := range parts {
		if !polygonCoordRe.MatchString(p) {
			return errors.New("polygon contains invalid coordinate value")
		}
	}
	return nil
}

// validateDistrict checks that q.District is non-empty, within length
// bounds, and whitelist-clean to prevent QL injection. It returns an error
// describing the invalid field, or nil.
func validateDistrict(q types.SearchQuery) error {
	if q.District == "" {
		return errors.New("mode=district requires a district parameter")
	}
	if len(q.District) > 200 {
		return errors.New("district name too long (max 200 characters)")
	}
	if !districtRe.MatchString(q.District) {
		return errors.New("district contains invalid characters")
	}
	return nil
}
