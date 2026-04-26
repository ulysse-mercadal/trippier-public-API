// Package ticketmaster implements the Provider interface for the Ticketmaster Discovery API v2.
// Documentation: https://developer.ticketmaster.com/products-and-docs/apis/discovery-api/v2/
package ticketmaster

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/trippier/poi-api/pkg/types"
)

const (
	defaultAPIURL  = "https://app.ticketmaster.com/discovery/v2/events.json"
	defaultTimeout = 10 * time.Second
	minRadiusKm    = 50
	maxRadiusKm    = 100
)

type tmResponse struct {
	Embedded *struct {
		Events []tmEvent `json:"events"`
	} `json:"_embedded"`
}

type tmEvent struct {
	ID     string    `json:"id"`
	Name   string    `json:"name"`
	URL    string    `json:"url"`
	Info   string    `json:"info"`
	Images []tmImage `json:"images"`
	Dates  struct {
		Start struct {
			DateTime string `json:"dateTime"`
		} `json:"start"`
		End struct {
			DateTime string `json:"dateTime"`
		} `json:"end"`
	} `json:"dates"`
	Embedded *struct {
		Venues []tmVenue `json:"venues"`
	} `json:"_embedded"`
}

type tmImage struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Ratio  string `json:"ratio"`
}

type tmVenue struct {
	Name     string `json:"name"`
	Location struct {
		Latitude  string `json:"latitude"`
		Longitude string `json:"longitude"`
	} `json:"location"`
}

// Provider fetches events from the Ticketmaster Discovery API.
type Provider struct {
	client  *http.Client
	apiKey  string
	baseURL string
}

// New returns a Provider authenticated with the given Ticketmaster consumer key.
func New(apiKey string) *Provider {
	return &Provider{
		client:  &http.Client{Timeout: defaultTimeout},
		apiKey:  apiKey,
		baseURL: defaultAPIURL,
	}
}

// NewWithURL returns a Provider targeting a custom endpoint. Intended for tests.
func NewWithURL(baseURL, apiKey string) *Provider {
	return &Provider{
		client:  &http.Client{Timeout: defaultTimeout},
		apiKey:  apiKey,
		baseURL: baseURL,
	}
}

// Name implements providers.Provider.
func (p *Provider) Name() types.Provider { return types.ProviderTicketmaster }

// SupportsMode implements providers.Provider.
func (p *Provider) SupportsMode(mode types.SearchMode) bool {
	return mode == types.ModeRadius || mode == types.ModeDistrict
}

// Search implements providers.Provider.
// Radius is clamped to [minRadiusKm, maxRadiusKm] km to protect the daily API quota.
func (p *Provider) Search(ctx context.Context, q types.SearchQuery) ([]types.RawPoi, error) {
	radiusKm := q.Radius / 1000
	if radiusKm < minRadiusKm {
		radiusKm = minRadiusKm
	}
	if radiusKm > maxRadiusKm {
		radiusKm = maxRadiusKm
	}

	now := time.Now().UTC()
	params := url.Values{
		"apikey":        {p.apiKey},
		"latlong":       {fmt.Sprintf("%.6f,%.6f", q.Lat, q.Lng)},
		"radius":        {strconv.Itoa(radiusKm)},
		"unit":          {"km"},
		"size":          {"100"},
		"startDateTime": {now.Format("2006-01-02T15:04:05Z")},
		"endDateTime":   {now.AddDate(0, 6, 0).Format("2006-01-02T15:04:05Z")},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("ticketmaster: build request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ticketmaster: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ticketmaster: unexpected status %d", resp.StatusCode)
	}

	var result tmResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ticketmaster: decode: %w", err)
	}

	if result.Embedded == nil {
		return nil, nil
	}
	return p.toRawPois(result.Embedded.Events), nil
}

// Ping implements providers.Pingable by calling the lightweight classifications endpoint.
// This does not consume search quota and verifies the API key is valid.
func (p *Provider) Ping(ctx context.Context) error {
	params := url.Values{"apikey": {p.apiKey}, "size": {"1"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		strings.Replace(p.baseURL, "events.json", "classifications.json", 1)+"?"+params.Encode(), nil)
	if err != nil {
		return err
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ticketmaster ping: status %d", resp.StatusCode)
	}
	return nil
}

// toRawPois converts Ticketmaster events to RawPoi records, skipping entries without a venue location.
func (p *Provider) toRawPois(events []tmEvent) []types.RawPoi {
	pois := make([]types.RawPoi, 0, len(events))
	for _, ev := range events {
		if ev.Name == "" {
			continue
		}

		lat, lng, ok := p.venueCoords(ev)
		if !ok {
			continue
		}

		poi := types.RawPoi{
			ID:          fmt.Sprintf("ticketmaster:%s", ev.ID),
			Name:        ev.Name,
			Type:        types.TypeEvent,
			Provider:    types.ProviderTicketmaster,
			Description: ev.Info,
			Thumbnail:   p.pickThumbnail(ev.Images),
			Coords:      &types.Coordinates{Lat: lat, Lng: lng},
			Contact:     types.Contact{Website: ev.URL},
		}

		if t, err := time.Parse(time.RFC3339, ev.Dates.Start.DateTime); err == nil {
			poi.DateStart = &t
		}
		if t, err := time.Parse(time.RFC3339, ev.Dates.End.DateTime); err == nil {
			poi.DateEnd = &t
		}

		pois = append(pois, poi)
	}
	return pois
}

// venueCoords extracts lat/lng from the first venue of an event.
func (p *Provider) venueCoords(ev tmEvent) (lat, lng float64, ok bool) {
	if ev.Embedded == nil || len(ev.Embedded.Venues) == 0 {
		return 0, 0, false
	}
	v := ev.Embedded.Venues[0]
	lat, err1 := strconv.ParseFloat(strings.TrimSpace(v.Location.Latitude), 64)
	lng, err2 := strconv.ParseFloat(strings.TrimSpace(v.Location.Longitude), 64)
	return lat, lng, err1 == nil && err2 == nil
}

// pickThumbnail selects the best image: prefers 16:9 ratio at width ≥ 640, else first available.
func (p *Provider) pickThumbnail(images []tmImage) string {
	for _, img := range images {
		if img.Ratio == "16_9" && img.Width >= 640 {
			return img.URL
		}
	}
	if len(images) > 0 {
		return images[0].URL
	}
	return ""
}
