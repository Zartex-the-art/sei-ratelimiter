package integration

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/Zartex-the-art/sei-ratelimiter/internal/algorithms"
	"github.com/Zartex-the-art/sei-ratelimiter/internal/store"
)

func TestFixedWindow_RedisBacked(t *testing.T) {
	conn, err := net.DialTimeout("tcp", "localhost:6379", 1*time.Second)
	if err != nil {
		t.Skipf("Redis not available - skipping: %v", err)
	}
	conn.Close()

	ctx := context.Background()
	redisStore := store.NewRedisStore("localhost:6379")
	redisStore.Del(ctx, "fw:client-1")

	limiter := algorithms.NewFixedWindow(3, 60, redisStore)

	for i := 1; i <= 3; i++ {
		allowed, remaining, err := limiter.Allow(ctx, "client-1")
		if err != nil {
			t.Fatalf("request %d: error: %v", i, err)
		}
		if !allowed {
			t.Errorf("request %d: expected allowed", i)
		}
		if remaining != 3-i {
			t.Errorf("request %d: remaining=%d, want %d", i, remaining, 3-i)
		}
	}

	allowed, remaining, err := limiter.Allow(ctx, "client-1")
	if err != nil {
		t.Fatalf("request 4: error: %v", err)
	}
	if allowed {
		t.Error("request 4: expected blocked")
	}
	if remaining != 0 {
		t.Errorf("request 4: remaining=%d, want 0", remaining)
	}

	redisStore.Del(ctx, "fw:client-1")
}
