// Package geonames implements the Provider interface for the GeoNames geographical database.
// Documentation: https://www.geonames.org/export/web-services.html
// Requires a free GeoNames account: https://www.geonames.org/login
package geonames

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/trippier/poi-api/pkg/types"
)

const (
	apiURL         = "http://api.geonames.org"
	defaultTimeout = 10 * time.Second
	maxRows        = 100
)

var fcodeTypeMap = map[string]types.PoiType{
	"MUS":  types.TypeSee,
	"MNMT": types.TypeSee,
	"CSTL": types.TypeSee,
	"MSQE": types.TypeSee,
	"CH":   types.TypeSee,
	"RSRT": types.TypeSleep,
	"HTL":  types.TypeSleep,
	"HSTP": types.TypeSleep,
	"PRK":  types.TypeDo,
	"GRDN": types.TypeDo,
	"RECG": types.TypeDo,
	"RSTN": types.TypeGeneric,
}

type geonamesResponse struct {
	Geonames []geonameItem `json:"geonames"`
}

type geonameItem struct {
	GeonameID   int    `json:"geonameId"`
	Name        string `json:"name"`
	Lat         string `json:"lat"`
	Lng         string `json:"lng"`
	FcodeName   string `json:"fcodeName"`
	Fcode       string `json:"fcode"`
	CountryCode string `json:"countryCode"`
	Distance    string `json:"distance"`
}

// Provider fetches nearby geographic features from the GeoNames API.
type Provider struct {
	client   *http.Client
	username string
	baseURL  string
}

// New returns a Provider authenticated with the given GeoNames username.
func New(username string) *Provider {
	return &Provider{
		client:   &http.Client{Timeout: defaultTimeout},
		username: username,
		baseURL:  apiURL,
	}
}

// NewWithURL returns a Provider targeting a custom API endpoint.
// Intended for testing against a local httptest server.
func NewWithURL(baseURL, username string) *Provider {
	return &Provider{
		client:   &http.Client{Timeout: defaultTimeout},
		username: username,
		baseURL:  baseURL,
	}
}

// Name implements providers.Provider.
func (p *Provider) Name() types.Provider { return types.ProviderGeoNames }

// SupportsMode implements providers.Provider.
func (p *Provider) SupportsMode(mode types.SearchMode) bool {
	return mode == types.ModeRadius || mode == types.ModeDistrict
}

// Search implements providers.Provider.
func (p *Provider) Search(ctx context.Context, q types.SearchQuery) ([]types.RawPoi, error) {
	var endpoint string
	var params url.Values

	switch q.Mode {
	case types.ModeRadius:
		endpoint = "/findNearbyJSON"
		params = url.Values{
			"lat":      {fmt.Sprintf("%.6f", q.Lat)},
			"lng":      {fmt.Sprintf("%.6f", q.Lng)},
			"radius":   {strconv.Itoa(q.Radius / 1000)},
			"maxRows":  {strconv.Itoa(maxRows)},
			"username": {p.username},
		}
	case types.ModeDistrict:
		if q.Lat != 0 || q.Lng != 0 {
			endpoint = "/findNearbyJSON"
			radius := q.Radius / 1000
			if radius == 0 {
				radius = 5
			}
			params = url.Values{
				"lat":      {fmt.Sprintf("%.6f", q.Lat)},
				"lng":      {fmt.Sprintf("%.6f", q.Lng)},
				"radius":   {strconv.Itoa(radius)},
				"maxRows":  {strconv.Itoa(maxRows)},
				"username": {p.username},
			}
		} else {
			endpoint = "/searchJSON"
			params = url.Values{
				"q":        {q.District},
				"maxRows":  {strconv.Itoa(maxRows)},
				"username": {p.username},
			}
		}
	default:
		return nil, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("geonames: build request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("geonames: do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geonames: unexpected status %d", resp.StatusCode)
	}

	var result geonamesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("geonames: decode response: %w", err)
	}

	return p.toRawPois(result.Geonames), nil
}

// toRawPois converts GeoNames items to RawPoi records, skipping entries with unknown feature codes or invalid coordinates.
func (p *Provider) toRawPois(items []geonameItem) []types.RawPoi {
	pois := make([]types.RawPoi, 0, len(items))
	for _, item := range items {
		poiType, known := fcodeTypeMap[item.Fcode]
		if !known {
			continue
		}
		lat, err1 := strconv.ParseFloat(item.Lat, 64)
		lng, err2 := strconv.ParseFloat(item.Lng, 64)
		if err1 != nil || err2 != nil {
			continue
		}
		pois = append(pois, types.RawPoi{
			ID:       fmt.Sprintf("geonames:%d", item.GeonameID),
			Name:     item.Name,
			Type:     poiType,
			Provider: types.ProviderGeoNames,
			Coords:   &types.Coordinates{Lat: lat, Lng: lng},
		})
	}
	return pois
}
