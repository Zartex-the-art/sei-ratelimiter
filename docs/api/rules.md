# API Reference: POST /rules and GET /rules

## POST /rules — Create a Rate Limit Rule

Stores a rate limiting rule in Redis. Rules are used by POST /check
in Day 13 to automatically apply limits without specifying them per request.

### Endpoint

POST /rules

### Request Payload

| Field | Type | Required | Description |
|-------------|--------|----------|-------------|
| client_id | string | Yes | Client identifier to apply this rule to |
| algorithm | string | Yes | fixed_window, sliding_window, or token_bucket |
| limit | int | Yes | Max requests allowed. Must be > 0 |
| window_secs | int | Yes | Window duration in seconds. Must be > 0 |
| enabled | bool | No | Whether the rule is active. Defaults to false if omitted |

### Request Example

```json
{
  "client_id": "user-alice",
  "algorithm": "fixed_window",
  "limit": 100,
  "window_secs": 60,
  "enabled": true
}
```

### Response (201 Created)

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "client_id": "user-alice",
  "algorithm": "fixed_window",
  "limit": 100,
  "window_secs": 60,
  "enabled": true,
  "created_at": "2026-06-05T09:00:00Z"
}
```

### Error Response (400 Bad Request)

```json
{ "error": "client_id is required" }
```

```json
{ "error": "limit must be > 0" }
```

```json
{ "error": "unknown algorithm \"banana\"" }
```

### curl Example

```bash
curl -X POST http://localhost:8080/rules \
-H "Content-Type: application/json" \
-d '{"client_id":"user-alice","algorithm":"fixed_window","limit":100,"window_secs":60,"enabled":true}'
```

---

## GET /rules — List All Rules

Returns all stored rate limit rules.

### Endpoint

GET /rules

### Response (200 OK)

```json
{
  "rules": [
    {
      "id": "550e8400-...",
      "client_id": "user-alice",
      "algorithm": "fixed_window",
      "limit": 100,
      "window_secs": 60,
      "enabled": true,
      "created_at": "2026-06-05T09:00:00Z"
    }
  ]
}
```

### Empty Response

```json
{
  "rules": []
}
```

### curl Example

```bash
curl http://localhost:8080/rules
```

## Redis Storage Design

Each rule is stored as:

```text
rule:{id}              — Hash with all rule fields
rules:index            — Set of all rule IDs (for listing)
rule:by-client:{cid}   — String pointing to rule ID (for /check resolution, Day 13)
```