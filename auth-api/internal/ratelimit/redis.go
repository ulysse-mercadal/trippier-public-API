// Package ratelimit manages token-bucket state in Redis.
package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisKey builds the shared per-user rate-limit bucket key for userID.
func RedisKey(userID string) string {
	return "rl:user:" + userID
}

// InitBucket creates the bucket for userID with the given limit and ttl only
// if it doesn't already exist, so issuing a new API key doesn't reset
// existing token counts. It returns an error if the operation fails.
func InitBucket(ctx context.Context, rdb *redis.Client, userID string, limit int, ttl time.Duration) error {
	return rdb.SetArgs(ctx, RedisKey(userID), limit, redis.SetArgs{
		Mode:    "NX",
		TTL:     ttl,
		Get:     false,
		KeepTTL: false,
	}).Err()
}

// SetTokens force-resets userID's bucket to limit tokens with the given ttl;
// use only for admin resets, not normal key creation (see InitBucket). It
// returns an error if the operation fails.
func SetTokens(ctx context.Context, rdb *redis.Client, userID string, limit int, ttl time.Duration) error {
	return rdb.Set(ctx, RedisKey(userID), limit, ttl).Err()
}

// GetUsage reads the current bucket state for userID. It returns the
// remaining tokens (-1 if no bucket exists in Redis), the number of seconds
// until the bucket expires, and an error if the read fails.
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

// deductScript atomically deducts cost tokens, returning [remaining, ttl], [-1,0] if absent, or [-2, ttl] if insufficient.
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

// Deduct atomically subtracts cost tokens from userID's bucket. It returns
// the remaining tokens after deduction, the seconds until the bucket
// expires, notFound if no bucket exists, insufficient if there weren't
// enough tokens to cover cost, and err if the script execution fails.
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
