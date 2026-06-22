package store_test

import (
	"context"
	"os"
	"testing"

	"github.com/Zartex-the-art/sei-ratelimiter/internal/store"

	"github.com/redis/go-redis/v9"
)

// redisClient creates a test Redis client.
// It skips the test if Redis is unavailable.
func redisClient(t *testing.T) *redis.Client {
	t.Helper()

	addr := os.Getenv("REDIS_URL")
	if addr == "" {
		addr = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		t.Skipf("Redis not available at %s — skipping integration test", addr)
	}

	t.Cleanup(func() {
		client.Close()
	})

	return client
}

func TestRedisConnection_Ping(t *testing.T) {
	client := redisClient(t)

	err := client.Ping(context.Background()).Err()
	if err != nil {
		t.Fatalf("PING failed: %v", err)
	}
}

func TestRedisConnection_SetAndGet(t *testing.T) {
	client := redisClient(t)

	ctx := context.Background()
	key := "test:day4:setget"

	t.Cleanup(func() {
		client.Del(ctx, key)
	})

	if err := client.Set(ctx, key, "hello", 0).Err(); err != nil {
		t.Fatalf("SET failed: %v", err)
	}

	val, err := client.Get(ctx, key).Result()
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}

	if val != "hello" {
		t.Errorf("got %q, want hello", val)
	}
}

func TestRedisConnection_IncrAndExpire(t *testing.T) {
	client := redisClient(t)

	ctx := context.Background()
	key := "test:day4:incr"

	t.Cleanup(func() {
		client.Del(ctx, key)
	})

	count, err := client.Incr(ctx, key).Result()
	if err != nil {
		t.Fatalf("INCR failed: %v", err)
	}

	if count != 1 {
		t.Errorf("got %d, want 1", count)
	}

	client.Expire(ctx, key, 60000000000)

	count, err = client.Incr(ctx, key).Result()
	if err != nil {
		t.Fatalf("second INCR failed: %v", err)
	}

	if count != 2 {
		t.Errorf("got %d, want 2", count)
	}
}
func TestRedisStore_Increment(t *testing.T) {
	client := redisClient(t)

	rs := store.NewRedisStore(client.Options().Addr)

	ctx := context.Background()
	key := "test:store:increment"

	t.Cleanup(func() {
		client.Del(ctx, key)
	})

	count, err := rs.Increment(ctx, key, 60)
	if err != nil {
		t.Fatalf("Increment failed: %v", err)
	}

	if count != 1 {
		t.Fatalf("got %d, want 1", count)
	}

	count, err = rs.Increment(ctx, key, 60)
	if err != nil {
		t.Fatalf("second Increment failed: %v", err)
	}

	if count != 2 {
		t.Fatalf("got %d, want 2", count)
	}
}

func TestRedisStore_HSetAndHGetAll(t *testing.T) {
	client := redisClient(t)

	rs := store.NewRedisStore(client.Options().Addr)

	ctx := context.Background()
	key := "test:store:hash"

	t.Cleanup(func() {
		client.Del(ctx, key)
	})

	err := rs.HSet(ctx, key, map[string]interface{}{
		"name": "alice",
		"role": "admin",
	})

	if err != nil {
		t.Fatalf("HSet failed: %v", err)
	}

	fields, err := rs.HGetAll(ctx, key)
	if err != nil {
		t.Fatalf("HGetAll failed: %v", err)
	}

	if fields["name"] != "alice" {
		t.Fatalf("got %q, want alice", fields["name"])
	}
}

func TestRedisStore_ZAddAndZCount(t *testing.T) {
	client := redisClient(t)

	rs := store.NewRedisStore(client.Options().Addr)

	ctx := context.Background()
	key := "test:store:zset"

	t.Cleanup(func() {
		client.Del(ctx, key)
	})

	if err := rs.ZAdd(ctx, key, 1, "a"); err != nil {
		t.Fatalf("ZAdd failed: %v", err)
	}

	if err := rs.ZAdd(ctx, key, 2, "b"); err != nil {
		t.Fatalf("second ZAdd failed: %v", err)
	}

	count, err := rs.ZCount(ctx, key, 0, 10)
	if err != nil {
		t.Fatalf("ZCount failed: %v", err)
	}

	if count != 2 {
		t.Fatalf("got %d, want 2", count)
	}
}

func TestRedisStore_Del(t *testing.T) {
	client := redisClient(t)

	rs := store.NewRedisStore(client.Options().Addr)

	ctx := context.Background()
	key := "test:store:delete"

	if err := rs.HSet(ctx, key, map[string]interface{}{
		"value": "x",
	}); err != nil {
		t.Fatalf("HSet failed: %v", err)
	}

	if err := rs.Del(ctx, key); err != nil {
		t.Fatalf("Del failed: %v", err)
	}

	fields, err := rs.HGetAll(ctx, key)
	if err != nil {
		t.Fatalf("HGetAll failed: %v", err)
	}

	if len(fields) != 0 {
		t.Fatalf("expected deleted hash")
	}
}

func TestRedisStore_PingAndClient(t *testing.T) {
	client := redisClient(t)

	rs := store.NewRedisStore(client.Options().Addr)

	if err := rs.Ping(context.Background()); err != nil {
		t.Fatalf("Ping failed: %v", err)
	}

	if rs.Client() == nil {
		t.Fatal("Client returned nil")
	}
}
