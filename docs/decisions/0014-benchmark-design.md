# ADR-014: Load Test Benchmark Design

## Status
Accepted

## Context
Phase 5 requires a benchmark table with p50/p95/p99 at 1K, 5K, 10K RPS.
Decisions are needed on: which algorithm, which k6 executor, what thresholds.

## Decisions

### Algorithm: Fixed Window
Fixed window uses one EVALSHA per request (O(1)).
Sliding window uses ZADD + ZREMRANGEBYSCORE + ZCOUNT (O(log N)).
Token bucket uses HGETALL + HSET (O(1) but with hash parsing overhead).
Fixed window isolates the HTTP+Redis path for pure throughput measurement.

### k6 Executor: constant-arrival-rate
Sends exactly N requests/second regardless of server response time.
Exposes backpressure naturally.
More representative of production load than constant-vus.

### Threshold: p99 < 5ms
Industry standard for rate limiting middleware.
At 5ms p99, the rate limiter adds minimal latency to the request path.
Above 10ms p99, the rate limiter becomes a noticeable bottleneck.

## Consequences
Fixed window benchmarks understate latency for sliding window workloads.
A separate sliding window benchmark may be added after Phase 5.