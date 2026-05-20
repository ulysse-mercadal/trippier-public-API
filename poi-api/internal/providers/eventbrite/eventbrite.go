// Package eventbrite implements the Provider interface for the Eventbrite API v3.
// Documentation: https://www.eventbrite.com/platform/api
//
// This provider uses a BYOK (Bring Your Own Key) pattern: no private token is
// stored server-side. Callers must inject their Eventbrite private token into the
// request context via byok.EventbriteKey before invoking Search. If the token is
// absent, Search returns nil, nil (the provider is silently skipped).
package eventbrite

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/trippier/poi-api/internal/byok"
	"github.com/trippier/poi-api/pkg/types"
)

const (
	defaultAPIURL  = "https://www.eventbriteapi.com/v3/events/search/"
	defaultTimeout = 10 * time.Second
	minRadiusKm    = 50
	maxRadiusKm    = 100
)

type ebResponse struct {
	Events []ebEvent `json:"events"`
}

type ebEvent struct {
	ID          string   `json:"id"`
	URL         string   `json:"url"`
	Name        ebText   `json:"name"`
	Description ebText   `json:"description"`
	Start       ebTime   `json:"start"`
	End         ebTime   `json:"end"`
	Logo        *ebLogo  `json:"logo"`
	Venue       *ebVenue `json:"venue"`
}

type ebText struct {
	Text string `json:"text"`
}

type ebTime struct {
	UTC string `json:"utc"`
}

type ebLogo struct {
	URL string `json:"url"`
}

type ebVenue struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}

// Provider fetches events from the Eventbrite API.
// The private token is read per-request from the context (BYOK pattern).
type Provider struct {
	client  *http.Client
	baseURL string
}

// New returns a Provider. No private token is stored — callers supply their own
// token per request via the X-Eventbrite-Token header (injected into context by the handler).
func New() *Provider {
	return &Provider{
		client:  &http.Client{Timeout: defaultTimeout},
		baseURL: defaultAPIURL,
	}
}

// NewWithURL returns a Provider targeting a custom endpoint. Intended for tests.
func NewWithURL(baseURL string) *Provider {
	return &Provider{
		client:  &http.Client{Timeout: defaultTimeout},
		baseURL: baseURL,
	}
}

// Name implements providers.Provider.
func (p *Provider) Name() types.Provider { return types.ProviderEventbrite }
func (p *Provider) IsByok() bool         { return true }

// SupportsMode implements providers.Provider.
func (p *Provider) SupportsMode(mode types.SearchMode) bool {
	return mode == types.ModeRadius || mode == types.ModeDistrict
}

// Search implements providers.Provider.
// Returns nil, nil when no Eventbrite private token is present in ctx (BYOK absent).
// Radius is clamped to [minRadiusKm, maxRadiusKm] km to protect the daily API quota.
func (p *Provider) Search(ctx context.Context, q types.SearchQuery) ([]types.RawPoi, error) {
	token := byok.EventbriteToken(ctx)
	if token == "" {
		return nil, nil
	}

	radiusKm := q.Radius / 1000
	if radiusKm < minRadiusKm {
		radiusKm = minRadiusKm
	}
	if radiusKm > maxRadiusKm {
		radiusKm = maxRadiusKm
	}

	params := url.Values{
		"location.latitude":  {fmt.Sprintf("%.6f", q.Lat)},
		"location.longitude": {fmt.Sprintf("%.6f", q.Lng)},
		"location.within":    {fmt.Sprintf("%dkm", radiusKm)},
		"expand":             {"venue"},
		"page_size":          {"100"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("eventbrite: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("eventbrite: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("eventbrite: unexpected status %d", resp.StatusCode)
	}

	var result ebResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("eventbrite: decode: %w", err)
	}

	return p.toRawPois(result.Events), nil
}

// toRawPois converts Eventbrite events to RawPoi records, skipping entries without a venue location.
func (p *Provider) toRawPois(events []ebEvent) []types.RawPoi {
	pois := make([]types.RawPoi, 0, len(events))
	for _, ev := range events {
		if ev.Name.Text == "" {
			continue
		}

		lat, lng, ok := p.venueCoords(ev)
		if !ok {
			continue
		}

		thumbnail := ""
		if ev.Logo != nil {
			thumbnail = ev.Logo.URL
		}

		poi := types.RawPoi{
			ID:          fmt.Sprintf("eventbrite:%s", ev.ID),
			Name:        ev.Name.Text,
			Type:        types.TypeEvent,
			Provider:    types.ProviderEventbrite,
			Description: ev.Description.Text,
			Thumbnail:   thumbnail,
			Coords:      &types.Coordinates{Lat: lat, Lng: lng},
			Contact:     types.Contact{Website: ev.URL},
		}

		if t, err := time.Parse(time.RFC3339, ev.Start.UTC); err == nil {
			poi.DateStart = &t
		}
		if t, err := time.Parse(time.RFC3339, ev.End.UTC); err == nil {
			poi.DateEnd = &t
		}

		pois = append(pois, poi)
	}
	return pois
}

// venueCoords extracts lat/lng from the event venue.
func (p *Provider) venueCoords(ev ebEvent) (lat, lng float64, ok bool) {
	if ev.Venue == nil || ev.Venue.Latitude == "" || ev.Venue.Longitude == "" {
		return 0, 0, false
	}
	lat, err1 := strconv.ParseFloat(strings.TrimSpace(ev.Venue.Latitude), 64)
	lng, err2 := strconv.ParseFloat(strings.TrimSpace(ev.Venue.Longitude), 64)
	return lat, lng, err1 == nil && err2 == nil
}
