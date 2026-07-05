# DailyNiche Architecture

## High-Level Overview

DailyNiche is a three-tier system:

1. **Backend (Go)** - REST API server + CLI feed fetcher
2. **Frontend (SvelteKit)** - Magazine UI
3. **Database (SQLite)** - Persistent storage

## Data Flow

```
RSS Feeds
   ↓
Fetcher (Go CLI)
   ↓ (parses, deduplicates)
SQLite Database
   ↓
API Server (Go)
   ↓ (serves data by date)
Frontend (SvelteKit)
   ↓
User (browser)
```

### Daily Workflow

1. **Cron job** runs once per day (e.g., 3 AM)
2. **Fetcher** connects to all subscribed feeds
3. **Fetcher** extracts new posts since last run
4. **Fetcher** stores posts in database with `fetched_at = today`
5. **User** opens browser → frontend queries API
6. **API** returns posts for that date
7. **Frontend** renders magazine layout

## Technology Choices

| Component | Tech | Why |
|-----------|------|-----|
| Backend API | Go + net/http | Lightweight, fast, single binary, good for learning |
| Feed Fetcher | Go CLI | Same language as API, efficient, easy to schedule |
| Frontend | SvelteKit | Modern, good DX, reactive, supports both client & server |
| Database | SQLite | Zero setup, single file, perfect for personal projects |
| Deployment | Docker + Compose | Portable, Pi-friendly, declarative |
| Remote Access | Cloudflare Tunnel | Free, no port forwarding, secure |

## Database Schema

### `feeds` table
- `id` (INTEGER PRIMARY KEY)
- `name` (TEXT) - user-facing feed title
- `url` (TEXT UNIQUE) - RSS/Atom feed URL
- `created_at` (TIMESTAMP)
- `updated_at` (TIMESTAMP)

### `posts` table
- `id` (INTEGER PRIMARY KEY)
- `feed_id` (INTEGER FOREIGN KEY) - references feeds.id
- `title` (TEXT) - post headline
- `url` (TEXT) - link to full article
- `content_summary` (TEXT) - excerpt or summary
- `published_at` (TIMESTAMP) - when post was published by feed
- `fetched_at` (TIMESTAMP) - when we discovered this post (used for daily snapshots)
- `guid` (TEXT UNIQUE) - feed-provided GUID, prevents duplicates
- `created_at` (TIMESTAMP)

**Indexes:**
- `posts(feed_id)` - filter posts by feed
- `posts(published_at)` - sort posts by publish date
- `posts(fetched_at)` - find posts from a specific day

## Directory Structure

```
DailyNiche/
├── api/
│   ├── cmd/
│   │   ├── api/
│   │   │   └── main.go        # REST API server entry point
│   │   └── fetcher/
│   │       └── main.go        # CLI feed fetcher entry point
│   ├── internal/
│   │   ├── db/
│   │   │   ├── db.go          # Database connection and migrations
│   │   │   └── schema.sql     # SQL schema definitions
│   │   ├── models/
│   │   │   └── models.go      # Feed and Post structs
│   │   ├── repos/
│   │   │   ├── feed_repo.go   # Feed CRUD operations
│   │   │   └── post_repo.go   # Post CRUD operations
│   │   ├── feeds/
│   │   │   └── parser.go      # RSS/Atom feed parsing
│   │   └── handlers/
│   │       ├── feeds_handler.go  # Feed API endpoints
│   │       └── posts_handler.go  # Posts API endpoints
│   ├── go.mod
│   └── go.sum
├── web/
│   ├── src/
│   │   ├── routes/
│   │   │   ├── +layout.svelte    # App shell, navigation
│   │   │   ├── +page.svelte      # Home: daily magazine view
│   │   │   └── feeds/
│   │   │       └── +page.svelte  # Feed management page
│   │   ├── lib/
│   │   │   └── api.js            # API client library
│   │   └── components/
│   │       ├── AboveTheFold.svelte     # Top 10 articles (2 cols)
│   │       ├── BelowTheFold.svelte     # Next articles (4 cols)
│   │       ├── BottomNews.svelte       # Text list (1 col)
│   │       └── FeedManager.svelte      # Add/remove feeds
│   ├── package.json
│   └── svelte.config.js
├── docs/
│   ├── ARCHITECTURE.md   # This file
│   └── API.md            # API endpoint documentation
├── Dockerfile            # (Phase 9) Container image
├── docker-compose.yml    # (Phase 9) Orchestration
├── .gitignore
├── README.md
└── CLAUDE.md             # Project implementation guide
```

## API Endpoints (Overview)

See `docs/API.md` for full specification.

- `GET /health` - health check
- `GET /api/feeds` - list all feeds
- `POST /api/feeds` - add feed
- `DELETE /api/feeds/:id` - remove feed
- `GET /api/posts?date=YYYY-MM-DD&feed_id=N` - get posts for a date

## Development Workflow

1. **Backend first** (Phases 0-5) - build API locally, test with curl
2. **Frontend second** (Phases 6-7) - connect to running API
3. **Polish** (Phase 8) - optimize, test, document
4. **Deployment** (Phase 9) - dockerize for Pi + Cloudflare Tunnel

See [CLAUDE.md](../CLAUDE.md) for detailed task breakdown.

## Key Design Decisions

### Why Go for the backend?
- Fast, efficient, minimal dependencies
- Single binary deployment (easy on Pi)
- Good for learning while shipping
- Perfect for a simple REST API

### Why SQLite?
- Zero setup, single file
- Sufficient for personal scale
- Easy backups (just copy the file)
- Can migrate to Postgres later if needed

### Why daily snapshots instead of infinite feeds?
- Reduces decision fatigue ("what should I read?")
- Creates coherent "issues" like a magazine
- Aligns with cron job (daily fetch)
- Better UX for small blog aggregation

### Why separate fetcher CLI?
- Decoupled from API server (cleaner architecture)
- Easy to schedule with cron (no daemon complexity)
- Can run independently for testing
- API doesn't block on network timeouts

## Performance Considerations

- **Caching:** API responses have Cache-Control headers (static per day)
- **Database:** Simple queries with indexes on frequently filtered columns
- **Frontend:** SvelteKit handles code splitting and lazy loading
- **Scalability:** SQLite sufficient for personal use; Postgres migration path if needed

## Monitoring & Observability

**Development:**
- Console logs from fetcher
- Browser dev tools for frontend

**Production (Phase 9):**
- Fetcher logs to stdout + file
- Can monitor via Cloudflare dashboard
- Simple health check endpoint

## Future Enhancements (Out of Scope for MVP)

- User accounts and multi-user support
- Feed categorization/folders
- Search and filtering
- Mobile app (backend already REST-ready)
- Feed-specific settings (update frequency, disable)
- Post starring/bookmarking
- OPML export
