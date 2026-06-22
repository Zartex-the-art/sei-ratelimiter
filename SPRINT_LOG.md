# Sprint Log — Zartex SEI Project 1

---

## Day 1 — May 18, 2026

**Phase:** 1 — Foundation
**Goal:** Go fundamentals + HTTP /health server
**Schedule:** All 5 members, same tasks, no role split today

### Completed

- All 5 members completed tour.golang.org sections 1–13
  (variables, types, functions, structs, pointers, error handling, methods)
- All 5 members read Go by Example: error handling, closures, interfaces
- All 5 members built HTTP server: GET /health returns {"status": "ok"}
- All 5 members tested server with curl — all returning correct JSON
- EOD confidence ratings posted in group chat

### What Was Harder Than Expected

- Pointers — the difference between value and address took time to click
- Understanding why pointer receivers are needed on struct methods
- Interfaces — the concept of implicit implementation (no 'implements' keyword)

### What Was Easier Than Expected

- Setting up the HTTP server — Go's net/http package is straightforward
- JSON encoding — json.NewEncoder(w).Encode() just works

### Phase 1 Gate Status

- [ ] All 5 members: HTTP server returns {"status":"ok"} on GET /health
- [ ] All 5 members: Go fundamentals tour complete

---

## Day 2 — May 19, 2026

**Phase:** 1 — Foundation
**Goal:** Full dev environment + GitHub repo + role-specific tasks
**Schedule:** All 5 members

### Environment Setup (All Members)

- [ ] WSL2 Ubuntu 22.04 installed and running
- [ ] Go 1.22.4 installed and verified (go version)
- [ ] VS Code connected to WSL2 (WSL: Ubuntu in bottom-left)
- [ ] Git configured (name, email, defaultBranch=main)
- [ ] GitHub SSH key generated and added to account
- [ ] sei-ratelimiter repo cloned successfully

### Role Tasks

- [ ] Abhishek: k6 installed, healthHandler refactored, 4 tests passing, PR submitted
- [ ] Madhu: Limiter interface, Config struct, FixedWindow impl, PR submitted
- [ ] Gayathri: 5 handler tests written and passing, PR submitted
- [ ] Hari: Docker Desktop installed, Redis CLI commands practiced, PR submitted
- [ ] Vishnu: README skeleton, ADR template, ADR-001, Sprint Log, PR submitted

### PRs Submitted Today

| Member | Branch | Status |
|--------|--------|--------|
| Abhishek | day2/abhishek-setup | |
| Madhu | day2/madhu-interface | |
| Gayathri | day2/gayathri-tests | |
| Hari | day2/hari-docker-redis | |
| Vishnu | day2/vishnu-docs | |

### Notes

<!-- Fill in after EOD -->
<!-- What was harder than expected? -->
<!-- Did everyone get environment set up? Any blockers? -->

---

## Day 2 — May 19, 2026

**Phase:** Documentation & Architecture

### Completed

- Created README structure
- Created ADR template
- Created ADR-001
- Created initial architecture diagram
- Practiced GitHub PR workflow


## Day 3 — May 20, 2026

**Phase:** Foundation

**Goal:** Go concurrency — goroutines, mutex, race conditions

### Completed

- All 5 members: goroutines and channels practice completed
- Abhishek: race condition demo and worker pool pattern
- Madhu: concurrent Allow() test with race checks
- Gayathri: concurrent correctness tests and race verification
- Hari: Docker networking exercise
- Vishnu: CONCURRENCY.md, ADR-002 draft, Sprint Log Day 2 update

### Notes

<!-- Hardest concepts and blockers -->

## Day 4 — May 21, 2026

**Phase:** Foundation

**Goal:** Docker fundamentals — Dockerfile, multi-stage builds, Redis CLI, Compose

### Completed

- Abhishek: multi-stage Dockerfile, Redis CLI deep dive, first k6 smoke test
- Madhu: Redis data structure simulations for all 3 algorithms, key design doc
- Gayathri: Redis test connection pattern, skip-not-fail pattern established
- Hari: full 3-service docker-compose.yml with health checks and volumes
- Vishnu: DOCKER_CONCEPTS.md, ADR-002, README How To Run section

