// Package middleware provides gin HTTP middleware handlers.
package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// incrScript atomically increments a counter and sets its TTL on first creation, avoiding an INCR/EXPIRE race.
var incrScript = redis.NewScript(`
local count = redis.call('INCR', KEYS[1])
if count == 1 then
    redis.call('EXPIRE', KEYS[1], ARGV[1])
end
return count
`)

// IPRateLimit builds a gin handler that rate-limits requests per client IP
// via Redis, rejecting with 503 if Redis is unavailable. rdb is the Redis
// client used to track counters, limit is the max requests allowed per
// window, and window is the duration of the rate-limit window. It returns a
// gin.HandlerFunc that enforces the rate limit.
func IPRateLimit(rdb *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	windowSecs := int(window.Seconds())
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := fmt.Sprintf("rl:ip:%s:%s", c.FullPath(), ip)
		ctx := c.Request.Context()

		res, err := incrScript.Run(ctx, rdb, []string{key}, windowSecs).Int64()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "rate-limit service unavailable",
			})
			return
		}

		if res > int64(limit) {
			ttl, _ := rdb.TTL(ctx, key).Result()
			c.Header("Retry-After", fmt.Sprintf("%d", int(ttl.Seconds())))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many attempts — try again later",
			})
			return
		}
		c.Next()
	}
}
