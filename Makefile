.PHONY: test test-race test-smoke test-integration test-e2e test-full test-db-up test-db-down verify

TEST_DATABASE_DSN ?= postgres://postgres:postgres@localhost:55432/redbutton_test?sslmode=disable

test:
	go test ./...

test-race:
	go test -race ./...

test-smoke:
	go test -tags=smoke ./internal/bot

test-integration:
	TEST_DATABASE_DSN='$(TEST_DATABASE_DSN)' go test -count=1 -tags=integration ./internal/repository/postgres

test-e2e:
	TEST_DATABASE_DSN='$(TEST_DATABASE_DSN)' go test -count=1 -tags=e2e ./tests/e2e

test-db-up:
	docker compose -f docker-compose.test.yaml up -d --wait

test-db-down:
	docker compose -f docker-compose.test.yaml down -v

verify:
	go vet ./...
	go test -race -coverprofile=coverage.out ./...
	go test -tags=smoke ./internal/bot

test-full:
	./scripts/test-full.sh
