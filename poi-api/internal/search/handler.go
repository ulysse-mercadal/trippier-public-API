// Package search contains the HTTP handler that exposes the search service via the Gin router.
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

// RegisterRoutes attaches all POI routes to the given router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/search", h.search)
	rg.GET("/search/slim", h.searchSlim)
	rg.GET("/providers", h.providers)
	rg.GET("/:id", h.getByID)
}

// search godoc.
// @Summary     Search for points of interest
// @Tags        pois
// @Produce     json
// @Param       mode      query  string   true  "Search mode: radius | polygon | district"
// @Param       lat       query  number   false "Latitude (required for mode=radius)"
// @Param       lng       query  number   false "Longitude (required for mode=radius)"
// @Param       radius    query  integer  false "Search radius in meters (default 5000)"
// @Param       polygon   query  string   false "GeoJSON polygon string (mode=polygon)"
// @Param       district  query  string   false "District or city name (mode=district)"
// @Param       providers query  []string false "Data providers to query"
// @Param       types     query  []string false "POI types to include"
// @Param       weights   query  string   false "JSON map of type weights e.g. {\"see\":2,\"eat\":1}"
// @Param       lang      query  string   false "Language code (default en)"
// @Param       limit     query  integer  false "Max results (default 20, max 100)"
// @Param       offset    query  integer  false "Pagination offset"
// @Param       min_score query  number   false "Minimum score 0-100"
// @Success     200  {object}  types.SearchResult
// @Failure     400  {object}  errorResponse
// @Failure     500  {object}  errorResponse
// @Router      /pois/search [get]
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
	q.Weights = weights

	applyQueryDefaults(&q)

	if err := Validate(q); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	result, err := h.service.Search(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// searchSlim godoc.
// @Summary     Search POIs — lightweight projection (name, type, coords only)
// @Tags        pois
// @Produce     json
// @Param       mode      query  string   false "Search mode: radius | polygon | district (default radius)"
// @Param       lat       query  number   false "Latitude"
// @Param       lng       query  number   false "Longitude"
// @Param       radius    query  integer  false "Search radius in meters (default 5000)"
// @Param       polygon   query  string   false "GeoJSON polygon string"
// @Param       district  query  string   false "District or city name"
// @Param       types     query  []string false "POI types to include"
// @Param       limit     query  integer  false "Max results (default 20, max 100)"
// @Param       offset    query  integer  false "Pagination offset"
// @Success     200  {object}  types.SlimResult
// @Failure     400  {object}  errorResponse
// @Router      /pois/search/slim [get]
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
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}

	slim := make([]types.SlimPoi, len(result.Results))
	for i, p := range result.Results {
		slim[i] = types.SlimPoi{Name: p.Name, Type: p.Type, Coords: p.Coords}
	}
	c.JSON(http.StatusOK, types.SlimResult{Total: result.Total, Results: slim})
}

// providers godoc.
// @Summary  List available data providers and their status
// @Tags     pois
// @Produce  json
// @Success  200  {array}  types.ProviderStatus
// @Router   /pois/providers [get]
func (h *Handler) providers(c *gin.Context) {
	statuses := h.service.ProvidersStatus(c.Request.Context())
	c.JSON(http.StatusOK, statuses)
}

// getByID godoc.
// @Summary  Retrieve a single POI by its namespaced ID
// @Tags     pois
// @Produce  json
// @Param    id  path  string  true  "Namespaced POI ID e.g. overpass:123456"
// @Success  200  {object}  types.EnrichedPoi
// @Failure  404  {object}  errorResponse
// @Router   /pois/{id} [get]
func (h *Handler) getByID(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, errorResponse{Error: "not implemented"})
}

// applyQueryDefaults fills in missing mode so callers can omit it when lat/lng/radius are provided.
func applyQueryDefaults(q *types.SearchQuery) {
	if q.Mode == "" {
		q.Mode = types.ModeRadius
	}
}

type errorResponse struct {
	Error string `json:"error"`
}
