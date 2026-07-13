// Package middleware provides reusable Gin middlewares for the poi-api server.
package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

const headerRequestID = "X-Request-Id"

// RequestID returns a Gin middleware handler that injects a unique request
// identifier into every request context and response header. Callers can
// retrieve it with RequestIDFromCtx.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(headerRequestID)
		if id == "" {
			id = newRequestID()
		}
		c.Set(headerRequestID, id)
		c.Header(headerRequestID, id)
		c.Next()
	}
}

// RequestIDFromCtx retrieves the request ID stored in c by the RequestID
// middleware, returning an empty string if none was set.
func RequestIDFromCtx(c *gin.Context) string {
	if id, ok := c.Get(headerRequestID); ok {
		if s, ok := id.(string); ok {
			return s
		}
	}
	return ""
}

// newRequestID generates and returns a random 8-byte hex-encoded string for
// use as a request identifier.
func newRequestID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
