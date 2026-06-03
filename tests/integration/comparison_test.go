package integration

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/Zartex-the-art/sei-ratelimiter/internal/algorithms"
	"github.com/Zartex-the-art/sei-ratelimiter/internal/store"
)

func TestFixedWindow_InMemoryVsRedis(t *testing.T) {
	ctx := context.Background()
	memStore := store.NewMemoryStore()
	memLimiter := algorithms.NewFixedWindow(memStore, 5, 60)

	conn, err := net.DialTimeout("tcp", "localhost:6379", 1*time.Second)
	if err != nil {
		t.Skipf("Redis not available - skipping comparison")
	}
	conn.Close()

	redisStore := store.NewRedisStore("localhost:6379")
	redisStore.Del(ctx, "test:compare:client-1")
	redisLimiter := algorithms.NewFixedWindow(redisStore, 5, 60)

	for i := 1; i <= 5; i++ {
		memAllowed, memRemaining, _ := memLimiter.Allow(ctx, "client-1")
		redisAllowed, redisRemaining, _ := redisLimiter.Allow(ctx, "client-1")
		if memAllowed != redisAllowed {
			t.Errorf("request %d: allowed mismatch", i)
		}
		if memRemaining != redisRemaining {
			t.Errorf("request %d: remaining mismatch", i)
		}
	}

	redisStore.Del(ctx, "test:compare:client-1")
}
