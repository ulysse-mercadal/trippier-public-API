// Package overpass implements the Provider interface for the OpenStreetMap Overpass API.
// Documentation: https://wiki.openstreetmap.org/wiki/Overpass_API
package overpass

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/trippier/poi-api/internal/providers"
	"github.com/trippier/poi-api/pkg/types"
)

const defaultTimeout = 10 * time.Second

// defaultAPIURLs lists the public Overpass mirrors we try in order. Each has
// its own per-IP quota, so rotating between them on transport / 429 / 5xx
// errors effectively multiplies the available headroom without paying for a
// self-hosted instance. Order is "best-effort first":
//   - overpass-api.de: the canonical instance.
//   - overpass.kumi.systems: privately-operated mirror, very reliable.
//   - lz4.overpass-api.de: alternate edge of the canonical pool.
//
// Keep additions conservative — a banned mirror still costs one connect
// attempt per call before we fall through.
var defaultAPIURLs = []string{
	"https://overpass-api.de/api/interpreter",
	"https://overpass.kumi.systems/api/interpreter",
	"https://lz4.overpass-api.de/api/interpreter",
}

var osmTagMap = map[string]types.PoiType{
	"museum":      types.TypeSee,
	"gallery":     types.TypeSee,
	"artwork":     types.TypeSee,
	"monument":    types.TypeSee,
	"castle":      types.TypeSee,
	"ruins":       types.TypeSee,
	"viewpoint":   types.TypeSee,
	"attraction":  types.TypeSee,
	"restaurant":  types.TypeEat,
	"cafe":        types.TypeEat,
	"fast_food":   types.TypeEat,
	"food_court":  types.TypeEat,
	"bar":         types.TypeDrink,
	"pub":         types.TypeDrink,
	"nightclub":   types.TypeDrink,
	"biergarten":  types.TypeDrink,
	"hotel":       types.TypeSleep,
	"hostel":      types.TypeSleep,
	"guest_house": types.TypeSleep,
	"motel":       types.TypeSleep,
	"camp_site":   types.TypeSleep,
	"theme_park":  types.TypeDo,
	"zoo":         types.TypeDo,
	"aquarium":    types.TypeDo,
}

var typeToOsmFilters = map[types.PoiType][]string{
	// artwork excluded: thousands of minor street sculptures pollute results.
	types.TypeSee:   {`["tourism"~"museum|gallery|attraction|viewpoint|castle|ruins|monument"]`, `["historic"~"monument|memorial|castle|ruins|fort|battlefield|archaeological_site"]`},
	types.TypeEat:   {`["amenity"~"restaurant|cafe|fast_food|food_court"]`},
	types.TypeDrink: {`["amenity"~"bar|pub|nightclub|biergarten"]`},
	types.TypeSleep: {`["tourism"~"hotel|hostel|guest_house|motel|camp_site"]`},
	// park excluded: every urban green space matches, returning hundreds of minor squares.
	types.TypeDo: {`["tourism"~"theme_park|zoo|aquarium"]`, `["leisure"~"sports_centre|water_park|golf_course|marina"]`},
	// ["shop"] without a value restriction returns every shop in the city.
	types.TypeBuy: {`["shop"~"mall|department_store|market|souvenir|gift|bookstore"]`, `["amenity"="marketplace"]`},
}

type overpassResponse struct {
	Elements []overpassElement `json:"elements"`
}

type overpassElement struct {
	Type   string            `json:"type"`
	ID     int64             `json:"id"`
	Lat    float64           `json:"lat"`
	Lon    float64           `json:"lon"`
	Center *overpassCenter   `json:"center,omitempty"`
	Tags   map[string]string `json:"tags"`
}

type overpassCenter struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// Provider fetches POIs from the OpenStreetMap Overpass API.
//
// `apiURLs` holds an ordered list of Overpass endpoints — the first reachable
// one wins. On a transport error / 429 / 5xx the next URL is tried, so a
// banned IP on one mirror or a transient outage doesn't surface as a hard
// failure to callers.
type Provider struct {
	client  *http.Client
	apiURLs []string
}

