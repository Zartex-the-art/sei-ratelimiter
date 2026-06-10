# ADR-0011: Lua Scripts for Atomic Redis Operations

## Status
Accepted

## Context

The fixed window rate limiter uses two Redis commands:
INCR fw:{clientID} — increment counter
EXPIRE fw:{clientID} N — set TTL on new window
These are separate commands and NOT atomic.
Two race conditions exist:
1. A node can crash between INCR and EXPIRE, leaving a key
   that never expires — the client is permanently blocked.
2. Two nodes can both call INCR with count=1 (brand new key),
   and both call EXPIRE — the second EXPIRE resets the window
   start time, allowing more requests than intended.

## Decision
Replace the INCR + EXPIRE pipeline with a Lua script executed
via redis.NewScript():
local count = redis.call('INCR', KEYS[1])
if count == 1 then
    redis.call('EXPIRE', KEYS[1], tonumber(ARGV[1]))
end
return count
Redis executes Lua scripts atomically and single-threaded.
No other command can interleave between INCR and EXPIRE.

## Why redis.NewScript() Over Raw EVAL
redis.NewScript() computes the SHA1 hash at startup.
First call: EVAL (sends full script — one-time cost)
Subsequent calls: EVALSHA (40-byte hash instead of full script)
NOSCRIPT error: automatic fallback to EVAL
At 10K RPS, EVALSHA vs EVAL saves significant bandwidth per request.

## Scope of Change
Only internal/store/redis_store.go changes.
The Store interface (Increment signature) is unchanged.
FixedWindow.Allow() is unchanged.
All existing tests pass without modification.
One new test added: TestFixedWindow_AtomicUnderConcurrency.

## Consequences

Good:
  Eliminates both race conditions described above
  Keys always get a TTL — no permanently stuck counters
  EVALSHA is more efficient than pipeline for high-frequency calls
  Implementation change is invisible to callers (same interface)
  
Bad:
  Lua scripts are harder to debug than plain Redis commands
  SCRIPT FLUSH or Redis restart requires re-loading scripts
  (handled automatically by redis.NewScript() fallback)
  Lua execution is single-threaded — complex scripts block Redis
  (our script is 3 lines — negligible impact)
