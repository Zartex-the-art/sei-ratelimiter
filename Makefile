.PHONY: build vet test test-race test-integration test-full test-factory phase2 coverage up down

build:
	go build ./...

vet:
	go vet ./...

test:
	go test -race -v -count=1 ./...

test-race:
	go test -race -v -count=3 ./...

test-integration:
	REDIS_URL=localhost:6379 go test -race -v -count=1 ./...

test-factory:
	REDIS_URL=localhost:6379 go test -race -v -count=1 -run 'TestNewLimiter|TestFactory' ./internal/algorithms/...

test-full:
	./scripts/run_tests.sh

phase2:
	./scripts/run_tests.sh

coverage:
	go test -race -coverprofile=coverage.out ./internal/...
	go tool cover -func=coverage.out | grep total

up:
	docker-compose up -d --build

down:
	docker-compose down
