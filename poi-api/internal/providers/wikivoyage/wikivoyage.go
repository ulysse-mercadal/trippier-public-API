// Package wikivoyage implements the Provider interface for the Wikivoyage MediaWiki API.
// Documentation: https://en.wikivoyage.org/w/api.php
package wikivoyage

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/trippier/poi-api/internal/providers"
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

// NewWithURL returns a Provider targeting a custom API endpoint.
// Intended for testing against a local httptest server.
func NewWithURL(baseURL string) *Provider {
	return &Provider{
		client:  &http.Client{Timeout: defaultTimeout},
		baseURL: baseURL,
	}
}

// Name implements providers.Provider.
func (p *Provider) Name() types.Provider { return types.ProviderWikivoyage }

// zoneURL returns the canonical Wikivoyage article URL for a zone title.
// Derived from the configured MediaWiki endpoint so tests against a local
// httptest server produce predictable URLs.
func (p *Provider) zoneURL(zone string) string {
	i := strings.Index(p.baseURL, "/w/api.php")
	if i <= 0 {
		return ""
	}
	return fmt.Sprintf("%s/wiki/%s", p.baseURL[:i], url.PathEscape(zone))
}

// listingURL returns the article URL with a fragment pointing at the listing
// inside the page. Wikivoyage listing templates render each entry as a span
// with id="<listing name>" (spaces become underscores, MediaWiki convention),
// so the fragment scrolls the reader straight to the row.
func (p *Provider) listingURL(zone, name string) string {
	base := p.zoneURL(zone)
	if base == "" || name == "" {
		return base
	}
	anchor := strings.ReplaceAll(name, " ", "_")
	return base + "#" + url.PathEscape(anchor)
}

// @param v raw value of a listing's wikipedia= field (e.g. "Eiffel Tower", "fr:Tour Eiffel", or a full URL).
// @return a Wikipedia article URL, defaulting to the Wikivoyage language edition when the value has no interwiki prefix, or empty when v is empty.
func (p *Provider) wikipediaURL(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	if strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://") {
		return v
	}
	lang := p.langCode()
	if i := strings.Index(v, ":"); i > 0 && i <= 5 {
		lang = strings.ToLower(v[:i])
		v = v[i+1:]
	}
	v = strings.ReplaceAll(strings.TrimSpace(v), " ", "_")
	if v == "" {
		return ""
	}
	return fmt.Sprintf("https://%s.wikipedia.org/wiki/%s", lang, url.PathEscape(v))
}

// langCode extracts the language subdomain from the configured baseURL
// (e.g. "https://fr.wikivoyage.org/w/api.php" → "fr"). Falls back to "en".
func (p *Provider) langCode() string {
	if u, err := url.Parse(p.baseURL); err == nil {
		if i := strings.Index(u.Host, "."); i > 0 {
			return u.Host[:i]
		}
	}
	return "en"
}

// SupportsMode implements providers.Provider.
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

// zoneSearchRadius is fixed at the MediaWiki geosearch maximum (10 000 m) to always find nearby articles.
const zoneSearchRadius = 10_000

