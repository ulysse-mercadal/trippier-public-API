// Package search orchestrates all providers in parallel and returns a merged,
// scored, and paginated list of EnrichedPoi results.
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
)

// Service orchestrates the full POI search pipeline.
type Service struct {
	providers       map[types.Provider]providers.Provider
	providerTimeout time.Duration
}

// NewService returns a Service wired with the given provider implementations.
func NewService(pp []providers.Provider, providerTimeout time.Duration) *Service {
	m := make(map[types.Provider]providers.Provider, len(pp))
	for _, p := range pp {
		m[p.Name()] = p
	}
	return &Service{providers: m, providerTimeout: providerTimeout}
}

// Search runs all requested providers in parallel, merges and scores results,
// then applies pagination and minimum score filtering.
func (s *Service) Search(ctx context.Context, q types.SearchQuery) (*types.SearchResult, error) {
	s.applyDefaults(&q)

	selected := s.selectProviders(q)
	rawPois := s.fetchAll(ctx, q, selected)

	rawPois = geo.SetDistances(rawPois, q.Lat, q.Lng)
	if q.Mode == types.ModeRadius {
		rawPois = geo.FilterByRadius(rawPois, q.Lat, q.Lng, float64(q.Radius))
	}

	merged := dedup.Merge(rawPois)

	for i := range merged {
		merged[i].Score = scoring.Score(merged[i], q)
	}

	sort.Slice(merged, func(i, j int) bool {
		return merged[i].Score > merged[j].Score
	})

	filtered := s.applyFilters(merged, q)
	total := len(filtered)

	start := q.Offset
	if start > total {
		start = total
	}
	end := start + q.Limit
	if end > total {
		end = total
	}

	return &types.SearchResult{
		Query:   q,
		Total:   total,
		Results: filtered[start:end],
	}, nil
}

// ProvidersStatus returns availability and latency for each registered provider.
func (s *Service) ProvidersStatus(ctx context.Context) []types.ProviderStatus {
	statuses := make([]types.ProviderStatus, 0, len(s.providers))
	var mu sync.Mutex
	var wg sync.WaitGroup

	probe := types.SearchQuery{
		Mode:   types.ModeRadius,
		Lat:    48.8566,
		Lng:    2.3522,
		Radius: 500,
		Limit:  1,
	}

	for name, p := range s.providers {
		wg.Add(1)
		go func(name types.Provider, p providers.Provider) {
			defer wg.Done()
			start := time.Now()
			tctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			_, err := p.Search(tctx, probe)
			status := types.ProviderStatus{
				Name:      name,
				Available: err == nil,
				LatencyMs: time.Since(start).Milliseconds(),
			}
			if err != nil {
				status.Error = err.Error()
			}
			mu.Lock()
			statuses = append(statuses, status)
			mu.Unlock()
		}(name, p)
	}

	wg.Wait()
	return statuses
}

func (s *Service) applyDefaults(q *types.SearchQuery) {
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
		q.Providers = types.AllProviders
	}
}

func (s *Service) selectProviders(q types.SearchQuery) []providers.Provider {
	var selected []providers.Provider
	for _, name := range q.Providers {
		p, ok := s.providers[name]
		if ok && p.SupportsMode(q.Mode) {
			selected = append(selected, p)
		}
	}
	return selected
}

func (s *Service) fetchAll(ctx context.Context, q types.SearchQuery, pp []providers.Provider) []types.RawPoi {
	type result struct {
		pois []types.RawPoi
	}

	results := make([]result, len(pp))
	var wg sync.WaitGroup

	for i, p := range pp {
		wg.Add(1)
		go func(i int, p providers.Provider) {
			defer wg.Done()
			pctx, cancel := context.WithTimeout(ctx, s.providerTimeout)
			defer cancel()
			pois, err := p.Search(pctx, q)
			if err == nil {
				results[i] = result{pois: pois}
			}
		}(i, p)
	}

	wg.Wait()

	var all []types.RawPoi
	for _, r := range results {
		all = append(all, r.pois...)
	}
	return all
}

func (s *Service) applyFilters(pois []types.EnrichedPoi, q types.SearchQuery) []types.EnrichedPoi {
	result := pois[:0]
	for _, p := range pois {
		if p.Score >= q.MinScore {
			result = append(result, p)
		}
	}
	return result
}

// ParseWeights deserialises the "weights" query parameter from its JSON form.
func ParseWeights(raw string) (map[types.PoiType]float64, error) {
	if raw == "" {
		return nil, nil
	}
	var weights map[types.PoiType]float64
	if err := json.Unmarshal([]byte(raw), &weights); err != nil {
		return nil, fmt.Errorf("weights: invalid JSON: %w", err)
	}
	return weights, nil
}
