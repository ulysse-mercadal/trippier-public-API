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
)

// cacheWriter wraps gin.ResponseWriter to capture the response body.
type cacheWriter struct {
	gin.ResponseWriter
	buf    bytes.Buffer
	status int
}

func (w *cacheWriter) Write(b []byte) (int, error) {
	w.buf.Write(b)
	return w.ResponseWriter.Write(b)
}

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
		// Only cache GET requests.
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		// Allow clients to bypass the cache.
		if c.GetHeader("X-No-Cache") != "" {
			c.Next()
			return
		}

		key := cacheKey(c)
		ctx := c.Request.Context()

		// Cache hit — serve stored response.
		if cached, err := rdb.Get(ctx, key).Bytes(); err == nil {
			c.Header("X-Cache", "HIT")
			c.Data(http.StatusOK, "application/json; charset=utf-8", cached)
			c.Abort()
			return
		}

		// Cache miss — proxy to handler and store result.
		cw := &cacheWriter{ResponseWriter: c.Writer, status: http.StatusOK}
		c.Writer = cw
		c.Header("X-Cache", "MISS")

		c.Next()

		if cw.status == http.StatusOK && cw.buf.Len() > 0 {
			_ = rdb.Set(ctx, key, cw.buf.Bytes(), ttl).Err()
		}
	}
}

// cacheKey returns a deterministic SHA-256 key for a request based on
// its path and sorted query parameters.
func cacheKey(c *gin.Context) string {
	params := c.Request.URL.Query()
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString(c.Request.URL.Path)
	sb.WriteByte('?')
	for i, k := range keys {
		if i > 0 {
			sb.WriteByte('&')
		}
		sb.WriteString(k)
		sb.WriteByte('=')
		sb.WriteString(strings.Join(params[k], ","))
	}

	h := sha256.Sum256([]byte(sb.String()))
	return "poi:cache:" + hex.EncodeToString(h[:])
}
