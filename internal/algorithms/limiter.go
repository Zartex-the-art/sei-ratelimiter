package algorithms

import "context"

// Limiter interface for all rate limiting algorithms.
//
// All algorithms (FixedWindow, SlidingWindow, TokenBucket)
// must implement this interface so they can be used
// interchangeably via the factory.
type Limiter interface {
	// Allow checks if a request for the given clientID
	// is allowed under the rate limit.
	//
	// Returns:
	// - allowed (bool): whether request is permitted
	// - remaining (int): remaining requests/tokens
	// - error: if any issue occurs (e.g., Redis failure)
	Allow(ctx context.Context, clientID string) (bool, int, error)
}