### Notes

<!-- Add observations here -->

---

## Phase 1 Summary — Days 1–5

**All 5 members achieved:**

- Go 1.22.4 installed, WSL2 running, VS Code connected
- GitHub SSH working, repo cloned, branch protection on main
- HTTP /health server built and tested
- Go concurrency understood: goroutines, mutex, race detector
- Docker: Dockerfile, multi-stage builds, Compose, Redis CLI
- CI pipeline live on GitHub Actions
- Package structure and interfaces defined
- Full 3-service stack running with one command


## Day 6 — Phase 2 Kick-Off

### Vishnu
- Updated Fixed Window documentation
- Added algorithm comparison table
- Created fixed-window-sequence diagram
- Added ADR-004
- Updated sprint documentation

### Team Progress
- Fixed Window implementation started
- Test suite preparation ongoing
- Docker hardening ongoing
- k6 integration testing ongoing



## Day 6 - May 23, 2026

Phase: Core Algorithms Day 1 of 5
Goal: Fixed window in-memory, test harness, infra hardening

Deliverables:
Abhishek: scripts/run_tests.sh, Makefile with 6 targets, k6 smoke test hitting both nodes, CORRECTNESS.md
Madhu: FixedWindow in-memory with per-client window tracking and Reset() Gayathri: 8 tests (6 unit + 2 concurrent), -race-count=3 all passing
Hari: resource limits (CPU/RAM), Redis maxmemory + LRU, COMPOSE_GUIDE.md
Vishnu: fixed window README, algorithm comparison table, ADR-004

Phase2: coordination running smoothly:
Madhu + Gayathri morning sync: 30 min, spec written, parallel work
Abhishek midday + EOD check:
all 5 on track
Hari stack: stable all day under resource limits
Vishnu PR gate: all 5 PRs reviewed before merge


## Day 7 — May 25, 2026

Phase: Core Algorithms — Day 2 of 5

Goal: Fixed window Redis-backed, dependency injection, two-node correctness

Deliverables:

- Madhu: RedisStore + FakeStore + DI
- Gayathri: 6 unit tests + Redis tests
- Abhishek: TestTwoNodes
- Hari: restart scenarios
- Vishnu: README Redis impl, ADR-005, architecture-v5

Key result:

TestTwoNodes_LimitEnforcedGlobally PASSES.


## Day 8 — May 26, 2026
Phase: Core Algorithms — Day 3 of 5
Goal: Sliding window algorithm, auto-restart verification
Deliverables:
  Madhu: SlidingWindow — ZADD/ZREMRANGEBYSCORE/ZCOUNT/ZREM, ms timestamps
  Gayathri: 6 sliding window tests — NoBoundaryBurst is the critical one
  Abhishek: boundary burst comparison tests (FW vs SW), harness step 6
  Hari: auto-restart verified (kill app1 → restart < 15s, app2 unaffected)
  Vishnu: sliding window docs, comparison table col2, ADR-006, boundary burst diagram
Key result:
  TestSlidingWindow_NoBoundaryBurst PASSES.
  TestBoundaryBurst_Comparison: swBurst == 0.
  Sliding window boundary burst prevention confirmed.



  ## Day 10- June 1, 2026
Phase: Core Algorithms Day 5 of 5
Goal: Algorithm factory, full regression, Phase 2 gate
Deliverables:
Madhu: NewLimiter factory, ValidAlgorithms, startup sanity check
Gayathri: 8 factory tests, full regression zero failures
Abhishek: Phase 2 final harness, all pending PRS merged, report posted
Hari: clean clone verified, STARTUP_WALKTHROUGH.md
Vishnu: final architecture diagram, ADR-008, comparison table complete

## Phase 2 Retrospective
### What We Built
Three Redis-backed rate limiting algorithms:
Fixed Window: INCR + EXPIRE, 0(1) memory, boundary burst documented
Sliding Window: ZADD + ZREMRANGEBYSCORE + ZCOUNT, no boundary burst
Token Bucket: HGETALL + HSet, continuous refill, burst allowance
Algorithm factory: single entry point for all three algorithms.
Dependency injection: Store interface enables FakeStore for unit tests.
Two-node correctness: limits enforced globally across Redis.
Boundary burst prevention: confirmed by TestBoundaryBurst_Comparison.

