# ADR-0007: Token Bucket Implementation Using Redis Hash
## Status
Accepted
## Context
We need a third rate limiting algorithm that handles bursty traffic patterns.
Fixed window and sliding window enforce hard request counts per window.
Some clients (SDKs, batch jobs, mobile) send bursts then go idle —
these clients are penalised unfairly by window-based algorithms.
## Decision
Implement token bucket using a Redis hash with two fields:
  tokens:      current available tokens (float64)
  last_refill: timestamp of last token calculation (int64 unix ms)
Rate = limit / windowSecs tokens per second.
Refill is computed lazily on each request, not on a background timer.
## Why Lazy Refill
No background process required.
Redis has no built-in timer/cron — lazy computation on each request
is simpler and more reliable than a separate refill worker.
## Alternatives Considered
Leaky bucket:
  Similar concept but models output rate, not input burst.
  Token bucket is more intuitive for API rate limiting.
Fixed refill intervals:
  Refill tokens every N seconds via a background job.
  Requires a separate process, adds complexity.
  Rejected in favour of lazy computation.
## Consequences
Good:
  Burst allowance — clients can use accumulated tokens instantly
  O(1) memory per client regardless of request rate
  Lazy refill — no background process needed
Bad:
  HGETALL + HSet are two separate commands — not atomic
  Lua script required in Phase 4 for correctness under concurrency
  Float arithmetic for tokens — minor precision issues possible
  (Lua script uses integer math to avoid this — addressed in Phase 4)