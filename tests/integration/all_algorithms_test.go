package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/Zartex-the-art/sei-ratelimiter/internal/algorithms"
	"github.com/Zartex-the-art/sei-ratelimiter/internal/store"
	"github.com/Zartex-the-art/sei-ratelimiter/internal/testhelpers"
)

func TestAllAlgorithms_BurstBehaviour(t *testing.T) {
	client := testhelpers.RedisClient(t)

	defer func() {
		testhelpers.FlushKeys(t, client, "fw:burst-cmp")
		testhelpers.FlushKeys(t, client, "sw:burst-cmp")
		testhelpers.FlushKeys(t, client, "tb:burst-cmp")
	}()

	rs := store.NewRedisStore(redisAddr())
	ctx := context.Background()

	const limit = 5

	fw := algorithms.NewFixedWindow(limit, 60, rs)
	sw := algorithms.NewSlidingWindow(limit, 60, rs)
	tb := algorithms.NewTokenBucket(limit, 60, rs)

	var fwAllowed, swAllowed, tbAllowed int

	for i := 0; i < limit; i++ {
		ok, _, _ := fw.Allow(ctx, "burst-cmp")
		if ok {
			fwAllowed++
		}

		ok, _, _ = sw.Allow(ctx, "burst-cmp")
		if ok {
			swAllowed++
		}

		ok, _, _ = tb.Allow(ctx, "burst-cmp")
		if ok {
			tbAllowed++
		}
	}

	t.Logf("Burst of %d requests: FW=%d SW=%d TB=%d allowed",
		limit, fwAllowed, swAllowed, tbAllowed)

	if fwAllowed != limit {
		t.Errorf("FW: got %d, want %d", fwAllowed, limit)
	}
	if swAllowed != limit {
		t.Errorf("SW: got %d, want %d", swAllowed, limit)
	}
	if tbAllowed != limit {
		t.Errorf("TB: got %d, want %d", tbAllowed, limit)
	}
}
func TestTokenBucket_vs_FixedWindow_AfterIdle(t *testing.T) {
	client := testhelpers.RedisClient(t)

	defer func() {
		testhelpers.FlushKeys(t, client, "fw:idle-cmp")
		testhelpers.FlushKeys(t, client, "tb:idle-cmp")
	}()

	rs := store.NewRedisStore(redisAddr())
	ctx := context.Background()

	// rate = 3/3 = 1 token per second
	fw := algorithms.NewFixedWindow(3, 3, rs)
	tb := algorithms.NewTokenBucket(3, 3, rs)

	// Exhaust both
	for i := 0; i < 3; i++ {
		fw.Allow(ctx, "idle-cmp")
		tb.Allow(ctx, "idle-cmp")
	}

	// Wait 2 seconds
	time.Sleep(2100 * time.Millisecond)

	okFW, _, _ := fw.Allow(ctx, "idle-cmp")
	okTB, _, _ := tb.Allow(ctx, "idle-cmp")

	t.Logf("After 2s idle: FW allowed=%v, TB allowed=%v", okFW, okTB)

	if !okTB {
		t.Error("Token bucket should be partially refilled after 2s")
	}

	t.Logf("Fixed window result depends on window timing: %v", okFW)
}