### What Was Harder Than Expected
Sliding window timestamp precision and ZREM-on-block step.
Token bucket float arithmetic parsing from Redis hash fields.
Concurrent tests getting exactly limit allowed, not limit+1.
Morning syncs consistently ran over 30 minutes.

### What Was Easier Than Expected
Dependency injection - mechanical once understood.
FakeStore - simple to write, immediately useful.
Docker restart scenarios - restart: unless-stopped just worked.

### What We Would Do Differently
Design the factory on Day 5 alongside the interfaces.
45-minute morning syncs, not 30.
Token bucket before sliding window simpler algorithm first.

### Phase 3 Readiness
Factory returns correct algorithm by string.
Store interface complete for all three algorithms.
HTTP handlers (Phase 3) call NewLimiter with algorithm from request.



## Day 11 -June 2, 2026

Phase: REST API Layer Day 1 of 4
Goal: POST /check endpoint wired to algorithm factory

Deliverables:
Madhu: internal/models/check.go, internal/handlers/check.go, route wired
Gayathri: 6 handler unit tests using httptest and FakeStore
Abhishek: 8 integration tests against live Docker (check_test.go)
Hari: Docker verification, curl commands against both nodes documented
Vishnu: docs/api/check.md - full API reference with curl examples

Key result:
POST /check accessible on : 8080 and : 8081.
All 3 algorithms reachable through REST.
Blocked requests return 200 with allowed=false (not 429).



## Day 12 — June 5, 2026
Phase: REST API Layer — Day 2 of 4
Goal: POST /rules + GET /rules — rule storage in Redis

Deliverables:
   Madhu: CreateRuleHandler, ListRulesHandler, RedisStore.Client() accessor,
         google/uuid dependency, rules:index + rule:by-client secondary index
  Gayathri: 5 handler unit tests — CreateReturns201, FullObject, ListEmpty,
            ListAll, ValidationErrors (6 invalid input cases)
  Abhishek: 4 integration tests — CreateAndList, InvalidAlgorithm,
            MissingClientID, PersistsAcrossNodes
  Hari: config/redis.conf mounted, AOF persistence verified, 5 rules survive
        Redis restart and full stack restart, scripts/test_api.sh,
        scripts/inspect_redis.sh
  Vishnu: docs/api/rules.md, ADR-009, sprint log Day 11

Key result:
  Rules persist across Redis restarts — AOF confirmed working.
  Rules created on node1 visible on node2 immediately — shared Redis confirmed.



  ## Day 14 — June 8, 2026

Phase: REST API Layer — Day 4 of 4 (PHASE 3 COMPLETE)
Goal: Code cleanup, edge case tests, CI pipeline update, README completion

## Phase 3 Retrospective

### What We Built

REST API layer on top of the Phase 2 algorithm library:
  POST /check  — rate limit evaluation with config resolution
  POST /rules  — create stored rate limit rules
  GET /rules   — list all rules
  GET /rules/:id — get one rule
  DELETE /rules/:id — delete a rule and all Redis keys
Config resolution: callers send just {client_id}, server applies stored rule.
Go 1.22 path parameters: no external router dependency.
Google UUID: proper UUID generation without external crypto complexity.

### What Was Harder Than Expected

Config resolution edge cases — stored rule override behaviour needed extra
tests and a careful priority decision (stored rule always wins).
DELETE cleanup — forgetting rule:by-client: key was caught by the
RemovesClientIndex test. Good test coverage saved hours of debugging.
Consistent error response struct — different handlers had different
error formats, discovered in Madhu's cleanup pass.

### What Was Easier Than Expected

Go 1.22 path parameters — cleaner than expected, no gorilla/mux needed.
Handler unit tests with httptest — FakeStore made tests fast and clean.
The dependency injection pattern from Phase 2 paid off immediately —
handlers accept Store and *redis.Client, easy to test with fake/real.

### Phase 4 Readiness

All 5 endpoints working and tested.
Config resolution confirmed working across nodes.
CI runs integration tests with real Redis.
Phase 4 (Lua scripts for atomicity) starts tomorrow.



