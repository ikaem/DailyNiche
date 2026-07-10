# DailyNiche - Implementation Guide

## Project Overview

**DailyNiche** is a personal RSS magazine service that:
- Reads RSS/Atom feeds you provide
- Scans feeds once per day, storing new posts to a database
- Presents a magazine-like UI with posts from those feeds
- Allows browsing previous daily issues (archive)

**Tech Stack:**
- Backend: Go (REST API + CLI feed fetcher)
- Frontend: SvelteKit
- Database: SQLite
- Monorepo structure (go/, web/, docs/)

**Deployment Model:**
- **Development:** Native (`go run`, `npm run dev`) on your local machine
- **Eventually:** Docker + docker-compose on Raspberry Pi
- **Remote Access:** Cloudflare Tunnel (free, no port forwarding needed)
- **Current Focus:** Build the service first (Phases 0-8), deployment second (Phase 9)

**Key Features (MVP):**
- Dashboard to add/remove feeds (OPML or individual URLs)
- Daily snapshot: magazine layout with visual hierarchy
- Above the fold: top 10 posts (2 columns, time-based sizing)
- Below the fold: next articles (4 columns)
- Bottom section: summary list (1 column)
- Archive of previous issues

---

## Monorepo Structure

```
DailyNiche/
├── api/
│   ├── cmd/
│   │   ├── api/          # REST API server
│   │   └── fetcher/      # Daily feed fetcher CLI
│   ├── internal/
│   │   ├── db/           # Database init & schema
│   │   ├── models/       # Data structures
│   │   ├── repos/        # Repository layer (CRUD)
│   │   ├── feeds/        # Feed parsing
│   │   └── handlers/     # HTTP handlers
│   ├── go.mod
│   └── go.sum
├── web/
│   ├── src/
│   │   ├── routes/       # SvelteKit pages
│   │   ├── lib/          # API client & utilities
│   │   └── components/   # Svelte components
│   ├── package.json
│   └── svelte.config.js
├── docs/
│   ├── API.md            # API documentation
│   └── ARCHITECTURE.md   # Architecture details
├── .gitignore
├── README.md
└── CLAUDE.md             # This file
```

---

## Commit Workflow

**How we work on each task:**

### Before Starting a Step
I briefly outline what I'm about to do, including:
1. **Task ID** (e.g., 1.1)
2. **What** - the specific change or component being built
3. **Context** - how this fits into the larger task/phase and why it matters
4. **Where** - which file(s) will be modified

Example: "Task 1.1 (Database Schema): Creating Feed and Post structs in internal/models/models.go. These are the core data structures that the database, repositories, and API will all use. This is foundational for Phase 1."

### During the Work
- I make focused, atomic changes using Edit/Write tools
- One logical change = one commit
- No "everything in one go" - break it into reviewable chunks
- Each change is small enough that you can understand it quickly

### After Changes
I summarize:
1. **What was done** - brief description of the change
2. **Files modified** - which files were touched
3. **The diff** - show you the actual changes (using `git diff` or `git status` output)

### Code Review & Approval
1. You review the changes in your IDE or via the diff I show
2. You approve ("looks good") or request changes ("fix X" or "add Y")
3. If approved, you make the commit:
   ```bash
   git add <files>
   git commit -m "<message>"
   ```
   OR ask me to commit it

### Then We Repeat
Once the commit is made, I start the next step with its context.

### Commit Message Format
- **Prefix:** `feat:` (feature), `fix:` (bug fix), `docs:` (documentation), `refactor:`, `test:`, `chore:`
- **Message:** Clear, active voice, explains the *what* not the *why* (why goes in PR description)
- Examples:
  - `feat: add Feed and Post model structs`
  - `feat: implement feed repository CRUD operations`
  - `test: add unit tests for feed parser`

### Rules
- **No auto-commits** - I never commit without your explicit approval
- **Atomic changes** - each commit is one logical piece
- **Reviewable** - each step should take <5 min to review
- **Context matters** - always explain how it fits into the bigger picture

### Testing (TDD is mandatory)
- **TDD is required** - tests are not optional polish tacked on in Phase 8, they're written as each piece of logic is built. Applies to the frontend too (via Vitest), not just the Go backend - confirmed as a standing policy once the frontend's first real logic (`toPostModel`/`formatDate`) was built.
- **Tests + the code they cover can live in the same commit** - unlike other changes, you don't need to split "write test" and "write implementation" into two atomic commits. One commit covering both is fine, since they're really one logical unit of work.
- Applies to repos, parsers, handlers, frontend mapping/formatting functions - anything with real logic. Skeleton/wiring code (e.g. an empty `main()`, a purely static Svelte component with no logic) doesn't need tests.
- **Structure every test body with given/when/then comments** marking the setup, the action under test, and the assertion. Neither Go nor Vitest have a built-in BDD syntax that replaces this, so these comments are how test intent stays readable at a glance regardless of language.

