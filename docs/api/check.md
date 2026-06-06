# API Reference: POST /check

Evaluates whether a given client is allowed to make a request based on a specific rate-limiting algorithm.

## Endpoint

POST /check

## Request Payload (JSON)

| Field | Type | Required | Description |
|---------|---------|---------|---------|
| client_id | string | Yes | Unique identifier for the client |
| algorithm | string | Yes | fixed_window, sliding_window, token_bucket |
| limit | int | Yes | Maximum requests allowed |
| window_secs | int | Yes | Window size in seconds |

## Request Example

```json
{
  "client_id": "user-1",
  "algorithm": "fixed_window",
  "limit": 5,
  "window_secs": 60
}
```

## Response Payload

### Success (200 OK)

```json
{
  "allowed": true,
  "remaining": 4,
  "algorithm": "fixed_window",
  "client_id": "user-1"
}
```

### Error Response (400 Bad Request)

```json
{
  "error": "unknown algorithm"
}
```

## Curl Example

```bash
curl -X POST http://localhost:8080/check \
-H "Content-Type: application/json" \
-d '{
  "client_id": "user-1",
  "algorithm": "fixed_window",
  "limit": 5,
  "window_secs": 60
}'
```


## Config Resolution
POST /check supports two modes:

### Mode 1: Explicit (no stored rule)
Caller provides algorithm, limit, and window_secs in the request body.
Used when no stored rule exists for the client.

```json
{
  "client_id": "user-1",
  "algorithm": "fixed_window",
  "limit": 10,
  "window_secs": 60
}
```

### Mode 2: Stored Rule (config resolution)
Caller provides only client_id.
The server looks up the stored rule for that client and applies its algorithm/limit/window_secs automatically.

```json
{
  "client_id": "user-1"
}
```

Response includes rule_id to show which rule was applied:

```json
{
  "allowed": true,
  "remaining": 9,
  "algorithm": "fixed_window",
  "client_id": "user-1",
  "rule_id": "550e8400-..."
}
```

### Priority
If a stored rule exists for client_id, it ALWAYS takes priority over any algorithm/limit/window_secs in the request body.

### Error: No Rule + No Fields
If no stored rule exists and algorithm is missing from the body:

```text
HTTP 400 — no stored rule for this client — algorithm is required
```