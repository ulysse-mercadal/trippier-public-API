// Package wikipedia implements two Provider adapters for the Wikipedia MediaWiki API:
//   - Provider: geo-located Wikipedia articles filtered to physical places.
//   - EventProvider: geo-located Wikipedia articles filtered to cultural festivals.
//
// Both use the Wikipedia Geosearch API to find articles near a location, enrich
// them with extract and thumbnail, and apply a Wikidata SPARQL classification
// filter to keep only the relevant entity class.
//
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

	"github.com/trippier/poi-api/internal/providers"
	"github.com/trippier/poi-api/pkg/types"
)

const (
	defaultTimeout = 10 * time.Second
	sparqlTimeout  = 5 * time.Second

	// wikidataSPARQL is the Wikidata Query Service SPARQL endpoint.
	wikidataSPARQL = "https://query.wikidata.org/sparql"

	// festivalClass is the root Wikidata class for festivals (music, cultural,
	// film, food, etc.). The events provider keeps only articles whose P31 chain
	// reaches this class.
	festivalClass = "Q132241"
)

// ── Shared infrastructure ─────────────────────────────────────────────────────

type base struct {
	client    *http.Client
	baseURL   string
	sparqlURL string
}

type geosearchPage struct {
	PageID int     `json:"pageid"`
	Title  string  `json:"title"`
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
	Dist   float64 `json:"dist"`
	Type   string  `json:"type"`
}

