package algorithms

import (
	"context"
	"testing"
)

type fakeStore struct {
	count int64
}

func (f *fakeStore) Increment(ctx context.Context, key string, windowSecs int) (int64, error) {
	f.count++
	return f.count, nil
}

func (f *fakeStore) ZAdd(ctx context.Context, key string, score float64, member string) error {
	return nil
}

func (f *fakeStore) ZRemRangeByScore(ctx context.Context, key string, min, max float64) error {
	return nil
}

func (f *fakeStore) ZCount(ctx context.Context, key string, min, max float64) (int64, error) {
	return 0, nil
}

func (f *fakeStore) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return map[string]string{}, nil
}

func (f *fakeStore) HSet(ctx context.Context, key string, values map[string]interface{}) error {
	return nil
}

func (f *fakeStore) Del(ctx context.Context, keys ...string) error {
	return nil
}
func (f *fakeStore) SlidingWindowAllow(
	ctx context.Context,
	key string,
	nowMs, windowMs int64,
	member string,
	limit int,
) (int64, error) {
	f.count++
	return f.count, nil
}

func (f *fakeStore) TokenBucketAllow(
	ctx context.Context,
	key string,
	nowMs int64,
	limit, windowSecs int,
) (int, error) {
	return limit - 1, nil
}

func TestFixedWindow_AllowsUnderLimit(t *testing.T) {
	store := &fakeStore{}
	fw := NewFixedWindow(5, 60, store)

	ctx := context.Background()

	allowed, remaining, err := fw.Allow(ctx, "client-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !allowed {
		t.Error("expected allowed=true for first request")
	}

	if remaining != 4 {
		t.Errorf("expected remaining=4, got %d", remaining)
	}
}

func TestFixedWindow_BlocksAtLimit(t *testing.T) {
	store := &fakeStore{}
	fw := NewFixedWindow(2, 60, store)

	ctx := context.Background()

	fw.Allow(ctx, "client-1")
	fw.Allow(ctx, "client-1")

	allowed, _, err := fw.Allow(ctx, "client-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if allowed {
		t.Error("expected allowed=false when limit exceeded")
	}
}
