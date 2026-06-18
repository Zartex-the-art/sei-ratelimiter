package store_test

import (
	"context"
	"testing"

	"github.com/Zartex-the-art/sei-ratelimiter/internal/store"
)

func TestFakeStore_IncrementAndReset(t *testing.T) {
	fs := store.NewFakeStore()

	count, err := fs.Increment(context.Background(), "user1", 60)
	if err != nil {
		t.Fatalf("Increment error: %v", err)
	}

	if count != 1 {
		t.Fatalf("got %d, want 1", count)
	}

	fs.Reset("user1")

	count, err = fs.Increment(context.Background(), "user1", 60)
	if err != nil {
		t.Fatalf("Increment after reset error: %v", err)
	}

	if count != 1 {
		t.Fatalf("got %d, want 1 after reset", count)
	}
}

func TestFakeStore_Del(t *testing.T) {
	fs := store.NewFakeStore()

	fs.Increment(context.Background(), "a", 60)

	if err := fs.Del(context.Background(), "a"); err != nil {
		t.Fatalf("Del error: %v", err)
	}

	count, _ := fs.Increment(context.Background(), "a", 60)

	if count != 1 {
		t.Fatalf("expected deleted key to restart at 1, got %d", count)
	}
}

func TestFakeStore_SlidingWindowAllow(t *testing.T) {
	fs := store.NewFakeStore()

	for i := 0; i < 3; i++ {
		count, err := fs.SlidingWindowAllow(
			context.Background(),
			"sw",
			0,
			1000,
			"member",
			3,
		)

		if err != nil {
			t.Fatalf("error: %v", err)
		}

		if count != int64(i+1) {
			t.Fatalf("got %d, want %d", count, i+1)
		}
	}

	count, err := fs.SlidingWindowAllow(
		context.Background(),
		"sw",
		0,
		1000,
		"member",
		3,
	)

	if err != nil {
		t.Fatalf("error: %v", err)
	}

	if count != 4 {
		t.Fatalf("got %d, want 4", count)
	}
}

func TestFakeStore_TokenBucketAllow(t *testing.T) {
	fs := store.NewFakeStore()

	remaining, err := fs.TokenBucketAllow(
		context.Background(),
		"tb",
		0,
		3,
		60,
	)

	if err != nil {
		t.Fatalf("error: %v", err)
	}

	if remaining != 2 {
		t.Fatalf("got %d, want 2", remaining)
	}
}

func TestErrorStore_ReturnsErrors(t *testing.T) {
	es := store.NewErrorStore()

	if _, err := es.Increment(context.Background(), "x", 60); err == nil {
		t.Fatal("expected error")
	}

	if err := es.ZAdd(context.Background(), "x", 1, "m"); err == nil {
		t.Fatal("expected error")
	}

	if err := es.Del(context.Background(), "x"); err == nil {
		t.Fatal("expected error")
	}

	if _, err := es.HGetAll(context.Background(), "x"); err == nil {
		t.Fatal("expected error")
	}
}
func TestErrorStore_AllMethodsReturnError(t *testing.T) {
	es := store.NewErrorStore()

	if err := es.ZRemRangeByScore(context.Background(), "k", 0, 1); err == nil {
		t.Fatal("expected error")
	}

	if _, err := es.ZCount(context.Background(), "k", 0, 1); err == nil {
		t.Fatal("expected error")
	}

	if err := es.HSet(context.Background(), "k", map[string]interface{}{"a": 1}); err == nil {
		t.Fatal("expected error")
	}

	if _, err := es.SlidingWindowAllow(
		context.Background(),
		"k",
		1,
		1000,
		"m",
		5,
	); err == nil {
		t.Fatal("expected error")
	}

	if _, err := es.TokenBucketAllow(
		context.Background(),
		"k",
		1,
		5,
		60,
	); err == nil {
		t.Fatal("expected error")
	}
}
