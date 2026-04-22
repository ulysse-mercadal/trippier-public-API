// Package wikivoyage implements the Provider interface for the Wikivoyage MediaWiki API.
// Documentation: https://en.wikivoyage.org/w/api.php
package wikivoyage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/trippier/poi-api/pkg/types"
)

const defaultTimeout = 10 * time.Second

var listingTypeMap = map[string]types.PoiType{
	"see":   types.TypeSee,
	"do":    types.TypeDo,
	"eat":   types.TypeEat,
	"drink": types.TypeDrink,
	"buy":   types.TypeBuy,
	"sleep": types.TypeSleep,
}

// Provider fetches POIs from Wikivoyage by parsing wikitext listing templates.
type Provider struct {
	client  *http.Client
	baseURL string
}

// New returns a Provider targeting the given language edition (e.g. "en", "fr").
func New(lang string) *Provider {
	return &Provider{
		client:  &http.Client{Timeout: defaultTimeout},
		baseURL: fmt.Sprintf("https://%s.wikivoyage.org/w/api.php", lang),
	}
}

// Name implements providers.Provider.
func (p *Provider) Name() types.Provider { return types.ProviderWikivoyage }

// SupportsMode implements providers.Provider.
// Wikivoyage works best with district and radius (zone resolution); polygon is approximate.
func (p *Provider) SupportsMode(mode types.SearchMode) bool {
	return mode == types.ModeDistrict || mode == types.ModeRadius
}

// Search implements providers.Provider.
func (p *Provider) Search(ctx context.Context, q types.SearchQuery) ([]types.RawPoi, error) {
	var pageTitle string
	var err error

	switch q.Mode {
	case types.ModeDistrict:
		pageTitle = q.District
	case types.ModeRadius:
		pageTitle, err = p.resolveZone(ctx, q.Lat, q.Lng, q.Radius)
		if err != nil {
			return nil, fmt.Errorf("wikivoyage: resolve zone: %w", err)
		}
	default:
		return nil, nil
	}

	wikitext, err := p.fetchWikitext(ctx, pageTitle)
	if err != nil {
		return nil, fmt.Errorf("wikivoyage: fetch wikitext for %q: %w", pageTitle, err)
	}

	return p.parseListings(wikitext, pageTitle), nil
}

func (p *Provider) resolveZone(ctx context.Context, lat, lng float64, radius int) (string, error) {
	params := url.Values{
		"action":      {"query"},
		"list":        {"geosearch"},
		"gscoord":     {fmt.Sprintf("%.6f|%.6f", lat, lng)},
		"gsradius":    {strconv.Itoa(radius)},
		"gslimit":     {"1"},
		"gsnamespace": {"0"},
		"format":      {"json"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"?"+params.Encode(), nil)
	if err != nil {
		return "", err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Query struct {
			Geosearch []struct {
				Title string `json:"title"`
			} `json:"geosearch"`
		} `json:"query"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if len(result.Query.Geosearch) == 0 {
		return "", fmt.Errorf("no zone found near (%.4f, %.4f)", lat, lng)
	}
	return result.Query.Geosearch[0].Title, nil
}

func (p *Provider) fetchWikitext(ctx context.Context, title string) (string, error) {
	params := url.Values{
		"action": {"parse"},
		"page":   {title},
		"prop":   {"wikitext"},
		"format": {"json"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"?"+params.Encode(), nil)
	if err != nil {
		return "", err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Parse struct {
			Wikitext struct {
				Content string `json:"*"`
			} `json:"wikitext"`
		} `json:"parse"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Parse.Wikitext.Content, nil
}

var listingRe = regexp.MustCompile(`(?i)\{\{(see|do|eat|drink|buy|sleep|listing)\s*\|([^}]+)\}\}`)
var fieldRe = regexp.MustCompile(`(\w+)\s*=\s*([^|}\n]+)`)

func (p *Provider) parseListings(wikitext, zone string) []types.RawPoi {
	var pois []types.RawPoi

	for _, match := range listingRe.FindAllStringSubmatch(wikitext, -1) {
		kind := strings.ToLower(match[1])
		fields := p.parseFields(match[2])

		name := strings.TrimSpace(fields["name"])
		if name == "" {
			continue
		}

		poi := types.RawPoi{
			ID:          fmt.Sprintf("wikivoyage:%s:%s", zone, name),
			Name:        name,
			Type:        listingTypeMap[kind],
			Provider:    types.ProviderWikivoyage,
			Description: strings.TrimSpace(fields["content"]),
			Contact:     types.Contact{
				Website: strings.TrimSpace(fields["url"]),
				Phone:   strings.TrimSpace(fields["phone"]),
				Hours:   strings.TrimSpace(fields["hours"]),
			},
			Zone: &types.Zone{Name: zone, Source: types.ProviderWikivoyage},
		}

		if lat, lng, ok := p.parseCoords(fields); ok {
			poi.Coords = &types.Coordinates{Lat: lat, Lng: lng}
			poi.Zone = nil
		} else {
			poi.Coords = &types.Coordinates{Approximate: true}
		}

		pois = append(pois, poi)
	}

	return pois
}

func (p *Provider) parseFields(raw string) map[string]string {
	fields := map[string]string{}
	for _, m := range fieldRe.FindAllStringSubmatch(raw, -1) {
		fields[strings.TrimSpace(m[1])] = strings.TrimSpace(m[2])
	}
	return fields
}

func (p *Provider) parseCoords(fields map[string]string) (lat, lng float64, ok bool) {
	latStr, lngStr := fields["lat"], fields["long"]
	if latStr == "" || lngStr == "" {
		return 0, 0, false
	}
	lat, err1 := strconv.ParseFloat(latStr, 64)
	lng, err2 := strconv.ParseFloat(lngStr, 64)
	return lat, lng, err1 == nil && err2 == nil
}
