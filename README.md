# sei-ratelimiter

Distributed Rate Limiter as a Service — Zartex SEI Project 1

---

## Architecture

![Fixed Window Sequence](docs/diagrams/fixed-window-sequence.png)

---

## Algorithms

- Fixed Window
- Sliding Window
- Token Bucket

---

## API Reference

### POST /check

Checks whether request is allowed.

### POST /rules

Create a new rule.

### GET /rules

Get all rules.

### GET /rules/:id

Get rule by ID.

### DELETE /rules/:id

Delete rule.

---

## How To Run

Coming soon.

---

## How To Run Tests

Coming soon.

---

## Benchmarks

Coming soon.

---

## Failure Modes

Coming soon.

---

## What We Would Do at 10x Scale

Coming soon.


## How To Run

### Prerequisites

- Docker Desktop
- WSL2 enabled
- Git

### Start the full stack

```bash
git clone git@github.com:Zartex-the-art/sei-ratelimiter.git
cd sei-ratelimiter
docker compose up --build
```

This starts:

- App Node 1 → http://localhost:8080
- App Node 2 → http://localhost:8081
- Redis → localhost:6379

### Verify

```bash
curl http://localhost:8080/health
curl http://localhost:8081/health
```

Both should return:

```json
{"status":"ok"}
```

### Stop

```bash
docker compose down
```

## Project Structure

```text
sei-ratelimiter/
├── cmd/
│   └── server/                  # Entry point — wires packages, starts server
│       └── main.go
│
├── internal/
│   ├── algorithms/              # Rate limiting logic
│   │   ├── limiter.go           # Limiter interface
│   │   ├── fixed_window.go      # Fixed Window algorithm
│   │   ├── fixed_window_test.go # Fixed Window tests
│   │   ├── sliding_window.go    # Sliding Window (Phase 2)
│   │   └── token_bucket.go      # Token Bucket (Phase 2)
│   │
│   ├── api/                     # HTTP handlers
│   ├── config/                  # Environment variable loading
│   ├── store/                   # Redis client/store layer
│   └── testhelpers/             # Shared test utilities
│
├── tests/
│   └── load/                    # k6 load test scripts
│
├── docs/
│   ├── decisions/               # Architecture Decision Records
│   │   ├── 0000-template.md
│   │   ├── 0001-go-language-choice.md
│   │   ├── 0002-concurrency-model.md
│   │   ├── 0002-infrastructure-tooling.md
│   │   ├── 0003-package-structure.md
│   │   └── 0004-fixed-window-first.md
│   │
│   ├── diagrams/                # Architecture diagrams
│   │   ├── architecture-v1.png
│   │   ├── architecture-v2.png
│   │   ├── architecture-v3.png
│   │   ├── architecture-v4.png
│   │   └── fixed-window-sequence.png
│   │
│   ├── ARCHITECTURE.md
│   ├── CONCURRENCY.md
│   ├── DOCKER_CONCEPTS.md
│   ├── REDIS_KEY_DESIGN.md
│   ├── SHARED_STATE.md
│   ├── TEST_CONVENTIONS.md
│   └── TEST_STRATEGY.md
│
├── .github/
│   └── workflows/               # GitHub Actions CI
│
├── pkg/
├── practice/
│   ├── goroutine.go
│   ├── mutex.go
│   └── race.go
│
├── .env.example
├── .gitignore
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── README.md
├── server_test.go
└── SPRINT_LOG.md
```


## Fixed Window Algorithm

The Fixed Window algorithm limits requests within a fixed time duration.

Example:
- Limit: 5 requests
- Window: 60 seconds

Flow:
1. Client sends request
2. Redis counter increments using INCR
3. Redis EXPIRE sets TTL on first request
4. If count <= limit:
   request allowed
5. Else:
   request blocked

Redis Commands:
- INCR
- EXPIRE

Advantages:
- Simple
- Fast
- Low memory usage

Tradeoff:
Fixed Window suffers from the Boundary Burst problem.
A client can send requests at the end of one window and beginning of another window, effectively doubling request rate.

Best Use Cases:
- Simple APIs
- Low complexity systems
- Basic rate limiting

### Fixed Window — Redis Implementation (Day 7)

What changed from in-memory (Day 6):

- Counter stored in Redis — shared by all app nodes
- Both nodes read from and write to the same key: fw:{clientID}
- State survives app restarts (Redis persists via AOF)
- Global limit now enforced correctly across the cluster