---

## Task Checklist

### PHASE 0: Project Initialization (~1.5 hours)

- [x] **0.1: Initialize monorepo structure** (30 min)
  - [x] Create folder hierarchy (go/, web/, docs/)
  - [x] Initialize Git repo with .gitignore
  - [x] Create root README.md with project overview
  - [x] Create docs/ARCHITECTURE.md
  - PR: "chore: initialize monorepo structure"

- [x] **0.2: Initialize Go module and project layout** (30 min)
  - [x] Run `go mod init github.com/karlo/dailyniche`
  - [x] Create cmd/api/main.go and cmd/fetcher/main.go skeletons
  - [x] Create internal/models/models.go
  - [x] Verify: `go build ./cmd/api` works
  - PR: "chore: initialize Go project structure"
  - Note: directory renamed from `go/` to `api/` for clarity (purpose-based, not language-based naming)

- [x] **0.3: Initialize SvelteKit project** (30 min)
  - [x] Create web/ with SvelteKit scaffolding
  - [x] `npm install` and `npm run dev` works
  - [x] Configure API_URL env var
  - [x] Basic landing page with "Coming soon"
  - PR: "chore: initialize SvelteKit project"
  - Note: scaffolded via `sv create` (Svelte 5 + runes mode forced on, TypeScript, Prettier, ESLint, Vitest + vitest-browser-svelte + Playwright for component tests). No Tailwind, default `adapter-auto` - both deliberately deferred. `svelte.config.js` doesn't exist separately in this SvelteKit version - adapter/compilerOptions live directly in the `sveltekit()` plugin call inside `vite.config.ts`.
  - Note: `PUBLIC_API_URL` set via `.env` (gitignored) + `.env.example` (committed) - SvelteKit requires the `PUBLIC_` prefix for any env var exposed to client-side code.

- [x] **0.4: Create Makefile for development commands** (15 min)
  - [x] Create Makefile in project root
  - [x] Add targets: `make api`, `make fetcher`, `make fetcher-dry`, `make test_api`, `make build`, `make web-dev`
  - [x] Document common development tasks
  - PR: "chore: add Makefile for development workflow"

