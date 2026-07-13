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
	"github.com/trippier/poi-api/internal/tilecache"
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

// Provider fetches POIs from the OpenStreetMap Overpass API, rotating across mirrors on transport errors, 429s, or 5xxs.
type Provider struct {
	client  *http.Client
	apiURLs []string
}

// New creates a Provider rotating across the public Overpass mirrors,
// returning a provider using the default mirror list.
func New() *Provider {
	urls := make([]string, len(defaultAPIURLs))
	copy(urls, defaultAPIURLs)
	return &Provider{
		client:  &http.Client{Timeout: defaultTimeout},
		apiURLs: urls,
	}
}

// NewWithURL creates a Provider targeting a single custom Overpass endpoint u,
// intended for testing against a local httptest server. It returns a
// provider bound to the given endpoint.
func NewWithURL(u string) *Provider {
	return &Provider{
		client:  &http.Client{Timeout: defaultTimeout},
		apiURLs: []string{u},
	}
}

// NewWithURLs creates a Provider rotating across the given Overpass endpoint
// urls, returning a provider using the given mirror list.
func NewWithURLs(urls []string) *Provider {
	cp := make([]string, len(urls))
	copy(cp, urls)
	return &Provider{
		client:  &http.Client{Timeout: defaultTimeout},
		apiURLs: cp,
	}
}

// Name returns this provider's identifier.
func (p *Provider) Name() types.Provider { return types.ProviderOverpass }

// SupportsMode reports whether Overpass supports the given search mode; it
// always returns true, since Overpass supports every mode.
func (p *Provider) SupportsMode(_ types.SearchMode) bool { return true }

// Search queries Overpass for POIs matching q, trying each mirror in order
// (using ctx for cancellation and deadlines) and retrying only on transient
// failures (transport errors, 429, 5xx). It returns the matching POIs, or an
// error if all mirrors failed.
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

// searchOnce performs a single POST of body against apiURL, using ctx for
// cancellation and deadlines. It returns the matching POIs, whether the
// caller should retry another mirror, and an error.
func (p *Provider) searchOnce(ctx context.Context, apiURL, body string) ([]types.RawPoi, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(body))
	if err != nil {
		return nil, false, fmt.Errorf("overpass: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	providers.SetUserAgent(req)

	resp, err := p.client.Do(req)
	if err != nil {
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

// escapeOQLString escapes the raw string s for safe embedding in an Overpass
// QL double-quoted context, returning the escaped string.
func escapeOQLString(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}

// buildQuery assembles the full Overpass QL query string for the search mode
// and parameters in q, returning the complete Overpass QL query string.
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

// buildFilters returns the Overpass tag filter expressions for the requested
// poiTypes, or a broad default list when none are given.
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

// buildStatements builds Overpass node and way statements for query q using
// the given filters, keeping them separate so buildQuery can apply
// independent output limits. It returns the node statements and way
// statements.
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

// toRawPois converts the raw Overpass elements to RawPoi records,
// deduplicating by element type+ID, and returns the resulting records.
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
				Website: providers.SafeURL(el.Tags["website"]),
				Phone:   el.Tags["phone"],
				Hours:   el.Tags["opening_hours"],
			},
			WikidataID: el.Tags["wikidata"],
			Images:     osmImages(el.Tags),
			SourceURL:  osmURL(el.Type, el.ID),
		})
	}
	return pois
}

// osmImages gathers up to 3 image URLs from the image, image:url, and
// wikimedia_commons entries in tags, returning them or nil if none are found.
func osmImages(tags map[string]string) []string {
	out := make([]string, 0, 3)
	seen := make(map[string]bool, 3)
	for _, k := range []string{"image", "image:url"} {
		v := strings.TrimSpace(tags[k])
		if v == "" || seen[v] {
			continue
		}
		if !strings.HasPrefix(v, "http://") && !strings.HasPrefix(v, "https://") {
			continue
		}
		out = append(out, v)
		seen[v] = true
	}
	if wc := strings.TrimSpace(tags["wikimedia_commons"]); wc != "" {
		if u := providers.CommonsFileURL(wc); u != "" && !seen[u] {
			out = append(out, u)
			seen[u] = true
		}
	}
	if len(out) > 3 {
		out = out[:3]
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// osmURL returns the canonical openstreetmap.org browse URL for the element
// of the given elementType ("node", "way", or "relation") and id, or ""
// for unrecognized element types.
func osmURL(elementType string, id int64) string {
	switch elementType {
	case "node", "way", "relation":
		return fmt.Sprintf("https://www.openstreetmap.org/%s/%d", elementType, id)
	}
	return ""
}

// resolveType maps the given OSM element tags to a PoiType by checking the
// tourism, amenity, leisure, and shop keys in order, returning the resolved
// type or TypeGeneric if none match.
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

// init registers this provider with the providers registry.
func init() {
	providers.Register(types.ProviderOverpass, func(cfg providers.BuildConfig) (providers.Provider, error) {
		return tilecache.NewCachedProvider(New(), cfg.Redis, cfg.CacheTTL, cfg.Log), nil
	})
}
