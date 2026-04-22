// Package wikipedia implements the Provider interface for the Wikipedia MediaWiki API.
// It uses the geosearch feature to find articles near a location and enriches
// each result with an extract and thumbnail.
// Documentation: https://www.mediawiki.org/wiki/API:Geosearch
package wikipedia

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/trippier/poi-api/pkg/types"
)

const defaultTimeout = 10 * time.Second

// Provider fetches geo-located Wikipedia articles and maps them to POIs.
type Provider struct {
	client  *http.Client
	baseURL string
}

// New returns a Provider targeting the given language edition (e.g. "en", "fr").
func New(lang string) *Provider {
	return &Provider{
		client:  &http.Client{Timeout: defaultTimeout},
		baseURL: fmt.Sprintf("https://%s.wikipedia.org/w/api.php", lang),
	}
}

// NewWithBaseURL returns a Provider using a custom API base URL.
// Intended for testing against a local httptest server.
func NewWithBaseURL(baseURL string) *Provider {
	return &Provider{
		client:  &http.Client{Timeout: defaultTimeout},
		baseURL: baseURL,
	}
}

// Name implements providers.Provider.
func (p *Provider) Name() types.Provider { return types.ProviderWikipedia }

// SupportsMode implements providers.Provider.
func (p *Provider) SupportsMode(mode types.SearchMode) bool {
	return mode == types.ModeRadius || mode == types.ModeDistrict
}

// Search implements providers.Provider.
func (p *Provider) Search(ctx context.Context, q types.SearchQuery) ([]types.RawPoi, error) {
	pages, err := p.geosearch(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("wikipedia: geosearch: %w", err)
	}
	if len(pages) == 0 {
		return nil, nil
	}
	return p.enrichPages(ctx, pages), nil
}

type geosearchPage struct {
	PageID int     `json:"pageid"`
	Title  string  `json:"title"`
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
	Dist   float64 `json:"dist"`
}

func (p *Provider) geosearch(ctx context.Context, q types.SearchQuery) ([]geosearchPage, error) {
	params := url.Values{
		"action":      {"query"},
		"list":        {"geosearch"},
		"gscoord":     {fmt.Sprintf("%.6f|%.6f", q.Lat, q.Lng)},
		"gsradius":    {strconv.Itoa(q.Radius)},
		"gslimit":     {"50"},
		"gsnamespace": {"0"},
		"format":      {"json"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Query struct {
			Geosearch []geosearchPage `json:"geosearch"`
		} `json:"query"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Query.Geosearch, nil
}

func (p *Provider) enrichPages(ctx context.Context, pages []geosearchPage) []types.RawPoi {
	ids := make([]string, 0, len(pages))
	index := make(map[int]geosearchPage, len(pages))
	for _, pg := range pages {
		ids = append(ids, strconv.Itoa(pg.PageID))
		index[pg.PageID] = pg
	}

	params := url.Values{
		"action":      {"query"},
		"pageids":     {strings.Join(ids, "|")},
		"prop":        {"extracts|pageimages|pageprops"},
		"exintro":     {"1"},
		"exsentences": {"3"},
		"piprop":      {"thumbnail"},
		"pithumbsize": {"400"},
		"format":      {"json"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"?"+params.Encode(), nil)
	if err != nil {
		return p.pagesWithoutEnrichment(pages)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return p.pagesWithoutEnrichment(pages)
	}
	defer resp.Body.Close()

	var result struct {
		Query struct {
			Pages map[string]struct {
				PageID    int    `json:"pageid"`
				Title     string `json:"title"`
				Extract   string `json:"extract"`
				Thumbnail *struct {
					Source string `json:"source"`
				} `json:"thumbnail"`
				PageProps map[string]string `json:"pageprops"`
			} `json:"pages"`
		} `json:"query"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return p.pagesWithoutEnrichment(pages)
	}

	pois := make([]types.RawPoi, 0, len(result.Query.Pages))
	for _, wp := range result.Query.Pages {
		geo, ok := index[wp.PageID]
		if !ok {
			continue
		}
		poi := types.RawPoi{
			ID:          fmt.Sprintf("wikipedia:%d", wp.PageID),
			Name:        wp.Title,
			Type:        types.TypeSee,
			Provider:    types.ProviderWikipedia,
			Description: wp.Extract,
			Coords:      &types.Coordinates{Lat: geo.Lat, Lng: geo.Lon},
			Distance:    geo.Dist,
			WikidataID:  wp.PageProps["wikibase_item"],
		}
		if wp.Thumbnail != nil {
			poi.Thumbnail = wp.Thumbnail.Source
		}
		pois = append(pois, poi)
	}
	return pois
}

func (p *Provider) pagesWithoutEnrichment(pages []geosearchPage) []types.RawPoi {
	pois := make([]types.RawPoi, 0, len(pages))
	for _, pg := range pages {
		pois = append(pois, types.RawPoi{
			ID:       fmt.Sprintf("wikipedia:%d", pg.PageID),
			Name:     pg.Title,
			Type:     types.TypeSee,
			Provider: types.ProviderWikipedia,
			Coords:   &types.Coordinates{Lat: pg.Lat, Lng: pg.Lon},
			Distance: pg.Dist,
		})
	}
	return pois
}

