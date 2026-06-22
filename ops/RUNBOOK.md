# RUNBOOK

## Startup Procedure

Start the stack:

```bash
docker compose up -d --build
# Operational Runbook

## Two Node Verification

### Start Services

```bash
docker compose up -d
## Restart Policy Verification

1. Verify containers are running

docker compose ps

2. Stop app1

docker stop app1

3. Verify app1 stopped

docker ps

Observation:
- app1 stopped
- app2 running
- redis running

4. Verify endpoint

curl http://localhost:8080/health

Observed:

curl: (7) Failed to connect to localhost port 8080 after 1 ms: Could not connect to server

Conclusion:
Docker restart policy `unless-stopped` does not restart containers that are manually stopped by the user.
## Redis Reconnect Verification

1. Stop Redis

docker stop sei-ratelimiter-redis-1

2. Send request

curl -X POST http://localhost:8080/check \
-H "Content-Type: application/json" \
-d '{"client_id":"reconnect-test","algorithm":"fixed_window","limit":5,"window_secs":60}'

Observed:

{"error":"internal server error"}

3. Restart Redis

docker start sei-ratelimiter-redis-1

4. Retry request

curl -X POST http://localhost:8080/check \
-H "Content-Type: application/json" \
-d '{"client_id":"reconnect-test","algorithm":"fixed_window","limit":5,"window_secs":60}'

Observed:
Request succeeded after Redis became available.

Conclusion:
The application resumes normal operation once Redis connectivity is restored.
