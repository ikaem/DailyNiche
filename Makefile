.PHONY: help api fetcher fetcher-dry seed db-reset test_api build build-api build-fetcher web-dev clean

help:
	@echo "DailyNiche - available commands:"
	@echo "  make api            Run the API server (go run)"
	@echo "  make fetcher        Run the feed fetcher once, verbose"
	@echo "  make fetcher-dry    Run the feed fetcher once, dry-run (no DB writes)"
	@echo "  make seed           Seed the database with sample feeds/posts (dev only)"
	@echo "  make db-reset       Delete the local database file"
	@echo "  make test_api       Run all API tests"
	@echo "  make build          Build both api and fetcher binaries"
	@echo "  make web-dev        Run the frontend dev server"
	@echo "  make clean          Remove built binaries"

api:
	cd api && go run ./cmd/api

fetcher:
	cd api && go run ./cmd/fetcher -once -verbose

fetcher-dry:
	cd api && go run ./cmd/fetcher -once -verbose -dry-run

seed:
	cd api && go run ./cmd/seed

db-reset:
	rm -f api/dailyniche.db

test_api:
	cd api && go test ./... -v

build: build-api build-fetcher

build-api:
	cd api && go build -o bin/api ./cmd/api

build-fetcher:
	cd api && go build -o bin/fetcher ./cmd/fetcher

web-dev:
	cd web && npm run dev

clean:
	rm -rf api/bin
