package algorithms

import (
	"context"
	"fmt"
	"math"
	"strconv"
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
	// Step 1: Read current bucket state
	fields, err := tb.store.HGetAll(ctx, key)
	if err != nil {
		return false, 0, fmt.Errorf("TokenBucket.Allow HGetAll: %w", err)
	}
	var tokens float64
	var lastRefill int64
	// Step 2: Initialise or compute refill
	if len(fields) == 0 {
		// First request — bucket starts full
		tokens = float64(tb.limit)
		lastRefill = nowMs
	} else {
		// Parse existing state
		tokens, err = strconv.ParseFloat(fields["tokens"], 64)
		if err != nil {
			return false, 0, fmt.Errorf("TokenBucket.Allow parse tokens: %w", err)
		}
		lastRefill, err = strconv.ParseInt(fields["last_refill"], 10, 64)
		if err != nil {
			return false, 0, fmt.Errorf("TokenBucket.Allow parse last_refill: %w", err)
		}
		// Compute how many tokens to add based on elapsed time
		elapsedSecs := float64(nowMs-lastRefill) / 1000.0
		tokensToAdd := elapsedSecs * tb.rate
		tokens = math.Min(tokens+tokensToAdd, float64(tb.limit))
	}
	// Step 3: Decision
	if tokens < 1.0 {
		// Bucket empty — blocked. Do NOT update state on blocked requests.
		return false, 0, nil
	}
	// Step 4: Consume one token and persist
	tokens -= 1.0
	err = tb.store.HSet(ctx, key, map[string]interface{}{
		"tokens":      strconv.FormatFloat(tokens, 'f', 6, 64),
		"last_refill": strconv.FormatInt(nowMs, 10),
	})
	if err != nil {
		return false, 0, fmt.Errorf("TokenBucket.Allow HSet: %w", err)
	}
	return true, int(tokens), nil
}
