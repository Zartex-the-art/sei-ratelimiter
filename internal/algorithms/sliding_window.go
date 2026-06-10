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
	count, err := sw.store.SlidingWindowAllow(ctx, key, nowMs, windowMs, member,
		sw.limit)
	if err != nil {
		return false, 0, fmt.Errorf("SlidingWindow.Allow: %w", err)
	}
	if int(count) > sw.limit {
		return false, 0, nil
	}
	remaining := sw.limit - int(count)
	return true, remaining, nil
}
