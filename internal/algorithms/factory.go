package algorithms

import (
	"fmt"

	"github.com/Zartex-the-art/sei-ratelimiter/internal/store"
)

// Algorithm string constants — DO NOT use custom types


// NewLimiter creates the correct Limiter based on algorithm string
func NewLimiter(algorithm string, s store.Store, limit, windowSecs int) (Limiter, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be > 0, got %d", limit)
	}

	if windowSecs <= 0 {
		return nil, fmt.Errorf("windowSecs must be > 0, got %d", windowSecs)
	}

	switch algorithm {
	case AlgorithmFixedWindow:
		return NewFixedWindow(limit, windowSecs, s), nil
	case AlgorithmSlidingWindow:
		return NewSlidingWindow(limit, windowSecs, s), nil
	case AlgorithmTokenBucket:
		return NewTokenBucket(limit, windowSecs, s), nil
	default:
		return nil, fmt.Errorf("unknown algorithm %q", algorithm)
	}
}

// ValidAlgorithms returns all supported algorithms
func ValidAlgorithms() []string {
	return []string{
		AlgorithmFixedWindow,
		AlgorithmSlidingWindow,
		AlgorithmTokenBucket,
	}
}
