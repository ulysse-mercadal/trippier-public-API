// Package middleware provides Gin HTTP middleware handlers.
package middleware

import "github.com/gin-gonic/gin"

// SecureHeaders returns a Gin handler that sets defensive HTTP response
// headers (X-Content-Type-Options, X-Frame-Options, Referrer-Policy) on
// every response before continuing the middleware chain.
func SecureHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}
