package search

import (
	"context"

	"go.uber.org/zap"

	"github.com/trippier/poi-api/internal/mathutil"
	"github.com/trippier/poi-api/internal/providers"
	"github.com/trippier/poi-api/pkg/types"
)

// enrichRaw runs every registered providers.Enricher over the raw POI list.
// For each enricher we fetch its POIs once and let it apply itself to every
// non-self target within its declared EnrichmentRadius. The core stays
// provider-agnostic: provider-specific borrowing rules live inside each
// enricher's implementation.
//
// Enrichers are skipped silently when they fail or return nothing — partial
// enrichment is always preferable to a hard error on a non-essential step.
//
// @param ctx parent request context.
// @param raw POIs returned by the request's selected providers.
// @param q the search query, forwarded to each enricher's Search.
// @return raw with each enricher's contributions applied in place.
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

// applyEnricher fetches one enricher's data and lets it modify each target
// POI that has a viable nearest source within its radius. Targets belonging
// to the enricher's own provider are skipped — an enricher cannot enrich
// itself.
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

// closestNeighbour returns the POI in sources closest to target, or nil when
// none falls within radius. Sources without usable coordinates are skipped.
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
