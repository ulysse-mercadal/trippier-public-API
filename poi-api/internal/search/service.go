// Package search orchestrates POI search across providers: selection, fetching, dedup, scoring, and pagination.
package search

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/trippier/poi-api/internal/dedup"
	"github.com/trippier/poi-api/internal/geo"
	"github.com/trippier/poi-api/internal/providers"
	"github.com/trippier/poi-api/internal/registry"
	"github.com/trippier/poi-api/internal/scoring"
	"github.com/trippier/poi-api/pkg/types"
)

// autoSelectThreshold is the minimum composite score for a provider to be auto-selected.
const autoSelectThreshold = 0.10

// Service orchestrates the full POI search pipeline.
type Service struct {
	providers       map[types.Provider]providers.Provider
	providerTimeout time.Duration
	log             *zap.Logger
	// selectionCache memoises selectByCountry results for the common
	// no-override path (no per-request weight overrides, no exclusions). Cache
	// values are immutable []types.Provider slices keyed by the requested
	// (country, types-set, forEvents) tuple. Sized by registry × type-subset
	// cardinality, no eviction needed at realistic scale.
	selectionCache sync.Map
}

// NewService builds a Service backed by the given providers pp, keyed
// internally by name, using timeout as the per-provider search timeout and
// log for pipeline diagnostics. It returns the constructed Service.
func NewService(pp []providers.Provider, timeout time.Duration, log *zap.Logger) *Service {
	m := make(map[types.Provider]providers.Provider, len(pp))
	for _, p := range pp {
		m[p.Name()] = p
	}
	return &Service{providers: m, providerTimeout: timeout, log: log}
}

// selectionCacheKey is the immutable key selectByCountry results are memoised under.
type selectionCacheKey struct {
	cc    string
	types string
	kind  types.PointKind
}

// Search runs the full POI pipeline with geo-aware provider auto-selection,
// using ctx as the request context and q as the search query. When the
// caller supplies ?providers=…, that explicit list is used as-is. Otherwise
// providers are chosen from the registry by country + category scores. It
// returns the paginated POI search result.
func (s *Service) Search(ctx context.Context, q types.SearchQuery) (*types.SearchResult, error) {
	userSpecified := len(q.Providers) > 0
	applyDefaults(&q, defaultProviders())
	if !userSpecified {
		if selected := s.autoSelectProviders(ctx, q, types.KindPOI); len(selected) > 0 {
			q.Providers = selected
		}
	}
	merged := filterByKind(s.pipeline(ctx, &q), types.KindPOI)
	filtered := applyFilters(merged, q)
	return paginate(filtered, q), nil
}

// SearchCustom is the fully-controllable variant of Search, running with ctx
// as the request context and q as the search query, respecting
// q.CountryHint, q.ExcludeProviders and q.ProviderWeights overrides. It
// returns the paginated POI search result.
func (s *Service) SearchCustom(ctx context.Context, q types.SearchQuery) (*types.SearchResult, error) {
	userSpecified := len(q.Providers) > 0
	applyDefaults(&q, defaultProviders())
	if !userSpecified {
		if selected := s.autoSelectProviders(ctx, q, types.KindPOI); len(selected) > 0 {
			q.Providers = selected
		}
	}
	q.Providers = filterExcluded(q.Providers, q.ExcludeProviders)
	merged := filterByKind(s.pipeline(ctx, &q), types.KindPOI)
	filtered := applyFilters(merged, q)
	return paginate(filtered, q), nil
}

// SearchEvents runs the POI pipeline restricted to event providers, using
// ctx as the request context and q as the search query, stretching the
// radius to the maximum MinRadius declared by any selected provider. It
// returns the paginated event search result.
func (s *Service) SearchEvents(ctx context.Context, q types.SearchQuery) (*types.SearchResult, error) {
	applyDefaults(&q, defaultEventProviders())
	clampRadiusToMin(&q)
	merged := filterByKind(s.pipeline(ctx, &q), types.KindEvent)
	return paginate(merged, q), nil
}

