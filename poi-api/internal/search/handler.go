package search

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/trippier/poi-api/internal/byok"
	"github.com/trippier/poi-api/internal/registry"
	"github.com/trippier/poi-api/pkg/types"
)

// Handler exposes the search service over HTTP.
type Handler struct {
	service *Service
}

// NewHandler returns a Handler backed by the given Service.
func NewHandler(svc *Service) *Handler {
	return &Handler{service: svc}
}

// RegisterRoutes attaches the core POI search and provider routes to the given group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/search", h.search)
	rg.GET("/search/slim", h.searchSlim)
	rg.GET("/search/custom", h.searchCustom)
	rg.GET("/search/custom/slim", h.searchCustomSlim)
	rg.GET("/providers", h.providers)
	rg.GET("/providers/catalog", h.providersCatalog)
	rg.GET("/providers/recommend", h.providersRecommend)
}

// RegisterEventRoutes attaches the event search routes to a separate router group.
func (h *Handler) RegisterEventRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.events)
	rg.GET("/slim", h.eventsSlim)
	rg.GET("/custom", h.eventsCustom)
	rg.GET("/custom/slim", h.eventsCustomSlim)
}

// ── Standard search routes ─────────────────────────────────────────────────────

// search returns merged, scored, paginated POIs with geo-aware provider auto-selection.
func (h *Handler) search(c *gin.Context) {
	q, ok := parseQuery(c)
	if !ok {
		return
	}
	result, err := h.service.Search(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// searchSlim returns a lightweight projection (name, type, coords) suitable for map rendering.
func (h *Handler) searchSlim(c *gin.Context) {
	q, ok := parseQuery(c)
	if !ok {
		return
	}
	result, err := h.service.Search(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}
	slim := make([]types.SlimPoi, len(result.Results))
	for i, p := range result.Results {
		slim[i] = types.SlimPoi{Name: p.Name, Type: p.Type, Coords: p.Coords}
	}
	c.JSON(http.StatusOK, types.SlimResult{Total: result.Total, Results: slim})
}

// ── Custom search routes ───────────────────────────────────────────────────────

// searchCustom gives full control over provider selection.
// Additional params: country_hint, provider_weights (JSON), exclude_providers.
func (h *Handler) searchCustom(c *gin.Context) {
	q, ok := parseCustomQuery(c)
	if !ok {
		return
	}
	result, err := h.service.SearchCustom(allByokContext(c), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// searchCustomSlim is the slim projection variant of searchCustom.
func (h *Handler) searchCustomSlim(c *gin.Context) {
	q, ok := parseCustomQuery(c)
	if !ok {
		return
	}
	result, err := h.service.SearchCustom(allByokContext(c), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}
	slim := make([]types.SlimPoi, len(result.Results))
	for i, p := range result.Results {
		slim[i] = types.SlimPoi{Name: p.Name, Type: p.Type, Coords: p.Coords}
	}
	c.JSON(http.StatusOK, types.SlimResult{Total: result.Total, Results: slim})
}

// ── Standard event routes ──────────────────────────────────────────────────────

// events returns events powered by Wikipedia/SPARQL and optional BYOK providers.
func (h *Handler) events(c *gin.Context) {
	q, ok := parseQuery(c)
	if !ok {
		return
	}
	result, err := h.service.SearchEvents(byokContext(c), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// eventsSlim returns a lightweight projection of events.
func (h *Handler) eventsSlim(c *gin.Context) {
	q, ok := parseQuery(c)
	if !ok {
		return
	}
	result, err := h.service.SearchEvents(byokContext(c), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}
	slim := make([]types.SlimEvent, len(result.Results))
	for i, e := range result.Results {
		slim[i] = types.SlimEvent{
			Name:      e.Name,
			Coords:    e.Coords,
			DateStart: e.DateStart,
			DateEnd:   e.DateEnd,
			Recurring: e.Recurring,
		}
	}
	c.JSON(http.StatusOK, types.SlimEventResult{Total: result.Total, Results: slim})
}

// ── Custom event routes ────────────────────────────────────────────────────────

// eventsCustom gives full control over event provider selection.
func (h *Handler) eventsCustom(c *gin.Context) {
	q, ok := parseCustomQuery(c)
	if !ok {
		return
	}
	result, err := h.service.SearchEventsCustom(allByokContext(c), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// eventsCustomSlim is the slim projection variant of eventsCustom.
func (h *Handler) eventsCustomSlim(c *gin.Context) {
	q, ok := parseCustomQuery(c)
	if !ok {
		return
	}
	result, err := h.service.SearchEventsCustom(allByokContext(c), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}
	slim := make([]types.SlimEvent, len(result.Results))
	for i, e := range result.Results {
		slim[i] = types.SlimEvent{
			Name:      e.Name,
			Coords:    e.Coords,
			DateStart: e.DateStart,
			DateEnd:   e.DateEnd,
			Recurring: e.Recurring,
		}
	}
	c.JSON(http.StatusOK, types.SlimEventResult{Total: result.Total, Results: slim})
}

// ── Provider info routes ───────────────────────────────────────────────────────

// providers probes each registered provider and returns availability and latency.
func (h *Handler) providers(c *gin.Context) {
	statuses := h.service.ProvidersStatus(c.Request.Context())
	c.JSON(http.StatusOK, statuses)
}

// providersCatalog returns the full registry with country/category scores and implementation status.
func (h *Handler) providersCatalog(c *gin.Context) {
	c.JSON(http.StatusOK, h.service.ProvidersCatalog())
}

// providersRecommend returns scored provider recommendations for a location.
// Query params: lat, lng (required), types (comma-separated, optional),
//
//	for_events (bool, default false), limit (int, default 10).
func (h *Handler) providersRecommend(c *gin.Context) {
	lat, err1 := strconv.ParseFloat(c.Query("lat"), 64)
	lng, err2 := strconv.ParseFloat(c.Query("lng"), 64)
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "lat and lng are required and must be valid numbers"})
		return
	}

	var requestedTypes []types.PoiType
	if raw := c.Query("types"); raw != "" {
		for _, t := range strings.Split(raw, ",") {
			requestedTypes = append(requestedTypes, types.PoiType(strings.TrimSpace(t)))
		}
	}

	forEvents := c.Query("for_events") == "true" || c.Query("for_events") == "1"

	limit := 10
	if raw := c.Query("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			limit = n
		}
	}

	result := h.service.ProvidersRecommend(c.Request.Context(), lat, lng, forEvents, requestedTypes, limit)
	c.JSON(http.StatusOK, result)
}

// ── Query parsing helpers ──────────────────────────────────────────────────────

// parseQuery binds standard query params, parses weights, applies defaults, and validates.
func parseQuery(c *gin.Context) (types.SearchQuery, bool) {
	var q types.SearchQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return q, false
	}

	weights, err := ParseWeights(c.Query("weights"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return q, false
	}
	if len(weights) > 0 && len(q.Types) > 0 {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "types and weights are mutually exclusive: use types to filter, or weights to reorder"})
		return q, false
	}
	q.Weights = weights

	applyQueryDefaults(&q)

	if err := Validate(q); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return q, false
	}
	return q, true
}

// parseCustomQuery extends parseQuery with the custom-route params:
// country_hint, provider_weights (JSON), exclude_providers.
func parseCustomQuery(c *gin.Context) (types.SearchQuery, bool) {
	q, ok := parseQuery(c)
	if !ok {
		return q, false
	}

	// country_hint is already bound via ShouldBindQuery (form:"country_hint")

	providerWeights, err := ParseProviderWeights(c.Query("provider_weights"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return q, false
	}
	q.ProviderWeights = providerWeights

	return q, true
}

// applyQueryDefaults sets mode=radius when no mode is provided by the caller.
func applyQueryDefaults(q *types.SearchQuery) {
	if q.Mode == "" {
		q.Mode = types.ModeRadius
	}
}

// ── BYOK context helpers ───────────────────────────────────────────────────────

// byokContext injects the legacy Ticketmaster and Eventbrite BYOK keys into ctx.
// Used by the standard event routes to preserve backward compatibility.
func byokContext(c *gin.Context) context.Context {
	ctx := c.Request.Context()
	if k := c.GetHeader("X-Ticketmaster-Key"); k != "" {
		ctx = context.WithValue(ctx, byok.TicketmasterKey, k)
	}
	if t := c.GetHeader("X-Eventbrite-Token"); t != "" {
		ctx = context.WithValue(ctx, byok.EventbriteKey, t)
	}
	return ctx
}

// allByokContext injects all BYOK keys from the registry into ctx.
// Used by the custom routes so every known provider can receive its key.
// Legacy Ticketmaster and Eventbrite keys are injected under both the old specific
// constant AND the generic provider-keyed slot, ensuring all providers find their key.
func allByokContext(c *gin.Context) context.Context {
	ctx := c.Request.Context()
	for id, meta := range registry.All {
		if !meta.Byok || meta.ByokHeader == "" {
			continue
		}
		v := c.GetHeader(meta.ByokHeader)
		if v == "" {
			continue
		}
		// Generic slot — used by new providers.
		ctx = byok.WithProviderKey(ctx, id, v)
		// Legacy slots — keep existing Ticketmaster/Eventbrite providers working.
		switch meta.ByokHeader {
		case "X-Ticketmaster-Key":
			ctx = context.WithValue(ctx, byok.TicketmasterKey, v)
		case "X-Eventbrite-Token":
			ctx = context.WithValue(ctx, byok.EventbriteKey, v)
		}
	}
	return ctx
}

type errorResponse struct {
	Error string `json:"error"`
}
