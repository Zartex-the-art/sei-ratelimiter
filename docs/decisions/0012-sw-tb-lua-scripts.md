# ADR-012: Lua Scripts for Sliding Window and Token Bucket

## Status
Accepted

## Context
Same atomicity problem as fixed window (ADR-011), but more complex:
Sliding window requires 3-4 separate Redis commands per request.
Token bucket requires HGETALL + compute + HSET per request.
None of these multi-step flows are atomic without Lua.

## Decision
Add SlidingWindowAllow() and TokenBucketAllow() methods to the Store interface.
RedisStore implements them via Lua scripts.
FakeStore implements them in-memory for unit tests.
Algorithm implementations use one method call instead of 3-4.

## Consequences
Good:
  True atomicity for all three algorithms
  Fewer round trips (3-4 Redis calls → 1 EVALSHA)
  Store interface methods map 1:1 to algorithm operations
  Unit tests remain fast via FakeStore

Bad:
  Token bucket Lua script is complex (float arithmetic, HGETALL parsing)
  If Lua has a bug, it affects the entire token bucket algorithm
  Harder to debug than plain Redis commands
  (Mitigated by comprehensive atomicity tests)