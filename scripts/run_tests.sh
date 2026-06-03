#!/usr/bin/env bash
set -e

echo '==> Starting Docker Compose stack...'
docker compose up -d --build

echo '==> Waiting for services to be healthy...'
sleep 5

echo '==> Running unit tests...'
go test -race -v -count=1 ./internal/algorithms/... || { docker compose down; exit 1; }

echo '==> Running integration tests...'
REDIS_URL=localhost:6379 go test -race -v -count=1 ./internal/store/... || { docker compose down; exit 1; }

echo '==> All tests passed.'
docker compose down