// New returns a Provider rotating across the public Overpass mirrors listed
// in {@link defaultAPIURLs}.
func New() *Provider {
	urls := make([]string, len(defaultAPIURLs))
	copy(urls, defaultAPIURLs)
	return &Provider{
		client:  &http.Client{Timeout: defaultTimeout},
		apiURLs: urls,
	}
}

// NewWithURL returns a Provider targeting a single custom Overpass endpoint.
// Intended for testing against a local httptest server.
func NewWithURL(u string) *Provider {
	return &Provider{
		client:  &http.Client{Timeout: defaultTimeout},
		apiURLs: []string{u},
	}
}

// NewWithURLs returns a Provider rotating across an explicit list of
// Overpass endpoints — used when a caller wants to override the public
// mirror list (e.g. a self-hosted instance first, public mirrors as fallback).
func NewWithURLs(urls []string) *Provider {
	cp := make([]string, len(urls))
	copy(cp, urls)
	return &Provider{
		client:  &http.Client{Timeout: defaultTimeout},
		apiURLs: cp,
	}
}

// Name implements providers.Provider.
func (p *Provider) Name() types.Provider { return types.ProviderOverpass }

// SupportsMode implements providers.Provider. Overpass supports all search modes.
func (p *Provider) SupportsMode(_ types.SearchMode) bool { return true }

// Search implements providers.Provider. Walks the configured mirror list and
// returns the first successful response. Only retries on transient failures
// (transport errors, HTTP 429, HTTP 5xx) — a 4xx is the same query failing
// everywhere, so we fail fast.
func (p *Provider) Search(ctx context.Context, q types.SearchQuery) ([]types.RawPoi, error) {
	body := url.Values{"data": {p.buildQuery(q)}}.Encode()

	var lastErr error
	for _, apiURL := range p.apiURLs {
		pois, retryable, err := p.searchOnce(ctx, apiURL, body)
		if err == nil {
			return pois, nil
		}
		lastErr = err
		if ctx.Err() != nil {
			return nil, fmt.Errorf("overpass: %w", ctx.Err())
		}
		if !retryable {
			return nil, err
		}
	}
	if lastErr == nil {
		return nil, fmt.Errorf("overpass: no mirrors configured")
	}
	return nil, fmt.Errorf("overpass: all mirrors failed: %w", lastErr)
}

