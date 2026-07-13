// Package ticketmaster implements the Provider interface for the Ticketmaster Discovery API v2.
// Uses a BYOK pattern: callers must inject the API key into ctx via byok.WithProviderKey; if absent, Search returns nil, nil.
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

	"github.com/trippier/poi-api/internal/byok"
	"github.com/trippier/poi-api/internal/geo"
	"github.com/trippier/poi-api/internal/providers"
	"github.com/trippier/poi-api/pkg/types"
)

const (
	defaultAPIURL  = "https://app.ticketmaster.com/discovery/v2/events.json"
	defaultTimeout = 10 * time.Second
	minRadiusKm    = 50
	maxRadiusKm    = 100
)

// tmResponse is the top-level shape of a Ticketmaster Discovery API events response.
type tmResponse struct {
	Embedded *struct {
		Events []tmEvent `json:"events"`
	} `json:"_embedded"`
}

// tmEvent is a single event as returned by the Ticketmaster Discovery API.
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

// tmImage is an image asset associated with a Ticketmaster event.
type tmImage struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Ratio  string `json:"ratio"`
}

// tmVenue is a venue location associated with a Ticketmaster event.
type tmVenue struct {
	Name     string `json:"name"`
	Location struct {
		Latitude  string `json:"latitude"`
		Longitude string `json:"longitude"`
	} `json:"location"`
}

// Provider fetches events from the Ticketmaster Discovery API using a per-request BYOK API key.
type Provider struct {
	client  *http.Client
	baseURL string
}

// New creates a Provider using the default Ticketmaster API endpoint.
func New() *Provider {
	return &Provider{
		client:  &http.Client{Timeout: defaultTimeout},
		baseURL: defaultAPIURL,
	}
}

// NewWithURL creates a Provider targeting the given baseURL instead of the
// default endpoint; intended for tests.
func NewWithURL(baseURL string) *Provider {
	return &Provider{
		client:  &http.Client{Timeout: defaultTimeout},
		baseURL: baseURL,
	}
}

// Name returns the provider identifier for Ticketmaster.
func (p *Provider) Name() types.Provider { return types.ProviderTicketmaster }

// IsByok reports that Ticketmaster always requires a caller-supplied API key,
// returning true.
func (p *Provider) IsByok() bool { return true }

// SupportsMode reports whether mode is supported by this provider, returning
// true if mode is radius or district.
func (p *Provider) SupportsMode(mode types.SearchMode) bool {
	return mode == types.ModeRadius || mode == types.ModeDistrict
}

// Search queries the Ticketmaster Discovery API for events matching q, using
// ctx for the request and to carry the BYOK API key. It returns the matching
// raw POIs, or nil if no API key is present, along with any error
// encountered.
func (p *Provider) Search(ctx context.Context, q types.SearchQuery) ([]types.RawPoi, error) {
	apiKey := byok.GetProviderKey(ctx, types.ProviderTicketmaster)
	if apiKey == "" {
		return nil, nil
	}

	radiusKm := q.Radius / 1000
	if radiusKm < minRadiusKm {
		radiusKm = minRadiusKm
	}
	if radiusKm > maxRadiusKm {
		radiusKm = maxRadiusKm
	}

	now := time.Now().UTC()
	startDate := now
	if q.Date != "" {
		if parsed, err := time.Parse("2006-01-02", q.Date); err == nil {
			startDate = parsed.UTC()
		}
	}
	startOfDay := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
	params := url.Values{
		"apikey":        {apiKey},
		"radius":        {strconv.Itoa(radiusKm)},
		"unit":          {"km"},
		"size":          {"100"},
		"startDateTime": {startOfDay.Format("2006-01-02T15:04:05Z")},
		"endDateTime":   {startOfDay.AddDate(1, 0, 0).Format("2006-01-02T15:04:05Z")},
	}

	if city, err := geo.ReverseGeocode(ctx, q.Lat, q.Lng); err == nil {
		params.Set("city", city)
	} else {
		params.Set("latlong", fmt.Sprintf("%.6f,%.6f", q.Lat, q.Lng))
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
	return p.toRawPois(result.Embedded.Events, q.Lat, q.Lng), nil
}

// toRawPois converts events into RawPoi records, falling back to
// (centerLat, centerLng) as an approximate location for venues lacking
// coordinates. It returns the converted raw POIs.
func (p *Provider) toRawPois(events []tmEvent, centerLat, centerLng float64) []types.RawPoi {
	pois := make([]types.RawPoi, 0, len(events))
	for _, ev := range events {
		if ev.Name == "" {
			continue
		}

		lat, lng, ok := p.venueCoords(ev)
		approximate := false
		if !ok {
			lat, lng, approximate = centerLat, centerLng, true
		}

		poi := types.RawPoi{
			ID:          fmt.Sprintf("ticketmaster:%s", ev.ID),
			Name:        ev.Name,
			Type:        types.TypeEvent,
			Provider:    types.ProviderTicketmaster,
			Description: ev.Info,
			Thumbnail:   p.pickThumbnail(ev.Images),
			Images:      p.pickImages(ev.Images),
			Coords:      &types.Coordinates{Lat: lat, Lng: lng, Approximate: approximate},
			Contact:     types.Contact{Website: ev.URL},
			SourceURL:   ev.URL,
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

// venueCoords extracts the latitude and longitude from the first venue of
// ev. It returns lat, lng, and ok indicating whether valid coordinates were
// found.
func (p *Provider) venueCoords(ev tmEvent) (lat, lng float64, ok bool) {
	if ev.Embedded == nil || len(ev.Embedded.Venues) == 0 {
		return 0, 0, false
	}
	v := ev.Embedded.Venues[0]
	lat, err1 := strconv.ParseFloat(strings.TrimSpace(v.Location.Latitude), 64)
	lng, err2 := strconv.ParseFloat(strings.TrimSpace(v.Location.Longitude), 64)
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	if lat == 0 && lng == 0 {
		return 0, 0, false
	}
	return lat, lng, true
}

// pickThumbnail selects the best thumbnail from images, preferring 16:9
// ratio at width >= 640. It returns the chosen thumbnail URL, or an empty
// string if none are available.
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

// pickImages selects up to 3 distinct image URLs from images, prioritising
// 16:9 ratio at width >= 640. It returns the selected image URLs, or nil if
// none are available.
func (p *Provider) pickImages(images []tmImage) []string {
	const max = 3
	out := make([]string, 0, max)
	seen := make(map[string]bool, max)
	for _, img := range images {
		if len(out) >= max {
			break
		}
		if img.URL == "" || seen[img.URL] {
			continue
		}
		if img.Ratio == "16_9" && img.Width >= 640 {
			out = append(out, img.URL)
			seen[img.URL] = true
		}
	}
	for _, img := range images {
		if len(out) >= max {
			break
		}
		if img.URL == "" || seen[img.URL] {
			continue
		}
		out = append(out, img.URL)
		seen[img.URL] = true
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// init registers the Ticketmaster provider factory.
func init() {
	providers.Register(types.ProviderTicketmaster, func(_ providers.BuildConfig) (providers.Provider, error) {
		return New(), nil
	})
}
