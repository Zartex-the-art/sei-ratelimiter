# Restart Scenarios and Expected Behaviour

## Scenario 1: Redis Restart
Trigger:
docker compose restart redis

Expected:
- Redis reloads data from AOF
- Rate limit counters preserved
- Both app nodes reconnect automatically
- Health checks pass after reconnect

## Scenario 2: App Node Restart
Trigger:
docker compose stop app1
docker compose start app1

Expected:
- app2 continues serving traffic
- Redis data unaffected
- app1 reconnects successfully

## Scenario 3: Full Stack Restart
Trigger:
docker compose down
docker compose up -d

Expected:
- Redis volume preserves data
- Both apps reconnect
- Health endpoints return 200