// SearchEventsCustom is the fully-controllable variant of SearchEvents,
// running with ctx as the request context and q as the search query with
// optional overrides. It returns the paginated event search result.
func (s *Service) SearchEventsCustom(ctx context.Context, q types.SearchQuery) (*types.SearchResult, error) {
	userSpecified := len(q.Providers) > 0
	applyDefaults(&q, defaultEventProviders())
	clampRadiusToMin(&q)
	if !userSpecified {
		if selected := s.autoSelectProviders(ctx, q, types.KindEvent); len(selected) > 0 {
			q.Providers = selected
		}
	}
	q.Providers = filterExcluded(q.Providers, q.ExcludeProviders)
	merged := filterByKind(s.pipeline(ctx, &q), types.KindEvent)
	return paginate(merged, q), nil
}

// defaultProviders lists non-BYOK, non-event registry providers with a
// non-zero country-wildcard score. It returns sorted default provider IDs.
func defaultProviders() []types.Provider {
	out := make([]types.Provider, 0, len(registry.All))
	for id, meta := range registry.All {
		if meta.Byok || !meta.Provides(types.KindPOI) {
			continue
		}
		if meta.CountryScore("*") <= 0 {
			continue
		}
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return string(out[i]) < string(out[j]) })
	return out
}

// defaultEventProviders lists all registry event providers, including BYOK
// ones (each checks its own key). It returns sorted default event provider
// IDs.
func defaultEventProviders() []types.Provider {
	out := make([]types.Provider, 0, len(registry.All))
	for id, meta := range registry.All {
		if !meta.Provides(types.KindEvent) {
			continue
		}
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return string(out[i]) < string(out[j]) })
	return out
}

// clampRadiusToMin stretches q.Radius up to the maximum MinRadius declared
// by any provider in q.Providers, mutating q in place.
func clampRadiusToMin(q *types.SearchQuery) {
	maxMin := 0
	for _, id := range q.Providers {
		if m := registry.All[id].MinRadius; m > maxMin {
			maxMin = m
		}
	}
	if q.Radius < maxMin {
		q.Radius = maxMin
	}
}

// ProvidersStatus lists registered providers with their static metadata;
// upstreams are never probed (avoids burning paid quota). The context
// argument is unused. It returns the provider statuses.
func (s *Service) ProvidersStatus(_ context.Context) []types.ProviderStatus {
	statuses := make([]types.ProviderStatus, 0, len(s.providers))
	for name, p := range s.providers {
		_, isByok := p.(providers.ByokProvider)
		statuses = append(statuses, types.ProviderStatus{Name: name, Byok: isByok})
	}
	return statuses
}

