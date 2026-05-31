package middleware

import (
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/trippier/poi-api/internal/registry"
)

// SecureHeaders adds defensive HTTP response headers to every response.
func SecureHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// allowedHeaders is computed once at startup from the registry plus the
// always-on auth headers. Adding a new BYOK provider only requires editing
// the registry — the CORS preflight picks the new header automatically.
var allowedHeaders = buildAllowedHeaders()

func buildAllowedHeaders() string {
	base := []string{"Authorization", "X-API-Key", "X-Internal-Auth", "Content-Type"}
	seen := make(map[string]bool, len(base))
	for _, h := range base {
		seen[h] = true
	}
	for _, meta := range registry.All {
		if meta.ByokHeader == "" || seen[meta.ByokHeader] {
			continue
		}
		base = append(base, meta.ByokHeader)
		seen[meta.ByokHeader] = true
	}
	sort.Strings(base)
	return strings.Join(base, ", ")
}

// CORS allows all origins — this is a public API.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", allowedHeaders)

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
