# ADR-0005: Redis-Backed State for Rate Limit Counters

## Status
Accepted

## Context

The in-memory FixedWindow from Day 6 works on a single node.

With two nodes, each node has its own counter.

A client hitting both nodes can exceed the limit — incorrect behaviour.

## Decision

Store rate limit counters in Redis shared by all app nodes.

Use dependency injection: FixedWindow accepts a Store interface.

RedisStore is the production implementation.

FakeStore is the test implementation.

## Why Dependency Injection

Unit tests use FakeStore — no Redis, no Docker, runs anywhere.

Integration tests use RedisStore — validates real Redis behaviour.

Algorithm code is decoupled from the Redis client library.

Future: swap Redis for another store without changing algorithm code.

## Alternatives Considered

Hard-code redis.Client inside FixedWindow:

Rejected — untestable without a running Redis instance.

Gossip protocol between nodes:

Rejected — too complex, eventual consistency not suitable for rate limiting.

## Consequences

### Good

- Global limit enforced across all nodes
- Algorithm code fully unit-testable without Redis
- State survives app restarts

### Bad

- Redis is a single point of failure
- INCR + EXPIRE pipeline not atomic
- Integration tests require Docker Compose