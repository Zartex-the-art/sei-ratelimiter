# Distributed Correctness — How We Guarantee It

## The Problem
A rate limiter that runs on a single machine is simple:
one process, one counter, one source of truth.
A rate limiter across multiple nodes is hard:- Node 1 reads counter: 4 (at limit=5)- Node 2 reads counter: 4 (same value!)- Node 1 increments: 5 — allowed- Node 2 increments: 5 — allowed- Result: 6 requests served, limit was 5 — WRONG

## Why Redis Alone Is Not Enough
INCR is atomic. Each INCR returns a unique value.
But INCR + EXPIRE together are NOT atomic.
If a node crashes between them, the key never expires.
If two nodes both call INCR with count=1, both call EXPIRE
and the second EXPIRE resets the window — extending it incorrectly.

## The Solution: Lua Scripts
Redis executes Lua scripts single-threaded and atomically.
No other Redis command can run between any two lines of a Lua script.
Fixed window script: INCR + conditional EXPIRE = one atomic unit
Sliding window script: ZADD + ZREMRANGEBYSCORE + ZCOUNT + optional ZREM
Token bucket script: HGETALL + compute + conditional HSET

## Verification
Three levels of verification:
Level 1: go test -race (catches in-process data races)
Level 2: TestXxx_AtomicUnderConcurrency (300 goroutines, limit=50)
Level 3: k6 correctness test (2 nodes, 10 VUs, 60 seconds)
All three must pass for each algorithm before Phase 4 is closed.