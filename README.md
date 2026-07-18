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
