package geo_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/trippier/poi-api/internal/geo"
)

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func nominatimServer(t *testing.T, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

// ── CountryCode ───────────────────────────────────────────────────────────────

func TestCountryCode_Success(t *testing.T) {
	srv := nominatimServer(t, `{"address":{"country_code":"fr"}}`)
	old := geo.NominatimReverseURL
	geo.NominatimReverseURL = srv.URL
	t.Cleanup(func() { geo.NominatimReverseURL = old })

	cc, err := geo.CountryCode(context.Background(), 48.8566, 2.3522)
	if err != nil {
		t.Fatalf("CountryCode: %v", err)
	}
	if cc != "FR" {
		t.Errorf("CountryCode = %q, want %q", cc, "FR")
	}
}

func TestCountryCode_UppercasesResult(t *testing.T) {
	srv := nominatimServer(t, `{"address":{"country_code":"jp"}}`)
	old := geo.NominatimReverseURL
	geo.NominatimReverseURL = srv.URL
	t.Cleanup(func() { geo.NominatimReverseURL = old })

	cc, err := geo.CountryCode(context.Background(), 35.68, 139.69)
	if err != nil {
		t.Fatalf("CountryCode: %v", err)
	}
	if cc != "JP" {
		t.Errorf("CountryCode = %q, want %q", cc, "JP")
	}
}

func TestCountryCode_EmptyResult(t *testing.T) {
	srv := nominatimServer(t, `{"address":{}}`)
	old := geo.NominatimReverseURL
	geo.NominatimReverseURL = srv.URL
	t.Cleanup(func() { geo.NominatimReverseURL = old })

	_, err := geo.CountryCode(context.Background(), 0, 0)
	if err == nil {
		t.Error("expected error for empty country_code")
	}
}

func TestCountryCode_InvalidJSON(t *testing.T) {
	srv := nominatimServer(t, `not-json`)
	old := geo.NominatimReverseURL
	geo.NominatimReverseURL = srv.URL
	t.Cleanup(func() { geo.NominatimReverseURL = old })

	_, err := geo.CountryCode(context.Background(), 48.8566, 2.3522)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestCountryCode_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	old := geo.NominatimReverseURL
	geo.NominatimReverseURL = srv.URL
	t.Cleanup(func() { geo.NominatimReverseURL = old })

	// 500 response leads to JSON decode error (empty body) or other failure.
	_, err := geo.CountryCode(context.Background(), 48.8566, 2.3522)
	if err == nil {
		t.Error("expected error on server 500")
	}
}

// ── GeocodeDistrict ───────────────────────────────────────────────────────────

func TestGeocodeDistrict_Success(t *testing.T) {
	srv := nominatimServer(t, `[{"lat":"48.8566","lon":"2.3522"}]`)
	old := geo.NominatimSearchURL
	geo.NominatimSearchURL = srv.URL
	t.Cleanup(func() { geo.NominatimSearchURL = old })

	place, err := geo.GeocodeDistrict(context.Background(), "Paris")
	if err != nil {
		t.Fatalf("GeocodeDistrict: %v", err)
	}
	if abs(place.Lat-48.8566) > 1e-4 || abs(place.Lng-2.3522) > 1e-4 {
		t.Errorf("GeocodeDistrict coords = (%.4f, %.4f), want (48.8566, 2.3522)", place.Lat, place.Lng)
	}
}

func TestGeocodeDistrict_NoResult(t *testing.T) {
	srv := nominatimServer(t, `[]`)
	old := geo.NominatimSearchURL
	geo.NominatimSearchURL = srv.URL
	t.Cleanup(func() { geo.NominatimSearchURL = old })

	_, err := geo.GeocodeDistrict(context.Background(), "ZZZ-Nonexistent")
	if err == nil {
		t.Error("expected error for empty results")
	}
}

func TestGeocodeDistrict_InvalidCoords(t *testing.T) {
	srv := nominatimServer(t, `[{"lat":"not-a-float","lon":"also-bad"}]`)
	old := geo.NominatimSearchURL
	geo.NominatimSearchURL = srv.URL
	t.Cleanup(func() { geo.NominatimSearchURL = old })

	_, err := geo.GeocodeDistrict(context.Background(), "Bad")
	if err == nil {
		t.Error("expected error for unparseable coordinates")
	}
}

func TestGeocodeDistrict_InvalidJSON(t *testing.T) {
	srv := nominatimServer(t, `not-json`)
	old := geo.NominatimSearchURL
	geo.NominatimSearchURL = srv.URL
	t.Cleanup(func() { geo.NominatimSearchURL = old })

	_, err := geo.GeocodeDistrict(context.Background(), "Paris")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// ── ReverseGeocode ────────────────────────────────────────────────────────────

func TestReverseGeocode_City(t *testing.T) {
	srv := nominatimServer(t, `{"address":{"city":"Paris","country_code":"fr"}}`)
	old := geo.NominatimReverseURL
	geo.NominatimReverseURL = srv.URL
	t.Cleanup(func() { geo.NominatimReverseURL = old })

	name, err := geo.ReverseGeocode(context.Background(), 48.8566, 2.3522)
	if err != nil {
		t.Fatalf("ReverseGeocode: %v", err)
	}
	if name != "Paris" {
		t.Errorf("ReverseGeocode = %q, want %q", name, "Paris")
	}
}

func TestReverseGeocode_FallsBackToTown(t *testing.T) {
	srv := nominatimServer(t, `{"address":{"town":"Versailles","country_code":"fr"}}`)
	old := geo.NominatimReverseURL
	geo.NominatimReverseURL = srv.URL
	t.Cleanup(func() { geo.NominatimReverseURL = old })

	name, err := geo.ReverseGeocode(context.Background(), 48.80, 2.13)
	if err != nil {
		t.Fatalf("ReverseGeocode: %v", err)
	}
	if name != "Versailles" {
		t.Errorf("ReverseGeocode = %q, want %q", name, "Versailles")
	}
}

func TestReverseGeocode_FallsBackToVillage(t *testing.T) {
	srv := nominatimServer(t, `{"address":{"village":"Giverny","country_code":"fr"}}`)
	old := geo.NominatimReverseURL
	geo.NominatimReverseURL = srv.URL
	t.Cleanup(func() { geo.NominatimReverseURL = old })

	name, err := geo.ReverseGeocode(context.Background(), 49.07, 1.53)
	if err != nil {
		t.Fatalf("ReverseGeocode: %v", err)
	}
	if name != "Giverny" {
		t.Errorf("ReverseGeocode = %q, want %q", name, "Giverny")
	}
}

func TestReverseGeocode_NoCity(t *testing.T) {
	srv := nominatimServer(t, `{"address":{}}`)
	old := geo.NominatimReverseURL
	geo.NominatimReverseURL = srv.URL
	t.Cleanup(func() { geo.NominatimReverseURL = old })

	_, err := geo.ReverseGeocode(context.Background(), 0, 0)
	if err == nil {
		t.Error("expected error when no city/town/village found")
	}
}

func TestReverseGeocode_InvalidJSON(t *testing.T) {
	srv := nominatimServer(t, `not-json`)
	old := geo.NominatimReverseURL
	geo.NominatimReverseURL = srv.URL
	t.Cleanup(func() { geo.NominatimReverseURL = old })

	_, err := geo.ReverseGeocode(context.Background(), 48.8566, 2.3522)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
