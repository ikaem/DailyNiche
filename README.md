# DailyNiche

A personal RSS magazine service that transforms your favorite feeds into a daily magazine-like interface.

## What it does

- **Reads your RSS feeds** - add feeds via dashboard or OPML import
- **Daily snapshots** - scans feeds once per day, creates a "magazine issue" with new posts
- **Magazine layout** - beautiful visual hierarchy: featured articles above the fold, supporting articles below
- **Browse archives** - go back and read previous days' issues

## Why DailyNiche?

Most RSS readers overwhelm you with infinite feeds. DailyNiche gives you a curated daily snapshot—one coherent "issue" per day. Perfect for staying connected to small blogs and the indie web without the noise.

## Quick Start (Development)

### Prerequisites
- Go 1.22+
- Node.js 20+ and npm
- Git

### Setup

```bash
# Clone or enter the repo
cd DailyNiche

# Backend
cd api
go mod download

# Frontend
cd ../web
npm install
```

### Run Locally

**Terminal 1 - API server:**
```bash
cd api
go run ./cmd/api
```

**Terminal 2 - Feed fetcher (one-shot):**
```bash
cd api
go run ./cmd/fetcher -verbose
```

**Terminal 3 - Frontend dev server:**
```bash
cd web
npm run dev
```

Visit `http://localhost:5173` in your browser.

### Automating the Fetcher (Cron)

The fetcher is a one-shot command - something else has to invoke it on a schedule. Locally, that's your system's `cron`. (This is separate from the Raspberry Pi deployment's cron setup in [CLAUDE.md - Phase 9](./CLAUDE.md), which invokes the fetcher inside a Docker container instead.)

Add an entry via `crontab -e` to run the fetcher daily, e.g. at 3am:

```
0 3 * * * cd /path/to/DailyNiche/api && go run ./cmd/fetcher -verbose
```

Or, against a built binary (see `make build`, which outputs to `api/bin/`):

```
0 3 * * * /path/to/DailyNiche/api/bin/fetcher -verbose
```

You can also trigger a fetch on demand from the dashboard's "Fetch now" button, without waiting for the next scheduled run.

**On the Raspberry Pi (Docker deployment):** cron runs on the Pi's host OS, not inside the container - it invokes the fetcher binary inside the already-running `api` container via `docker compose exec`:

```
0 3 * * * cd /srv/dailyniche && /usr/bin/docker compose exec -T api /app/fetcher -verbose >> /srv/dailyniche/cron-fetch.log 2>&1
```

`-T` disables the pseudo-TTY `docker compose exec` allocates by default - needed here since cron has no terminal to attach one to. `/usr/bin/docker` is spelled out in full because cron jobs run with a minimal `PATH` that may not include it. `3am` here is the Pi's local time (`Europe/Zagreb`), not UTC - this only controls when the daily fetch runs, unrelated to the app's own UTC-everywhere data convention.

### Fetcher Logging

Every run logs structured (`key=value`) output to both stdout and a log file, so a cron-triggered run's history is still available later even though nothing is watching its stdout live. The log file's path comes from the `LOG_PATH` env var, defaulting to `fetcher.log` in the working directory the fetcher was started from:

```
LOG_PATH=/path/to/DailyNiche/api/fetcher.log
```

Pass `-verbose` for `Debug`-level detail (per-feed fetch attempts, dry-run notices); without it, only the run's start/completion summary and any warnings/errors are logged.

If the fetcher receives `SIGTERM` or `SIGINT` (e.g. a system shutdown, or a manually cancelled run), it stops cleanly before starting its next feed rather than being killed mid-fetch, logs a warning noting the early stop, and exits with code 130.

**The API server logs the same way**, for the same reason: an on-demand fetch triggered from the dashboard's "Fetch now" button runs inside the `api` server process itself, not the separate `fetcher` binary, so without this its output would only ever reach the container's ephemeral stdout - lost the next time the container gets recreated (i.e. every deploy). Its own `LOG_PATH` (default `api.log`) captures this, plus the server's regular startup/request logging as a side effect of sharing the same underlying logger. In the Docker deployment, `LOG_PATH` is set to `/data/api.log` - inside the same already-persisted volume as the SQLite database, not the container's throwaway filesystem.

## Project Structure

See [CLAUDE.md](./CLAUDE.md) for the full implementation guide and task breakdown.

- `api/` - Backend: REST API and feed fetcher CLI
- `web/` - Frontend: SvelteKit magazine UI
- `docs/` - Architecture and API documentation

## Features (MVP)

- Dashboard to add/remove feeds
- Daily magazine layout with visual hierarchy
- Archive of previous issues
- Responsive design (mobile, tablet, desktop)

## Deployment

Eventually runs on a Raspberry Pi with Cloudflare Tunnel for remote access. See [CLAUDE.md - Phase 9](./CLAUDE.md) for deployment setup.

For now, development is local only.

## License

Personal project.
