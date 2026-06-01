package limiter

import (
	"sync"
	"time"
)

type FixedWindowLimiter struct {
	mu          sync.Mutex
	limit       int
	window      time.Duration
	counts      map[string]int
	windowStart time.Time
}

func NewFixedWindowLimiter(limit int, window time.Duration) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		limit:       limit,
		window:      window,
		counts:      make(map[string]int),
		windowStart: time.Now(),
	}
}

func (f *FixedWindowLimiter) Allow(key string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	now := time.Now()

	if now.Sub(f.windowStart) >= f.window {
		f.counts = make(map[string]int)
		f.windowStart = now
	}

	if f.counts[key] >= f.limit {
		return false
	}

	f.counts[key]++
	return true
}
