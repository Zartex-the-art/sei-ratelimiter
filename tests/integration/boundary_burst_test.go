package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/Zartex-the-art/sei-ratelimiter/internal/algorithms"
	"github.com/Zartex-the-art/sei-ratelimiter/internal/store"
	"github.com/Zartex-the-art/sei-ratelimiter/internal/testhelpers"
)
// TestBoundaryBurst_FixedWindow_AllowsDoubleLimit documents the
// known boundary burst problem in fixed window.
// This test PASSES — it records the bug, not reports a failure.
func TestBoundaryBurst_FixedWindow_AllowsDoubleLimit(t *testing.T) {
	client := testhelpers.RedisClient(t)
	defer testhelpers.FlushKeys(t, client, "fw:burst-fw")

	const limit = 5

	fw := algorithms.NewFixedWindow(limit, 1, store.NewRedisStore(redisAddr()))
	ctx := context.Background()

	var w1 int
	for i := 0; i < limit; i++ {
		ok, _, _ := fw.Allow(ctx, "burst-fw")
		if ok {
			w1++
		}
	}

	time.Sleep(1100 * time.Millisecond)

	var w2 int
	for i := 0; i < limit; i++ {
		ok, _, _ := fw.Allow(ctx, "burst-fw")
		if ok {
			w2++
		}
	}

	total := w1 + w2

	t.Logf("Fixed Window: w1=%d w2=%d total=%d (limit=%d per window)", w1, w2, total, limit)
	t.Logf("KNOWN BUG: %d requests passed in ~1s against limit of %d", total, limit)

	if total < limit*2 {
		t.Logf("Note: expected %d across boundary, got %d", limit*2, total)
	}
}
// TestBoundaryBurst_SlidingWindow_BlocksAtBoundary proves sliding window
// prevents the burst that fixed window allows.
func TestBoundaryBurst_SlidingWindow_BlocksAtBoundary(t *testing.T) {
	client := testhelpers.RedisClient(t)
	defer testhelpers.FlushKeys(t, client, "sw:burst-sw")

	const limit = 5

	sw := algorithms.NewSlidingWindow(limit, 2, store.NewRedisStore(redisAddr()))
	ctx := context.Background()

	// Send limit requests at t=0
	for i := 0; i < limit; i++ {
		ok, _, _ := sw.Allow(ctx, "burst-sw")
		if !ok {
			t.Fatalf("req %d should be allowed", i+1)
		}
	}

	// Wait 1 second — midpoint of the 2-second window
	time.Sleep(1000 * time.Millisecond)

	// Must be blocked
	ok, _, _ := sw.Allow(ctx, "burst-sw")

	if ok {
		t.Error("FAIL: sliding window allowed burst — boundary burst not prevented")
	} else {
		t.Log("PASS: burst correctly blocked — original requests still in window")
	}
}
// TestBoundaryBurst_Comparison runs both algorithms and compares results.
func TestBoundaryBurst_Comparison(t *testing.T) {
	client := testhelpers.RedisClient(t)

	defer func() {
		testhelpers.FlushKeys(t, client, "fw:compare")
		testhelpers.FlushKeys(t, client, "sw:compare")
	}()

	const limit = 3

	fw := algorithms.NewFixedWindow(limit, 1, store.NewRedisStore(redisAddr()))
	sw := algorithms.NewSlidingWindow(limit, 2, store.NewRedisStore(redisAddr()))

	ctx := context.Background()

	// Fixed window
	for i := 0; i < limit; i++ {
		fw.Allow(ctx, "compare")
	}

	time.Sleep(1100 * time.Millisecond)

	var fwBurst int
	for i := 0; i < limit; i++ {
		ok, _, _ := fw.Allow(ctx, "compare")
		if ok {
			fwBurst++
		}
	}

	// Sliding window
	for i := 0; i < limit; i++ {
		sw.Allow(ctx, "compare")
	}

	time.Sleep(1000 * time.Millisecond)

	var swBurst int
	for i := 0; i < limit; i++ {
		ok, _, _ := sw.Allow(ctx, "compare")
		if ok {
			swBurst++
		}
	}

	t.Logf("Fixed Window burst: %d/%d — no protection", fwBurst, limit)
	t.Logf("Sliding Window burst: %d/%d — should be 0", swBurst, limit)

	if swBurst > 0 {
		t.Errorf("Sliding window allowed %d burst requests — want 0", swBurst)
	}
}
