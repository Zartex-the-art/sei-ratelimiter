package store

import "context"

type Store interface {
	Increment(ctx context.Context, key string, windowSecs int) (int64, error)

	ZAdd(ctx context.Context, key string, score float64, member string) error
	ZRemRangeByScore(ctx context.Context, key string, min, max float64) error
	ZCount(ctx context.Context, key string, min, max float64) (int64, error)

	HGetAll(ctx context.Context, key string) (map[string]string, error)
	HSet(ctx context.Context, key string, values map[string]interface{}) error

	Del(ctx context.Context, keys ...string) error

	// Day 16
	SlidingWindowAllow(
		ctx context.Context,
		key string,
		nowMs, windowMs int64,
		member string,
		limit int,
	) (int64, error)

	TokenBucketAllow(
		ctx context.Context,
		key string,
		nowMs int64,
		limit, windowSecs int,
	) (int, error)
}