// resolveZone finds the nearest Wikivoyage article title for the given coordinates via MediaWiki geosearch.
func (p *Provider) resolveZone(ctx context.Context, lat, lng float64, _ int) (string, error) {
	params := url.Values{
		"action":      {"query"},
		"list":        {"geosearch"},
		"gscoord":     {fmt.Sprintf("%.6f|%.6f", lat, lng)},
		"gsradius":    {strconv.Itoa(zoneSearchRadius)},
		"gslimit":     {"1"},
		"gsnamespace": {"0"},
		"format":      {"json"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"?"+params.Encode(), nil)
	if err != nil {
		return "", err
	}
	providers.SetUserAgent(req)

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

// fetchWikitext retrieves the raw wikitext of a Wikivoyage page by title.
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
	providers.SetUserAgent(req)

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

// wikiLinkRe matches complete [[Target]] or [[Target|Display]] links.
var wikiLinkRe = regexp.MustCompile(`\[\[[^\]]*\]\]`)

// wikiFragmentRe matches broken [[ fragments (truncated by the | field delimiter).
var wikiFragmentRe = regexp.MustCompile(`\[\[.*`)

// Markup patterns used by stripDescriptionMarkup. Order of application matters
// (refs and templates first to remove their inner content before generic HTML
// tag stripping picks up the bare angle brackets).
var (
	refPairRe    = regexp.MustCompile(`(?is)<ref[^>]*>.*?</ref>`)
	refSelfRe    = regexp.MustCompile(`(?i)<ref[^>]*/>`)
	templateRe   = regexp.MustCompile(`\{\{[^{}]*\}\}`)
	htmlTagRe    = regexp.MustCompile(`<[^>]+>`)
	boldRe       = regexp.MustCompile(`'''([^']*)'''`)
	italicRe     = regexp.MustCompile(`''([^']*)''`)
	multiSpaceRe = regexp.MustCompile(`\s+`)
)

// @param s raw wikitext value from a listing's content field.
// @return s with refs, templates, HTML tags, entities, wiki links and bold/italic markup resolved to readable text.
func stripDescriptionMarkup(s string) string {
	s = refPairRe.ReplaceAllString(s, "")
	s = refSelfRe.ReplaceAllString(s, "")
	for {
		out := templateRe.ReplaceAllString(s, "")
		if out == s {
			break
		}
		s = out
	}
	s = wikiLinkRe.ReplaceAllStringFunc(s, func(m string) string {
		inner := m[2 : len(m)-2]
		if i := strings.LastIndex(inner, "|"); i >= 0 {
			return inner[i+1:]
		}
		return inner
	})
	s = wikiFragmentRe.ReplaceAllString(s, "")
	s = htmlTagRe.ReplaceAllString(s, " ")
	s = boldRe.ReplaceAllString(s, "$1")
	s = italicRe.ReplaceAllString(s, "$1")
	s = html.UnescapeString(s)
	s = multiSpaceRe.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// stripWikiMarkup resolves [[Target|Display]] → Display, drops broken [[ fragments.
func stripWikiMarkup(s string) string {
	s = wikiLinkRe.ReplaceAllStringFunc(s, func(m string) string {
		inner := m[2 : len(m)-2]
		if i := strings.LastIndex(inner, "|"); i >= 0 {
			return strings.TrimSpace(inner[i+1:])
		}
		if i := strings.Index(inner, "#"); i >= 0 {
			inner = inner[:i]
		}
		if i := strings.LastIndex(inner, "/"); i >= 0 {
			inner = inner[i+1:]
		}
		return strings.TrimSpace(inner)
	})
	s = wikiFragmentRe.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

// parseListings extracts listing templates from wikitext and converts them to RawPoi records.
func (p *Provider) parseListings(wikitext, zone string) []types.RawPoi {
	matches := listingRe.FindAllStringSubmatch(wikitext, -1)
	pois := make([]types.RawPoi, 0, len(matches))

	for _, match := range matches {
		kind := strings.ToLower(match[1])
		fields := p.parseFields(match[2])

		name := stripWikiMarkup(strings.TrimSpace(fields["name"]))
		if name == "" {
			continue
		}

		var images []string
		if u := providers.CommonsFileURL(fields["image"]); u != "" {
			images = []string{u}
		}

		poi := types.RawPoi{
			ID:          fmt.Sprintf("wikivoyage:%s:%s", zone, name),
			Name:        name,
			Type:        listingTypeMap[kind],
			Provider:    types.ProviderWikivoyage,
			Description: stripDescriptionMarkup(fields["content"]),
			Contact: types.Contact{
				Website: strings.TrimSpace(fields["url"]),
				Phone:   strings.TrimSpace(fields["phone"]),
				Hours:   strings.TrimSpace(fields["hours"]),
			},
			Images:    images,
			Zone:      &types.Zone{Name: zone, Source: types.ProviderWikivoyage},
			SourceURL: p.listingURL(zone, name),
		}

		if lat, lng, ok := p.parseCoords(fields); ok {
			poi.Coords = &types.Coordinates{Lat: lat, Lng: lng}
			poi.Zone = nil
		} else {
			poi.Coords = &types.Coordinates{Approximate: true}
		}

		pois = append(pois, poi)

		if wikiURL := p.wikipediaURL(fields["wikipedia"]); wikiURL != "" {
			pois = append(pois, types.RawPoi{
				ID:        fmt.Sprintf("wikipedia:%s:%s", zone, name),
				Name:      name,
				Type:      poi.Type,
				Provider:  types.ProviderWikipedia,
				Coords:    poi.Coords,
				Zone:      poi.Zone,
				SourceURL: wikiURL,
			})
		}
	}

	return pois
}

// parseFields extracts key=value pairs from a listing template's parameter string.
func (p *Provider) parseFields(raw string) map[string]string {
	fields := map[string]string{}
	for _, m := range fieldRe.FindAllStringSubmatch(raw, -1) {
		fields[strings.TrimSpace(m[1])] = strings.TrimSpace(m[2])
	}
	return fields
}

// parseCoords extracts lat/long from a listing's field map and returns ok=false if either is missing or invalid.
func (p *Provider) parseCoords(fields map[string]string) (lat, lng float64, ok bool) {
	latStr, lngStr := fields["lat"], fields["long"]
	if latStr == "" || lngStr == "" {
		return 0, 0, false
	}
	lat, err1 := strconv.ParseFloat(latStr, 64)
	lng, err2 := strconv.ParseFloat(lngStr, 64)
	return lat, lng, err1 == nil && err2 == nil
}
