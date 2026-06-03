.PHONY: test test-integration test-full load-smoke test-tokenbucket test-all-algorithms

test:
	go test -race -v -count=1 ./...

test-integration:
	REDIS_URL=localhost:6379 go test -race -v -count=1 ./...

test-full:
	./scripts/run_tests.sh

load-smoke:
	k6 run tests/load/smoke.js
test-tokenbucket:
	REDIS_URL=localhost:6379 go test -race -v -count=1 \
	-run TestTokenBucket ./internal/algorithms/...

test-all-algorithms:
	REDIS_URL=localhost:6379 go test -race -v -count=1 \
	-run "TestAllAlgorithms|TestTokenBucket" \
	./internal/algorithms/... ./tests/integration/...
