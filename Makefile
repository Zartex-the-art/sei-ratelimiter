.PHONY: test test-integration test-full load-smoke

test:
	go test -race -v -count=1 ./...

test-integration:
	REDIS_URL=localhost:6379 go test -race -v -count=1 ./...

test-full:
	./scripts/run_tests.sh

load-smoke:
	k6 run tests/load/smoke.js
