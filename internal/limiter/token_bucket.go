package limiter

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// luaScript is the Lua script to handle token bucket logic atomically in Redis.
// Keys:
// 1. rate_limiter:{key} (Hash) - Stores "tokens" and "last_refilled"
// Args:
// 1. capacity (int)
// 2. rate (float64) - tokens per second
// 3. now (int64) - current unix timestamp in microseconds (or milliseconds, let's use microseconds for precision)
// 4. requested (int) - tokens requested (default 1)
const luaScript = `
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local requested = tonumber(ARGV[4])

local fill_time = capacity / rate
local ttl = fill_time * 2

local info = redis.call("HMGET", key, "tokens", "last_refilled")
local tokens = tonumber(info[1])
local last_refilled = tonumber(info[2])

if not tokens then
    tokens = capacity
    last_refilled = now
end

local elapsed = (now - last_refilled) / 1000000 -- Convert microseconds to seconds
local added = elapsed * rate
tokens = math.min(capacity, tokens + added)

if tokens >= requested then
    tokens = tokens - requested
    last_refilled = now
    redis.call("HMSET", key, "tokens", tokens, "last_refilled", last_refilled)
    redis.call("EXPIRE", key, math.ceil(ttl))
    return 1
else
    return 0
end
`

// TokenBucketLimiter implements a rate limiter using the Token Bucket algorithm on Redis.
type TokenBucketLimiter struct {
	client   *redis.Client
	capacity int
	rate     float64
	script   *redis.Script
}

// NewTokenBucketLimiter creates a new TokenBucketLimiter.
// client: Redis client
// capacity: Maximum number of tokens in the bucket
// rate: Token refill rate per second
func NewTokenBucketLimiter(client *redis.Client, capacity int, rate float64) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		client:   client,
		capacity: capacity,
		rate:     rate,
		script:   redis.NewScript(luaScript),
	}
}

// Allow checks if a request is allowed for the given key.
// It effectively consumes 1 token.
func (l *TokenBucketLimiter) Allow(ctx context.Context, key string) (bool, error) {
	now := time.Now().UnixMicro()
	keys := []string{"rate_limiter:" + key}
	args := []interface{}{l.capacity, l.rate, now, 1}

	result, err := l.script.Run(ctx, l.client, keys, args...).Result()
	if err != nil {
		return false, err
	}

	return result.(int64) == 1, nil
}
