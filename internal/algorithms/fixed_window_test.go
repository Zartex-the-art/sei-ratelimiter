package algorithms_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Zartex-the-art/sei-ratelimiter/internal/algorithms"
	"github.com/Zartex-the-art/sei-ratelimiter/internal/store"
	"github.com/Zartex-the-art/sei-ratelimiter/internal/testhelpers"
)

var ctx = context.Background()

// newFakeWindow creates a FixedWindow with FakeStore — no Redis needed.
func newFakeWindow(limit, windowSecs int) *algorithms.FixedWindow {
	return algorithms.NewFixedWindow(
		limit,
		windowSecs,
		store.NewFakeStore(),
	)
}

// newRedisWindow creates a FixedWindow with real Redis.
// Skips if Redis is unavailable.
func newRedisWindow(
	t *testing.T,
	limit,
	windowSecs int,
) (*algorithms.FixedWindow, *testhelpers.Cleaner) {

	t.Helper()

	client := testhelpers.RedisClient(t)

	rs := store.NewRedisStore("localhost:6379")

	fw := algorithms.NewFixedWindow(
		limit,
		windowSecs,
		rs,
	)

	return fw, testhelpers.NewCleaner(t, client)
}

// ─── Unit Tests — FakeStore, no Redis, no Docker ────────────────────────────

func TestFixedWindow_AllowsUnderLimit(t *testing.T) {
	fw := newFakeWindow(10, 60)

	for i := 0; i < 5; i++ {
		ok, remaining, err := fw.Allow(ctx, "alice")

		if err != nil {
			t.Fatalf("req %d: error: %v", i+1, err)
		}

		if !ok {
			t.Fatalf("req %d: should be allowed", i+1)
		}

		want := 10 - (i + 1)

		if remaining != want {
			t.Errorf(
				"req %d: remaining got %d, want %d",
				i+1,
				remaining,
				want,
			)
		}
	}
}

func TestFixedWindow_BlocksAtLimit(t *testing.T) {
	fw := newFakeWindow(10, 60)

	for i := 0; i < 10; i++ {
		ok, _, _ := fw.Allow(ctx, "alice")

		if !ok {
			t.Fatalf("req %d should be allowed", i+1)
		}
	}

	ok, remaining, err := fw.Allow(ctx, "alice")

	if err != nil {
		t.Fatalf("error: %v", err)
	}

	if ok {
		t.Error("11th request should be blocked")
	}

	if remaining != 0 {
		t.Errorf("remaining: got %d, want 0", remaining)
	}
}

func TestFixedWindow_BlocksOverLimit(t *testing.T) {
	fw := newFakeWindow(10, 60)

	var allowed, blocked int

	for i := 0; i < 20; i++ {
		ok, _, _ := fw.Allow(ctx, "alice")

		if ok {
			allowed++
		} else {
			blocked++
		}
	}

	if allowed != 10 {
		t.Errorf("allowed: got %d, want 10", allowed)
	}

	if blocked != 10 {
		t.Errorf("blocked: got %d, want 10", blocked)
	}
}

func TestFixedWindow_RemainingDecrement(t *testing.T) {
	fw := newFakeWindow(5, 60)

	wants := []int{4, 3, 2, 1, 0}

	for i, want := range wants {
		_, got, _ := fw.Allow(ctx, "alice")

		if got != want {
			t.Errorf(
				"req %d: remaining got %d, want %d",
				i+1,
				got,
				want,
			)
		}
	}
}

func TestFixedWindow_MultipleClientsAreIndependent(t *testing.T) {
	fw := newFakeWindow(2, 60)

	fw.Allow(ctx, "alice")
	fw.Allow(ctx, "alice")

	okA, _, _ := fw.Allow(ctx, "alice")

	if okA {
		t.Error("alice should be blocked")
	}

	okB, _, _ := fw.Allow(ctx, "bob")

	if !okB {
		t.Error("bob should be allowed")
	}
}

func TestFixedWindow_ConcurrentAllows_ExactCount(t *testing.T) {
	const limit = 100
	const goroutines = 500

	fw := newFakeWindow(limit, 60)

	var wg sync.WaitGroup
	var allowed int64

	for i := 0; i < goroutines; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			ok, _, _ := fw.Allow(ctx, "concurrent")

			if ok {
				atomic.AddInt64(&allowed, 1)
			}
		}()
	}

	wg.Wait()

	if int(allowed) != limit {
		t.Errorf(
			"allowed: got %d, want %d",
			allowed,
			limit,
		)
	}
}

func TestFixedWindow_ConcurrentMultipleClients(t *testing.T) {
	const limit = 50

	fw := newFakeWindow(limit, 60)

	var wg sync.WaitGroup

	var allowedA int64
	var allowedB int64

	for i := 0; i < 200; i++ {
		wg.Add(2)

		go func() {
			defer wg.Done()

			ok, _, _ := fw.Allow(ctx, "A")

			if ok {
				atomic.AddInt64(&allowedA, 1)
			}
		}()

		go func() {
			defer wg.Done()

			ok, _, _ := fw.Allow(ctx, "B")

			if ok {
				atomic.AddInt64(&allowedB, 1)
			}
		}()
	}

	wg.Wait()

	if int(allowedA) != limit {
		t.Errorf("A: got %d, want %d", allowedA, limit)
	}

	if int(allowedB) != limit {
		t.Errorf("B: got %d, want %d", allowedB, limit)
	}
}

// ─── Redis Integration Tests — require docker compose up ─────────────────────

func TestFixedWindow_Redis_AllowsUnderLimit(t *testing.T) {
	fw, c := newRedisWindow(t, 10, 60)

	defer c.Del("fw:redis-under")

	for i := 0; i < 5; i++ {
		ok, _, err := fw.Allow(ctx, "redis-under")

		if err != nil {
			t.Fatalf("req %d: error: %v", i+1, err)
		}

		if !ok {
			t.Fatalf("req %d: should be allowed", i+1)
		}
	}
}

func TestFixedWindow_Redis_BlocksAtLimit(t *testing.T) {
	fw, c := newRedisWindow(t, 5, 60)

	defer c.Del("fw:redis-limit")

	for i := 0; i < 5; i++ {
		ok, _, _ := fw.Allow(ctx, "redis-limit")

		if !ok {
			t.Fatalf("req %d should be allowed", i+1)
		}
	}

	ok, _, _ := fw.Allow(ctx, "redis-limit")

	if ok {
		t.Error("6th request should be blocked")
	}
}

func TestFixedWindow_Redis_WindowExpiry(t *testing.T) {
	// 1-second window — tests real Redis TTL expiry

	fw, c := newRedisWindow(t, 3, 1)

	defer c.Del("fw:redis-expiry")

	for i := 0; i < 3; i++ {
		ok, _, _ := fw.Allow(ctx, "redis-expiry")

		if !ok {
			t.Fatalf("req %d should be allowed", i+1)
		}
	}

	ok, _, _ := fw.Allow(ctx, "redis-expiry")

	if ok {
		t.Fatal("4th request should be blocked")
	}

	// Wait for Redis TTL expiry
	time.Sleep(1100 * time.Millisecond)

	ok, remaining, err := fw.Allow(ctx, "redis-expiry")

	if err != nil {
		t.Fatalf("after expiry: error: %v", err)
	}

	if !ok {
		t.Fatal("should be allowed after window expiry")
	}

	if remaining != 2 {
		t.Errorf(
			"remaining after expiry: got %d, want 2",
			remaining,
		)
	}
}
