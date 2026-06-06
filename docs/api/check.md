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