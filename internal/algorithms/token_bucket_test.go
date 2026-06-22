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

func newTB(t *testing.T, limit, windowSecs int) (*algorithms.TokenBucket, *testhelpers.Cleaner) {
	t.Helper()
	client := testhelpers.RedisClient(t)
	rs := store.NewRedisStore("localhost:6379")
	tb := algorithms.NewTokenBucket(limit, windowSecs, rs)
	return tb, testhelpers.NewCleaner(t, client)
}

// TestTokenBucket_StartsWithFullBucket verifies that the first request
// sees a full bucket and gets tokens = limit - 1 remaining.
func TestTokenBucket_StartsWithFullBucket(t *testing.T) {
	tb, c := newTB(t, 5, 60)
	defer c.Del("tb:tb-full")
	ok, remaining, err := tb.Allow(context.Background(), "tb-full")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !ok {
		t.Fatal("first request should be allowed")
	}
	if remaining != 4 {
		t.Errorf("remaining: got %d, want 4 (limit-1)", remaining)
	}
}

// TestTokenBucket_ExhaustsBucket verifies that after limit requests,
// the next request is blocked.
func TestTokenBucket_ExhaustsBucket(t *testing.T) {
	tb, c := newTB(t, 5, 60)
	defer c.Del("tb:tb-exhaust")
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		ok, _, _ := tb.Allow(ctx, "tb-exhaust")
		if !ok {
			t.Fatalf("req %d should be allowed", i+1)
		}
	}

	ok, remaining, err := tb.Allow(ctx, "tb-exhaust")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if ok {
		t.Error("6th request should be blocked — bucket empty")
	}
	if remaining != 0 {
		t.Errorf("remaining: got %d, want 0", remaining)
	}
}

// TestTokenBucket_RemainingDecrement verifies remaining decrements on each allow.
func TestTokenBucket_RemainingDecrement(t *testing.T) {
	tb, c := newTB(t, 5, 60)
	defer c.Del("tb:tb-decrement")
	ctx := context.Background()
	wants := []int{4, 3, 2, 1, 0}
	for i, want := range wants {
		_, got, _ := tb.Allow(ctx, "tb-decrement")
		if got != want {
			t.Errorf("req %d: remaining got %d, want %d", i+1, got, want)
		}
	}
}

// TestTokenBucket_Refills verifies that tokens refill after waiting.
// rate = 5/5 = 1 token per second.
func TestTokenBucket_Refills(t *testing.T) {
	tb, c := newTB(t, 5, 5)
	defer c.Del("tb:tb-refill")
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		tb.Allow(ctx, "tb-refill")
	}
	ok, _, _ := tb.Allow(ctx, "tb-refill")
	if ok {
		t.Fatal("should be blocked after exhausting bucket")
	}
	// Wait 3 seconds — 3 tokens should refill
	time.Sleep(3100 * time.Millisecond)
	// 3 requests should now be allowed
	for i := 0; i < 3; i++ {
		ok, _, err := tb.Allow(ctx, "tb-refill")
		if err != nil {
			t.Fatalf("refill req %d: error: %v", i+1, err)
		}
		if !ok {
			t.Fatalf("refill req %d: should be allowed after refill", i+1)
		}
	}
	// 4th request should be blocked — only 3 tokens refilled
	ok, _, _ = tb.Allow(ctx, "tb-refill")
	if ok {
		t.Error("4th request should be blocked — only 3 tokens refilled in 3s")
	}
}

// TestTokenBucket_BurstBehaviour verifies that all limit requests
// succeed immediately when bucket is full. This is the key feature
// of token bucket vs fixed/sliding window.
func TestTokenBucket_BurstBehaviour(t *testing.T) {
	tb, c := newTB(t, 10, 60)
	defer c.Del("tb:tb-burst")
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		ok, _, err := tb.Allow(ctx, "tb-burst")
		if err != nil {
			t.Fatalf("burst req %d: error: %v", i+1, err)
		}
		if !ok {
			t.Fatalf("burst req %d: should be allowed — bucket was full",
				i+1)
		}
	}
	// 11th is blocked — bucket empty
	ok, _, _ := tb.Allow(ctx, "tb-burst")
	if ok {
		t.Error("11th request should be blocked — bucket exhausted")
	}
}

// TestTokenBucket_MultipleClientsAreIndependent verifies each client
// has their own independent token bucket.
func TestTokenBucket_MultipleClientsAreIndependent(t *testing.T) {
	tb, c := newTB(t, 3, 60)
	defer func() {
		c.Del("tb:tb-multi-a")
		c.Del("tb:tb-multi-b")
	}()
	ctx := context.Background()
	// Exhaust client-a
	for i := 0; i < 3; i++ {
		tb.Allow(ctx, "tb-multi-a")
	}
	okA, _, _ := tb.Allow(ctx, "tb-multi-a")
	if okA {
		t.Error("client-a should be blocked")
	}
	// client-b unaffected
	okB, remaining, _ := tb.Allow(ctx, "tb-multi-b")
	if !okB {
		t.Error("client-b should be allowed")
	}

	if remaining != 2 {
		t.Errorf("client-b remaining: got %d, want 2", remaining)
	}
}

// TestTokenBucket_ConcurrentRequests verifies that under concurrent load,
// total allowed does not exceed the limit.
func TestTokenBucket_ConcurrentRequests(t *testing.T) {
	tb, c := newTB(t, 50, 60)
	defer c.Del("tb:tb-concurrent")
	ctx := context.Background()
	var wg sync.WaitGroup
	var allowed int64
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ok, _, _ := tb.Allow(ctx, "tb-concurrent")
			if ok {
				atomic.AddInt64(&allowed, 1)
			}
		}()
	}
	wg.Wait()
	if int(allowed) > 50 {
		t.Errorf("concurrent: allowed %d requests, must not exceed 50", allowed)
	}
	t.Logf("concurrent: %d/50 allowed from 200 goroutines", allowed)
}
func TestTokenBucket_AtomicUnderConcurrency(t *testing.T) {
	client := testhelpers.RedisClient(t)
	defer testhelpers.FlushKeys(t, client, "tb:tb-atomic")
	const limit = 50
	const goroutines = 300
	rs := store.NewRedisStore("localhost:6379")
	tb := algorithms.NewTokenBucket(limit, 60, rs)
	var wg sync.WaitGroup
	var allowed int64
	var blocked int64
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ok, _, err := tb.Allow(context.Background(), "tb-atomic")
			if err != nil {
				t.Errorf("error: %v", err)
				return
			}
			if ok {
				atomic.AddInt64(&allowed, 1)
			} else {
				atomic.AddInt64(&blocked, 1)
			}
		}()
	}
	wg.Wait()
	// Token bucket starts with 'limit' tokens.
	// Under concurrent load, total allowed must never exceed limit.
	if int(allowed) > limit {
		t.Errorf("OVER-COUNTING: allowed %d, limit %d — Lua script not atomic", allowed,
			limit)
	}
	// Exactly limit should be allowed (full bucket at start)
	if int(allowed) != limit {
		t.Errorf("expected exactly %d allowed (full bucket), got %d", limit, allowed)
	}
	t.Logf("Token bucket atomic: %d/%d allowed, %d blocked", allowed, limit, blocked)
}
