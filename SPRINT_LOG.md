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