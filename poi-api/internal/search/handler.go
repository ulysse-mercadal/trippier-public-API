package search

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
	rg.GET("/providers", h.providers)
}

// RegisterEventRoutes attaches the event search routes to a separate router group.
// This group must carry a higher rate-limit cost because it fans out to
// quota-constrained providers (Ticketmaster, Eventbrite, Wikipedia/Wikidata).
func (h *Handler) RegisterEventRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.events)
	rg.GET("/slim", h.eventsSlim)
}

// search returns merged, scored, paginated POIs for a given location or district.
// Supports mode=radius (lat/lng/radius), mode=polygon, mode=district (name geocoded via Nominatim).
// Optional types filter and weights map control what categories are returned and how they rank.
func (h *Handler) search(c *gin.Context) {
	var q types.SearchQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	weights, err := ParseWeights(c.Query("weights"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	if len(weights) > 0 && len(q.Types) > 0 {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "types and weights are mutually exclusive: use types to filter, or weights to reorder"})
		return
	}
	q.Weights = weights

	applyQueryDefaults(&q)

	if err := Validate(q); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
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
	var q types.SearchQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	applyQueryDefaults(&q)

	if err := Validate(q); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
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

// events returns cultural festivals and recurring events powered by Wikipedia/Wikidata SPARQL.
// Supports mode=radius and mode=district; filters to Wikidata class Q132241 (festival).
func (h *Handler) events(c *gin.Context) {
	var q types.SearchQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	applyQueryDefaults(&q)

	if err := Validate(q); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	result, err := h.service.SearchEvents(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// eventsSlim returns a lightweight projection (name, coords, dates, recurring) of events.
func (h *Handler) eventsSlim(c *gin.Context) {
	var q types.SearchQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	applyQueryDefaults(&q)

	if err := Validate(q); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	result, err := h.service.SearchEvents(c.Request.Context(), q)
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

// providers probes each registered provider and returns availability and latency.
func (h *Handler) providers(c *gin.Context) {
	statuses := h.service.ProvidersStatus(c.Request.Context())
	c.JSON(http.StatusOK, statuses)
}

// applyQueryDefaults sets mode=radius when no mode is provided by the caller.
func applyQueryDefaults(q *types.SearchQuery) {
	if q.Mode == "" {
		q.Mode = types.ModeRadius
	}
}

type errorResponse struct {
	Error string `json:"error"`
}
