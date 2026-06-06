# ADR-010: Config Resolution in POST /check

## Status
Accepted

## Context
Callers initially must send algorithm/limit/window_secs on every /check request.
This is repetitive for clients with stable rate limit configurations.
Stored rules (Days 12-13) allow pre-configuring limits per client.
The /check endpoint needs to use stored rules automatically.

## Decision
Config resolution priority:

1. Stored rule for client_id (lookup via rule:by-client: {clientID})
2. Fields in request body (algorithm, limit, window_secs)
3. Neither exists →400 Bad Request

Stored rule always takes priority over request body fields.
Response includes rule_id when a stored rule was used.

## Alternatives Considered
Request body always takes priority:
Rejected stored rules would be ignored by any caller that sends body fields.
Merge stored rule and request body (request body overrides specific fields): Rejected complex merging logic, ambiguous behaviour, hard to test.

## Consequences

Good:
Callers can send just {client_id} after setting up a rule once
rule_id in response lets callers audit which rule was applied
Priority is clear -stored rule wins, always

Bad:
Callers cannot override a stored rule without deleting it first
If a client's stored rule is wrong, all their /check calls use wrong config
(Addressed by DELETE /rules/:id)