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

// apiTemplate is the MediaWiki API endpoint pattern; %s is the language edition subdomain.
const apiTemplate = "https://%s.wikivoyage.org/w/api.php"

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
	// apiTemplate, when non-empty, lets Search retarget baseURL to the
	// per-request language edition. It is empty for the test constructor,
	// which pins a fixed baseURL.
	apiTemplate string
}

// New creates a Provider for the given Wikivoyage language edition, using
// lang (e.g. "en", "fr") to build the API endpoint. It returns the
// configured Provider.
func New(lang string) *Provider {
	return &Provider{
		client:      &http.Client{Timeout: defaultTimeout},
		baseURL:     fmt.Sprintf(apiTemplate, lang),
		apiTemplate: apiTemplate,
	}
}

// NewWithURL creates a Provider targeting the custom MediaWiki API endpoint
// baseURL, for tests. It returns the configured Provider.
func NewWithURL(baseURL string) *Provider {
	return &Provider{
		client:  &http.Client{Timeout: defaultTimeout},
		baseURL: baseURL,
	}
}

// forLang returns a copy of p whose baseURL targets the requested language
// edition, or p unchanged when lang is empty/invalid or the endpoint is pinned
// (tests). The shared http.Client is reused.
//
// @param lang the caller-supplied language code (e.g. from ?lang=fr).
// @return the Provider to use for this request.
func (p *Provider) forLang(lang string) *Provider {
	code := providers.NormalizeLang(lang)
	if code == "" || p.apiTemplate == "" {
		return p
	}
	rp := *p
	rp.baseURL = fmt.Sprintf(p.apiTemplate, code)
	return &rp
}

// Name returns the Wikivoyage provider identifier.
func (p *Provider) Name() types.Provider { return types.ProviderWikivoyage }

// zoneURL builds the canonical Wikivoyage article URL for the given zone
// title. It returns the canonical article URL, or an empty string if it
// cannot be resolved.
func (p *Provider) zoneURL(zone string) string {
	i := strings.Index(p.baseURL, "/w/api.php")
	if i <= 0 {
		return ""
	}
	return fmt.Sprintf("%s/wiki/%s", p.baseURL[:i], url.PathEscape(zone))
}

// listingURL builds an article URL for zone with a fragment anchored to the
// listing named name. It returns the article URL with anchor, or the base
// URL if name is empty.
func (p *Provider) listingURL(zone, name string) string {
	base := p.zoneURL(zone)
	if base == "" || name == "" {
		return base
	}
	anchor := strings.ReplaceAll(name, " ", "_")
	return base + "#" + url.PathEscape(anchor)
}

// wikipediaURL converts v, a listing's raw wikipedia= field value
// (optionally in "lang:Title" form), into a full Wikipedia article URL. It
// returns the full article URL, or an empty string if v is empty.
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

// langCode extracts the language subdomain from the configured base URL. It
// returns the language code, defaulting to "en".
func (p *Provider) langCode() string {
	if u, err := url.Parse(p.baseURL); err == nil {
		if i := strings.Index(u.Host, "."); i > 0 {
			return u.Host[:i]
		}
	}
	return "en"
}

// SupportsMode reports whether mode is a search mode supported by this
// provider.
func (p *Provider) SupportsMode(mode types.SearchMode) bool {
	return mode == types.ModeDistrict || mode == types.ModeRadius
}

// Search fetches POIs for query q, resolved either by district or by
// radius lookup, using ctx for cancellation. It returns the matching POIs,
// or an error on failure.
func (p *Provider) Search(ctx context.Context, q types.SearchQuery) ([]types.RawPoi, error) {
	rp := p.forLang(q.Lang)

	var pageTitle string
	var err error

	switch q.Mode {
	case types.ModeDistrict:
		pageTitle = q.District
	case types.ModeRadius:
		pageTitle, err = rp.resolveZone(ctx, q.Lat, q.Lng, q.Radius)
		if err != nil {
			return nil, fmt.Errorf("wikivoyage: resolve zone: %w", err)
		}
	default:
		return nil, nil
	}

	wikitext, err := rp.fetchWikitext(ctx, pageTitle)
	if err != nil {
		return nil, fmt.Errorf("wikivoyage: fetch wikitext for %q: %w", pageTitle, err)
	}

	return rp.parseListings(wikitext, pageTitle), nil
}

// zoneSearchRadius is fixed at the MediaWiki geosearch maximum (10 000 m) to always find nearby articles.
const zoneSearchRadius = 10_000

// resolveZone finds the nearest Wikivoyage article title via MediaWiki
// geosearch, searching near lat/lng using ctx for cancellation; the third
// parameter is unused. It returns the nearest article title, or an error if
// none is found.
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

// fetchWikitext retrieves the raw wikitext of the Wikivoyage page named
// title, using ctx for cancellation. It returns the raw wikitext content,
// or an error on failure.
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

var fieldRe = regexp.MustCompile(`(\w+)\s*=\s*([^|}\n]+)`)

var listingKinds = map[string]bool{
	"see": true, "do": true, "eat": true, "drink": true,
	"buy": true, "sleep": true, "listing": true,
}

type listingMatch struct {
	kind string
	body string
}

