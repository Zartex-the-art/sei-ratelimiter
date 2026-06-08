package algorithms

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Zartex-the-art/sei-ratelimiter/internal/store"
)

// SlidingWindow implements the sliding window log rate limiting algorithm.
//
// Each request is recorded in a Redis sorted set with the current
// timestamp (milliseconds) as the score. On each Allow() call, entries
// outside the window are pruned before counting.
//
// Eliminates the boundary burst problem present in FixedWindow.
// Memory: O(requests in current window) per client.
// Not atomic — Lua script replaces this in Phase 4.
type SlidingWindow struct {
	limit      int
	windowSecs int
	store      store.Store
}

// NewSlidingWindow creates a SlidingWindow backed by the provided Store.
func NewSlidingWindow(limit, windowSecs int, s store.Store) *SlidingWindow {
	return &SlidingWindow{
		limit:      limit,
		windowSecs: windowSecs,
		store:      s,
	}
}

// Allow checks if clientID is within the sliding window rate limit.
//
// Redis key: sw:{clientID}
// Score: unix timestamp in milliseconds
// Member: nanosecond timestamp string (unique per request)
func (sw *SlidingWindow) Allow(ctx context.Context, clientID string) (bool, int, error) {
	key := fmt.Sprintf("sw:%s", clientID)
	nowMs := time.Now().UnixMilli()
	windowMs := int64(sw.windowSecs) * 1000
	member := strconv.FormatInt(time.Now().UnixNano(), 10)
	// Step 1: Record this request
	if err := sw.store.ZAdd(ctx, key, float64(nowMs), member); err != nil {
	return false, 0, fmt.Errorf("sliding window zadd: %w", err)
}
	// Step 2: Prune entries outside the window
	pruneMax := float64(nowMs - windowMs)
	if err := sw.store.ZRemRangeByScore(ctx, key, 0, pruneMax); err != nil {
	return false, 0, fmt.Errorf("sliding window zremrangebyscore: %w", err)
}
	// Step 3: Count requests currently in the window
	// Use nowMs+1000 as upper bound to include all entries up to 1s in the future
	// (handles any clock skew between the ZADD and ZCOUNT calls)
	count, err := sw.store.ZCount(ctx, key, 0, float64(nowMs+1000))
if err != nil {
	return false, 0, fmt.Errorf("sliding window zcount: %w", err)
}
	// Step 4: Decision
	if int(count) > sw.limit {
		// Remove the entry we just added — clean state on blocked requests
		_ = sw.store.ZRemRangeByScore(ctx, key, float64(nowMs), float64(nowMs))
		return false, 0, nil
	}
	remaining := sw.limit - int(count)
	return true, remaining, nil
}