## Day 14 — June 8, 2026

Phase: REST API Layer — Day 4 of 4 (PHASE 3 COMPLETE)
Goal: Code cleanup, edge cases, CI with Redis service, README completion

Deliverables:

  Madhu: consistent error struct, all exports documented, go vet + staticcheck clean
  Gayathri: concurrent /check test (50 goroutines), Unicode clientID, large limit
  Abhishek: coverage report, all test files clean, CI confirmed green
  Hari: CI pipeline updated with Redis service, integration tests run in GitHub Actions
  Vishnu: README How To sections complete, Phase 3 retrospective, ADR review
Phase 3 gate cleared. All 5 endpoints working, CI green, coverage > 70%.
Phase 4 starts: distributed correctness via Lua scripts.

## Phase 4 Kick-Off — Day 15 — June 9, 2026

Phase: Distributed Correctness — Day 1 of 3
Goal: Lua script for fixed window — replace pipeline with atomic INCR+EXPIRE



## Day 15 — June 9, 2026
Phase: Distributed Correctness — Day 1 of 3
Goal: Lua script for fixed window — atomic INCR+EXPIRE

Deliverables:
  Madhu: fixedWindowScript using redis.NewScript(), Increment() replaced
  Gayathri: TestFixedWindow_AtomicUnderConcurrency — 300 goroutines, count==50 exactly
  Abhishek: docs/RACE_CONDITION_ANALYSIS.md, CI verified green with Lua changes
  Hari: EVALSHA confirmed in redis-cli MONITOR, ops/LUA_MONITORING.md
  Vishnu: ADR-011, Distributed Correctness section in README, architecture diagram 
updated

Key result:
  EVALSHA appears in Redis monitor — Lua scripts executing correctly.
  TestFixedWindow_AtomicUnderConcurrency: 50/50 across 5 runs — perfect atomicity.
  All existing tests pass unchanged — Lua changes implementation not behaviour.



  ## Day 17 — June 11, 2026

Phase: Distributed Correctness — Day 3 of 3 (PHASE 4 COMPLETE)

Goal:
Failure simulation, graceful degradation, RUNBOOK.md

Deliverables:

- Madhu: Redis failure handling, retry config, 503 responses
- Gayathri: Redis failure tests
- Abhishek: Failure simulation and k6 testing
- Hari: RUNBOOK.md and restart verification
- Vishnu: Failure Modes README, Phase 4 retrospective

## Phase 4 Retrospective

### What We Built

- Atomic Redis operations via Lua scripts
- Graceful degradation on Redis failure
- go-redis retry configuration
- Distributed correctness validation
- Failure simulation and recovery testing

### What Was Hardest

- Token bucket Lua script implementation
- Distinguishing infrastructure errors from logic errors

### What Surprised Us

- Redis reconnects very quickly (~3-5s)
- Lua scripts eliminated race conditions completely

### Phase 5 Readiness

- All algorithms atomic
- Correctness verified
- Graceful degradation implemented
- Operational documentation complete

Ready for Phase 5 load testing.



## Day 17 — June 11, 2026
Phase: Distributed Correctness — Day 3 of 3 (PHASE 4 COMPLETE)
Goal: Failure simulation, graceful degradation, RUNBOOK.md

Deliverables:
  Madhu: isInfraError(), 503 on Redis failure, retry + pool config
  Gayathri: ErrorStore, 4 Redis failure handler tests
  Abhishek: failure_sim.js, failure_simulation.sh, results documented
  Hari: RUNBOOK.md all 4 scenarios, kill vs stop verified, reconnect timed
  Vishnu: Failure Modes README, Phase 4 retrospective, ADR-013
  
Key results:
  Redis down: 503 within 500ms confirmed.
  Auto-reconnection: ~5 seconds after Redis restarts.
  Failure simulation: app1 kill + Redis kill both recovered cleanly.
  Phase 4 gate: all items checked. Phase 5 starts.



  ## Day 19 — June 13, 2026
Phase: Load Testing — Day 2 of 2 (PHASE 5 COMPLETE)
Goal: 10K RPS sustained, final benchmark table

## Phase 5 Retrospective

