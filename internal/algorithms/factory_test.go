package algorithms_test

import (
	"context"
	"testing"

	"github.com/Zartex-the-art/sei-ratelimiter/internal/algorithms"
	"github.com/Zartex-the-art/sei-ratelimiter/internal/store"
)

// ✅ First test
func TestNewLimiter_FixedWindow_ReturnsFW(t *testing.T) {
	l, err := algorithms.NewLimiter(algorithms.AlgorithmFixedWindow, store.NewFakeStore(), 10, 60)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l == nil {
		t.Fatal("expected non-nil Limiter")
	}

	ok, remaining, err := l.Allow(context.Background(), "test")
	if err != nil {
		t.Fatalf("Allow error: %v", err)
	}
	if !ok {
		t.Error("first request should be allowed")
	}
	if remaining != 9 {
		t.Errorf("remaining: got %d, want 9", remaining)
	}
}

// ✅ Unknown algorithm
func TestNewLimiter_UnknownAlgorithm_ReturnsError(t *testing.T) {
	unknowns := []string{"unknown", "", "fixedwindow"}

	for _, algo := range unknowns {
		t.Run(algo, func(t *testing.T) {
			l, err := algorithms.NewLimiter(algo, store.NewFakeStore(), 10, 60)
			if err == nil {
				t.Errorf("expected error for %q, got nil", algo)
			}
			if l != nil {
				t.Errorf("expected nil limiter for %q", algo)
			}
		})
	}
}

// ✅ Invalid limit
func TestNewLimiter_InvalidLimit_ReturnsError(t *testing.T) {
	for _, limit := range []int{0, -1, -100} {
		_, err := algorithms.NewLimiter(algorithms.AlgorithmFixedWindow, store.NewFakeStore(), limit, 60)
		if err == nil {
			t.Errorf("limit=%d: expected error, got nil", limit)
		}
	}
}

// ✅ Invalid window
func TestNewLimiter_InvalidWindowSecs_ReturnsError(t *testing.T) {
	for _, ws := range []int{0, -1, -60} {
		_, err := algorithms.NewLimiter(algorithms.AlgorithmFixedWindow, store.NewFakeStore(), 10, ws)
		if err == nil {
			t.Errorf("windowSecs=%d: expected error, got nil", ws)
		}
	}
}

// ✅ Valid algorithms
func TestNewLimiter_ValidAlgorithms_ReturnsAll3(t *testing.T) {
	valid := algorithms.ValidAlgorithms()

	if len(valid) != 3 {
		t.Errorf("expected 3 valid algorithms, got %d: %v", len(valid), valid)
	}

	expected := map[string]bool{
		"fixed_window":   true,
		"sliding_window": true,
		"token_bucket":   true,
	}

	for _, algo := range valid {
		if !expected[algo] {
			t.Errorf("unexpected algorithm: %q", algo)
		}
	}
}
