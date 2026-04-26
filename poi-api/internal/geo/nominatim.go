package geo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const nominatimURL = "https://nominatim.openstreetmap.org/search"

var nominatimClient = &http.Client{Timeout: 5 * time.Second}

// Place holds a geocoded location.
type Place struct {
	Lat float64
	Lng float64
}

// GeocodeDistrict resolves a place name to coordinates via the Nominatim OSM API.
func GeocodeDistrict(ctx context.Context, name string) (Place, error) {
	params := url.Values{
		"q":      {name},
		"format": {"json"},
		"limit":  {"1"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		nominatimURL+"?"+params.Encode(), nil)
	if err != nil {
		return Place{}, fmt.Errorf("nominatim: build request: %w", err)
	}
	req.Header.Set("User-Agent", "trippier-poi-api/1.0 (github.com/trippier/poi-api)")

	resp, err := nominatimClient.Do(req)
	if err != nil {
		return Place{}, fmt.Errorf("nominatim: request: %w", err)
	}
	defer resp.Body.Close()

	var results []struct {
		Lat string `json:"lat"`
		Lon string `json:"lon"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return Place{}, fmt.Errorf("nominatim: decode: %w", err)
	}
	if len(results) == 0 {
		return Place{}, fmt.Errorf("nominatim: no result for %q", name)
	}

	lat, err1 := strconv.ParseFloat(results[0].Lat, 64)
	lng, err2 := strconv.ParseFloat(results[0].Lon, 64)
	if err1 != nil || err2 != nil {
		return Place{}, fmt.Errorf("nominatim: invalid coordinates in response")
	}
	return Place{Lat: lat, Lng: lng}, nil
}
