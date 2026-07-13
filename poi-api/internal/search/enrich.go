// Package search provides POI search and enrichment logic.
package search

import (
	"context"

	"go.uber.org/zap"

	"github.com/trippier/poi-api/internal/mathutil"
	"github.com/trippier/poi-api/internal/providers"
	"github.com/trippier/poi-api/pkg/types"
)

// enrichRaw runs every registered enricher over the raw POI list, skipping
// failures silently. ctx is the request context, raw is the POI slice
// enriched in place, and q is the original search query. It returns the
// enriched POI slice.
func (s *Service) enrichRaw(ctx context.Context, raw []types.RawPoi, q types.SearchQuery) []types.RawPoi {
	if len(raw) == 0 {
		return raw
	}
	for _, p := range s.providers {
		e, ok := p.(providers.Enricher)
		if !ok {
			continue
		}
		if !p.SupportsMode(q.Mode) {
			continue
		}
		s.applyEnricher(ctx, raw, q, p, e)
	}
	return raw
}

// applyEnricher enriches each in-radius target POI, skipping targets from
// the enricher's own provider. ctx is the request context, raw holds the
// target POIs enriched in place, q is the original search query, p is the
// provider supplying enrichment sources, and e is the enricher applying
// data to targets.
func (s *Service) applyEnricher(ctx context.Context, raw []types.RawPoi, q types.SearchQuery, p providers.Provider, e providers.Enricher) {
	pctx, cancel := context.WithTimeout(ctx, s.providerTimeout)
	defer cancel()
	sources, err := p.Search(pctx, q)
	if err != nil {
		s.log.Warn("enrichRaw: provider search failed", zap.String("provider", string(p.Name())), zap.Error(err))
		return
	}
	if len(sources) == 0 {
		return
	}
	radius := e.EnrichmentRadius()
	for i := range raw {
		if raw[i].Provider == p.Name() {
			continue
		}
		if raw[i].Coords == nil || raw[i].Coords.Approximate {
			continue
		}
		nearest := closestNeighbour(raw[i], sources, radius)
		if nearest == nil {
			continue
		}
		e.EnrichTarget(&raw[i], *nearest)
	}
}

// closestNeighbour returns the nearest source POI to target within radius,
// or nil if none qualifies. target is the POI to find a neighbour for,
// sources are the candidate POIs to search, and radius is the maximum
// allowed distance. It returns the nearest qualifying source POI, or nil.
func closestNeighbour(target types.RawPoi, sources []types.RawPoi, radius float64) *types.RawPoi {
	bestDist := radius + 1
	var best *types.RawPoi
	for i := range sources {
		s := &sources[i]
		if s.Coords == nil || s.Coords.Approximate {
			continue
		}
		d := mathutil.Haversine(target.Coords.Lat, target.Coords.Lng, s.Coords.Lat, s.Coords.Lng)
		if d < bestDist {
			bestDist = d
			best = s
		}
	}
	if best == nil || bestDist > radius {
		return nil
	}
	return best
}
