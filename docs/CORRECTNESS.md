# Correctness Contract --- sei-ratelimiter

## Exact Contract

- **N allowed, not N+1**: A client with limit=N is allowed exactly N requests per window. The (N+1)th request is blocked.
- **remaining = max(0, limit - count)**: When count == limit, remaining is 0. When count > limit, remaining stays 0.
- **Error vs Blocked**: A blocked request returns `(false, 0, nil)`. An infrastructure failure returns `(false, 0, error)` with HTTP 503.

## Concurrency Requirement

- All algorithm implementations must be safe for concurrent use by multiple goroutines.
- The race detector (`go test -race`) must pass on every test run.
- No data races on shared state (maps, counters, Redis connections).

## Multi-Node Requirement

- With 2 app nodes + 1 Redis, a client must not exceed the limit regardless of which node serves the request.
- Redis operations must be atomic (Lua scripts in Phase 4).
- No over-counting when two nodes serve the same client simultaneously.

## Failure Modes That Break Correctness

| Failure | Impact | Detection |
|---------|--------|-----------|
| Redis unavailable | All requests return 503 (not blocked) | Integration test |
| Clock skew between nodes | Window boundaries misaligned | Not yet tested |
| Lua script not atomic | Over-counting under concurrent load | Phase 4 test |
| Memory leak in Redis | Keys accumulate without expiry | Monitoring |

## Verification

- Unit tests: `go test -race -count=3 ./internal/algorithms/...`
- Integration tests: `go test -race -v ./internal/store/...` with Redis running
- Load tests: `k6 run tests/load/smoke.js` against running stack
