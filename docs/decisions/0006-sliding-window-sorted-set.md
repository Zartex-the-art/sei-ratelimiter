# ADR-0006: Sliding Window Using Redis Sorted Sets

## Status

Accepted

## Context

Fixed window has a boundary burst problem:

clients can make 2x the configured limit by sending requests at a window boundary.

A more precise algorithm is needed for security-critical use cases.

## Decision

Implement sliding window log using Redis sorted sets.

- Score = unix timestamp in milliseconds
- Member = nanosecond timestamp string (unique per request)
- Prune = ZREMRANGEBYSCORE on every Allow() call
- Count = ZCOUNT after pruning

## Why Sorted Sets

Natural fit: timestamps as scores, efficient range operations.

ZREMRANGEBYSCORE is O(log N + M).

ZCOUNT is O(log N).

## Alternatives Considered

### Sliding window counter

Approximate.

Rejected for billing/security.

### In-memory sorted structure

Not shared across nodes.

Rejected.

## Consequences

### Good

- Eliminates boundary burst
- Precise to milliseconds
- Shared across nodes

### Bad

- Higher memory usage
- More Redis commands
- Not atomic yet