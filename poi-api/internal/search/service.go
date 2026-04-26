package search

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/trippier/poi-api/internal/dedup"
	"github.com/trippier/poi-api/internal/geo"
	"github.com/trippier/poi-api/internal/providers"
	"github.com/trippier/poi-api/internal/scoring"
	"github.com/trippier/poi-api/pkg/types"
	"go.uber.org/zap"
)

// Service orchestrates the full POI search pipeline.
type Service struct {
	providers       map[types.Provider]providers.Provider
	providerTimeout time.Duration
	log             *zap.Logger
}

// NewService returns a Service backed by the given providers.
func NewService(pp []providers.Provider, timeout time.Duration, log *zap.Logger) *Service {
	m := make(map[types.Provider]providers.Provider, len(pp))
	for _, p := range pp {
		m[p.Name()] = p
	}
	return &Service{providers: m, providerTimeout: timeout, log: log}
}

// Search runs the full POI pipeline for the given query and returns paginated, scored results.
func (s *Service) Search(ctx context.Context, q types.SearchQuery) (*types.SearchResult, error) {
	applyDefaults(&q, types.AllProviders)
	merged := s.pipeline(ctx, &q)
	filtered := applyFilters(merged, q)
	return paginate(filtered, q), nil
}

// SearchEvents runs the POI pipeline restricted to event providers and returns paginated results.
// The radius is forced to a minimum of 50 km because Ticketmaster and Eventbrite return
// few or no results at smaller radii.
func (s *Service) SearchEvents(ctx context.Context, q types.SearchQuery) (*types.SearchResult, error) {
	applyDefaults(&q, types.AllEventProviders)
	if q.Radius < 50_000 {
		q.Radius = 50_000
	}
	merged := s.pipeline(ctx, &q)
	return paginate(merged, q), nil
}

// ProvidersStatus probes each provider and returns availability + latency.
func (s *Service) ProvidersStatus(ctx context.Context) []types.ProviderStatus {
	probe := types.SearchQuery{Mode: types.ModeRadius, Lat: 48.8566, Lng: 2.3522, Radius: 500, Limit: 1}

	statuses := make([]types.ProviderStatus, 0, len(s.providers))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for name, p := range s.providers {
		wg.Add(1)
		go func(name types.Provider, p providers.Provider) {
			defer wg.Done()
			start := time.Now()
			tctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			var err error
			if pp, ok := p.(providers.Pingable); ok {
				err = pp.Ping(tctx)
			} else {
				_, err = p.Search(tctx, probe)
			}
			st := types.ProviderStatus{Name: name, Available: err == nil, LatencyMs: time.Since(start).Milliseconds()}
			if err != nil {
				st.Error = err.Error()
			}
			mu.Lock()
			statuses = append(statuses, st)
			mu.Unlock()
		}(name, p)
	}

	wg.Wait()
	return statuses
}

// pipeline fetches from all providers, geocodes district queries, deduplicates, scores, sorts.
// q is passed as a pointer so that district geocoding is reflected in the caller's query
// (and therefore in the SearchResult.Query returned to the client).
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
	merged := dedup.Merge(raw)
	for i := range merged {
		merged[i].Score = scoring.Score(merged[i], *q)
	}
	sort.Slice(merged, func(i, j int) bool { return merged[i].Score > merged[j].Score })
	return merged
}

// fetchAll fans out the search query to all selected providers concurrently and collects raw results.
func (s *Service) fetchAll(ctx context.Context, q types.SearchQuery) []types.RawPoi {
	selected := s.selectProviders(q)
	results := make([][]types.RawPoi, len(selected))
	var wg sync.WaitGroup

	for i, p := range selected {
		wg.Add(1)
		go func(i int, p providers.Provider) {
			defer wg.Done()
			pctx, cancel := context.WithTimeout(ctx, s.providerTimeout)
			defer cancel()
			pois, err := p.Search(pctx, q)
			if err != nil {
				s.log.Warn("provider error", zap.String("provider", string(p.Name())), zap.Error(err))
				return
			}
			results[i] = pois
		}(i, p)
	}

	wg.Wait()
	var all []types.RawPoi
	for _, r := range results {
		all = append(all, r...)
	}
	return all
}

// selectProviders filters the registered providers to those listed in q.Providers that support q.Mode.
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

// applyFilters removes POIs that do not match the requested types or fall below the minimum score.
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

// paginate slices the scored list according to q.Offset and q.Limit and wraps it in a SearchResult.
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
