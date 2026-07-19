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

### Fetcher Logging

Every run logs structured (`key=value`) output to both stdout and a log file, so a cron-triggered run's history is still available later even though nothing is watching its stdout live. The log file's path comes from the `LOG_PATH` env var, defaulting to `fetcher.log` in the working directory the fetcher was started from:

```
LOG_PATH=/path/to/DailyNiche/api/fetcher.log
```

Pass `-verbose` for `Debug`-level detail (per-feed fetch attempts, dry-run notices); without it, only the run's start/completion summary and any warnings/errors are logged.

If the fetcher receives `SIGTERM` or `SIGINT` (e.g. a system shutdown, or a manually cancelled run), it stops cleanly before starting its next feed rather than being killed mid-fetch, logs a warning noting the early stop, and exits with code 130.

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