Redis commands:

- INCR fw:{clientID} increments counter, returns new value
- EXPIRE fw:{clientID} {windowSecs} sets TTL (batched in pipeline)

Pipeline note:

- INCR and EXPIRE are sent together in one round-trip (pipeline).
- This is not atomic — see boundary burst limitation.
- Lua script replaces this in Phase 4 for true atomicity.

Design — Dependency Injection:

- FixedWindow does not create a Redis client directly.
- It accepts a Store interface — injected at startup.
- Unit tests inject FakeStore (no Redis needed).
- Production injects RedisStore.

This pattern is used for all three algorithms.


### Sliding Window Log

Eliminates the boundary burst problem by tracking exact request timestamps.

Every request in the last N seconds is counted precisely.

How it works:

1. ZADD sw:{clientID} {timestamp_ms} {unique_member} — record request
2. ZREMRANGEBYSCORE sw:{clientID} 0 {now-window_ms} — prune old entries
3. ZCOUNT sw:{clientID} 0 +inf — count in window
4. count <= limit: allow. count > limit: ZREM + block.

Redis key: sw:{clientID}

Score: unix timestamp in milliseconds (float64)

Member: nanosecond timestamp string (unique per request)

Why no boundary burst:

At t=61s with a 60s window, requests from t=1s to t=61s are all counted.

There is no clock reset moment. The window slides continuously.

A client cannot exploit a boundary — every second is treated equally.

Memory tradeoff:

Each request in the current window occupies one sorted set entry.

High-traffic clients: 1000 req/min x 60s window = 1000 entries.

Fixed window: always 1 entry per client regardless of traffic.

When to use:

Billing APIs, security-critical endpoints, any place where a 2x burst would cause real harm or incorrect billing.

Known limitation:

ZADD + ZREMRANGEBYSCORE + ZCOUNT are 3 separate commands.

Not atomic — Phase 4 Lua scripts fix this.


## Algorithm Comparison

| Property | Fixed Window | Sliding Window | Token Bucket |
|----------|---------------|----------------|---------------|
| Redis data type | String | Sorted Set | Hash |
| Memory per client | O(1) | O(requests in window) | O(1) |
| Boundary burst | Yes (bug) | No | No |
| Burst allowance | No | No | Yes |
| Complexity | Low | Medium | Medium |
| Best for | Simple APIs | Precision-critical | Bursty clients |
| Implemented | Day 6-7 | Day 8 | Day 9 |


| Property | Fixed Window | Sliding Window |
|---|---|---|
| Redis data type | String (INCR + EXPIRE) | Sorted Set |
| Memory per client | O(1) | O(requests in window) |
| Boundary burst | YES | NO |
| Window reset | Fixed clock tick | Continuous |
| Implementation | Simple | Medium |
| Best for | Simple APIs | Billing/Security |



| Property          | Fixed Window | Sliding Window        | Token Bucket         |
| ----------------- | ------------ | --------------------- | -------------------- |
| Redis data type   | String       | Sorted Set            | Hash                 |
| Memory per client | O(1)         | O(requests in window) | O(1)                 |
| Boundary burst    | Yes          | No                    | No                   |
| Burst allowance   | No           | No                    | Yes — up to limit    |
| Refill model      | Window reset | Continuous pruning    | Continuous token add |
| Implementation    | Simple       | Medium                | Medium               |
| Best for          | Simple APIs  | Precision-critical    | Bursty API clients   |


| Property          | Fixed Window    | Sliding Window        | Token Bucket         |
| ----------------- | --------------- | --------------------- | -------------------- |
| Redis data type   | String          | Sorted Set            | Hash                 |
| Memory per client | O(1)            | O(requests in window) | O(1)                 |
| Boundary burst    | Yes             | No                    | No                   |
| Burst allowance   | No              | No                    | Yes                  |
| Refill model      | Window reset    | Continuous pruning    | Continuous token add |
| Implementation    | Simple          | Medium                | Medium               |
| Best for          | Simple APIs     | Precision-critical    | Bursty API clients   |
| Phase 4 atomicity | Lua INCR+EXPIRE | Lua ZADD+ZREM+ZCOUNT  | Lua HGETALL+HSET     |
| Status            | Done            | Done                  | Done                 |


Why Token Bucket?
Fixed Window is simple and memory-efficient but allows boundary burst problems.
Sliding Window improves precision by tracking request timestamps, but memory usage grows with traffic.
Token Bucket allows controlled bursts while maintaining an average request rate over time.
This makes Token Bucket useful for real-world APIs where short bursts are acceptable but long-term abuse must still be prevented.