// geosearch calls the Wikipedia Geosearch API and returns pages near the query coordinates.
func (b *base) geosearch(ctx context.Context, q types.SearchQuery) ([]geosearchPage, error) {
	params := url.Values{
		"action":      {"query"},
		"list":        {"geosearch"},
		"gscoord":     {fmt.Sprintf("%.6f|%.6f", q.Lat, q.Lng)},
		"gsradius":    {strconv.Itoa(q.Radius)},
		"gslimit":     {"50"},
		"gsnamespace": {"0"},
		"gsprop":      {"type"},
		"format":      {"json"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, b.baseURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	providers.SetUserAgent(req)

	resp, err := b.client.Do(req)
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

type enrichedPage struct {
	PageID    int
	Title     string
	Extract   string
	Thumbnail string
	WikidataID string
	Geo       geosearchPage
}

// enrich fetches extracts, thumbnails, and Wikidata IDs for a batch of geosearch pages.
// Falls back to enrichWithoutAPI if the batch request fails.
func (b *base) enrich(ctx context.Context, pages []geosearchPage) []enrichedPage {
	if len(pages) == 0 {
		return nil
	}

	ids := make([]string, len(pages))
	index := make(map[int]geosearchPage, len(pages))
	for i, pg := range pages {
		ids[i] = strconv.Itoa(pg.PageID)
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, b.baseURL+"?"+params.Encode(), nil)
	if err != nil {
		return b.enrichWithoutAPI(pages)
	}
	providers.SetUserAgent(req)

	resp, err := b.client.Do(req)
	if err != nil {
		return b.enrichWithoutAPI(pages)
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
		return b.enrichWithoutAPI(pages)
	}

	out := make([]enrichedPage, 0, len(result.Query.Pages))
	for _, wp := range result.Query.Pages {
		geo, ok := index[wp.PageID]
		if !ok {
			continue
		}
		ep := enrichedPage{
			PageID:     wp.PageID,
			Title:      wp.Title,
			Extract:    wp.Extract,
			WikidataID: wp.PageProps["wikibase_item"],
			Geo:        geo,
		}
		if wp.Thumbnail != nil {
			ep.Thumbnail = wp.Thumbnail.Source
		}
		out = append(out, ep)
	}
	return out
}

// enrichWithoutAPI builds minimal enrichedPage records (title + geo only) when the batch API call is unavailable.
func (b *base) enrichWithoutAPI(pages []geosearchPage) []enrichedPage {
	out := make([]enrichedPage, len(pages))
	for i, pg := range pages {
		out[i] = enrichedPage{
			PageID: pg.PageID,
			Title:  pg.Title,
			Geo:    pg,
		}
	}
	return out
}

// wikidataClassMembers queries the Wikidata SPARQL service and returns the
// subset of the given Wikidata IDs that are instances of wikidataClass (or any
// subclass thereof, via P31/P279*).
//
// On error it returns nil so callers can fail open or closed depending on context.
func (b *base) wikidataClassMembers(ctx context.Context, ids []string, wikidataClass string) map[string]bool {
	if len(ids) == 0 {
		return nil
	}

	values := make([]string, len(ids))
	for i, id := range ids {
		values[i] = "wd:" + id
	}
	query := fmt.Sprintf(
		"SELECT DISTINCT ?item WHERE { VALUES ?item { %s } ?item wdt:P31/wdt:P279* wd:%s . }",
		strings.Join(values, " "),
		wikidataClass,
	)

	sparqlCtx, cancel := context.WithTimeout(ctx, sparqlTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(sparqlCtx, http.MethodGet,
		b.sparqlURL+"?query="+url.QueryEscape(query), nil)
	if err != nil {
		return nil
	}
	req.Header.Set("Accept", "application/sparql-results+json")
	providers.SetUserAgent(req)

	resp, err := b.client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var result struct {
		Results struct {
			Bindings []struct {
				Item struct {
					Value string `json:"value"`
				} `json:"item"`
			} `json:"bindings"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil
	}

	members := make(map[string]bool, len(result.Results.Bindings))
	for _, b := range result.Results.Bindings {
		if i := strings.LastIndex(b.Item.Value, "/"); i >= 0 {
			members[b.Item.Value[i+1:]] = true
		}
	}
	return members
}

// toRawPoi converts an enriched Wikipedia page to a RawPoi of the given type.
func toRawPoi(ep enrichedPage, poiType types.PoiType) types.RawPoi {
	return types.RawPoi{
		ID:          fmt.Sprintf("wikipedia:%d", ep.PageID),
		Name:        ep.Title,
		Type:        poiType,
		Provider:    types.ProviderWikipedia,
		Description: ep.Extract,
		Thumbnail:   ep.Thumbnail,
		Coords:      &types.Coordinates{Lat: ep.Geo.Lat, Lng: ep.Geo.Lon},
		Distance:    ep.Geo.Dist,
		WikidataID:  ep.WikidataID,
	}
}

// ── Provider (physical places) ────────────────────────────────────────────────

// Provider fetches geo-located Wikipedia articles and keeps only those that
// represent physical places (by excluding articles whose Wikidata entity is an
// instance of eventClass, Q1190554). The geosearch type=event articles are
// pre-filtered before the Wikidata check.
type Provider struct{ base }

// New returns a Provider targeting the given Wikipedia language edition.
func New(lang string) *Provider {
	return &Provider{base{
		client:    &http.Client{Timeout: defaultTimeout},
		baseURL:   fmt.Sprintf("https://%s.wikipedia.org/w/api.php", lang),
		sparqlURL: wikidataSPARQL,
	}}
}

// NewWithURLs returns a Provider using custom Wikipedia and SPARQL URLs.
// Intended for testing both endpoints against local httptest servers.
func NewWithURLs(baseURL, sparqlURL string) *Provider {
	return &Provider{base{
		client:    &http.Client{Timeout: defaultTimeout},
		baseURL:   baseURL,
		sparqlURL: sparqlURL,
	}}
}

// Name implements providers.Provider.
func (p *Provider) Name() types.Provider { return types.ProviderWikipedia }

// SupportsMode implements providers.Provider.
func (p *Provider) SupportsMode(mode types.SearchMode) bool {
	return mode == types.ModeRadius || mode == types.ModeDistrict
}

// Search implements providers.Provider.
// NOTE: this provider is not included in AllProviders and is therefore not
// called during standard place searches. It is available for explicit use
// (e.g. enrichment pipelines) and for testing.
func (p *Provider) Search(ctx context.Context, q types.SearchQuery) ([]types.RawPoi, error) {
	pages, err := p.base.geosearch(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("wikipedia: geosearch: %w", err)
	}
	if len(pages) == 0 {
		return nil, nil
	}
	enriched := p.base.enrich(ctx, pages)
	pois := make([]types.RawPoi, 0, len(enriched))
	for _, ep := range enriched {
		pois = append(pois, toRawPoi(ep, types.TypeSee))
	}
	return pois, nil
}

// ── EventProvider (cultural festivals) ───────────────────────────────────────

// EventProvider fetches geo-located Wikipedia articles and keeps only those
// that represent cultural festivals (Wikidata class Q132241 and subclasses:
// music festivals, film festivals, food festivals, carnivals, etc.).
type EventProvider struct{ base }

// NewEventProvider returns an EventProvider targeting the given Wikipedia language edition.
func NewEventProvider(lang string) *EventProvider {
	return &EventProvider{base{
		client:    &http.Client{Timeout: defaultTimeout},
		baseURL:   fmt.Sprintf("https://%s.wikipedia.org/w/api.php", lang),
		sparqlURL: wikidataSPARQL,
	}}
}

// NewEventProviderWithURLs returns an EventProvider using custom Wikipedia and SPARQL URLs.
// Intended for testing both endpoints against local httptest servers.
func NewEventProviderWithURLs(baseURL, sparqlURL string) *EventProvider {
	return &EventProvider{base{
		client:    &http.Client{Timeout: defaultTimeout},
		baseURL:   baseURL,
		sparqlURL: sparqlURL,
	}}
}

// Name implements providers.Provider.
func (p *EventProvider) Name() types.Provider { return types.ProviderWikipediaEvents }

// SupportsMode implements providers.Provider.
func (p *EventProvider) SupportsMode(mode types.SearchMode) bool {
	return mode == types.ModeRadius || mode == types.ModeDistrict
}

// Search implements providers.Provider.
// Returns only articles classified as cultural festivals in Wikidata.
// Articles without a Wikidata ID are dropped (cannot be classified).
func (p *EventProvider) Search(ctx context.Context, q types.SearchQuery) ([]types.RawPoi, error) {
	pages, err := p.base.geosearch(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("wikipedia_events: geosearch: %w", err)
	}
	if len(pages) == 0 {
		return nil, nil
	}

	enriched := p.base.enrich(ctx, pages)

	wikidataIDs := make([]string, 0, len(enriched))
	for _, ep := range enriched {
		if ep.WikidataID != "" {
			wikidataIDs = append(wikidataIDs, ep.WikidataID)
		}
	}

	festivalIDs := p.base.wikidataClassMembers(ctx, wikidataIDs, festivalClass)
	if festivalIDs == nil {
		return nil, nil
	}

	pois := make([]types.RawPoi, 0)
	for _, ep := range enriched {
		if ep.WikidataID == "" || !festivalIDs[ep.WikidataID] {
			continue
		}
		poi := toRawPoi(ep, types.TypeEvent)
		poi.Provider = types.ProviderWikipediaEvents
		poi.Recurring = true
		pois = append(pois, poi)
	}
	return pois, nil
}
