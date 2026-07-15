// Package wikipedia implements Provider (physical places) and EventProvider
// (cultural festivals) adapters over the Wikipedia Geosearch API, classified
// via Wikidata SPARQL.
package wikipedia

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

	"github.com/trippier/poi-api/internal/providers"
	"github.com/trippier/poi-api/pkg/types"
)

// wikidataIDRe matches the canonical Wikidata entity ID shape (Q followed by digits); used to sanitize values before SPARQL interpolation.
var wikidataIDRe = regexp.MustCompile(`^Q\d+$`)

const (
	defaultTimeout = 10 * time.Second
	sparqlTimeout  = 5 * time.Second

	// wikidataSPARQL is the Wikidata Query Service SPARQL endpoint.
	wikidataSPARQL = "https://query.wikidata.org/sparql"

	// festivalClass is the root Wikidata class for festivals; events provider keeps only articles under it.
	festivalClass = "Q132241"

	// apiTemplate is the MediaWiki API endpoint pattern; %s is the language edition subdomain.
	apiTemplate = "https://%s.wikipedia.org/w/api.php"
)

type base struct {
	client  *http.Client
	baseURL string
	// apiTemplate, when non-empty, lets Search retarget baseURL to the
	// per-request language edition. It is empty for test constructors, which
	// pin a fixed baseURL.
	apiTemplate string
	sparqlURL   string
}

// forLang returns a copy of b whose baseURL targets the requested language
// edition, or b unchanged when lang is empty/invalid or the endpoint is pinned
// (tests). The shared http.Client is reused.
//
// @param lang the caller-supplied language code (e.g. from ?lang=fr).
// @return the base to use for this request.
func (b *base) forLang(lang string) *base {
	code := providers.NormalizeLang(lang)
	if code == "" || b.apiTemplate == "" {
		return b
	}
	nb := *b
	nb.baseURL = fmt.Sprintf(b.apiTemplate, code)
	return &nb
}

type geosearchPage struct {
	PageID int     `json:"pageid"`
	Title  string  `json:"title"`
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
	Dist   float64 `json:"dist"`
	Type   string  `json:"type"`
}

// geosearch calls the Wikipedia Geosearch API for pages near the query
// coordinates and radius given in q, using ctx for cancellation. It returns
// the matching geosearch pages, or an error.
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
	PageID     int
	Title      string
	Extract    string
	Thumbnail  string
	WikidataID string
	Geo        geosearchPage
}

// enrich fetches extracts, thumbnails, and Wikidata IDs for the given batch
// of geosearch pages, using ctx for cancellation, falling back to
// enrichWithoutAPI on failure. It returns the enriched pages.
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

// enrichWithoutAPI builds minimal enrichedPage records (title + geo only)
// from pages when the batch API call is unavailable. It returns the
// minimally enriched pages.
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

