// Package search implements HTTP handlers for POI and event search endpoints.
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

// NewHandler creates a Handler backed by the given search service svc and
// returns it ready to register routes.
func NewHandler(svc *Service) *Handler {
	return &Handler{service: svc}
}

// RegisterRoutes attaches the core POI search and provider routes to the
// given router group rg.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/search", h.search)
	rg.GET("/search/slim", h.searchSlim)
	rg.GET("/search/custom", h.searchCustom)
	rg.GET("/search/custom/slim", h.searchCustomSlim)
	rg.GET("/providers", h.providers)
	rg.GET("/providers/catalog", h.providersCatalog)
	rg.GET("/providers/recommend", h.providersRecommend)
}

// RegisterEventRoutes attaches the event search routes to the separate
// router group rg.
func (h *Handler) RegisterEventRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.events)
	rg.GET("/slim", h.eventsSlim)
	rg.GET("/custom", h.eventsCustom)
	rg.GET("/custom/slim", h.eventsCustomSlim)
}

// search handles GET requests for merged, scored, paginated POIs with
// geo-aware provider auto-selection, writing the result to c.
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

// searchSlim handles GET requests and writes a lightweight projection
// (name, type, coords) suitable for map rendering to c.
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

// searchCustom handles GET requests giving full control over provider
// selection (country_hint, provider_weights, exclude_providers), writing
// the result to c.
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

// searchCustomSlim is the slim projection variant of searchCustom, writing
// its result to c.
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

// events handles GET requests for events powered by Wikipedia/SPARQL and
// optional BYOK providers, writing the result to c.
func (h *Handler) events(c *gin.Context) {
	q, ok := parseQuery(c)
	if !ok {
		return
	}
	result, err := h.service.SearchEvents(allByokContext(c), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// eventsSlim handles GET requests and writes a lightweight projection of
// events to c.
func (h *Handler) eventsSlim(c *gin.Context) {
	q, ok := parseQuery(c)
	if !ok {
		return
	}
	result, err := h.service.SearchEvents(allByokContext(c), q)
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

// eventsCustom handles GET requests giving full control over event
// provider selection, writing the result to c.
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

// eventsCustomSlim is the slim projection variant of eventsCustom, writing
// its result to c.
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

// providers probes each registered provider and writes availability and
// latency for each to c.
func (h *Handler) providers(c *gin.Context) {
	statuses := h.service.ProvidersStatus(c.Request.Context())
	c.JSON(http.StatusOK, statuses)
}

// providersCatalog writes the full registry, with country/category scores
// and implementation status, to c.
func (h *Handler) providersCatalog(c *gin.Context) {
	c.JSON(http.StatusOK, h.service.ProvidersCatalog())
}

// providersRecommend reads lat, lng, types, for_events, and limit from c's
// query params and writes scored provider recommendations for that location
// to c.
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

	kind := types.KindPOI
	if c.Query("for_events") == "true" || c.Query("for_events") == "1" {
		kind = types.KindEvent
	}

	limit := 10
	if raw := c.Query("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			limit = n
		}
	}

	result := h.service.ProvidersRecommend(c.Request.Context(), lat, lng, kind, requestedTypes, limit)
	c.JSON(http.StatusOK, result)
}

// parseQuery binds standard query params from c, parses weights, applies
// defaults, and validates the result. It returns the parsed search query and
// whether parsing succeeded; on failure it has already written an error
// response to c.
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

	if err := Validate(q); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return q, false
	}
	return q, true
}

// parseCustomQuery extends parseQuery with the custom-route params from c:
// country_hint, provider_weights (JSON), and exclude_providers. It returns
// the parsed search query and whether parsing succeeded.
func parseCustomQuery(c *gin.Context) (types.SearchQuery, bool) {
	q, ok := parseQuery(c)
	if !ok {
		return q, false
	}

	providerWeights, err := ParseProviderWeights(c.Query("provider_weights"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return q, false
	}
	q.ProviderWeights = providerWeights

	return q, true
}

// allByokContext scans c's inbound headers for any registered BYOK provider
// keys and returns a context carrying the discovered keys, derived from
// c.Request.Context().
func allByokContext(c *gin.Context) context.Context {
	ctx := c.Request.Context()
	for id, meta := range registry.All {
		if !meta.Byok || meta.ByokHeader == "" {
			continue
		}
		if v := c.GetHeader(meta.ByokHeader); v != "" {
			ctx = byok.WithProviderKey(ctx, id, v)
		}
	}
	return ctx
}

// errorResponse is the JSON body returned for failed requests.
type errorResponse struct {
	Error string `json:"error"`
}
