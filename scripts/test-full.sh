#!/bin/sh
set -eu

project_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$project_root"

unformatted=$(gofmt -l .)
if [ -n "$unformatted" ]; then
  echo "These Go files are not formatted:"
  echo "$unformatted"
  exit 1
fi

go vet ./...
go test -race -coverprofile=coverage.out ./...
go test -count=1 -tags=smoke ./internal/bot

docker compose -f docker-compose.test.yaml up -d --wait
cleanup() {
  docker compose -f docker-compose.test.yaml down -v
}
trap cleanup EXIT INT TERM

export TEST_DATABASE_DSN="${TEST_DATABASE_DSN:-postgres://postgres:postgres@localhost:55432/redbutton_test?sslmode=disable}"
go test -count=1 -tags=integration ./internal/repository/postgres
go test -count=1 -tags=e2e ./tests/e2e
docker build -t redbutton-bot:test .
image_user=$(docker image inspect redbutton-bot:test --format '{{.Config.User}}')
if [ "$image_user" != "redbutton" ]; then
  echo "Production image user is $image_user, want redbutton"
  exit 1
fi
