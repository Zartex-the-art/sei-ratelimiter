package integration_test

import (
	"context"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/Zartex-the-art/sei-ratelimiter/internal/algorithms"
	"github.com/Zartex-the-art/sei-ratelimiter/internal/store"
	"github.com/Zartex-the-art/sei-ratelimiter/internal/testhelpers"
)

func redisAddr() string {
	if addr := os.Getenv("REDIS_URL"); addr != "" {
		return addr
	}
	return "localhost:6379"
}

// TestTwoNodes_ShareRedisCounter verifies that two FixedWindow instances
// connected to the same Redis share the same counter.
func TestTwoNodes_ShareRedisCounter(t *testing.T) {
	client := testhelpers.RedisClient(t)

	t.Cleanup(func() {
		testhelpers.FlushKeys(t, client, "fw:shared-client")
	})

	store1 := store.NewRedisStore(redisAddr())
	store2 := store.NewRedisStore(redisAddr())

	node1 := algorithms.NewFixedWindow(10, 60, store1)
	node2 := algorithms.NewFixedWindow(10, 60, store2)

	var wg sync.WaitGroup
	var allowed int64

	for i := 0; i < 10; i++ {
		wg.Add(2)

		go func() {
			defer wg.Done()

			ok, _, _ := node1.Allow(context.Background(), "shared-client")
			if ok {
				atomic.AddInt64(&allowed, 1)
			}
		}()

		go func() {
			defer wg.Done()

			ok, _, _ := node2.Allow(context.Background(), "shared-client")
			if ok {
				atomic.AddInt64(&allowed, 1)
			}
		}()
	}

	wg.Wait()

	if int(allowed) > 10 {
		t.Errorf("two-node total allowed: got %d, must not exceed 10", allowed)
	}

	t.Logf("two-node: %d/10 allowed from 20 concurrent requests", allowed)
}

// TestTwoNodes_LimitEnforcedGlobally verifies that exhausting the limit
// on node1 also blocks requests on node2 for the same client.
func TestTwoNodes_LimitEnforcedGlobally(t *testing.T) {
	client := testhelpers.RedisClient(t)

	t.Cleanup(func() {
		testhelpers.FlushKeys(t, client, "fw:global-limit")
	})

	const limit = 5

	node1 := algorithms.NewFixedWindow(limit, 60, store.NewRedisStore(redisAddr()))
	node2 := algorithms.NewFixedWindow(limit, 60, store.NewRedisStore(redisAddr()))

	for i := 0; i < limit; i++ {
		ok, _, _ := node1.Allow(context.Background(), "global-limit")
		if !ok {
			t.Fatalf("req %d on node1 should be allowed", i+1)
		}
	}

	ok, _, _ := node2.Allow(context.Background(), "global-limit")
	if ok {
		t.Error("node2 should be blocked: limit already reached on node1")
	}
}

// TestTwoNodes_IndependentClients verifies different clients have
// independent counters even across two nodes.
func TestTwoNodes_IndependentClients(t *testing.T) {
	client := testhelpers.RedisClient(t)

	t.Cleanup(func() {
		testhelpers.FlushKeys(t, client, "fw:node-client-a")
		testhelpers.FlushKeys(t, client, "fw:node-client-b")
	})

	node1 := algorithms.NewFixedWindow(3, 60, store.NewRedisStore(redisAddr()))
	node2 := algorithms.NewFixedWindow(3, 60, store.NewRedisStore(redisAddr()))

	for i := 0; i < 3; i++ {
		node1.Allow(context.Background(), "node-client-a")
	}

	okA, _, _ := node1.Allow(context.Background(), "node-client-a")
	if okA {
		t.Error("client-a should be blocked on node1")
	}

	okB, _, _ := node2.Allow(context.Background(), "node-client-b")
	if !okB {
		t.Error("client-b should be allowed on node2")
	}
}
