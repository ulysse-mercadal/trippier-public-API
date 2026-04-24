package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// IPRateLimit limits requests per remote IP using a sliding window counter in Redis.
// limit: max requests allowed in the window.
// window: duration of the window.
func IPRateLimit(rdb *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := fmt.Sprintf("rl:ip:%s:%s", c.FullPath(), ip)
		ctx := c.Request.Context()

		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			// Redis unavailable — fail open (allow request) to avoid locking users out
			c.Next()
			return
		}
		if count == 1 {
			rdb.Expire(ctx, key, window) //nolint:errcheck
		}

		if count > int64(limit) {
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
