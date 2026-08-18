.PHONY: help dev api fetcher fetcher-dry seed db-reset test_api build build-api build-fetcher web-dev clean docker-build-api docker-run-api docker-logs-api docker-fetcher-api docker-fetcher-dry-api docker-stop-api docker-clean-api docker-build-web docker-run-web docker-logs-web docker-stop-web docker-up docker-down docker-logs docker-clean deploy

help:
	@echo "DailyNiche - available commands:"
	@echo "  make dev              Run the API server and frontend dev server together"
	@echo "  make api              Run the API server (go run)"
	@echo "  make fetcher          Run the feed fetcher once, verbose"
	@echo "  make fetcher-dry      Run the feed fetcher once, dry-run (no DB writes)"
	@echo "  make seed             Seed the database with sample feeds/posts (dev only)"
	@echo "  make db-reset         Delete the local database file"
	@echo "  make test_api         Run all API tests"
	@echo "  make build            Build both api and fetcher binaries"
	@echo "  make web-dev          Run the frontend dev server"
	@echo "  make clean            Remove built binaries"
	@echo "  make docker-build-api Build the api Docker image (tag: dailyniche-api:test)"
	@echo "  make docker-run-api   Run the built image locally, port 8080, persisted test volume"
	@echo "  make docker-logs-api  Follow the running test container's logs"
	@echo "  make docker-fetcher-api      Run the fetcher inside the running test container, verbose"
	@echo "  make docker-fetcher-dry-api  Same, but dry-run (no DB writes)"
	@echo "  make docker-stop-api  Stop and remove the test container (keeps the test volume)"
	@echo "  make docker-clean-api Remove the test volume too (full reset)"
	@echo "  make docker-build-web Build the web Docker image (tag: dailyniche-web:test)"
	@echo "  make docker-run-web   Run the built image locally, port 3000"
	@echo "  make docker-logs-web  Follow the running test container's logs"
	@echo "  make docker-stop-web  Stop and remove the test container"
	@echo "  make docker-up        Build and start the full stack (api + web) via docker-compose.yml"
	@echo "  make docker-down      Stop the full stack (keeps the database volume)"
	@echo "  make docker-logs      Follow logs for both services"
	@echo "  make docker-clean     Stop the full stack and remove the database volume too"
	@echo "  make deploy           Deploy to the Pi: git pull, rebuild, prune old images + stale build cache"

# -j2 runs both targets concurrently in one make invocation - no extra
# process-manager dependency (e.g. concurrently/foreman) needed for just two
# processes. Ctrl+C stops both.
dev:
	@$(MAKE) -j2 api web-dev

api:
	cd api && go run ./cmd/api

fetcher:
	cd api && go run ./cmd/fetcher -verbose

fetcher-dry:
	cd api && go run ./cmd/fetcher -verbose -dry-run

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

# Docker (local image build/test, before deploying anything to the Pi) -
# these build and run the exact same Dockerfile that'll eventually run on
# the Pi, so problems surface here first rather than after a slower,
# Pi-side rebuild-and-redeploy cycle.

docker-build-api:
	cd api && docker build -t dailyniche-api:test .

# Runs the built image with a persistent named volume (survives container
# restarts/recreates, unlike an anonymous volume) and DB_PATH pointed at it -
# matching how the real deployment will mount a volume for the SQLite file.
docker-run-api:
	docker run -d --name dailyniche-api-test -p 8080:8080 \
		-e DB_PATH=/data/dailyniche.db \
		-v dailyniche-test-data:/data \
		dailyniche-api:test

docker-logs-api:
	docker logs -f dailyniche-api-test

# Runs the fetcher binary inside the already-running api container via
# docker exec - this is exactly how the real cron job on the Pi invokes it
# too (see CLAUDE.md's Task 9.3), so this is what actually verifies that
# invocation path works, not just that the binary compiles.
docker-fetcher-api:
	docker exec dailyniche-api-test /app/fetcher -verbose

docker-fetcher-dry-api:
	docker exec dailyniche-api-test /app/fetcher -verbose -dry-run

# Stops and removes the test container, but leaves the named volume (and
# its data) in place - use docker-clean-api for a full wipe.
docker-stop-api:
	docker stop dailyniche-api-test && docker rm dailyniche-api-test

# Full reset: also removes the persisted test database, so the next
# docker-run-api starts from a completely empty state.
docker-clean-api:
	docker volume rm dailyniche-test-data

docker-build-web:
	cd web && docker build -t dailyniche-web:test .

# Standalone smoke test only - API_URL points nowhere reachable, since a
# plain `docker run` container can't resolve "localhost" as the host or any
# other separately-run container by default (confirmed empirically - this
# is the same networking gap docker-compose.yml solves automatically via a
# shared network + service-name DNS). Expect pages to render correctly but
# show a handled "fetch failed" state (CLAUDE.md's Task 6.2 error-as-data
# pattern) - that's success for this target's actual purpose: proving the
# server itself starts and serves requests, not a real API integration test.
docker-run-web:
	docker run -d --name dailyniche-web-test -p 3000:3000 \
		-e API_URL=http://localhost:8080 \
		dailyniche-web:test

docker-logs-web:
	docker logs -f dailyniche-web-test

docker-stop-web:
	docker stop dailyniche-web-test && docker rm dailyniche-web-test

# Whole-stack commands using the root docker-compose.yml (api + web
# together, networked automatically) - distinct from the docker-*-api/
# docker-*-web targets above, which test one service in isolation via plain
# `docker run`.

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

# Also removes the named volume (dailyniche-data) - a full reset, unlike
# docker-down, which leaves it in place.
docker-clean:
	docker compose down -v

# Deploys to the Pi over the ssh alias (see ~/.ssh/config, routed through
# Cloudflare Tunnel). /srv/dailyniche is a real git clone of this repo's
# GitHub remote, not an rsync destination - only committed and pushed
# changes can be deployed, deliberately. `image prune -f` (not `-a`) clears
# only dangling/untagged images left behind by `--build`'s rebuild, never
# anything still tagged - see docs/PI_SETUP.md's Task 9.1 notes for why.
# `builder prune --filter until=168h` clears BuildKit's layer cache, but
# only entries older than 7 days - unlike image prune, nothing else touches
# this cache, so it grows forever otherwise (confirmed live: 2.8GB
# reclaimable after just a handful of deploys). The 7-day filter is safe to
# run on every deploy since it never removes cache from a build that just
# ran moments ago - only stale entries a future build wouldn't reuse anyway.
deploy:
	ssh kaempi5 "cd /srv/dailyniche && git pull && docker compose up -d --build && docker image prune -f && docker builder prune -f --filter until=168h"
