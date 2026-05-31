package search

import (
	"context"

	"go.uber.org/zap"

	"github.com/trippier/poi-api/internal/mathutil"
	"github.com/trippier/poi-api/pkg/types"
)

const wikidataProximityMeters = 50.0

// @param ctx.
// @param raw POIs to enrich in place.
// @param q search query forwarded to the Wikipedia provider.
// @return raw with WikidataID filled when missing, plus SourceURL and Description filled on GeoNames POIs from the nearest Wikipedia neighbour within wikidataProximityMeters.
func (s *Service) enrichWithWikidata(ctx context.Context, raw []types.RawPoi, q types.SearchQuery) []types.RawPoi {
	if len(raw) == 0 {
		return raw
	}
	wp, ok := s.providers[types.ProviderWikipedia]
	if !ok || !wp.SupportsMode(q.Mode) {
		return raw
	}
	if !hasPoiNeedingEnrichment(raw) {
		return raw
	}

	pctx, cancel := context.WithTimeout(ctx, s.providerTimeout)
	defer cancel()
	wikiPois, err := wp.Search(pctx, q)
	if err != nil {
		s.log.Warn("enrichWithWikidata: wikipedia search failed", zap.Error(err))
		return raw
	}
	if len(wikiPois) == 0 {
		return raw
	}

	for i := range raw {
		if !needsWikipediaEnrichment(raw[i]) {
			continue
		}
		nearest := closestWikipediaNeighbour(raw[i], wikiPois)
		if nearest == nil {
			continue
		}
		if raw[i].WikidataID == "" && nearest.WikidataID != "" {
			raw[i].WikidataID = nearest.WikidataID
		}
		if raw[i].Provider == types.ProviderGeoNames && nearest.SourceURL != "" {
			raw[i].SourceURL = nearest.SourceURL
		}
		if raw[i].Provider == types.ProviderGeoNames && raw[i].Description == "" && nearest.Description != "" {
			raw[i].Description = nearest.Description
		}
	}
	return raw
}

// @param p candidate raw POI.
// @return true when p has real coordinates and either lacks a WikidataID or is a GeoNames entry (which is always evaluated for SourceURL/Description swap).
func needsWikipediaEnrichment(p types.RawPoi) bool {
	if p.Provider == types.ProviderWikipedia || p.Provider == types.ProviderWikipediaEvents {
		return false
	}
	if p.Coords == nil || p.Coords.Approximate {
		return false
	}
	if p.WikidataID == "" {
		return true
	}
	return p.Provider == types.ProviderGeoNames
}

// @param raw POI batch.
// @return true when at least one POI satisfies needsWikipediaEnrichment.
func hasPoiNeedingEnrichment(raw []types.RawPoi) bool {
	for _, p := range raw {
		if needsWikipediaEnrichment(p) {
			return true
		}
	}
	return false
}

// @param p target POI used as the reference point.
// @param wikiPois Wikipedia provider output.
// @return the Wikipedia RawPoi within wikidataProximityMeters of p, nearest first, or nil.
func closestWikipediaNeighbour(p types.RawPoi, wikiPois []types.RawPoi) *types.RawPoi {
	bestDist := wikidataProximityMeters + 1
	var best *types.RawPoi
	for i := range wikiPois {
		w := &wikiPois[i]
		if w.Coords == nil || w.Coords.Approximate {
			continue
		}
		d := mathutil.Haversine(p.Coords.Lat, p.Coords.Lng, w.Coords.Lat, w.Coords.Lng)
		if d < bestDist {
			bestDist = d
			best = w
		}
	}
	if best == nil || bestDist > wikidataProximityMeters {
		return nil
	}
	return best
}