// scanListings extracts top-level listing templates from wikitext as
// (kind, body) pairs. It returns the matched listings with their kind and
// body text.
func scanListings(wikitext string) []listingMatch {
	var out []listingMatch
	for i := 0; i+1 < len(wikitext); i++ {
		if wikitext[i] != '{' || wikitext[i+1] != '{' {
			continue
		}
		j := i + 2
		for j < len(wikitext) && (wikitext[j] == ' ' || wikitext[j] == '\t' || wikitext[j] == '\n') {
			j++
		}
		k := j
		for k < len(wikitext) && isAlpha(wikitext[k]) {
			k++
		}
		kind := strings.ToLower(wikitext[j:k])
		if !listingKinds[kind] {
			continue
		}
		for k < len(wikitext) && (wikitext[k] == ' ' || wikitext[k] == '\t' || wikitext[k] == '\n') {
			k++
		}
		if k >= len(wikitext) || wikitext[k] != '|' {
			continue
		}
		bodyStart := k + 1
		depth := 1
		p := bodyStart
		for p+1 < len(wikitext) {
			if wikitext[p] == '{' && wikitext[p+1] == '{' {
				depth++
				p += 2
				continue
			}
			if wikitext[p] == '}' && wikitext[p+1] == '}' {
				depth--
				if depth == 0 {
					out = append(out, listingMatch{kind: kind, body: wikitext[bodyStart:p]})
					i = p + 1
					break
				}
				p += 2
				continue
			}
			p++
		}
	}
	return out
}

// isAlpha reports whether c is an ASCII letter.
func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

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

// markerStripper removes leftover wiki-link bracket markers without deleting
// the text inside them. Used after wikiLinkRe has consumed complete [[…]]
// links — anything still bracketed at that point is a truncated fragment
// whose readable text we want to keep.
var markerStripper = strings.NewReplacer("[[", "", "]]", "")

// stripDescriptionMarkup converts s, a listing's raw content field, into
// readable plain text. It returns the plain-text description.
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
	s = markerStripper.Replace(s)
	s = htmlTagRe.ReplaceAllString(s, " ")
	s = boldRe.ReplaceAllString(s, "$1")
	s = italicRe.ReplaceAllString(s, "$1")
	s = html.UnescapeString(s)
	s = multiSpaceRe.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// stripWikiMarkup resolves [[Target|Display]] links in s to their Display
// text and drops broken fragments. It returns plain text with wiki-link
// markup removed.
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

// parseListings extracts listing templates from wikitext and converts them
// to RawPoi records belonging to the given zone/article title zone. It
// returns the parsed POIs, including any linked Wikipedia entries.
func (p *Provider) parseListings(wikitext, zone string) []types.RawPoi {
	matches := scanListings(wikitext)
	pois := make([]types.RawPoi, 0, len(matches))

	for _, match := range matches {
		kind := match.kind
		fields := p.parseFields(match.body)

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
				Website: providers.SafeURL(fields["url"]),
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

// pipePlaceholder masks "|" characters inside [[…]] wiki links so the
// field-value regex below doesn't treat them as field separators. NUL is
// safe because wikitext never contains literal NUL bytes.
const pipePlaceholder = "\x00"

// parseFields extracts key=value pairs from raw, a listing template's
// parameter string. It returns a map of field name to trimmed value.
func (p *Provider) parseFields(raw string) map[string]string {
	masked := maskPipesInBrackets(raw)
	fields := map[string]string{}
	for _, m := range fieldRe.FindAllStringSubmatch(masked, -1) {
		key := strings.TrimSpace(m[1])
		val := strings.ReplaceAll(m[2], pipePlaceholder, "|")
		fields[key] = strings.TrimSpace(val)
	}
	return fields
}

// maskPipesInBrackets replaces "|" characters inside [[…]] wiki links in s
// with pipePlaceholder. It returns the string with in-link pipes masked.
func maskPipesInBrackets(s string) string {
	if !strings.Contains(s, "[[") || !strings.Contains(s, "|") {
		return s
	}
	out := make([]byte, 0, len(s))
	depth := 0
	for i := 0; i < len(s); i++ {
		if i+1 < len(s) && s[i] == '[' && s[i+1] == '[' {
			depth++
			out = append(out, '[', '[')
			i++
			continue
		}
		if i+1 < len(s) && s[i] == ']' && s[i+1] == ']' {
			if depth > 0 {
				depth--
			}
			out = append(out, ']', ']')
			i++
			continue
		}
		if s[i] == '|' && depth > 0 {
			out = append(out, pipePlaceholder[0])
			continue
		}
		out = append(out, s[i])
	}
	return string(out)
}

// parseCoords extracts latitude/longitude from fields, a parsed listing
// field map. It returns lat and lng, with ok false if either is missing or
// invalid.
func (p *Provider) parseCoords(fields map[string]string) (lat, lng float64, ok bool) {
	latStr, lngStr := fields["lat"], fields["long"]
	if latStr == "" || lngStr == "" {
		return 0, 0, false
	}
	lat, err1 := strconv.ParseFloat(latStr, 64)
	lng, err2 := strconv.ParseFloat(lngStr, 64)
	return lat, lng, err1 == nil && err2 == nil
}

// init registers the Wikivoyage provider factory.
func init() {
	providers.Register(types.ProviderWikivoyage, func(cfg providers.BuildConfig) (providers.Provider, error) {
		return New(cfg.Lang), nil
	})
}
