#!/usr/bin/env bash
set -euo pipefail

echo "====================================="
echo " sei-ratelimiter Full Test Harness"
echo "====================================="
echo ""

echo "==> [1/5] Starting Docker Compose stack..."
docker-compose up -d --build

echo "==> [2/5] Waiting for services to be healthy..."
sleep 8

curl -sf http://localhost:8080/health > /dev/null || {
  echo "FAIL: app1 not healthy"
  docker-compose down
  exit 1
}

curl -sf http://localhost:8081/health > /dev/null || {
  echo "FAIL: app2 not healthy"
  docker-compose down
  exit 1
}

echo "Both nodes healthy."

echo "==> [3/5] Running unit tests..."
go test -race -v -count=3 ./internal/algorithms/... || {
  docker-compose down
  exit 1
}

echo "==> [4/5] Running Redis integration tests..."
REDIS_URL=localhost:6379 go test -race -v -count=1 \
  ./internal/store/... \
  ./internal/algorithms/... \
  ./tests/integration/... || {
  docker-compose down
  exit 1
}

echo "==> [5/5] Generating coverage report..."
go test -race -coverprofile=coverage.out ./internal/...
go tool cover -func=coverage.out | grep total

echo ""
echo "====================================="
echo " ALL TESTS PASSED"
echo "====================================="

docker-compose down
