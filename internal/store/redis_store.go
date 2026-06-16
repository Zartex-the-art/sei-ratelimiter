package store

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// fixedWindowScript atomically increments a counter and sets expiry
// on the first increment (new window).
var fixedWindowScript = redis.NewScript(`
local count = redis.call('INCR', KEYS[1])
if count == 1 then
	redis.call('EXPIRE', KEYS[1], tonumber(ARGV[1]))
end
return count
`)

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(addr string) *RedisStore {
	client := redis.NewClient(&redis.Options{
		Addr: addr,

		// Connection timeouts
		DialTimeout:  2 * time.Second,
		ReadTimeout:  500 * time.Millisecond,
		WriteTimeout: 500 * time.Millisecond,

		// Retry on transient failures
		MaxRetries:      3,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,

		// Connection pool
		PoolSize:     100,
		MinIdleConns: 20,
		PoolTimeout:  2 * time.Second,
	})

	return &RedisStore{
		client: client,
	}
}

func (rs *RedisStore) Increment(
	ctx context.Context,
	key string,
	windowSecs int,
) (int64, error) {

	count, err := fixedWindowScript.Run(
		ctx,
		rs.client,
		[]string{key},
		windowSecs,
	).Int64()

	if err != nil {
		return 0, fmt.Errorf("Increment Lua: %w", err)
	}

	return count, nil
}

func (r *RedisStore) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

func (r *RedisStore) ZAdd(
	ctx context.Context,
	key string,
	score float64,
	member string,
) error {
	return r.client.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: member,
	}).Err()
}

func (r *RedisStore) ZRemRangeByScore(
	ctx context.Context,
	key string,
	min, max float64,
) error {
	return r.client.ZRemRangeByScore(
		ctx,
		key,
		fmt.Sprintf("%f", min),
		fmt.Sprintf("%f", max),
	).Err()
}

func (r *RedisStore) ZCount(
	ctx context.Context,
	key string,
	min, max float64,
) (int64, error) {
	return r.client.ZCount(
		ctx,
		key,
		fmt.Sprintf("%f", min),
		fmt.Sprintf("%f", max),
	).Result()
}

func (r *RedisStore) HGetAll(
	ctx context.Context,
	key string,
) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

func (r *RedisStore) HSet(
	ctx context.Context,
	key string,
	values map[string]interface{},
) error {
	return r.client.HSet(ctx, key, values).Err()
}

func (r *RedisStore) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (rs *RedisStore) Client() *redis.Client {
	return rs.client
}

// slidingWindowScript atomically manages the sliding window sorted set.
//
// KEYS[1] = sorted set key (e.g., "sw:alice")
// ARGV[1] = current timestamp in milliseconds
// ARGV[2] = window size in milliseconds (windowSecs * 1000)
// ARGV[3] = unique request member (nanosecond timestamp as string)
// ARGV[4] = rate limit (max requests per window)
//
// Returns: count of entries in window after operation.
// If count > limit, the member was removed (request blocked).
var slidingWindowScript = redis.NewScript(`
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local member = ARGV[3]
local limit = tonumber(ARGV[4])
-- Step 1: Record this request
redis.call('ZADD', key, now, member)
-- Step 2: Prune entries outside the sliding window
redis.call('ZREMRANGEBYSCORE', key, 0, now - window)
-- Step 3: Count remaining entries
local count = redis.call('ZCOUNT', key, '-inf', '+inf')
-- Step 4: If over limit, remove the entry we just added
if count > limit then
 redis.call('ZREM', key, member)
end
return count
`)

// tokenBucketScript atomically manages the token bucket.
//
// KEYS[1] = hash key (e.g., "tb:alice")
// ARGV[1] = current timestamp in milliseconds
// ARGV[2] = bucket capacity (limit)
// ARGV[3] = window size in seconds (used to compute refill rate)
//
// Returns: remaining token count (floor) if allowed, -1 if blocked.
// On allowed: updates tokens and last_refill in the hash.
// On blocked: does not modify state.
var tokenBucketScript = redis.NewScript(`
local key = KEYS[1]
local now = tonumber(ARGV[1])
local limit = tonumber(ARGV[2])
local windowSec = tonumber(ARGV[3])
-- Rate: tokens per millisecond
local ratePerMs = limit / (windowSec * 1000)
-- Read current state
local data = redis.call('HGETALL', key)
local tokens, lastRefill
if #data == 0 then
 -- First request: full bucket
 tokens = limit
 lastRefill = now
else
 -- Parse hash into a table
 local h = {}
 for i = 1, #data, 2 do
 h[data[i]] = data[i+1]
 end
 tokens = tonumber(h['tokens'])
 lastRefill = tonumber(h['last_refill'])
end
-- Compute token refill based on elapsed time
local elapsed = math.max(0, now - lastRefill)
local toAdd = elapsed * ratePerMs
tokens = tokens + toAdd
if tokens > limit then tokens = limit end
-- Consume one token if available
if tokens >= 1 then
 tokens = tokens - 1
 -- Persist updated state
 redis.call('HSET', key,
 'tokens', tostring(tokens),
 'last_refill', tostring(now))
 return math.floor(tokens)
end
-- Bucket empty — blocked, do not update state
return -1
`)

func (rs *RedisStore) SlidingWindowAllow(
	ctx context.Context,
	key string,
	nowMs, windowMs int64,
	member string,
	limit int,
) (int64, error) {

	count, err := slidingWindowScript.Run(
		ctx,
		rs.client,
		[]string{key},
		nowMs,
		windowMs,
		member,
		limit,
	).Int64()

	if err != nil {
		return 0, fmt.Errorf("SlidingWindowAllow Lua: %w", err)
	}

	return count, nil
}

func (rs *RedisStore) TokenBucketAllow(
	ctx context.Context,
	key string,
	nowMs int64,
	limit, windowSecs int,
) (int, error) {

	remaining, err := tokenBucketScript.Run(
		ctx,
		rs.client,
		[]string{key},
		nowMs,
		limit,
		windowSecs,
	).Int()

	if err != nil {
		return 0, fmt.Errorf("TokenBucketAllow Lua: %w", err)
	}

	return remaining, nil
}
