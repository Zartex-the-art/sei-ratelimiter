# ADR-013: Graceful Degradation on Redis Failure

## Status

Accepted

## Context

If Redis is unavailable, the rate limiter cannot function.

Two choices:

- fail-open
- fail-closed

The HTTP response must be clear and fast.

## Decision

Return HTTP 503 Service Unavailable when Redis is unreachable.

503 means backend unavailable.

500 means application bug.

Timeout:
500ms

Retries:
3 with exponential backoff

## Consequences

Good:

- Fast failure
- Correct HTTP semantics
- Auto recovery

Bad:

- Traffic uncontrolled during Redis outage
- Error detection uses string matching

Future:

- Typed errors
- Sentinel errors