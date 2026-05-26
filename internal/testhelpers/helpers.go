package testhelpers

import (
	"context"
	"os"
	"testing"

	"github.com/redis/go-redis/v9"
)

// RedisClient creates a test Redis client.
// Skips the calling test if Redis is not reachable.
// Never calls t.Fatal — skipping is the correct behaviour for integration tests.
//
// Usage:
//
//	client := testhelpers.RedisClient(t)
//	// client is ready to use — test will be skipped if Redis is down
func RedisClient(t *testing.T) *redis.Client {
	t.Helper()

	addr := os.Getenv("REDIS_URL")
	if addr == "" {
		addr = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		client.Close()
		t.Skipf("Redis not available at %s (set REDIS_URL env var): %v", addr, err)
	}
	t.Cleanup(func() {
		client.Close()
	})

	return client
}

// FlushKeys deletes all Redis keys matching pattern.
// Always call this in t.Cleanup() to prevent test state from leaking.
//
// Usage:
//
//	t.Cleanup(func() {
//	    testhelpers.FlushKeys(t, client, "test:*")
//	})
func FlushKeys(t *testing.T, client *redis.Client, pattern string) {
	t.Helper()
	ctx := context.Background()
	keys, err := client.Keys(ctx, pattern).Result()
	if err != nil || len(keys) == 0 {
		return
	}
	if err := client.Del(ctx, keys...).Err(); err != nil {
		t.Logf("warning: cleanup failed for pattern %q: %v", pattern, err)
	}
}

// AssertEqual fails the test if got != want.
// A simple helper to reduce boilerplate in table-driven tests.
func AssertEqual(t *testing.T, name string, got, want interface{}) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %v, want %v", name, got, want)
	}
}
// Cleaner tracks Redis keys to delete after a test.
type Cleaner struct {
    t      *testing.T
    client *redis.Client
}
func NewCleaner(t *testing.T, client *redis.Client) *Cleaner {
    return &Cleaner{t: t, client: client}
}
// Del deletes the specified Redis keys.
func (c *Cleaner) Del(keys ...string) {
    c.t.Helper()
    if err := c.client.Del(context.Background(), keys...).Err(); err != nil {
        c.t.Logf("cleanup warning: Del %v: %v", keys, err)
    }
}