In this project:
- Fixed Window demonstrates simple rate limiting
- Sliding Window demonstrates precise fairness
- Token Bucket demonstrates burst-tolerant rate limiting
These implementations together show tradeoffs between simplicity, precision, memory usage, and client experience.



## Architecture

![Phase 2 Architecture](docs/diagrams/architecture-phase2-final.png)
Two app nodes share one Redis instance.
All rate limit state is in Redis — nodes are stateless.
Algorithm factory selects the correct algorithm at runtime.



## API Documentation

### POST /check

See detailed API documentation:
- [POST /check API](docs/api/check.md)

- [POST /rules and GET /rules API](docs/api/rules.md)



## How To Run

### Prerequisites
- Docker Desktop with WSL2 integration enabled
- Git
- Go 1.22.4 (for running tests locally)

### Start the full stack
```bash
git clone git@github.com:Zartex-the-art/sei-ratelimiter.git
cd sei-ratelimiter
docker compose up --build
```
This starts:

App Node 1: http://localhost:8080
App Node 2: http://localhost:8081
Redis:       localhost:6379

Verify
curl http://localhost:8080/health
# {"status":"ok","node":"node-1"}

curl http://localhost:8081/health
# {"status":"ok","node":"node-2"}

Stop
docker compose down

How To Run Tests

Unit tests (no Redis required)
go test -race -v ./internal/algorithms/... ./internal/handlers/...

Integration tests (requires docker compose up)
docker compose up -d
REDIS_URL=localhost:6379 go test -race -v ./...
docker compose down
Full test harness (recommended)
./scripts/run_tests.sh

Coverage report
REDIS_URL=localhost:6379 go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

Manual API test
docker compose up -d
./scripts/test_api.sh
docker compose down



## Distributed Correctness

This rate limiter is designed to enforce limits correctly across
multiple nodes sharing one Redis instance. Correctness is guaranteed
at three levels:

### Level 1 — Single-Node Concurrency

Within a single node, multiple goroutines call Allow() concurrently.
Go's race detector verifies no data races.
All tests run with go test -race.

### Level 2 — Multi-Node Shared State

Both app nodes write to the same Redis keys.
Redis INCR is atomic — each node gets a unique return value.
No node can see a stale count.

### Level 3 — Atomic Read-Modify-Write (Phase 4)

INCR and EXPIRE must execute as one unit.
If INCR succeeds but EXPIRE fails (node crash), the key never expires.
Lua scripts fix this: Redis executes the full script atomically.
No other command can run between INCR and EXPIRE inside a Lua script.

### Verification
TestFixedWindow_AtomicUnderConcurrency: 300 goroutines, limit=50.
Exactly 50 allowed across all 300 concurrent requests.
Run 5 times: go test -race -count=5 -run TestFixedWindow_Atomic ./internal/
algorithms/...



## API Reference

### POST /check — Evaluate Rate Limit

Request:
```json
{"client_id": "user-1", "algorithm": "fixed_window", "limit": 10, "window_secs": 60}
```

Config resolution: if client_id has a stored rule, omit other fields:
```json
{"client_id": "user-1"}
```

Response (200):
```json
{"allowed": true, "remaining": 9, "algorithm": "fixed_window", "client_id": "user-1"}
```

POST /rules — Create Rate Limit Rule

Request:
```json
{"client_id": "user-1", "algorithm": "fixed_window", "limit": 10, "window_secs": 60, 
"enabled": true}
```

Response (201):
```json
{"id": "uuid", "client_id": "user-1", "algorithm": "fixed_window", "limit": 10, 
"window_secs": 60, "enabled": true, "created_at": "..."}
```

GET /rules — List All Rules
Response (200): {"rules": [...]}

GET /rules/:id — Get One Rule
Response (200): full rule object
Response (404): {"error": "rule not found"}

DELETE /rules/:id — Delete a Rule
Response (204): no body
Response (404): {"error": "rule not found"}



## Distributed Correctness Verification

### What We Test
With two nodes sharing one Redis, we verify:
1. The global limit is never exceeded regardless of which node serves requests
2. No over-counting — atomic Lua scripts prevent race conditions
3. No under-counting — every request is evaluated correctly