### What We Measured
1K RPS: p99=Xms — baseline confirmed
5K RPS: p99=Xms — tuning [needed/not needed]
10K RPS: p99=Xms — project target [MET/CLOSE]

### Bottlenecks Found
[Document what pprof showed — JSON encoding? Pool exhaustion? Redis ops?]

### Tuning Applied
[Document what was changed between Day 18 and Day 19]

### What Would Give 2x Improvement
Redis connection pool optimisation: batch requests where possible.
HTTP keep-alive: already enabled in net/http by default.
JSON encoding: switch to sonic for ~30% improvement.
Redis pipeline for non-atomic operations (rules listing).

### Phase 6 Readiness
All benchmarks complete. Final table in README.
Phase 6: gRPC stretch goal or REST hardening + final polish.



## Day 21 — June 19-20, 2026
Phase: Final Polish — PHASE 6 + PROJECT 1 COMPLETE
Goal: All Definition of Done items checked, project closed

Deliverables:
  Abhishek: 5× full harness runs, benchmark verification, final DoD walk-through,
            final project status report, all PRs reviewed and merged
  Madhu: complete codebase review, go vet + staticcheck clean
  Gayathri: final test audit, TEST_SUMMARY.md, -count=5 all green
  Hari: clean-clone startup verified, Makefile all targets, ops/ complete
  Vishnu: README 100%, all 14 ADRs Accepted, project retrospective

## Project 1 Retrospective

### What We Set Out To Build
A distributed rate limiter as a service:
3 algorithms, REST API, Redis backend, Lua atomicity, Docker multi-node,
10K RPS target. Built by a 5-person team in 21 days, starting from zero
knowledge of Go, Redis, Docker, and distributed systems.

### What We Actually Built
Everything in the plan — and we understood every line.
Fixed Window, Sliding Window, Token Bucket:
Not just implemented but understood deeply enough to explain
the boundary burst problem, the Lua atomicity guarantee, and
the float64 precision issues in token bucket refill arithmetic.
REST API with config resolution:
Not just CRUD — a properly designed service where stored rules
allow callers to send just {client_id} and get correct behaviour.
Two-node correctness:
TestTwoNodes_LimitEnforcedGlobally is not a trivial test.
It proves that our system genuinely distributes rate limiting
across nodes — which was the whole point.

### What Was Hardest
Token bucket Lua script — float arithmetic in Lua 5.1, HGETALL parsing,
handling the initial-state case (no hash exists). We spent half a day
debugging why remaining was always slightly off before understanding
the Unix millisecond precision issue.
The team falling behind in Phase 2 (Days 6-8) — combining two days
into one was stressful but proved we could execute under pressure.

### What Surprised Us
How much the morning sync between Madhu and Gayathri mattered.
On days they synced, PRs merged cleanly. On days they skipped,
test failures revealed spec mismatches that cost hours.
How useful the race detector was — found 3 real bugs we never
would have caught by reading the code.

### What We Learned
Abhishek:  Go concurrency, k6 load testing, being a real team lead
Madhu:     Lua scripting, Redis data structures, HTTP handler design
Gayathri:  Test-driven thinking, httptest patterns, concurrent correctness
Hari:      Docker Compose operations, GitHub Actions, Redis monitoring
Vishnu:    Architecture documentation, ADR discipline, technical writing

### Numbers
Lines of Go code:       [wc -l $(find . -name "*.go") | tail -1]
Total tests:            [N]
Test coverage:          [X]%
ADRs written:           14
PRs merged:             [N]
Days taken:             21 (Days 1-21, all 6 phases)
10K RPS p99:            [X]ms

### If We Did This Again
1. Build the factory in Week 1, not Week 5
2. Madhu+Gayathri morning sync is non-negotiable — never skip it
3. Write the ADR before the code, not after
4. Phase 4 (Lua) earlier — it should follow Phase 2, not Phase 3
5. Load test earlier — some bottlenecks found in Phase 5 would have
   changed Phase 2 design decisions

### Final Thought
We started with zero domain knowledge. We finished with a system
that correctly rate limits across distributed nodes using atomic Lua
scripts with measured p99 latency at 10K RPS.
That is not beginner work. That is real engineering.