- [ ] **0.5: Split Makefile into included per-area files** (DEFERRED - do later, not urgent)
  - **Do NOT build this yet.** The Makefile is still small enough to read in one screen. This is queued for whenever it grows further (e.g. once the `web-*` targets from Task 0.3's follow-up land, plus anything Phase 9 deployment adds).
  - [ ] Create `make/api.mk` - move `api`, `fetcher`, `fetcher-dry`, `seed`, `db-reset`, `test_api`, `build`, `build-api`, `build-fetcher` here
  - [ ] Create `make/web.mk` - move `web-dev` and any new `web-*` targets (`web-lint`, `web-test`, `web-build`, `web-check`, `web-format`) here
  - [ ] Root `Makefile` keeps `help` (manually-curated documentation, not derived - splitting it across files makes the full command list harder to find in one place) and `include make/api.mk` / `include make/web.mk`
  - [ ] Decide where `clean` lives: stays root-level if it should clean both api and web build artifacts, moves into `make/api.mk` if it stays api-only
  - PR: "chore: split Makefile into per-area included files"

---

### PHASE 1: Database & Core Models (~2-3 hours)

- [x] **1.1: Design and implement database schema** (1 hour)
  - [x] Create internal/db/schema.sql with:
    - `feeds` table: id, name, url, disabled_at (soft delete), created_at, updated_at
    - `posts` table: id, feed_id, title, url, content_summary, published_at, fetched_at, guid (unique), created_at
    - Indexes on feed_id, published_at, fetched_at
  - [x] Create internal/models/models.go with Feed and Post structs
  - [x] Document schema in docs/ARCHITECTURE.md
  - PR: "feat: define database schema and models"

- [x] **1.2: Implement database connection and migrations** (1.5 hours)
  - [x] Create internal/db/db.go with Init(), Migrate(), and Open() (Init+Migrate combined)
  - [x] Add `modernc.org/sqlite` to go.mod (pure Go, no CGO)
  - [x] Test: database auto-creates on startup
  - [x] Test: subsequent runs don't error
  - [x] Wire into cmd/api/main.go
  - PR: "feat: implement database initialization and migrations"

- [ ] **1.3: Add lightweight DB migration system** (DEFERRED - see trigger below)
  - **Do NOT build this yet.** `schema.sql` uses `CREATE TABLE IF NOT EXISTS`, which only works for a brand-new database - it silently does nothing to an existing one when columns are added/changed later (already happened once: adding `disabled_at` to `feeds`). Fine for now since no real database exists yet.
  - **Trigger to actually build this:** once the fetcher (3.3) is running regularly and `dailyniche.db` holds real archived issues you don't want to lose. At that point "just delete the db and restart" stops being an acceptable fix for schema changes.
  - [ ] Numbered migration files (e.g. `0001_init.sql`, `0002_add_disabled_at.sql`) instead of one full `schema.sql`
  - [ ] `schema_migrations` table tracking which migrations have run
  - [ ] On startup, apply any migrations not yet recorded, in order
  - [ ] Consider `pressly/goose` or hand-roll (simple enough at this scale)
  - PR: "feat: add lightweight database migration system"

- [ ] **1.4: Add `.env`-based configuration loading** (DEFERRED - see trigger below)
  - **Do NOT build this yet.** Right now `cmd/api/main.go` reads `DB_PATH` via `os.Getenv` and falls back to a hardcoded default (`"dailyniche.db"`) when unset. That default-in-code is a stopgap, not the final approach.
  - **Trigger to actually build this:** before/during Phase 4.1, when `PORT` and CORS origin config get added alongside `DB_PATH` - once there's more than one env var, it's worth loading them properly instead of scattering `os.Getenv` + fallback logic across entry points.
  - [ ] Add a `.env` file (gitignored) for local dev config, loaded at startup (e.g. via `github.com/joho/godotenv`, or hand-rolled)
  - [ ] Remove the in-code default for `DB_PATH` - require it to be set explicitly (fail fast with a clear error if missing, rather than silently defaulting)
  - [ ] Document required env vars in README (`DB_PATH`, later `PORT`, etc.)
  - [ ] Production (Phase 9, Docker/Pi) sets real env vars via `docker-compose`'s `environment:` section, not the `.env` file - `.env` is a local dev convenience only
  - PR: "feat: add .env-based configuration loading"

- [x] **1.5: Minimal HTTP server (learning exercise)** (30-45 min)
  - **Goal:** stand up the simplest possible Go HTTP server using only the standard library (`net/http`) - no router libraries, no middleware - to learn the core primitives before Phase 4 introduces CORS, logging middleware, and multiple resource handlers on top.
  - Pulled forward ahead of Phase 4 so `cmd/api` actually behaves like a server (listens on a port, respondable via `curl`) instead of just running the DB setup and exiting.
  - [x] Add a `/health` handler in `internal/handlers/health_handler.go` returning `{"status":"ok"}`
  - [x] Wire it into `cmd/api/main.go` with `http.HandleFunc` + `http.ListenAndServe` (`PORT` env var, default 8080)
  - [x] Test the handler with `net/http/httptest` (no need to spin up a real server/port for the test)
  - [x] Verify: `curl localhost:8080/health` returns 200 and the JSON body
  - PR: "feat: add minimal HTTP server with health check"
  - Note: uses net/http's global DefaultServeMux for now (fine for one route); Phase 4.1 will switch to an explicit http.NewServeMux()

---

### PHASE 2: Feed Parsing Infrastructure (~2-3 hours)

- [x] **2.1: Implement feed parser** (1.5 hours)
  - [x] Add `github.com/mmcdole/gofeed` to go.mod
  - [x] Create internal/feeds/parser.go:
    - `ParseFeed(url string)` - fetch and parse RSS/Atom/JSON
    - `ExtractItems(feed)` - convert to Post structs
  - [x] Handle missing fields gracefully
  - [x] Write unit tests with sample feeds
  - PR: "feat: implement feed parser with gofeed"
  - TODO (minor, not urgent): `ParseFeed` currently calls `gofeed.NewParser()` fresh on every call. Fine at our scale. If we ever need a custom HTTP client/timeout, or want to reuse connections across many feed fetches (e.g. fetching dozens of feeds in one fetcher run), build one `Parser` once and reuse it instead.
  - TODO: `gofeed`'s default `User-Agent` is the literal string `"Gofeed/1.0"` - some sites (confirmed live: a WordPress site running a security plugin) return `403 Forbidden` specifically for this signature, while a browser/curl/even bare Go UA all pass. Fix: set `parser.UserAgent` to something more neutral or honestly self-identifying (e.g. `"DailyNiche/1.0 (personal RSS reader)"`) before calling `ParseURL` - `gofeed.Parser` already exposes this field, no new dependency needed. Combine with the `Parser`-reuse TODO above when addressed, since both need the same "build one configured `Parser` instead of a fresh default one" change.

- [ ] **2.3: Add post images to the pipeline** (DEFERRED - not urgent)
  - **Do NOT build this yet.** The frontend design mockups (Task 6.0, `docs/design/issue/`) use dummy Unsplash placeholder images for every post. Real images aren't needed until those mockups get wired up to real Svelte components with live data (Phase 7) - until then, dummy images are fine.
  - **Context:** verified live that `gofeed` already parses an image URL into `item.Image.URL` when a feed provides one (confirmed working against a real WordPress feed) - but our code currently discards it entirely. `models.Post` has no image field, `ExtractItems` never reads `item.Image`, and the `posts` table has no `image_url` column. Not every feed has an image though (confirmed varies by feed/platform) - the model must tolerate an empty/missing image gracefully, no fallback HTML-scraping needed for MVP.
  - [ ] Add `image_url TEXT` column to `posts` in `schema.sql` (safe to edit directly still - no real archived data exists yet, per Task 1.3's migration-system trigger note)
  - [ ] Add `ImageURL string` field to `models.Post`
  - [ ] Update `ExtractItems` to populate it from `item.Image.URL` (empty string if the feed doesn't provide one)
  - [ ] Update `post_repo.go`'s `CreatePost`/`scanPosts` SQL to include the new column
  - [ ] Add `image_url` to `PostResponse` in `posts_handler.go`
  - PR: "feat: capture and serve post images"

- [x] **2.2: Create CLI fetcher scaffold** (1 hour)
  - [x] Create cmd/fetcher/main.go with:
    - `-once` flag (run and exit)
    - `-verbose` flag
    - `-dry-run` flag
    - Database initialization
    - Proper exit codes
  - [x] Verify: `go build ./cmd/fetcher` works
  - PR: "feat: scaffold fetcher CLI"

---

### PHASE 3: Repository Layer (Data Access) (~2-3 hours)

- [x] **3.1: Implement feed repository** (1.5 hours)
  - [x] Create internal/repos/feed_repo.go with:
    - CreateFeed, ListFeeds, GetFeed, UpdateFeed, DeleteFeed
  - [x] Use prepared statements (prevent SQL injection)
  - [x] Write unit tests for each operation
  - PR: "feat: implement feed repository CRUD operations"
  - Note: DeleteFeed is a soft delete (sets disabled_at), per the "Feed Deletion is a Soft Delete" note above

- [x] **3.2: Implement post repository** (1.5 hours)
  - [x] Create internal/repos/post_repo.go with:
    - CreatePost (with duplicate detection via GUID)
    - ListPostsByDate (for fetching today's posts)
    - ListPostsByFeed (filter by feed)
    - DeletePostsByDate (for cleanup)
  - [x] Write tests including duplicate detection
  - PR: "feat: implement post repository CRUD operations"
  - Note: CreatePost returns the new row's int64 ID (0 = skipped duplicate) rather than a bool, so callers get the ID for free on success. DeletePostsByDate is for deliberate manual cleanup only, never called by the normal fetch/serve flow.

- [x] **3.3: Integrate fetcher with repositories** (2 hours)
  - [x] Update cmd/fetcher/main.go:
    - Load feeds from DB
    - Fetch each feed using parser
    - Store posts using post repo
    - Log progress/errors
    - Skip invalid feeds, continue
  - [x] Implement dry-run mode
  - [x] Write integration test
  - [x] Test: can call repeatedly without issues
  - PR: "feat: integrate feed fetcher with database"
  - Note: `main()`'s logic was extracted into a testable `run(args []string, dbPath string) int` (returns an exit code instead of calling `os.Exit` directly) as part of this task, per the earlier TODO - this is exactly the point where the real branching logic (per-feed error handling, dry-run, disabled-feed skipping) made a testable `run()` worth it.

---

### PHASE 4: REST API (~2-3 hours)

- [ ] **4.1: Set up HTTP server and middleware** (1.5 hours)
  - [ ] Create cmd/api/main.go with:
    - HTTP server on port 8080 (configurable via env)
    - CORS middleware (allow localhost:5173)
    - Logging middleware
    - Error response formatting
    - Health check: GET /health
  - [ ] Switch from the global `http.HandleFunc`/`DefaultServeMux` (used since Task 1.5) to an explicit `mux := http.NewServeMux()` - needed to cleanly wrap routes with CORS/logging middleware, and to stop relying on shared global state now that there are 5 routes registered (`/health`, `/api/posts`, `GET/POST /api/feeds`, `DELETE /api/feeds/{id}`)
  - [ ] Use standard library net/http
  - [ ] Test: `/health` returns 200 with `{"status":"ok"}`
  - PR: "feat: initialize HTTP API server"

- [x] **4.2: Implement feed management API endpoints** (1.5 hours)
  - [x] Create internal/handlers/feeds_handler.go:
    - GET /api/feeds - list all
    - POST /api/feeds - create (validate URL, name)
    - DELETE /api/feeds/:id - delete
  - [x] Proper status codes (201, 204, 400, 404, plus 409 for duplicate URL)
  - [x] Write httptest tests
  - [x] Wire routes in cmd/api/main.go
  - PR: "feat: implement feed management API endpoints"
  - Note: added `repos.ErrDuplicateURL` sentinel error (CreateFeed translates SQLite's SQLITE_CONSTRAINT_UNIQUE into it) so POST /api/feeds can return 409 for a duplicate URL instead of a generic 500.
  - Note: DELETE /api/feeds/{id} uses Go 1.22's method+path-pattern ServeMux (`r.PathValue`) - no router library needed. GET/POST /api/feeds share a literal path, so both required explicit method prefixes to avoid a duplicate-pattern panic at startup.
  - Note: verified live end-to-end with curl (create, duplicate 409, list, soft-delete 204, delete-nonexistent 404, invalid id 400).

- [x] **4.3: Implement posts API endpoint** (1.5 hours)
  - [x] Create internal/handlers/posts_handler.go:
    - GET /api/posts - query params: date (YYYY-MM-DD, default today), feed_id (optional)
    - Return posts with feed info, sorted by published_at
  - [x] Write tests
  - [x] Wire route in cmd/api/main.go
  - PR: "feat: implement posts API endpoint"
  - Note: verified live end-to-end with a new dev-only `cmd/seed` tool (see `make seed`/`make db-reset` below) - seeds sample feeds/posts across today and yesterday for manual API/frontend testing without needing real RSS feeds.
  - TODO (minor): feed_id filtering currently fetches all of a date's posts then filters in Go, rather than pushing the filter into the SQL query. Fine at our scale; revisit if per-day post volume ever grows enough to matter.

---

### PHASE 5: Cron & Scheduling (~2 hours)

- [ ] **5.1: Optimize fetcher for cron** (1 hour)
  - [ ] Ensure cmd/fetcher runs cleanly in one shot
  - [ ] Document cron setup: `0 3 * * * /path/to/fetcher -once`
  - [ ] Verify: posts get today's date
  - PR: "docs: prepare fetcher for cron scheduling"

- [ ] **5.2: Add error handling and observability** (1 hour)
  - [ ] Add structured logging to fetcher
  - [ ] Log: start/end time, feeds processed, posts added, errors
  - [ ] Write logs to file and stdout
  - [ ] Graceful shutdown on SIGTERM
  - [ ] Update README with cron setup instructions
  - PR: "feat: add logging and error handling to fetcher"

---

### PHASE 6: Frontend - Basic Setup (~2 hours)

- [ ] **6.0: Define frontend pages and full design (ASCII wireframe -> dummy HTML)** (1-2 hours)
  - **Goal:** decide what pages the app has and design each one *fully*, before writing any Svelte code - not just a paragraph description, an actual concrete HTML artifact to build components from.
  - **Trigger for this task existing:** the original plan jumped straight into building specific pages/components (6.2, 7.1-7.4) without ever explicitly listing the full page inventory - and doing so surfaced a real gap: "Archive of previous issues" is a stated MVP feature (see top of this file) with no corresponding implementation task anywhere in Phase 6/7. This task's job is to close that gap deliberately, not rediscover it mid-build.
  - [ ] List every page/route the app needs, e.g.:
    - Today's issue (home, `/`) - the magazine view for today's posts
    - Archive/previous issue view - **currently missing from the plan entirely**; decide the navigation mechanism (date picker? prev/next day arrows? a list of past dates?) - the backend already supports this fully via `GET /api/posts?date=YYYY-MM-DD`, so this is purely a frontend decision
    - Feed management (`/feeds`) - add/list/remove feeds
  - For each page, design in two passes:
    - [ ] **Pass 1 - ASCII wireframe in chat:** a rough text/box sketch of the layout, discussed and agreed on before any code, to align quickly without investing in markup yet
    - [ ] **Pass 2 - dummy static HTML:** once the wireframe is agreed, build a real plain HTML file (plus simple CSS) implementing that design - no Svelte, no framework, just markup/styling standing in for the real page
  - [ ] These dummy HTML files become the concrete base that Phase 6.2/7.x tasks translate into real Svelte components wired to live data - not thrown away once Svelte work starts
  - [ ] Store the dummy HTML files in `docs/design/` (reference artifacts, kept alongside other docs - not part of the runtime `web/` app)
  - [ ] Update this checklist (Phase 6/7) to add any missing tasks the above surfaces - e.g. an explicit "Archive page" task in Phase 7
  - PR: "docs: add frontend page designs (ASCII wireframes + dummy HTML)"

- [ ] **6.1: Create API client library** (1 hour)
  - [ ] Create web/src/lib/api.js:
    - Base URL from env
    - getFeeds(), addFeed(name, url), deleteFeed(id)
    - getPostsByDate(date), getPostsToday()
    - Error handling
  - [ ] Create web/src/lib/stores.js (optional, if using Svelte stores)
  - PR: "feat: create API client library"

- [ ] **6.2: Create main page layout and navigation** (1 hour)
  - [ ] Update web/src/routes/+layout.svelte:
    - Navigation header with logo
    - Links to /posts and /feeds
    - Basic styling
  - [ ] Create web/src/routes/+page.svelte:
    - Fetch posts for today on mount
    - Display loading/error states
    - Render posts as simple list for now
  - PR: "feat: create main page layout and navigation"
  - TODO (not urgent): `src/lib/postModel.ts`'s `formatDate` hardcodes the `hr-HR` locale. Should eventually be derived from the user's actual browser/location settings (e.g. `navigator.language`) rather than a fixed value - fine for a single-user personal project for now.

---

### PHASE 7: Frontend - Magazine Layout (~3 hours)

- [ ] **7.1: Above-the-fold section (top 10, 2 columns)** (1.5 hours)
  - [ ] Create web/src/components/AboveTheFold.svelte:
    - Takes top 10 posts
    - 2-column grid layout
    - Time-based sizing: newer posts larger, older smaller
    - Show: headline, feed name, publish time
    - Click opens article link
  - [ ] Update +page.svelte to use it
  - PR: "feat: implement above-the-fold magazine section"

- [ ] **7.2: Below-the-fold section (4 columns)** (1 hour)
  - [ ] Create web/src/components/BelowTheFold.svelte:
    - Takes posts 11-N
    - 4-column grid, smaller cards
    - Sorted by publish time
    - Clearly differentiated from above-fold
  - [ ] Update +page.svelte to use it
  - PR: "feat: implement below-the-fold section"

- [ ] **7.3: Bottom section (1 column, single-line text)** (45 min)
  - [ ] Create web/src/components/BottomNews.svelte:
    - Next 4 posts as text list
    - Format: "[Feed] > Title" or similar
    - Links work, text truncated if needed
  - [ ] Update +page.svelte to use it
  - PR: "feat: implement bottom news section"

- [ ] **7.4: Feed management UI** (1.5 hours)
  - [ ] Create web/src/components/FeedManager.svelte:
    - List feeds, add form, delete buttons
    - Success/error feedback
  - [ ] Create web/src/routes/feeds/+page.svelte
  - [ ] Connect to API
  - [ ] Test: add, delete, immediate UI update
  - PR: "feat: implement feed management UI"
  - TODO (not urgent): add a "Preview" button alongside "Add Feed", backed by a new `GET /api/feeds/preview?url=...` endpoint - reuses `feeds.ParseFeed`/`feeds.ExtractItems` directly, writes nothing to the DB, just shows the feed's title + a few current posts before the user commits to adding it. Motivated by discovering some feeds can silently fail (e.g. a site blocking `gofeed`'s default User-Agent with a 403) - preview surfaces that immediately instead of the user only finding out after adding a dead feed and waiting for the next fetch.
  - TODO (not urgent): the dashboard design (Task 6.0, `docs/design/dashboard/`) shows disabled feeds in their own section with an "Enable" button - this feature doesn't exist yet. Needs a `repos.EnableFeed` (clears `disabled_at`) and a corresponding API route (e.g. `POST /api/feeds/{id}/enable`), plus wiring the button in `FeedManager.svelte`. The dashboard mockup shows the button visually muted/non-functional until this is built.

---

### PHASE 8: Polish & Optimization (~3 hours)

- [ ] **8.1: Responsive design and mobile optimization** (1.5 hours)
  - [ ] Test on: 375px (mobile), 768px (tablet), 1024px+ (desktop)
  - [ ] Adjust grid columns:
    - Mobile: above-fold 1 col, below-fold 2 cols, bottom 1 col
    - Tablet: above-fold 2 col, below-fold 3 cols, bottom 1 col
    - Desktop: 2, 4, 1 as designed
  - [ ] Ensure touch-friendly (48px+ tap targets)
  - PR: "feat: implement responsive design"

- [ ] **8.2: Performance and caching** (1.5 hours)
  - [ ] API: Add Cache-Control headers
  - [ ] Frontend: Implement client-side caching
  - [ ] Lazy load images, use srcset
  - [ ] Minimize bundle
  - [ ] Lighthouse score >= 80
  - PR: "perf: optimize performance and caching"

- [ ] **8.3: Testing and documentation** (2 hours)
  - [ ] Add unit tests (>70% coverage for internal/)
  - [ ] Add integration tests for API endpoints
  - [ ] Write docs/API.md (endpoints, request/response formats)
  - [ ] Update README with:
    - Setup instructions (backend & frontend)
    - How to add feeds
    - Cron job setup
    - Development workflow
  - [ ] Remove TODO comments from code
  - PR: "docs: complete testing and documentation"

---

## Dependency Graph

```
CRITICAL PATH (do these in order):
0.1 -> 0.2 -> 0.3 -> 0.4 -> 1.1 -> 1.2 -> 2.2 -> 3.1 -> 3.2 -> 3.3 -> 4.1 -> 4.2 -> 4.3

PARALLEL (after 1.1):
2.1 (parser) -> 3.3 (fetcher integration)

THEN FRONTEND:
0.3 -> 6.1 -> 6.2 -> 7.1, 7.2, 7.3 (can parallelize)
              7.4 (parallel with 7.1-7.3)

THEN POLISH:
8.1, 8.2, 8.3 (can parallelize)

OPTIONAL FOR MVP (DEFER):
5.1, 5.2 (just use system cron locally until later)

DEPLOYMENT (PHASE 9 - DO LAST):
Complete Phase 8 -> 9.1 -> 9.2 -> 9.3 -> 9.4 (optional)
(Only tackle this after service is working end-to-end locally)
```

**Next task:** 4.1 (CORS/logging middleware, extending the minimal server from 1.5) or Phase 5 (cron/observability polish for the fetcher). Phases 0-4 are all done - full feed CRUD plus the daily posts feed all work end-to-end over HTTP, verified live with curl. 0.3 (SvelteKit) was intentionally skipped for now in favor of a backend-first vertical slice; revisit whenever ready to build the UI.

---

## Key Implementation Notes

### Database Schema
- **GUID field** in posts: prevents duplicates (RSS feeds may republish)
- **fetched_at**: tracks when post was discovered, separate from published_at
- **Simple indexes**: on feed_id, published_at, fetched_at for query performance

### Layout Logic
- **Above the fold:** Top 10 posts, 2 columns
  - **Time-based sizing:** posts from last 24h larger; older posts smaller proportionally
  - Posts sorted by feed discovery order, not by feed source
- **Below the fold:** Remaining posts, 4 columns, sorted by published_at
- **Bottom:** 4 items, single-line text summary

### Feed Deletion is a Soft Delete (via `disabled_at`)
- Feeds are never hard-deleted. `feeds.disabled_at` is NULL for active feeds; deleting a feed via the dashboard sets `disabled_at` to the current date instead of removing the row.
- Posts must survive feed removal so past issues never change (see schema.sql - no ON DELETE CASCADE on posts.feed_id, and the feed row itself is preserved).
- The fetcher (Phase 3.3) must skip feeds where `disabled_at` is set when pulling new posts, but past issues keep resolving `feed_id` -> feed name normally since the row still exists.
- `DeleteFeed` (Phase 3/4) should be implemented as `UPDATE feeds SET disabled_at = ? WHERE id = ?`, not a SQL `DELETE`.

### Timestamps & Timezones
- **Always store and compare times in UTC.** When parsing a feed's published date, immediately convert with `.UTC()` before storing (e.g. `parsedTime.UTC()`).
- Reason: Go's `time.Now()` defaults to local machine time, and SQLite has no native timestamp type - mixing local/UTC times causes subtle sorting/comparison bugs. Normalizing to UTC everywhere sidesteps this entirely.
- This app only needs relative recency (sort by newest, "last 24h" sizing) - we don't need to preserve each feed's original timezone for display.

### Go Dependencies (add as needed)
- Phase 1: `modernc.org/sqlite` (pure Go, no CGO - simpler cross-compilation for the Pi later)
- Phase 2: `github.com/mmcdole/gofeed`
- Standard library for rest (net/http, encoding/json, log, time, database/sql)

### Development Workflow
1. **Backend first** (Phases 0-5): build locally, test with curl/Postman
2. **Frontend second** (Phases 6-7): connect to running API
3. **Polish last** (Phase 8): optimize and refine
4. **Each PR = one task**, clear scope, mergeable in 2-3 hour session

### Deployment Model (Deferred to Phase 9)

**Development Phase (Phases 0-8):**
- Run locally: `go run ./cmd/api` and `npm run dev`
- SQLite database file: `dailyniche.db` in repo root (local)
- No Docker, no deployment overhead - pure development focus

**Deployment Phase (Phase 9 - do this last):**
- **Target:** Raspberry Pi running 24/7
- **Access:** Cloudflare Tunnel (free, secure, no port forwarding)
- **Containerization:** Docker + docker-compose on Pi only
- **Database:** SQLite on Pi (single machine), daily backups optional
- **Automation:** Cron job for daily feed fetching

**Why defer deployment?**
- Focus on building a working service first
- Deployment decisions don't affect core architecture
- Docker can be added without code changes (just Dockerfile + compose)
- Cloudflare Tunnel is independent of your code

**When to move to Phase 9:**
- Service is complete and working locally (Phase 8 done)
- You're ready to run it 24/7 on Pi
- You want remote access via a permanent URL

---

### PHASE 9: Deployment to Raspberry Pi (~2-3 hours) [DEFER UNTIL LATER]

This phase is optional and should only be done after Phase 8 is complete. Focus on building the service first.

- [ ] **9.1: Create Dockerfile and docker-compose.yml** (1 hour)
  - [ ] Create `Dockerfile` for Go API service:
    - Multi-stage build (build in one image, run in smaller image)
    - Copy binary and assets
    - Expose port 8080
  - [ ] Create `docker-compose.yml`:
    - Service for API (golang image or built binary)
    - Volume mount for SQLite database (persistence)
    - Environment variables for config (API_URL, etc.)
    - Port mapping (8080 -> 8080)
  - [ ] Test locally: `docker-compose up` and verify API works
  - [ ] Document build/run process
  - PR: "feat: add Docker support for deployment"

- [ ] **9.2: Set up Raspberry Pi and Cloudflare Tunnel** (1.5 hours)
  - [ ] Install Docker and docker-compose on Pi
  - [ ] Clone repo to Pi (or use scp to copy files)
  - [ ] Install Cloudflare Tunnel client (`cloudflared`) on Pi
  - [ ] Create Cloudflare account (free) and configure tunnel:
    - Authenticate: `cloudflared tunnel login`
    - Create tunnel: `cloudflared tunnel create dailyniche`
    - Route traffic to localhost:8080
    - Assign domain: `https://dailyniche.yourdomain.com` (or cloudflare subdomain)
  - [ ] Test: Access from another device via the public URL
  - PR: "docs: document Raspberry Pi deployment and Cloudflare Tunnel setup"

- [ ] **9.3: Set up daily cron job on Pi** (1 hour)
  - [ ] SSH into Pi
  - [ ] Create cron entry to run fetcher daily:
    ```bash
    0 3 * * * cd /path/to/DailyNiche && api/cmd/fetcher -once -verbose
    ```
    Or if using Docker:
    ```bash
    0 3 * * * docker-compose -f /path/to/docker-compose.yml exec api /app/fetcher -once -verbose
    ```
  - [ ] Test: manually run the command, verify posts are fetched
  - [ ] Set up log rotation (optional, logs don't grow too fast for personal use)
  - [ ] Document cron setup in README
  - PR: "docs: document cron job setup for Pi"

- [ ] **9.4: Backup strategy** (optional, 30 min)
  - [ ] Decide: daily backup of SQLite to cloud storage (e.g., rsync to a backup server, or tar + upload)
  - [ ] Create backup script
  - [ ] Add cron entry to run backup daily
  - [ ] Document backup procedure
  - PR: "docs: add backup strategy" (optional)

---

## Running Locally (Development)

### Backend
```bash
cd api
go run ./cmd/api          # Start API server (port 8080)
go run ./cmd/fetcher -once -verbose  # Run feed fetcher once
go run ./cmd/seed         # Seed sample feeds/posts for manual testing (dev only)
```

Or via Makefile from the repo root: `make api`, `make fetcher`, `make seed`, `make db-reset`.

### Frontend
```bash
cd web
npm install
npm run dev               # Start dev server (port 5173)
```

### Database
- SQLite database file: `dailyniche.db` (auto-created)
- Reset: delete `dailyniche.db` and restart API server

---

## Completed Tasks

(Mark off as you complete each PR)

- [x] 0.1: Initialize monorepo structure
- [x] 0.2: Initialize Go module and project layout
- [x] 0.4: Create Makefile for development commands
- [x] 1.1: Design and implement database schema
- [x] 1.2: Implement database connection and migrations
- [x] 1.5: Minimal HTTP server (learning exercise) - /health route, real server via ListenAndServe
- [x] 2.1: Implement feed parser - ParseFeed + ExtractItems with GUID/content/date fallbacks
- [x] 2.2: Create CLI fetcher scaffold - flags, DB init, exit codes
- [x] 3.1: Implement feed repository - CreateFeed, ListFeeds, GetFeed, UpdateFeed, DeleteFeed (soft delete)
- [x] 3.2: Implement post repository - CreatePost (GUID dedup), ListPostsByDate, ListPostsByFeed, DeletePostsByDate
- [x] 3.3: Integrate fetcher with repositories - real fetch loop, dry-run, per-feed error isolation, disabled-feed skipping, run() extracted for testability
- [x] 4.3: Implement posts API endpoint - GET /api/posts (date + feed_id filters), feed name enrichment, cmd/seed dev tool for manual verification
- [x] 4.2: Implement feed management API endpoints - GET/POST/DELETE /api/feeds, ErrDuplicateURL sentinel (409), Go 1.22 method+path routing, soft delete