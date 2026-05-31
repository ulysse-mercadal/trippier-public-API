package geo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// NominatimSearchURL and NominatimReverseURL are vars so tests can point them at a local mock server.
var (
	NominatimSearchURL  = "https://nominatim.openstreetmap.org/search"
	NominatimReverseURL = "https://nominatim.openstreetmap.org/reverse"
)

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
		NominatimSearchURL+"?"+params.Encode(), nil)
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

// CountryCode resolves coordinates to an ISO 3166-1 alpha-2 country code (uppercase)
// via Nominatim /reverse at zoom=3 (country granularity).
func CountryCode(ctx context.Context, lat, lng float64) (string, error) {
	params := url.Values{
		"lat":    {strconv.FormatFloat(lat, 'f', 6, 64)},
		"lon":    {strconv.FormatFloat(lng, 'f', 6, 64)},
		"format": {"json"},
		"zoom":   {"3"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		NominatimReverseURL+"?"+params.Encode(), nil)
	if err != nil {
		return "", fmt.Errorf("nominatim country: build request: %w", err)
	}
	req.Header.Set("User-Agent", "trippier-poi-api/1.0 (github.com/trippier/poi-api)")

	resp, err := nominatimClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("nominatim country: request: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Address struct {
			CountryCode string `json:"country_code"`
		} `json:"address"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("nominatim country: decode: %w", err)
	}
	if result.Address.CountryCode == "" {
		return "", fmt.Errorf("nominatim country: no result for %.4f,%.4f", lat, lng)
	}
	return strings.ToUpper(result.Address.CountryCode), nil
}

// ReverseGeocode resolves coordinates to a city name via the Nominatim OSM API.
// Returns the most specific populated place name available (city > town > village > county).
func ReverseGeocode(ctx context.Context, lat, lng float64) (string, error) {
	params := url.Values{
		"lat":    {strconv.FormatFloat(lat, 'f', 6, 64)},
		"lon":    {strconv.FormatFloat(lng, 'f', 6, 64)},
		"format": {"json"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		NominatimReverseURL+"?"+params.Encode(), nil)
	if err != nil {
		return "", fmt.Errorf("nominatim reverse: build request: %w", err)
	}
	req.Header.Set("User-Agent", "trippier-poi-api/1.0 (github.com/trippier/poi-api)")

	resp, err := nominatimClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("nominatim reverse: request: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Address struct {
			City    string `json:"city"`
			Town    string `json:"town"`
			Village string `json:"village"`
			County  string `json:"county"`
		} `json:"address"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("nominatim reverse: decode: %w", err)
	}

	for _, name := range []string{result.Address.City, result.Address.Town, result.Address.Village, result.Address.County} {
		if name != "" {
			return name, nil
		}
	}
	return "", fmt.Errorf("nominatim reverse: no city found for %.4f,%.4f", lat, lng)
}