### Atomicity Tests (per algorithm)
Run 300 concurrent goroutines against a limit of 50:
go test -race -count=5 -run "TestFixedWindow_AtomicUnderConcurrency" ./...
go test -race -count=5 -run "TestSlidingWindow_AtomicUnderConcurrency" ./...
go test -race -count=5 -run "TestTokenBucket_AtomicUnderConcurrency" ./...
Expected: exactly 50 allowed across 300 goroutines, every run.

### k6 Correctness Test (cross-node)
Runs 10 VUs across both nodes for 60 seconds:
make correctness-test
Expected: total allowed <= (duration / window) × limit + 1 window buffer.




## Failure Modes

This section covers how the system behaves under each failure scenario.

See ops/RUNBOOK.md for step-by-step recovery procedures.

### App Node Failure

What happens:

- Requests to the failed node's port get connection refused
- The other node continues serving all traffic
- Docker auto-restarts the failed container within ~10 seconds
- Rate limit counts in Redis are unaffected

What holds:

- Rate limiting correctness (all state is in Redis)
- The other node's response time (no degradation)

What breaks:

- ~10 seconds of connection-refused on the failed node's port

Recovery:

Automatic (restart: unless-stopped policy).

---

### Redis Failure

What happens:

- Both nodes cannot evaluate rate limits
- All /check requests return HTTP 503 within 500ms
- go-redis retries 3 times before failing
- /health endpoint continues returning 200

What holds:

- HTTP server stays up
- Response time stays under 1 second
- No 500 errors
- No panics

What breaks:

- Rate limit enforcement is suspended
- Rules API returns 503

Recovery:

- Automatic reconnection when Redis restarts
- No app restart required
- Counters preserved via AOF persistence

Data loss risk: max 1 second (appendfsync everysec).

---

### Both Nodes + Redis Down

Complete outage.

Recovery:

1. docker compose up -d
2. Redis volume restores data
3. Limits resume from previous state

---

### Expected vs Unacceptable Behaviour

| Condition | Expected | Unacceptable |
|------------|------------|------------|
| app1 down | Connection refused on :8080 | app2 also fails |
| Redis down | 503 within 500ms | Hang, panic, 500 |
| Redis recovery | Auto-reconnect | Manual restart |
| app1 recovery | Auto-restart | Data loss |
| High load | Clean 200/429 | Crashes |




## Benchmarks

### Test Conditions
- Hardware: [document your machine]
- Docker resource limits: 0.5 CPU, 256MB RAM per app node
- Redis: 7-alpine, 128MB maxmemory, allkeys-lru
- k6 version: v0.51.0
- Algorithm: fixed_window (O(1) Redis operations)
- Both nodes serving (round-robin between :8080 and :8081)

### Results

| RPS    | p50 (ms) | p95 (ms) | p99 (ms) | Error Rate | Status |
|--------|----------|----------|----------|------------|--------|
| 1,000  |          |          |          |            |        |
| 5,000  |          |          |          |            |        |
| 10,000 | Day 19   |          |          |            |        |
*Fill in with actual measured values after each run.*
*Target: p99 < 5ms at 10K RPS.*