// ProvidersCatalog returns the full registry merged with runtime
// implementation status, as sorted provider catalog entries.
func (s *Service) ProvidersCatalog() []types.ProviderCatalogEntry {
	entries := make([]types.ProviderCatalogEntry, 0, len(registry.All))
	for id, meta := range registry.All {
		_, implemented := s.providers[id]
		entries = append(entries, types.ProviderCatalogEntry{
			ID:             id,
			Label:          meta.Label,
			Byok:           meta.Byok,
			ByokHeader:     meta.ByokHeader,
			Kinds:          meta.Kinds,
			Categories:     meta.Categories,
			CountryScores:  meta.CountryScores,
			CategoryScores: meta.CategoryScores,
			Implemented:    implemented,
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		return string(entries[i].ID) < string(entries[j].ID)
	})
	return entries
}

// ProvidersRecommend scores registry providers for the location (lat, lng)
// and requestedTypes, using ctx for country detection, restricting
// candidates to those yielding kind, and keeping at most limit entries (0
// means unlimited). It returns the ranked recommendation result for the
// detected country, sorted by composite score descending.
func (s *Service) ProvidersRecommend(ctx context.Context, lat, lng float64, kind types.PointKind, requestedTypes []types.PoiType, limit int) types.RecommendResult {
	cc, _ := geo.CountryCode(ctx, lat, lng)

	type scored struct {
		entry types.RecommendedProvider
		score float64
	}
	candidates := make([]scored, 0, len(registry.All))

	for id, meta := range registry.All {
		if !meta.Provides(kind) {
			continue
		}
		score := meta.Score(cc, requestedTypes)
		_, implemented := s.providers[id]
		candidates = append(candidates, scored{
			entry: types.RecommendedProvider{
				ID:          id,
				Label:       meta.Label,
				Score:       score,
				Byok:        meta.Byok,
				ByokHeader:  meta.ByokHeader,
				Kinds:       meta.Kinds,
				Implemented: implemented,
			},
			score: score,
		})
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	if limit > 0 && len(candidates) > limit {
		candidates = candidates[:limit]
	}
	result := make([]types.RecommendedProvider, len(candidates))
	for i, c := range candidates {
		result[i] = c.entry
	}
	return types.RecommendResult{CountryCode: cc, Providers: result}
}

// autoSelectProviders picks non-BYOK registry providers for q's location and
// types, using ctx for geocoding/country detection, respecting
// q.CountryHint and q.ProviderWeights, and restricting candidates to those
// yielding kind. It returns the selected provider IDs, best score first.
func (s *Service) autoSelectProviders(ctx context.Context, q types.SearchQuery, kind types.PointKind) []types.Provider {
	lat, lng := q.Lat, q.Lng
	if q.Mode == types.ModeDistrict && lat == 0 && lng == 0 {
		if place, err := geo.GeocodeDistrict(ctx, q.District); err == nil {
			lat, lng = place.Lat, place.Lng
		}
	}

	cc := strings.ToUpper(q.CountryHint)
	if cc == "" && (lat != 0 || lng != 0) {
		if detected, err := geo.CountryCode(ctx, lat, lng); err == nil {
			cc = detected
		} else {
			s.log.Debug("country detection failed, using global defaults", zap.Error(err))
		}
	}

	return s.selectByCountry(cc, q.Types, q.ProviderWeights, q.ExcludeProviders, kind)
}

// selectByCountry scores and filters registry providers for the detected or
// hinted country cc, using requestedTypes for category scoring,
// weightOverrides as per-provider score overrides, exclude as provider IDs
// to drop from the result, and kind to restrict candidates to those
// yielding that point kind. The no-override result path is memoised. It
// returns the selected provider IDs, best score first.
func (s *Service) selectByCountry(
	cc string,
	requestedTypes []types.PoiType,
	weightOverrides map[types.Provider]float64,
	exclude []types.Provider,
	kind types.PointKind,
) []types.Provider {
	cacheable := len(weightOverrides) == 0 && len(exclude) == 0
	var cacheK selectionCacheKey
	if cacheable {
		cacheK = makeSelectionKey(cc, requestedTypes, kind)
		if v, ok := s.selectionCache.Load(cacheK); ok {
			return v.([]types.Provider)
		}
	}

	excludeSet := make(map[types.Provider]bool, len(exclude))
	for _, e := range exclude {
		excludeSet[e] = true
	}

	type candidate struct {
		id    types.Provider
		score float64
	}
	list := make([]candidate, 0, len(registry.All))

	for id, meta := range registry.All {
		if !meta.Provides(kind) {
			continue
		}
		if meta.Byok {
			continue
		}
		if _, ok := s.providers[id]; !ok {
			continue
		}
		if excludeSet[id] {
			continue
		}

		countryScore := meta.CountryScore(cc)
		if ow, ok := weightOverrides[id]; ok {
			countryScore = ow
		}
		score := countryScore * meta.CategoryScore(requestedTypes)
		if score < autoSelectThreshold {
			continue
		}
		list = append(list, candidate{id, score})
	}

	sort.Slice(list, func(i, j int) bool { return list[i].score > list[j].score })

	out := make([]types.Provider, len(list))
	for i, c := range list {
		out[i] = c.id
	}
	if cacheable {
		s.selectionCache.Store(cacheK, out)
	}
	return out
}

// makeSelectionKey builds the stable cache key for selectByCountry from cc
// as the country code component, requestedTypes sorted before joining into
// the key (so caller order doesn't fragment the cache), and kind as the
// point kind component. It returns the resulting cache key.
func makeSelectionKey(cc string, requestedTypes []types.PoiType, kind types.PointKind) selectionCacheKey {
	parts := make([]string, len(requestedTypes))
	for i, t := range requestedTypes {
		parts[i] = string(t)
	}
	sort.Strings(parts)
	return selectionCacheKey{cc: cc, types: strings.Join(parts, ","), kind: kind}
}

// filterExcluded removes the providers in exclude from pp, returning pp
// with excluded providers removed.
func filterExcluded(pp []types.Provider, exclude []types.Provider) []types.Provider {
	if len(exclude) == 0 {
		return pp
	}
	ex := make(map[types.Provider]bool, len(exclude))
	for _, e := range exclude {
		ex[e] = true
	}
	out := pp[:0]
	for _, p := range pp {
		if !ex[p] {
			out = append(out, p)
		}
	}
	return out
}

// pipeline fetches from all providers, geocoding district queries with ctx
// and possibly mutating q with the resulting coordinates, then
// deduplicates, scores, and sorts the results. It returns the enriched,
// deduplicated, scored, and sorted POIs.
func (s *Service) pipeline(ctx context.Context, q *types.SearchQuery) []types.EnrichedPoi {
	if q.Mode == types.ModeDistrict {
		if place, err := geo.GeocodeDistrict(ctx, q.District); err == nil {
			q.Lat = place.Lat
			q.Lng = place.Lng
		} else {
			s.log.Warn("geocode district failed", zap.String("district", q.District), zap.Error(err))
		}
	}

	raw := s.fetchAll(ctx, *q)
	raw = geo.SetDistances(raw, q.Lat, q.Lng)
	if q.Mode == types.ModeRadius {
		raw = geo.FilterByRadius(raw, q.Lat, q.Lng, float64(q.Radius))
	}
	raw = s.enrichRaw(ctx, raw, *q)
	merged := dedup.Merge(raw)
	for i := range merged {
		merged[i].Score = scoring.Score(merged[i], *q)
	}
	sort.Slice(merged, func(i, j int) bool { return merged[i].Score > merged[j].Score })
	return merged
}

// fanOutLimit caps concurrent provider Search calls per request.
const fanOutLimit = 16

// fetchAll fans out the search query q to all selected providers
// concurrently under ctx, capped at fanOutLimit in-flight calls. It returns
// the raw POIs from the selected providers.
func (s *Service) fetchAll(ctx context.Context, q types.SearchQuery) []types.RawPoi {
	selected := s.selectProviders(q)
	results := make([][]types.RawPoi, len(selected))

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(fanOutLimit)

	for i, p := range selected {
		i, p := i, p
		g.Go(func() error {
			pctx, cancel := context.WithTimeout(gctx, s.providerTimeout)
			defer cancel()
			pois, err := p.Search(pctx, q)
			if err != nil {
				s.log.Warn("provider error", zap.String("provider", string(p.Name())), zap.Error(err))
				return nil //nolint:nilerr // we never want a single provider to fail the whole request
			}
			results[i] = tagKinds(pois, p.Name())
			return nil
		})
	}
	_ = g.Wait()

	total := 0
	for _, r := range results {
		total += len(r)
	}
	all := make([]types.RawPoi, 0, total)
	for _, r := range results {
		all = append(all, r...)
	}
	return filterToSelectedProviders(all, q.Providers)
}

// filterToSelectedProviders drops entries from raw whose Provider isn't in
// selected. It returns raw restricted to the selected providers.
func filterToSelectedProviders(raw []types.RawPoi, selected []types.Provider) []types.RawPoi {
	if len(raw) == 0 {
		return raw
	}
	allowed := make(map[types.Provider]bool, len(selected))
	for _, p := range selected {
		allowed[p] = true
	}
	out := raw[:0]
	for _, p := range raw {
		if allowed[p.Provider] {
			out = append(out, p)
		}
	}
	return out
}

// tagKinds stamps name's declared kind onto entries in pois that left Kind
// unset (single-kind providers only). It returns pois with Kind filled in
// where applicable.
func tagKinds(pois []types.RawPoi, name types.Provider) []types.RawPoi {
	kinds := registry.All[name].Kinds
	if len(kinds) != 1 {
		return pois
	}
	for i := range pois {
		if pois[i].Kind == "" {
			pois[i].Kind = kinds[0]
		}
	}
	return pois
}

// filterByKind keeps only the entries in pois matching kind, returning pois
// restricted to that kind.
func filterByKind(pois []types.EnrichedPoi, kind types.PointKind) []types.EnrichedPoi {
	out := pois[:0]
	for _, p := range pois {
		if p.Kind == kind {
			out = append(out, p)
		}
	}
	return out
}

// selectProviders filters registered providers to those named in
// q.Providers that support q.Mode, returning the matching registered
// providers.
func (s *Service) selectProviders(q types.SearchQuery) []providers.Provider {
	var out []providers.Provider
	for _, name := range q.Providers {
		if p, ok := s.providers[name]; ok && p.SupportsMode(q.Mode) {
			out = append(out, p)
		}
	}
	return out
}

// applyDefaults fills in zero-value fields of q with sensible defaults
// before the pipeline runs, mutating q in place and falling back to
// defaultProviders when q.Providers is empty.
func applyDefaults(q *types.SearchQuery, defaultProviders []types.Provider) {
	if q.Mode == "" {
		q.Mode = types.ModeRadius
	}
	if q.Radius == 0 {
		q.Radius = 5000
	}
	if q.Limit == 0 || q.Limit > 100 {
		q.Limit = 20
	}
	if q.Lang == "" {
		q.Lang = "en"
	}
	if len(q.Providers) == 0 {
		q.Providers = defaultProviders
	}
}

// applyFilters removes entries from pois that do not match q's requested
// types or fall below q.MinScore. It returns pois matching the requested
// types and minimum score.
func applyFilters(pois []types.EnrichedPoi, q types.SearchQuery) []types.EnrichedPoi {
	allowed := make(map[types.PoiType]bool, len(q.Types))
	for _, t := range q.Types {
		allowed[t] = true
	}
	out := pois[:0]
	for _, p := range pois {
		if len(allowed) > 0 && !allowed[p.Type] {
			continue
		}
		if p.Score >= q.MinScore {
			out = append(out, p)
		}
	}
	return out
}

// paginate slices pois according to q.Offset and q.Limit and wraps the
// slice in a SearchResult, returning the paginated search result.
func paginate(pois []types.EnrichedPoi, q types.SearchQuery) *types.SearchResult {
	total := len(pois)
	start := min(q.Offset, total)
	end := min(start+q.Limit, total)
	return &types.SearchResult{Query: q, Total: total, Results: pois[start:end]}
}

// ParseWeights deserialises raw, the JSON-encoded "weights" query param;
// all values must be in [0, 1]. It returns the parsed weights, or an error
// if raw is malformed or a value is out of range.
func ParseWeights(raw string) (map[types.PoiType]float64, error) {
	if raw == "" {
		return nil, nil
	}
	var weights map[types.PoiType]float64
	if err := json.Unmarshal([]byte(raw), &weights); err != nil {
		return nil, fmt.Errorf("weights: invalid JSON: %w", err)
	}
	for t, v := range weights {
		if v < 0 || v > 1 {
			return nil, fmt.Errorf("weights: %q must be in [0, 1], got %g", t, v)
		}
	}
	return weights, nil
}

// ParseProviderWeights deserialises raw, the JSON-encoded
// "provider_weights" query param; all values must be in [0, 1]. It returns
// the parsed weights, or an error if raw is malformed or a value is out of
// range.
func ParseProviderWeights(raw string) (map[types.Provider]float64, error) {
	if raw == "" {
		return nil, nil
	}
	var weights map[types.Provider]float64
	if err := json.Unmarshal([]byte(raw), &weights); err != nil {
		return nil, fmt.Errorf("provider_weights: invalid JSON: %w", err)
	}
	for p, v := range weights {
		if v < 0 || v > 1 {
			return nil, fmt.Errorf("provider_weights: %q must be in [0, 1], got %g", p, v)
		}
	}
	return weights, nil
}
