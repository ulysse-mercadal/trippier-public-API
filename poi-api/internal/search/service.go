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

// NewService returns a Service backed by the given providers.
func NewService(pp []providers.Provider, timeout time.Duration, log *zap.Logger) *Service {
	m := make(map[types.Provider]providers.Provider, len(pp))
	for _, p := range pp {
		m[p.Name()] = p
	}
	return &Service{providers: m, providerTimeout: timeout, log: log}
}

// selectionCacheKey is the immutable key under which a selectByCountry result
// is memoised. Types are joined sorted so request ordering doesn't matter.
type selectionCacheKey struct {
	cc        string
	types     string
	forEvents bool
}

// Search runs the full POI pipeline with geo-aware provider auto-selection.
// When the caller supplies ?providers=…, that explicit list is used as-is.
// Otherwise providers are chosen from the registry by country + category scores.
func (s *Service) Search(ctx context.Context, q types.SearchQuery) (*types.SearchResult, error) {
	userSpecified := len(q.Providers) > 0
	applyDefaults(&q, defaultProviders())
	if !userSpecified {
		if selected := s.autoSelectProviders(ctx, q, false); len(selected) > 0 {
			q.Providers = selected
		}
	}
	merged := s.pipeline(ctx, &q)
	filtered := applyFilters(merged, q)
	return paginate(filtered, q), nil
}

// SearchCustom is the fully-controllable variant of Search.
// In addition to the standard parameters it respects:
//   - q.CountryHint     — override geo-detected country code
//   - q.ExcludeProviders — blacklist applied after auto-selection or explicit list
//   - q.ProviderWeights  — per-provider score overrides for auto-selection
//
// When ?providers=… is supplied the explicit list is used (after exclusions).
// When omitted, geo-aware auto-selection runs using ProviderWeights overrides.
func (s *Service) SearchCustom(ctx context.Context, q types.SearchQuery) (*types.SearchResult, error) {
	userSpecified := len(q.Providers) > 0
	applyDefaults(&q, defaultProviders())
	if !userSpecified {
		if selected := s.autoSelectProviders(ctx, q, false); len(selected) > 0 {
			q.Providers = selected
		}
	}
	q.Providers = filterExcluded(q.Providers, q.ExcludeProviders)
	merged := s.pipeline(ctx, &q)
	filtered := applyFilters(merged, q)
	return paginate(filtered, q), nil
}

// SearchEvents runs the POI pipeline restricted to event providers. The
// radius is stretched to the maximum MinRadius declared by any selected
// provider in the registry — providers that need a wide geo (Ticketmaster,
// Eventbrite) opt in via meta.MinRadius without the orchestrator knowing
// their names.
func (s *Service) SearchEvents(ctx context.Context, q types.SearchQuery) (*types.SearchResult, error) {
	applyDefaults(&q, defaultEventProviders())
	clampRadiusToMin(&q)
	merged := s.pipeline(ctx, &q)
	return paginate(merged, q), nil
}

// SearchEventsCustom is the fully-controllable variant of SearchEvents.
func (s *Service) SearchEventsCustom(ctx context.Context, q types.SearchQuery) (*types.SearchResult, error) {
	userSpecified := len(q.Providers) > 0
	applyDefaults(&q, defaultEventProviders())
	clampRadiusToMin(&q)
	if !userSpecified {
		if selected := s.autoSelectProviders(ctx, q, true); len(selected) > 0 {
			q.Providers = selected
		}
	}
	q.Providers = filterExcluded(q.Providers, q.ExcludeProviders)
	merged := s.pipeline(ctx, &q)
	return paginate(merged, q), nil
}

