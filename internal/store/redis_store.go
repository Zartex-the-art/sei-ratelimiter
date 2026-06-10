package store

import (
	"context"
	"fmt"

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