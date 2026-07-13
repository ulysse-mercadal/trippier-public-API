// Package eventbrite implements the Provider interface for the Eventbrite API v3.
// Uses BYOK: callers inject a private token via byok.WithProviderKey before calling Search; Search returns nil, nil if absent.
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
	"github.com/trippier/poi-api/internal/providers"
	"github.com/trippier/poi-api/pkg/types"
)

const (
	defaultAPIURL  = "https://www.eventbriteapi.com/v3/events/search/"
	defaultTimeout = 10 * time.Second
	minRadiusKm    = 50
	maxRadiusKm    = 100
)

// ebResponse is the top-level Eventbrite search API response.
type ebResponse struct {
	Events []ebEvent `json:"events"`
}

// ebEvent is a single event returned by the Eventbrite search API.
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

// ebText holds a plain text field from the Eventbrite API.
type ebText struct {
	Text string `json:"text"`
}

// ebTime holds a UTC timestamp string from the Eventbrite API.
type ebTime struct {
	UTC string `json:"utc"`
}

// ebLogo holds an event's logo image URL.
type ebLogo struct {
	URL string `json:"url"`
}

// ebVenue holds a venue's coordinates as returned by the Eventbrite API.
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

// New creates a Provider using the default Eventbrite API URL. No token is stored;
// callers supply their own token per request via context (BYOK). It returns the
// new Provider.
func New() *Provider {
	return &Provider{
		client:  &http.Client{Timeout: defaultTimeout},
		baseURL: defaultAPIURL,
	}
}

// NewWithURL creates a Provider targeting the given custom Eventbrite API
// endpoint baseURL, for tests. It returns the new Provider.
func NewWithURL(baseURL string) *Provider {
	return &Provider{
		client:  &http.Client{Timeout: defaultTimeout},
		baseURL: baseURL,
	}
}

// Name returns the provider identifier for Eventbrite.
func (p *Provider) Name() types.Provider { return types.ProviderEventbrite }

// IsByok reports that Eventbrite requires bring-your-own-key, always returning true.
func (p *Provider) IsByok() bool { return true }

// SupportsMode reports whether Eventbrite supports the given search mode,
// returning true if mode is radius or district search.
func (p *Provider) SupportsMode(mode types.SearchMode) bool {
	return mode == types.ModeRadius || mode == types.ModeDistrict
}

// Search queries Eventbrite for events near the location described by q, using
// a BYOK token read from ctx. The requested radius is clamped to
// [minRadiusKm, maxRadiusKm] km. It returns the matching raw POIs, or nil, nil
// if no token is present in ctx.
func (p *Provider) Search(ctx context.Context, q types.SearchQuery) ([]types.RawPoi, error) {
	token := byok.GetProviderKey(ctx, types.ProviderEventbrite)
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

// toRawPois converts the given Eventbrite events to RawPoi records, skipping
// entries with no name or without a venue location, and returns the converted
// raw POIs.
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
		var images []string
		if ev.Logo != nil && ev.Logo.URL != "" {
			thumbnail = ev.Logo.URL
			images = []string{ev.Logo.URL}
		}

		poi := types.RawPoi{
			ID:          fmt.Sprintf("eventbrite:%s", ev.ID),
			Name:        ev.Name.Text,
			Type:        types.TypeEvent,
			Provider:    types.ProviderEventbrite,
			Description: ev.Description.Text,
			Thumbnail:   thumbnail,
			Images:      images,
			Coords:      &types.Coordinates{Lat: lat, Lng: lng},
			Contact:     types.Contact{Website: ev.URL},
			SourceURL:   ev.URL,
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

// venueCoords extracts the latitude and longitude from ev's venue, returning
// the parsed coordinates and whether both values parsed successfully.
func (p *Provider) venueCoords(ev ebEvent) (lat, lng float64, ok bool) {
	if ev.Venue == nil || ev.Venue.Latitude == "" || ev.Venue.Longitude == "" {
		return 0, 0, false
	}
	lat, err1 := strconv.ParseFloat(strings.TrimSpace(ev.Venue.Latitude), 64)
	lng, err2 := strconv.ParseFloat(strings.TrimSpace(ev.Venue.Longitude), 64)
	return lat, lng, err1 == nil && err2 == nil
}

// init registers the Eventbrite provider factory.
func init() {
	providers.Register(types.ProviderEventbrite, func(_ providers.BuildConfig) (providers.Provider, error) {
		return New(), nil
	})
}
