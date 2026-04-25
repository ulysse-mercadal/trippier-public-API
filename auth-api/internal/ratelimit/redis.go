// Package ratelimit manages token-bucket state in Redis.
package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisKey returns the Redis key for a user-level token bucket.
// Buckets are per-user, not per API key — all keys belonging to the same user
// draw from the same pool.
func RedisKey(userID string) string {
	return "rl:user:" + userID
}

// InitBucket primes the token bucket for a user only if it does not already
// exist in Redis (SETNX semantics). Creating a second API key does not reset
// the remaining tokens.
func InitBucket(ctx context.Context, rdb *redis.Client, userID string, limit int, ttl time.Duration) error {
	return rdb.SetArgs(ctx, RedisKey(userID), limit, redis.SetArgs{
		Mode:    "NX",
		TTL:     ttl,
		Get:     false,
		KeepTTL: false,
	}).Err()
}

// SetTokens force-sets (or resets) the token bucket for a user. Use this for
// administrative resets only — normal key creation should call InitBucket.
func SetTokens(ctx context.Context, rdb *redis.Client, userID string, limit int, ttl time.Duration) error {
	return rdb.Set(ctx, RedisKey(userID), limit, ttl).Err()
}

// GetUsage returns (remaining, ttlSecs). remaining == -1 means no bucket in Redis.
func GetUsage(ctx context.Context, rdb *redis.Client, userID string) (remaining int, ttlSecs int64, err error) {
	key := RedisKey(userID)

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
//	[-1, 0]            if bucket not in Redis,
//	[-2, ttlSecs]      if insufficient tokens.
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

// Deduct subtracts cost from the user's bucket.
// Returns (remaining, ttlSecs, notFound, insufficient, err).
func Deduct(ctx context.Context, rdb *redis.Client, userID string, cost int) (remaining int, ttlSecs int64, notFound bool, insufficient bool, err error) {
	key := RedisKey(userID)
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