### How to Reproduce
```bash
docker compose up -d --build
k6 run tests/load/load_1k.js  # 1K RPS
k6 run tests/load/load_5k.js  # 5K RPS
k6 run tests/load/load_10k.js # 10K RPS (Day 19)




### Benchmark Methodology

**Why fixed_window for benchmarks:**
Fixed window uses O(1) Redis operations (one EVALSHA per request).
Sliding window uses O(N) sorted set operations — memory grows with traffic.
For raw throughput benchmarks, fixed window isolates the HTTP+Redis path.

**Why constant-arrival-rate executor in k6:**
This executor sends exactly N requests per second regardless of response time.
It exposes backpressure — if the server is slow, VUs queue up.
This is more realistic than constant-vus which just sends requests as fast as VUs allow.

**What p99 < 5ms means:**
At 10K RPS, 10,000 requests/second are processed.
p99 < 5ms means 99% of those 10,000 requests/second complete in under 5ms.
The slowest 1% (100 requests/second) may take longer.
This is the production target for a rate limiting sidecar.



## Benchmarks
### Test Environment
- Hardware: [CPU model, RAM]
- OS: Ubuntu 22.04 (WSL2)
- Docker: resource limits 0.5–1.0 CPU, 256–512MB per node
- Redis: 7-alpine, 128–256MB maxmemory, allkeys-lru
- k6: v0.51.0, constant-arrival-rate executor
- Algorithm: fixed_window (O(1) Redis EVALSHA per request)
- Distribution: round-robin across app1 (:8080) and app2 (:8081)

### Results

| RPS    | p50 (ms) | p95 (ms) | p99 (ms) | Error Rate | Pass? |
|--------|----------|----------|----------|------------|-------|
| 1,000  | X.XX     | X.XX     | X.XX     | 0.000%     | ✅    |
| 5,000  | X.XX     | X.XX     | X.XX     | 0.000%     | ✅    |
| 10,000 | X.XX     | X.XX     | X.XX     | 0.000%     | ✅    |

**Target: p99 < 5ms at 10,000 RPS — [MET / CLOSE / DOCUMENTATION NOTE]**

### Performance Summary

At 10K RPS sustained for 60 seconds:
- Both nodes remained healthy (zero container restarts)
- Redis remained stable (zero OOM events)
- Rate limiting remained correct (verified post-benchmark)
- Peak Redis ops/sec: [measured]

### How to Reproduce
```bash
git clone git@github.com:Zartex-the-art/sei-ratelimiter.git
cd sei-ratelimiter
docker compose up -d --build
k6 run tests/load/load_1k.js  # 1K RPS, 60s
k6 run tests/load/load_5k.js  # 5K RPS, 60s
k6 run tests/load/load_10k.js # 10K RPS, 60s
```



## gRPC Interface
In addition to the REST API, sei-ratelimiter exposes a gRPC interface.

### Endpoint
app1: localhost:9090
app2: localhost:9091

### Proto Definition
See api/proto/ratelimiter.proto

### Usage (grpcurl)
```bash
grpcurl -plaintext -d '{
  "client_id": "user-1",
  "algorithm": "fixed_window",
  "limit": 10,
  "window_secs": 60
}' localhost:9090 ratelimiter.RateLimiter/Check
```

### Response
```json
{
  "allowed": true,
  "remaining": 9,
  "algorithm": "fixed_window",
  "clientId": "user-1"
}
```



## What We Would Do at 10x Scale
Current system: 2 nodes, 1 Redis, ~10K RPS target.
10x scale: ~100K RPS, global deployment, multiple regions.

### Redis Cluster
Single Redis becomes a bottleneck at ~50K ops/sec.
Replace with Redis Cluster (3 primary + 3 replica nodes).
Consistent hashing routes each clientID to a specific shard.
Each shard handles ~33K RPS independently.

### Horizontal Node Scaling
Add more app nodes behind a real load balancer (AWS ALB, nginx).
Each app node is stateless — adding nodes requires no coordination.
Lua scripts ensure correctness regardless of node count.

### Edge Rate Limiting
At 100K RPS, the Redis round trip (~1ms) becomes a bottleneck.
Solution: CDN-level rate limiting (Cloudflare Workers, AWS WAF).
Rough counts at the CDN, precise enforcement at the service layer.

### Client-Specific Sharding
High-traffic clients get their own Redis shard.
Key prefix routing: VIP clients → shard-1, standard → shard-2.
Prevents one hot client from exhausting a shared shard.

### Consistent Hashing for Rules
At 100K clients, GET /rules (SMEMBERS + N HGETALL) becomes slow.
Replace rules:index set with a dedicated Redis Hash or sorted set.
Add cursor-based pagination to GET /rules.

### Observability
Add Prometheus metrics: allowed_total, blocked_total, latency_histogram.
Alert on: error rate > 0.1%, p99 > 10ms, Redis connection pool exhaustion.
Distributed tracing (OpenTelemetry) to correlate /check with Redis calls.



## Project Retrospective — sei-ratelimiter

### What We Built
A production-grade distributed rate limiter as a service:
- 3 algorithms: fixed window, sliding window, token bucket
- REST API: 5 endpoints with config resolution
- gRPC interface (Phase 6)
- Lua atomic scripts for distributed correctness
- Docker Compose: 2 nodes + Redis, one-command startup
- GitHub Actions CI with Redis service
- Load tested to 10K RPS

### What We Are Most Proud Of
[Team fills this in — honest answers only]

### What We Would Do Differently
[Team fills this in — honest answers only]

### What We Learned
[Team fills this in — one point per person]

### Numbers
Total tests: [N]
Coverage: [X]%
Load tested to: [X]K RPS, p99=[X]ms
Days taken: 21 (Days 1-21, 3 buffer days remaining)
Lines of Go code: [count with wc -l]
PRs merged: [count]