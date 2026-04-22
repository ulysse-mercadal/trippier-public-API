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

	"github.com/trippier/poi-api/pkg/types"
)

const (
	defaultAPIURL  = "https://overpass-api.de/api/interpreter"
	defaultTimeout = 10 * time.Second
)

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
	types.TypeSee:   {`["tourism"~"museum|gallery|attraction|artwork|viewpoint|castle|ruins|monument"]`, `["historic"]`},
	types.TypeEat:   {`["amenity"~"restaurant|cafe|fast_food|food_court"]`},
	types.TypeDrink: {`["amenity"~"bar|pub|nightclub|biergarten"]`},
	types.TypeSleep: {`["tourism"~"hotel|hostel|guest_house|motel|camp_site"]`},
	types.TypeDo:    {`["tourism"~"theme_park|zoo|aquarium"]`, `["leisure"~"park|sports_centre"]`},
	types.TypeBuy:   {`["shop"]`, `["amenity"="marketplace"]`},
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
type Provider struct {
	client *http.Client
	apiURL string
}

// New returns a Provider targeting the public Overpass API.
func New() *Provider {
	return &Provider{
		client: &http.Client{Timeout: defaultTimeout},
		apiURL: defaultAPIURL,
	}
}

// NewWithURL returns a Provider targeting a custom Overpass endpoint.
// Intended for testing against a local httptest server.
func NewWithURL(url string) *Provider {
	return &Provider{
		client: &http.Client{Timeout: defaultTimeout},
		apiURL: url,
	}
}

// Name implements providers.Provider.
func (p *Provider) Name() types.Provider { return types.ProviderOverpass }

// SupportsMode implements providers.Provider. Overpass supports all search modes.
func (p *Provider) SupportsMode(_ types.SearchMode) bool { return true }

// Search implements providers.Provider.
func (p *Provider) Search(ctx context.Context, q types.SearchQuery) ([]types.RawPoi, error) {
	body := url.Values{"data": {p.buildQuery(q)}}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.apiURL, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("overpass: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("overpass: do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("overpass: unexpected status %d", resp.StatusCode)
	}

	var result overpassResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("overpass: decode response: %w", err)
	}

	return p.toRawPois(result.Elements), nil
}

func (p *Provider) buildQuery(q types.SearchQuery) string {
	filters := p.buildFilters(q.Types)
	statements := p.buildStatements(q, filters)
	return fmt.Sprintf("[out:json][timeout:25];(%s);out center;", strings.Join(statements, ""))
}

func (p *Provider) buildFilters(poiTypes []types.PoiType) []string {
	if len(poiTypes) == 0 {
		return []string{`["name"]["tourism"]`, `["name"]["amenity"~"restaurant|cafe|bar|pub"]`}
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

func (p *Provider) buildStatements(q types.SearchQuery, filters []string) []string {
	var stmts []string
	for _, f := range filters {
		switch q.Mode {
		case types.ModeRadius:
			stmts = append(stmts,
				fmt.Sprintf(`node(around:%d,%.6f,%.6f)%s["name"];`, q.Radius, q.Lat, q.Lng, f),
				fmt.Sprintf(`way(around:%d,%.6f,%.6f)%s["name"];`, q.Radius, q.Lat, q.Lng, f),
			)
		case types.ModePolygon:
			stmts = append(stmts,
				fmt.Sprintf(`node(poly:"%s")%s["name"];`, q.Polygon, f),
				fmt.Sprintf(`way(poly:"%s")%s["name"];`, q.Polygon, f),
			)
		case types.ModeDistrict:
			stmts = append(stmts,
				fmt.Sprintf(`area["name"="%s"]->.a;node(area.a)%s["name"];`, q.District, f),
				fmt.Sprintf(`area["name"="%s"]->.a;way(area.a)%s["name"];`, q.District, f),
			)
		}
	}
	return stmts
}

func (p *Provider) toRawPois(elements []overpassElement) []types.RawPoi {
	pois := make([]types.RawPoi, 0, len(elements))
	for _, el := range elements {
		name := el.Tags["name"]
		if name == "" {
			continue
		}
		lat, lng := el.Lat, el.Lon
		if el.Center != nil {
			lat, lng = el.Center.Lat, el.Center.Lon
		}
		pois = append(pois, types.RawPoi{
			ID:          fmt.Sprintf("overpass:%d", el.ID),
			Name:        name,
			Type:        p.resolveType(el.Tags),
			Provider:    types.ProviderOverpass,
			Coords:      &types.Coordinates{
				Lat:     lat,
				Lng:     lng,
			},
			Contact:     types.Contact{
				Website: el.Tags["website"],
				Phone:   el.Tags["phone"],
				Hours:   el.Tags["opening_hours"],
			},
			Tags:        el.Tags,
			WikidataID:  el.Tags["wikidata"],
		})
	}
	return pois
}

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
