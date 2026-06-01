package algorithms

// Algorithm constants — must be plain strings (no custom type)
const (
	AlgorithmFixedWindow   = "fixed_window"
	AlgorithmSlidingWindow = "sliding_window"
	AlgorithmTokenBucket   = "token_bucket"
)

// Config used for passing limiter configuration
type Config struct {
	Algorithm  string
	Limit      int
	WindowSecs int
}
