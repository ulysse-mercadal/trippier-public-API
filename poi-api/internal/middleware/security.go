package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
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

// allowedHeaders lists every request header that BYOK-aware clients may send.
// When adding a new provider, append its key header here.
const allowedHeaders = "Authorization, X-API-Key, Content-Type, " +
	// Global BYOK providers
	"X-Foursquare-Key, X-Here-Key, " +
	// China
	"X-Baidu-Key, X-Amap-Key, " +
	// Korea
	"X-Kakao-Key, " +
	// Japan
	"X-Navitime-Key, " +
	// India
	"X-Mappls-Key, " +
	// Southeast Asia
	"X-Grabmaps-Key, " +
	// Event providers
	"X-Ticketmaster-Key, X-Eventbrite-Token, X-Meetup-Token, X-OpenAgenda-Key"

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
