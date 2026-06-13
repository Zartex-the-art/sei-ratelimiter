package store

import (
	"context"
	"errors"
)

// ErrorStore is a Store implementation that always returns infrastructure errors.
// Used to test graceful degradation in HTTP handlers.
type ErrorStore struct {
	Err error // error to return on all calls
}

// NewErrorStore creates an ErrorStore with a Redis-like connection error.
func NewErrorStore() *ErrorStore {
	return &ErrorStore{
		Err: errors.New("connection refused: redis is down"),
	}
}
func (es *ErrorStore) Increment(ctx context.Context, key string, windowSecs int) (int64,
	error) {
	return 0, es.Err
}
func (es *ErrorStore) ZAdd(ctx context.Context, key string, score float64, member string) error {
	return es.Err
}
func (es *ErrorStore) ZRemRangeByScore(ctx context.Context, key string, min, max float64) error {
	return es.Err
}
func (es *ErrorStore) ZCount(ctx context.Context, key string, min, max float64) (int64, error) {
	return 0, es.Err
}
func (es *ErrorStore) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return nil, es.Err
}
func (es *ErrorStore) HSet(ctx context.Context, key string, values map[string]interface{}) error {
	return es.Err
}
func (es *ErrorStore) Del(ctx context.Context, keys ...string) error {
	return es.Err
}
func (es *ErrorStore) SlidingWindowAllow(ctx context.Context, key string, nowMs, windowMs int64, member string, limit int) (int64, error) {
	return 0, es.Err
}
func (es *ErrorStore) TokenBucketAllow(ctx context.Context, key string, nowMs int64, limit, windowSecs int) (int, error) {
	return 0, es.Err
}
