# ADR-0008: Algorithm Factory Pattern
## Status Accepted
## Context
The HTTP layer (Phase 3) needs to create a Limiter based on the algorithm field in the request body or stored rule.
Without a factory, HTTP handlers would need to import and construct each algorithm type individually tightly coupling the API layer to the algorithm implementations.

## Decision
Implement NewLimiter (algorithm string, store Store, limit, windowSecs int) as the single entry point for algorithm creation.
Centralise algorithm string constants in the factory file.
ValidAlgorithms() exposes the list for HTTP input validation.

## Consequences
Good:
HTTP handlers call NewLimiter never import individual algorithm types
Adding a fourth algorithm requires one new case in the switch statement
Input validation centralised limit and windowSecs validated once String constants prevent typos across the codebase
Bad:
All algorithms must satisfy the same Limiter interface
No per-algorithm configuration beyond limit and windowSecs
(Stored rules in Phase 3 handle this via the rules API)