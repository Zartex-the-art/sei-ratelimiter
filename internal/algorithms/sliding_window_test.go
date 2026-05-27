package algorithms_test

import (
	"context"
	"testing"
	"time"

	"github.com/Zartex-the-art/sei-ratelimiter/internal/algorithms"
	"github.com/Zartex-the-art/sei-ratelimiter/internal/store"
	"github.com/Zartex-the-art/sei-ratelimiter/internal/testhelpers"
)

// newSW creates a SlidingWindow backed by real Redis.
// Skips if Redis is not available.
func newSW(t *testing.T, limit, windowSecs int) *algorithms.SlidingWindow {
	t.Helper()

	testhelpers.RedisClient(t)

	rs := store.NewRedisStore("localhost:6379")
	sw := algorithms.NewSlidingWindow(limit, windowSecs, rs)

	return sw
}

func TestSlidingWindow_AllowsUnderLimit(t *testing.T) {
	sw := newSW(t, 10, 60)

	for i := 0; i < 5; i++ {
		ok, remaining, err := sw.Allow(context.Background(), "sw-under")

		if err != nil {
			t.Fatalf("req %d: error: %v", i+1, err)
		}

		if !ok {
			t.Fatalf("req %d: should be allowed", i+1)
		}

		want := 10 - (i + 1)

		if remaining != want {
			t.Errorf("req %d: remaining got %d, want %d", i+1, remaining, want)
		}
	}
}

func TestSlidingWindow_BlocksAtLimit(t *testing.T) {
	sw := newSW(t, 5, 60)

	ctx := context.Background()

	for i := 0; i < 5; i++ {
		ok, _, _ := sw.Allow(ctx, "sw-limit")

		if !ok {
			t.Fatalf("req %d should be allowed", i+1)
		}
	}

	ok, remaining, err := sw.Allow(ctx, "sw-limit")

	if err != nil {
		t.Fatalf("error: %v", err)
	}

	if ok {
		t.Error("6th request should be blocked")
	}

	if remaining != 0 {
		t.Errorf("remaining: got %d, want 0", remaining)
	}
}

func TestSlidingWindow_PrunesOldEntries(t *testing.T) {
	sw := newSW(t, 10, 1) // 1-second window

	ctx := context.Background()

	// 3 requests — all allowed
	for i := 0; i < 3; i++ {
		ok, _, _ := sw.Allow(ctx, "sw-prune")

		if !ok {
			t.Fatalf("req %d should be allowed", i+1)
		}
	}

	// Wait for window to slide past those entries
	time.Sleep(1100 * time.Millisecond)

	// 3 more — old ones pruned, fresh count
	for i := 0; i < 3; i++ {
		ok, _, err := sw.Allow(ctx, "sw-prune")

		if err != nil {
			t.Fatalf("after slide req %d: error: %v", i+1, err)
		}

		if !ok {
			t.Fatalf("after slide req %d: should be allowed — old entries pruned", i+1)
		}
	}
}

// TestSlidingWindow_NoBoundaryBurst is the KEY TEST for this algorithm.
// It proves sliding window prevents the burst that fixed window allows.
func TestSlidingWindow_NoBoundaryBurst(t *testing.T) {
	sw := newSW(t, 5, 2) // 2-second window

	ctx := context.Background()

	// Send full limit at t=0
	for i := 0; i < 5; i++ {
		ok, _, _ := sw.Allow(ctx, "sw-burst")

		if !ok {
			t.Fatalf("req %d should be allowed", i+1)
		}
	}

	// Wait 1 second — midpoint of 2-second window
	// The 5 original requests are still inside the window
	time.Sleep(1000 * time.Millisecond)

	// Must be BLOCKED — original 5 still counted
	ok, _, _ := sw.Allow(ctx, "sw-burst")

	if ok {
		t.Error("BOUNDARY BURST DETECTED: request allowed mid-window — sliding window must prevent this")
	}

	// Wait another 1.1 seconds — original 5 now slide out of the 2-second window
	time.Sleep(1100 * time.Millisecond)

	// Now allowed — window is clear
	ok, remaining, err := sw.Allow(ctx, "sw-burst")

	if err != nil {
		t.Fatalf("after full slide: error: %v", err)
	}

	if !ok {
		t.Fatal("should be allowed after window slides past original entries")
	}

	if remaining != 4 {
		t.Errorf("remaining: got %d, want 4", remaining)
	}
}

func TestSlidingWindow_MultipleClientsAreIndependent(t *testing.T) {
	sw := newSW(t, 3, 60)

	ctx := context.Background()

	for i := 0; i < 3; i++ {
		sw.Allow(ctx, "sw-ca")
	}

	okA, _, _ := sw.Allow(ctx, "sw-ca")

	if okA {
		t.Error("client-a should be blocked")
	}

	okB, remainingB, _ := sw.Allow(ctx, "sw-cb")

	if !okB {
		t.Error("client-b should be allowed")
	}

	if remainingB != 2 {
		t.Errorf("client-b remaining: got %d, want 2", remainingB)
	}
}

func TestSlidingWindow_CountsOnlyCurrentWindow(t *testing.T) {
	sw := newSW(t, 5, 1) // 1-second window

	ctx := context.Background()

	// Fill the limit
	for i := 0; i < 5; i++ {
		sw.Allow(ctx, "sw-window")
	}

	ok, _, _ := sw.Allow(ctx, "sw-window")

	if ok {
		t.Fatal("should be blocked at limit")
	}

	// Wait for full window expiry
	time.Sleep(1200 * time.Millisecond)

	// All previous requests are outside the window
	ok, remaining, err := sw.Allow(ctx, "sw-window")

	if err != nil {
		t.Fatalf("error: %v", err)
	}

	if !ok {
		t.Fatal("should be allowed — window expired")
	}

	if remaining != 4 {
		t.Errorf("remaining: got %d, want 4", remaining)
	}
}
