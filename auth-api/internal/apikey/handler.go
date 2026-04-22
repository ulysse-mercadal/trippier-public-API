package apikey

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	mw "github.com/trippier/auth-api/internal/middleware"
	"github.com/trippier/auth-api/internal/models"
	rl "github.com/trippier/auth-api/internal/ratelimit"
)

// Handler exposes API-key management and the internal rate-limit endpoint.
type Handler struct {
	svc *Service
}

// NewHandler creates a Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts routes. Key management requires jwtAuth; the internal
// subtree is protected by the caller (InternalAuth middleware).
func (h *Handler) RegisterRoutes(keys gin.IRouter, internal gin.IRouter, jwtAuth gin.HandlerFunc) {
	keys.Use(jwtAuth)
	keys.POST("", h.create)
	keys.GET("", h.list)
	keys.DELETE("/:id", h.revoke)

	internal.POST("/check-rate-limit", h.checkRateLimit)
}

func (h *Handler) create(c *gin.Context) {
	var body struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString(mw.UserIDKey)
	result, err := h.svc.Create(c.Request.Context(), userID, body.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create key"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"key":      result.PlaintextKey,
		"metadata": result.Key,
	})
}

func (h *Handler) list(c *gin.Context) {
	userID := c.GetString(mw.UserIDKey)
	keys, err := h.svc.List(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list keys"})
		return
	}
	if keys == nil {
		keys = []models.APIKeyWithUsage{}
	}
	c.JSON(http.StatusOK, gin.H{"keys": keys})
}

func (h *Handler) revoke(c *gin.Context) {
	userID := c.GetString(mw.UserIDKey)
	keyID := c.Param("id")

	if err := h.svc.Revoke(c.Request.Context(), userID, keyID); err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not revoke key"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "key revoked"})
}

func (h *Handler) checkRateLimit(c *gin.Context) {
	var body struct {
		APIKey string `json:"api_key" binding:"required"`
		Cost   int    `json:"cost"    binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sum := sha256.Sum256([]byte(body.APIKey))
	sha256Hash := hex.EncodeToString(sum[:])

	info, err := h.svc.ValidateBySHA256(c.Request.Context(), sha256Hash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if !info.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"allowed": false, "error": "invalid api key"})
		return
	}

	remaining, ttlSecs, notFound, insufficient, err := rl.Deduct(
		c.Request.Context(), h.svc.rdb, sha256Hash, body.Cost,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "rate-limit error"})
		return
	}

	if notFound {
		ttl := time.Duration(info.TokensResetIntervalSecs) * time.Second
		_ = rl.SetTokens(c.Request.Context(), h.svc.rdb, sha256Hash, info.TokensLimit, ttl)
		remaining, ttlSecs, _, insufficient, err = rl.Deduct(
			c.Request.Context(), h.svc.rdb, sha256Hash, body.Cost,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "rate-limit error"})
			return
		}
	}

	if insufficient {
		c.JSON(http.StatusOK, gin.H{
			"allowed":        false,
			"remaining":      0,
			"limit":          info.TokensLimit,
			"resets_in_secs": ttlSecs,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"allowed":        true,
		"remaining":      remaining,
		"limit":          info.TokensLimit,
		"resets_in_secs": ttlSecs,
	})
}
