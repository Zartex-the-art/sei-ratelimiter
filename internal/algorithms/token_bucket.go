package algorithms

import (
	"context"
	"fmt"
	"time"

	"github.com/Zartex-the-art/sei-ratelimiter/internal/store"
)

// TokenBucket implements the token bucket rate limiting algorithm.
//
// A bucket holds up to 'limit' tokens. Each request consumes one token.
// Tokens refill at rate = limit/windowSecs tokens per second.
// Allows burst traffic up to the bucket capacity.
//
// Not atomic — Lua script replaces this in Phase 4.
type TokenBucket struct {
	limit      int
	windowSecs int
	rate       float64 // tokens per second
	store      store.Store
}

// NewTokenBucket creates a TokenBucket backed by the provided Store.
// rate = limit / windowSecs tokens per second.
func NewTokenBucket(limit, windowSecs int, s store.Store) *TokenBucket {
	return &TokenBucket{
		limit:      limit,
		windowSecs: windowSecs,
		rate:       float64(limit) / float64(windowSecs),
		store:      s,
	}
}

// Allow checks if clientID is allowed to make a request.
//
// Redis key: tb:{clientID}
// Hash fields: tokens (float64), last_refill (int64 unix ms)
func (tb *TokenBucket) Allow(ctx context.Context, clientID string) (bool, int, error) {
	key := fmt.Sprintf("tb:%s", clientID)
	nowMs := time.Now().UnixMilli()
	remaining, err := tb.store.TokenBucketAllow(ctx, key, nowMs, tb.limit,
		tb.windowSecs)
	if err != nil {
		return false, 0, fmt.Errorf("TokenBucket.Allow: %w", err)
	}
	if remaining == -1 {
		return false, 0, nil // blocked
	}
	return true, remaining, nil
}
