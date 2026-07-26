# Build stage - compiles both the API server and fetcher CLI as static
# binaries. CGO_ENABLED=0 works because modernc.org/sqlite is pure Go, no
# CGO needed (see CLAUDE.md's Go Dependencies notes) - this is what lets the
# final image be a minimal, non-glibc-dependent base.
FROM golang:1.26-alpine AS builder

WORKDIR /build

# go.mod/go.sum copied first so `go mod download` is cached in its own layer,
# only re-run when dependencies actually change, not on every source edit.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/fetcher ./cmd/fetcher

# Runtime stage - small final image. Needs ca-certificates for the
# fetcher's outbound HTTPS requests to RSS feeds - TLS verification would
# fail without them on a truly bare-minimum base.
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /out/api /app/api
COPY --from=builder /out/fetcher /app/fetcher

EXPOSE 8080

# fetcher isn't run here - it's invoked on demand via
# `docker compose exec api /app/fetcher` (cron on the Pi's host, per Task 9.3).
ENTRYPOINT ["/app/api"]
