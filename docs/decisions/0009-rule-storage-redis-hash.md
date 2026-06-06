#ADR-0009: Rule Storage Using Redis Hash + Index Set

## Status Accepted

## Context

Rate limit rules need to be stored persistently so clients don't send algorithm/limit/window_secs on every/check request.
Rules must survive server restarts and be visible across all nodes.

## Decision
Store each rule as a Redis hash: rule: {uuid}
Maintain a Redis set for listing: rules:index
Maintain a client lookup key: rule:by-client: {clientID}

## Alternatives Considered
Store rules in PostgreSQL:
Requires a new dependency, migration system, and connection pooling.
Rejected Redis is already the shared datastore.
Store rules as a single JSON blob in one Redis key:
Simple but causes write conflicts when multiple nodes create rules simultaneously.
Rejected - hash per rule with SADD index is atomic and conflict-safe.

## Consequences
Good:
No new infrastructure -Redis already running
HSET is atomic -no partial writes
SADD to index is atomic -no duplicate IDs in listing
Rules visible on all nodes immediately (shared Redis)

Bad:
No query capability - cannot filter rules by algorithm or limit
No TTL on rules manual deletion required
SMEMBERS + N HGETALL for listing O(N) Redis calls for N rules
(Acceptable at the scale of this project - addressed in 10x section)