// defaultProviders returns the list of non-BYOK, non-event providers from the
// registry, with a non-zero country-wildcard score. Wikipedia is included via
// CountryScore=0 to keep it usable as enricher-only.
func defaultProviders() []types.Provider {
	out := make([]types.Provider, 0, len(registry.All))
	for id, meta := range registry.All {
		if meta.Byok || meta.ForEvents {
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

// defaultEventProviders returns the list of event providers from the registry.
// BYOK event providers (Ticketmaster, Eventbrite) ARE included by default
// because the orchestrator can call them safely: each provider checks for its
// own key and returns nil cleanly when absent.
func defaultEventProviders() []types.Provider {
	out := make([]types.Provider, 0, len(registry.All))
	for id, meta := range registry.All {
		if !meta.ForEvents {
			continue
		}
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return string(out[i]) < string(out[j]) })
	return out
}

// clampRadiusToMin stretches q.Radius up to the maximum MinRadius declared by
// any currently selected provider — providers that don't need a wide geo can
// leave MinRadius=0 and remain unaffected.
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

// ProvidersStatus returns the list of registered providers with their static
// metadata. We do not probe upstream — every probe historically consumed paid
// quota (Ticketmaster, Eventbrite) without offering a reliable signal, and
// callers who want true health information should rely on observability of
// real Search calls.
func (s *Service) ProvidersStatus(_ context.Context) []types.ProviderStatus {
	statuses := make([]types.ProviderStatus, 0, len(s.providers))
	for name, p := range s.providers {
		_, isByok := p.(providers.ByokProvider)
		statuses = append(statuses, types.ProviderStatus{Name: name, Byok: isByok})
	}
	return statuses
}

// ProvidersCatalog returns the full registry merged with runtime implementation status.
func (s *Service) ProvidersCatalog() []types.ProviderCatalogEntry {
	entries := make([]types.ProviderCatalogEntry, 0, len(registry.All))
	for id, meta := range registry.All {
		_, implemented := s.providers[id]
		entries = append(entries, types.ProviderCatalogEntry{
			ID:             id,
			Label:          meta.Label,
			Byok:           meta.Byok,
			ByokHeader:     meta.ByokHeader,
			ForEvents:      meta.ForEvents,
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

// ProvidersRecommend scores all registry providers for the given location and types,
// returning the top-n entries sorted by composite score (descending).
func (s *Service) ProvidersRecommend(ctx context.Context, lat, lng float64, forEvents bool, requestedTypes []types.PoiType, limit int) types.RecommendResult {
	cc, _ := geo.CountryCode(ctx, lat, lng)

	type scored struct {
		entry types.RecommendedProvider
		score float64
	}
	candidates := make([]scored, 0, len(registry.All))

	for id, meta := range registry.All {
		if meta.ForEvents != forEvents {
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
				ForEvents:   meta.ForEvents,
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

// autoSelectProviders picks providers from the registry for the query's location and types.
// Only implemented, non-BYOK providers are auto-selected (BYOK = explicit opt-in by user).
// forEvents=true restricts to event providers.
// q.CountryHint and q.ProviderWeights are respected when present.
func (s *Service) autoSelectProviders(ctx context.Context, q types.SearchQuery, forEvents bool) []types.Provider {
	// Resolve coordinates for district mode so we can detect the country.
	lat, lng := q.Lat, q.Lng
	if q.Mode == types.ModeDistrict && lat == 0 && lng == 0 {
		if place, err := geo.GeocodeDistrict(ctx, q.District); err == nil {
			lat, lng = place.Lat, place.Lng
		}
	}

	// Determine country code (hint takes precedence over geo-detection).
	cc := strings.ToUpper(q.CountryHint)
	if cc == "" && (lat != 0 || lng != 0) {
		if detected, err := geo.CountryCode(ctx, lat, lng); err == nil {
			cc = detected
		} else {
			s.log.Debug("country detection failed, using global defaults", zap.Error(err))
		}
	}

	return s.selectByCountry(cc, q.Types, q.ProviderWeights, q.ExcludeProviders, forEvents)
}

// selectByCountry scores and filters registered providers for a given country
// code. Results for the no-override path (no weight overrides, no exclusions)
// are memoised in selectionCache so we don't re-iterate registry.All + sort
// on every standard /pois/search call.
func (s *Service) selectByCountry(
	cc string,
	requestedTypes []types.PoiType,
	weightOverrides map[types.Provider]float64,
	exclude []types.Provider,
	forEvents bool,
) []types.Provider {
	cacheable := len(weightOverrides) == 0 && len(exclude) == 0
	var cacheK selectionCacheKey
	if cacheable {
		cacheK = makeSelectionKey(cc, requestedTypes, forEvents)
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
		if meta.ForEvents != forEvents {
			continue
		}
		if meta.Byok {
			continue // BYOK providers require explicit user opt-in
		}
		if _, ok := s.providers[id]; !ok {
			continue // backend not implemented
		}
		if excludeSet[id] {
			continue
		}

		// Country score — allow per-request override via ProviderWeights.
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

// makeSelectionKey builds the stable cache key for selectByCountry. Types are
// sorted so caller order doesn't fragment the cache.
func makeSelectionKey(cc string, requestedTypes []types.PoiType, forEvents bool) selectionCacheKey {
	parts := make([]string, len(requestedTypes))
	for i, t := range requestedTypes {
		parts[i] = string(t)
	}
	sort.Strings(parts)
	return selectionCacheKey{cc: cc, types: strings.Join(parts, ","), forEvents: forEvents}
}

// filterExcluded removes blacklisted providers from a list.
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

// pipeline fetches from all providers, geocodes district queries, deduplicates, scores, sorts.
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

// fanOutLimit caps how many provider Search calls run concurrently per
// request. Bounds file-descriptor pressure (~N per concurrent request) and
// keeps the load polite enough that upstream mirrors don't 429/IP-ban us.
// Provider errors are logged inside the goroutine and never bubble up — a
// single slow or broken provider must not fail the whole pipeline.
const fanOutLimit = 16

// fetchAll fans out the search query to all selected providers concurrently,
// capped at fanOutLimit in-flight calls.
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
			results[i] = pois
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

// @param raw POIs as returned by every selected provider's Search.
// @param selected the provider IDs the user actually asked for.
// @return raw with any RawPoi whose Provider isn't in selected dropped — protects against cross-provider hints (e.g. a Wikivoyage listing referencing a Wikipedia article) leaking into a response when the user did not ask for that provider.
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

// selectProviders filters registered providers to those in q.Providers that support q.Mode.
func (s *Service) selectProviders(q types.SearchQuery) []providers.Provider {
	var out []providers.Provider
	for _, name := range q.Providers {
		if p, ok := s.providers[name]; ok && p.SupportsMode(q.Mode) {
			out = append(out, p)
		}
	}
	return out
}

// applyDefaults fills in zero-value fields of q with sensible defaults before the pipeline runs.
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

// applyFilters removes POIs that do not match the requested types or fall below min_score.
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

// paginate slices the scored list and wraps it in a SearchResult.
func paginate(pois []types.EnrichedPoi, q types.SearchQuery) *types.SearchResult {
	total := len(pois)
	start := min(q.Offset, total)
	end := min(start+q.Limit, total)
	return &types.SearchResult{Query: q, Total: total, Results: pois[start:end]}
}

// ParseWeights deserialises the "weights" query param. All values must be in [0, 1].
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

// ParseProviderWeights deserialises the "provider_weights" query param.
// All values must be in [0, 1].
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