// wikidataClassMembers checks, using ctx for cancellation, which of ids are
// instances of wikidataClass (or a subclass thereof, via P31/P279*). It
// returns the set of ids that belong to wikidataClass, or nil on error.
func (b *base) wikidataClassMembers(ctx context.Context, ids []string, wikidataClass string) map[string]bool {
	if len(ids) == 0 {
		return nil
	}

	values := make([]string, 0, len(ids))
	for _, id := range ids {
		if !wikidataIDRe.MatchString(id) {
			continue
		}
		values = append(values, "wd:"+id)
	}
	if len(values) == 0 || !wikidataIDRe.MatchString(wikidataClass) {
		return nil
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

// toRawPoi converts the enriched Wikipedia page ep to a RawPoi, tagging it
// with poiType. It returns the converted RawPoi.
func (b *base) toRawPoi(ep enrichedPage, poiType types.PoiType) types.RawPoi {
	var images []string
	if ep.Thumbnail != "" {
		images = []string{ep.Thumbnail}
	}
	return types.RawPoi{
		ID:          fmt.Sprintf("wikipedia:%d", ep.PageID),
		Name:        ep.Title,
		Type:        poiType,
		Provider:    types.ProviderWikipedia,
		Description: ep.Extract,
		Thumbnail:   ep.Thumbnail,
		Images:      images,
		Coords:      &types.Coordinates{Lat: ep.Geo.Lat, Lng: ep.Geo.Lon},
		Distance:    ep.Geo.Dist,
		WikidataID:  ep.WikidataID,
		SourceURL:   articleURL(b.baseURL, ep.PageID),
	}
}

// articleURL derives the canonical article URL for pageID from the
// MediaWiki API endpoint mediawikiURL. It returns the canonical article
// URL, or "" if the host cannot be extracted.
func articleURL(mediawikiURL string, pageID int) string {
	if i := strings.Index(mediawikiURL, "/w/api.php"); i > 0 {
		return fmt.Sprintf("%s/?curid=%d", mediawikiURL[:i], pageID)
	}
	return ""
}

// Provider fetches geo-located Wikipedia articles and keeps only those representing physical places.
type Provider struct{ base }

// New returns a Provider targeting the Wikipedia language edition lang.
func New(lang string) *Provider {
	return &Provider{base{
		client:      &http.Client{Timeout: defaultTimeout},
		baseURL:     fmt.Sprintf(apiTemplate, lang),
		apiTemplate: apiTemplate,
		sparqlURL:   wikidataSPARQL,
	}}
}

// NewWithURLs returns a Provider using the given Wikipedia API base URL
// baseURL and SPARQL query endpoint sparqlURL, for testing against local
// httptest servers.
func NewWithURLs(baseURL, sparqlURL string) *Provider {
	return &Provider{base{
		client:    &http.Client{Timeout: defaultTimeout},
		baseURL:   baseURL,
		sparqlURL: sparqlURL,
	}}
}

// Name implements providers.Provider. It returns the provider identifier.
func (p *Provider) Name() types.Provider { return types.ProviderWikipedia }

// SupportsMode implements providers.Provider. It reports whether mode is
// supported.
func (p *Provider) SupportsMode(mode types.SearchMode) bool {
	return mode == types.ModeRadius || mode == types.ModeDistrict
}

// enrichmentRadiusMeters is the max distance to consider a Wikipedia article the same place as a POI from another provider.
const enrichmentRadiusMeters = 50.0

// EnrichmentRadius implements providers.Enricher. It returns the
// enrichment radius in meters.
func (p *Provider) EnrichmentRadius() float64 { return enrichmentRadiusMeters }

// EnrichTarget implements providers.Enricher. It fills a missing WikidataID
// on target and, for GeoNames targets, backfills SourceURL and Description
// on target from source, the Wikipedia POI providing enrichment data.
func (p *Provider) EnrichTarget(target *types.RawPoi, source types.RawPoi) {
	if target.WikidataID == "" && source.WikidataID != "" {
		target.WikidataID = source.WikidataID
	}
	if target.Provider == types.ProviderGeoNames && source.SourceURL != "" {
		target.SourceURL = source.SourceURL
	}
	if target.Provider == types.ProviderGeoNames && target.Description == "" && source.Description != "" {
		target.Description = source.Description
	}
}

// Search implements providers.Provider. It is not included in
// AllProviders; it is used explicitly (e.g. enrichment pipelines) and in
// tests. It searches near the coordinates and radius in q, using ctx for
// cancellation, and returns matching POIs, or an error.
func (p *Provider) Search(ctx context.Context, q types.SearchQuery) ([]types.RawPoi, error) {
	b := p.forLang(q.Lang)
	pages, err := b.geosearch(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("wikipedia: geosearch: %w", err)
	}
	if len(pages) == 0 {
		return nil, nil
	}
	enriched := b.enrich(ctx, pages)
	pois := make([]types.RawPoi, 0, len(enriched))
	for _, ep := range enriched {
		pois = append(pois, b.toRawPoi(ep, types.TypeSee))
	}
	return pois, nil
}

// EventProvider fetches geo-located Wikipedia articles and keeps only those representing cultural festivals.
type EventProvider struct{ base }

// NewEventProvider returns an EventProvider targeting the Wikipedia
// language edition lang.
func NewEventProvider(lang string) *EventProvider {
	return &EventProvider{base{
		client:      &http.Client{Timeout: defaultTimeout},
		baseURL:     fmt.Sprintf(apiTemplate, lang),
		apiTemplate: apiTemplate,
		sparqlURL:   wikidataSPARQL,
	}}
}

// NewEventProviderWithURLs returns an EventProvider using the given
// Wikipedia API base URL baseURL and SPARQL query endpoint sparqlURL, for
// testing against local httptest servers.
func NewEventProviderWithURLs(baseURL, sparqlURL string) *EventProvider {
	return &EventProvider{base{
		client:    &http.Client{Timeout: defaultTimeout},
		baseURL:   baseURL,
		sparqlURL: sparqlURL,
	}}
}

// Name implements providers.Provider. It returns the provider identifier.
func (p *EventProvider) Name() types.Provider { return types.ProviderWikipediaEvents }

// SupportsMode implements providers.Provider. It reports whether mode is
// supported.
func (p *EventProvider) SupportsMode(mode types.SearchMode) bool {
	return mode == types.ModeRadius || mode == types.ModeDistrict
}

// Search implements providers.Provider. It searches near the coordinates
// and radius in q, using ctx for cancellation, and returns only articles
// classified as cultural festivals in Wikidata (articles without a
// Wikidata ID are dropped), or an error.
func (p *EventProvider) Search(ctx context.Context, q types.SearchQuery) ([]types.RawPoi, error) {
	b := p.forLang(q.Lang)
	pages, err := b.geosearch(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("wikipedia_events: geosearch: %w", err)
	}
	if len(pages) == 0 {
		return nil, nil
	}

	enriched := b.enrich(ctx, pages)

	wikidataIDs := make([]string, 0, len(enriched))
	for _, ep := range enriched {
		if ep.WikidataID != "" {
			wikidataIDs = append(wikidataIDs, ep.WikidataID)
		}
	}

	festivalIDs := b.wikidataClassMembers(ctx, wikidataIDs, festivalClass)
	if festivalIDs == nil {
		return nil, nil
	}

	pois := make([]types.RawPoi, 0)
	for _, ep := range enriched {
		if ep.WikidataID == "" || !festivalIDs[ep.WikidataID] {
			continue
		}
		poi := b.toRawPoi(ep, types.TypeEvent)
		poi.Provider = types.ProviderWikipediaEvents
		poi.Recurring = true
		pois = append(pois, poi)
	}
	return pois, nil
}

// init registers the Wikipedia and Wikipedia-events providers with the
// global registry.
func init() {
	providers.Register(types.ProviderWikipedia, func(cfg providers.BuildConfig) (providers.Provider, error) {
		return New(cfg.Lang), nil
	})
	providers.Register(types.ProviderWikipediaEvents, func(cfg providers.BuildConfig) (providers.Provider, error) {
		return NewEventProvider(cfg.Lang), nil
	})
}