// searchOnce performs a single POST against `apiURL`. The boolean second
// return indicates whether the caller should try the next mirror — true for
// transport errors / 429 / 5xx, false for 4xx (bad query — won't work on
// any mirror) or a decode failure (a successful 200 with malformed JSON
// suggests we hit a wrong-route placeholder, not a typical overload).
func (p *Provider) searchOnce(ctx context.Context, apiURL, body string) ([]types.RawPoi, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(body))
	if err != nil {
		return nil, false, fmt.Errorf("overpass: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	providers.SetUserAgent(req)

	resp, err := p.client.Do(req)
	if err != nil {
		// Transport-level failure (connection refused, DNS, TLS, timeout).
		// Always retry on the next mirror.
		return nil, true, fmt.Errorf("overpass: do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		retryable := resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500
		return nil, retryable, fmt.Errorf("overpass: unexpected status %d", resp.StatusCode)
	}

	var result overpassResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, false, fmt.Errorf("overpass: decode response: %w", err)
	}
	return p.toRawPois(result.Elements), false, nil
}

// escapeOQLString escapes a string for safe embedding in an Overpass QL double-quoted context.
func escapeOQLString(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}

// buildQuery assembles the full Overpass QL query string for the given search mode.
// For district mode it constrains the area lookup using geocoded coordinates when available.
func (p *Provider) buildQuery(q types.SearchQuery) string {
	filters := p.buildFilters(q.Types)
	nodeStmts, wayStmts := p.buildStatements(q, filters)

	if q.Mode == types.ModeDistrict {
		district := escapeOQLString(q.District)
		if q.Lat != 0 || q.Lng != 0 {
			return fmt.Sprintf(
				`[out:json][timeout:7];area["name"="%s"](around:100000,%.6f,%.6f)->.a;(%s) -> .n;(%s) -> .w;.n out 400;.w out center 400;`,
				district, q.Lat, q.Lng,
				strings.Join(nodeStmts, ""),
				strings.Join(wayStmts, ""),
			)
		}
		return fmt.Sprintf(
			`[out:json][timeout:7];area["name"="%s"]->.a;(%s) -> .n;(%s) -> .w;.n out 400;.w out center 400;`,
			district,
			strings.Join(nodeStmts, ""),
			strings.Join(wayStmts, ""),
		)
	}
	return fmt.Sprintf(
		"[out:json][timeout:7];(%s) -> .n;(%s) -> .w;.n out 400;.w out center 400;",
		strings.Join(nodeStmts, ""),
		strings.Join(wayStmts, ""),
	)
}

// buildFilters returns the Overpass tag filters for the requested POI types.
// When no types are requested it returns a broad default covering the most relevant categories.
func (p *Provider) buildFilters(poiTypes []types.PoiType) []string {
	if len(poiTypes) == 0 {
		return []string{
			`["tourism"~"museum|gallery|attraction|viewpoint|castle|ruins|theme_park|zoo|aquarium|hotel|hostel|guest_house|motel|camp_site"]`,
			`["amenity"~"restaurant|cafe|fast_food|bar|pub|nightclub"]`,
		}
	}
	seen := map[string]bool{}
	var filters []string
	for _, t := range poiTypes {
		for _, f := range typeToOsmFilters[t] {
			if !seen[f] {
				filters = append(filters, f)
				seen[f] = true
			}
		}
	}
	return filters
}

// buildStatements returns two slices: node statements and way statements.
// They are kept separate so buildQuery can apply independent output limits.
func (p *Provider) buildStatements(q types.SearchQuery, filters []string) (nodeStmts, wayStmts []string) {
	for _, f := range filters {
		switch q.Mode {
		case types.ModeRadius:
			nodeStmts = append(nodeStmts, fmt.Sprintf(`node(around:%d,%.6f,%.6f)%s["name"];`, q.Radius, q.Lat, q.Lng, f))
			wayStmts = append(wayStmts, fmt.Sprintf(`way(around:%d,%.6f,%.6f)%s["name"];`, q.Radius, q.Lat, q.Lng, f))
		case types.ModePolygon:
			nodeStmts = append(nodeStmts, fmt.Sprintf(`node(poly:"%s")%s["name"];`, q.Polygon, f))
			wayStmts = append(wayStmts, fmt.Sprintf(`way(poly:"%s")%s["name"];`, q.Polygon, f))
		case types.ModeDistrict:
			nodeStmts = append(nodeStmts, fmt.Sprintf(`node(area.a)%s["name"];`, f))
			wayStmts = append(wayStmts, fmt.Sprintf(`way(area.a)%s["name"];`, f))
		}
	}
	return
}

// toRawPois converts Overpass elements to RawPoi records, deduplicating by element type+ID.
func (p *Provider) toRawPois(elements []overpassElement) []types.RawPoi {
	seen := make(map[string]bool, len(elements))
	pois := make([]types.RawPoi, 0, len(elements))
	for _, el := range elements {
		key := fmt.Sprintf("%s:%d", el.Type, el.ID)
		if seen[key] {
			continue
		}
		seen[key] = true
		name := el.Tags["name"]
		if name == "" {
			continue
		}
		lat, lng := el.Lat, el.Lon
		if el.Center != nil {
			lat, lng = el.Center.Lat, el.Center.Lon
		}
		pois = append(pois, types.RawPoi{
			ID:       fmt.Sprintf("overpass:%d", el.ID),
			Name:     name,
			Type:     p.resolveType(el.Tags),
			Provider: types.ProviderOverpass,
			Coords: &types.Coordinates{
				Lat: lat,
				Lng: lng,
			},
			Contact: types.Contact{
				Website: el.Tags["website"],
				Phone:   el.Tags["phone"],
				Hours:   el.Tags["opening_hours"],
			},
			WikidataID: el.Tags["wikidata"],
		})
	}
	return pois
}

// resolveType maps OSM tags to a PoiType by checking tourism, amenity, leisure, and shop keys in order.
func (p *Provider) resolveType(tags map[string]string) types.PoiType {
	for _, key := range []string{"tourism", "amenity", "leisure", "shop"} {
		if v, ok := tags[key]; ok {
			if t, ok := osmTagMap[v]; ok {
				return t
			}
			if key == "shop" {
				return types.TypeBuy
			}
		}
	}
	return types.TypeGeneric
}
