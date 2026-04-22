package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimit returns a Gin middleware that enforces a sliding-window rate limit
// per client IP. requestsPerMin is the maximum number of requests allowed in
// any 60-second window.
//
// The implementation uses an in-memory token bucket so no external dependency
// is required. For multi-instance deployments, replace with a Redis-backed
// counter (see cache.go for the Redis client setup).
func RateLimit(requestsPerMin int) gin.HandlerFunc {
	type entry struct {
		mu       sync.Mutex
		requests []time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*entry)
	)

	getOrCreate := func(ip string) *entry {
		mu.Lock()
		defer mu.Unlock()
		if e, ok := clients[ip]; ok {
			return e
		}
		e := &entry{}
		clients[ip] = e
		return e
	}

	window := time.Minute

	return func(c *gin.Context) {
		ip := c.ClientIP()
		e := getOrCreate(ip)

		e.mu.Lock()
		now := time.Now()
		cutoff := now.Add(-window)

		// Slide the window: discard requests older than 1 minute.
		valid := e.requests[:0]
		for _, t := range e.requests {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}
		e.requests = valid

		if len(e.requests) >= requestsPerMin {
			e.mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}

		e.requests = append(e.requests, now)
		e.mu.Unlock()

		c.Next()
	}
}
