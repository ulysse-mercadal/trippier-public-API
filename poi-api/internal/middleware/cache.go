package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/trippier/poi-api/internal/registry"
)

// cacheWriter wraps gin.ResponseWriter to capture the response body.
type cacheWriter struct {
	gin.ResponseWriter
	buf    bytes.Buffer
	status int
}

// Write tees the response body into the internal buffer so it can be cached.
func (w *cacheWriter) Write(b []byte) (int, error) {
	w.buf.Write(b)
	return w.ResponseWriter.Write(b)
}

// WriteHeader captures the status code before forwarding to the underlying writer.
func (w *cacheWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// Cache returns a Gin middleware that caches GET responses in Redis.
// Only 200-OK responses are stored. The cache key is derived from the
// sorted query-string parameters, so parameter order does not matter.
//
// Requests with a non-empty X-No-Cache header bypass the cache.
func Cache(rdb *redis.Client, ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		if c.GetHeader("X-No-Cache") != "" {
			c.Next()
			return
		}

		key := cacheKey(c)
		ctx := c.Request.Context()

		if cached, err := rdb.Get(ctx, key).Bytes(); err == nil {
			c.Header("X-Cache", "HIT")
			c.Data(http.StatusOK, "application/json; charset=utf-8", cached)
			c.Abort()
			return
		}

		cw := &cacheWriter{ResponseWriter: c.Writer, status: http.StatusOK}
		c.Writer = cw
		c.Header("X-Cache", "MISS")

		c.Next()

		if cw.status == http.StatusOK && cw.buf.Len() > 0 {
			_ = rdb.Set(ctx, key, cw.buf.Bytes(), ttl).Err()
		}
	}
}

// cacheKey returns a deterministic SHA-256 key for a request based on its path,
// sorted query parameters, and which BYOK provider keys are present (not their
// values). Two requests for the same location with different active providers
// get different keys.
//
// The presence suffix is built by iterating every registered BYOK provider and
// stamping its ID when its declared header is set on the request. New BYOK
// providers participate automatically — no code change to the cache layer is
// required when one is added to the registry.
func cacheKey(c *gin.Context) string {
	encoded := c.Request.URL.Query().Encode()
	suffix := byokPresenceSuffix(c)

	h := sha256.Sum256([]byte(c.Request.URL.Path + "?" + encoded + suffix))
	return "poi:cache:" + hex.EncodeToString(h[:])
}

// byokPresenceSuffix returns a stable, sorted ":id" list for every BYOK
// provider whose declared header is set on the request. The values of those
// headers are NEVER included — only the fact that they are present — so that
// per-user keys cannot leak through the cache key.
func byokPresenceSuffix(c *gin.Context) string {
	var ids []string
	for id, meta := range registry.All {
		if !meta.Byok || meta.ByokHeader == "" {
			continue
		}
		if c.GetHeader(meta.ByokHeader) != "" {
			ids = append(ids, string(id))
		}
	}
	if len(ids) == 0 {
		return ""
	}
	sort.Strings(ids)
	return ":" + strings.Join(ids, ":")
}
