// Package ratelimit manages token-bucket state in Redis.
package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisKey returns the Redis key for a given API-key SHA-256 hash.
func RedisKey(sha256Hash string) string {
	return "rl:" + sha256Hash
}

// SetTokens initialises (or resets) the token bucket for a key.
func SetTokens(ctx context.Context, rdb *redis.Client, sha256Hash string, limit int, ttl time.Duration) error {
	return rdb.Set(ctx, RedisKey(sha256Hash), limit, ttl).Err()
}

// GetUsage returns (remaining, ttlSecs). remaining == -1 means key not in Redis.
func GetUsage(ctx context.Context, rdb *redis.Client, sha256Hash string) (remaining int, ttlSecs int64, err error) {
	key := RedisKey(sha256Hash)

	pipe := rdb.Pipeline()
	getCmd := pipe.Get(ctx, key)
	ttlCmd := pipe.TTL(ctx, key)
	if _, err = pipe.Exec(ctx); err != nil && err != redis.Nil {
		return 0, 0, fmt.Errorf("pipeline: %w", err)
	}

	val, err := getCmd.Int()
	if err == redis.Nil {
		return -1, 0, nil
	}
	if err != nil {
		return 0, 0, err
	}

	ttl := ttlCmd.Val()
	if ttl < 0 {
		ttl = 0
	}
	return val, int64(ttl.Seconds()), nil
}

// deductScript atomically deducts cost tokens.
// Returns [remaining, ttlSecs] on success,
//
//	[-1, 0]             if key not in Redis,
//	[-2, ttlSecs]       if insufficient tokens.
var deductScript = redis.NewScript(`
local val = redis.call('GET', KEYS[1])
if val == false then return {-1, 0} end
local tokens = tonumber(val)
local cost   = tonumber(ARGV[1])
if tokens < cost then
  local ttl = redis.call('TTL', KEYS[1])
  if ttl < 0 then ttl = 0 end
  return {-2, ttl}
end
local remaining = tokens - cost
redis.call('SET', KEYS[1], remaining, 'KEEPTTL')
local ttl = redis.call('TTL', KEYS[1])
if ttl < 0 then ttl = 0 end
return {remaining, ttl}
`)

// Deduct subtracts cost from the bucket.
// Returns (remaining, ttlSecs, notFound, insufficient, err).
func Deduct(ctx context.Context, rdb *redis.Client, sha256Hash string, cost int) (remaining int, ttlSecs int64, notFound bool, insufficient bool, err error) {
	key := RedisKey(sha256Hash)
	res, err := deductScript.Run(ctx, rdb, []string{key}, cost).Int64Slice()
	if err != nil {
		return 0, 0, false, false, fmt.Errorf("lua: %w", err)
	}
	switch res[0] {
	case -1:
		return 0, 0, true, false, nil
	case -2:
		return 0, res[1], false, true, nil
	default:
		return int(res[0]), res[1], false, false, nil
	}
}
