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

### Atomic Development

This is the concrete practice that makes atomic commits possible - not just a description of the commits after the fact, but a rule for how I write the code in the first place:

- **Write code so that an atomic commit is possible.** Before starting a step, shape the work as one self-contained, reviewable unit - don't build ahead into the next logical change (e.g. don't write `api.ts` before you've reviewed and committed the type it depends on, even if I already know I'll need it next).
- **Each unit pairs a small, self-contained component with its TDD test, if the code has real logic.** Per the testing rule below, one commit can hold both the implementation and its test - but it should not hold multiple unrelated components, or a component plus the start of the next one.
- **Then, and only then, hand it off for review.** You review the code (in your IDE or via the diff I show).
- **You make the actual commit.** I never commit without your explicit approval (see Rules below) - this is what keeps each review small enough to actually review.

If a task naturally decomposes into several small pieces (e.g. a type, then the client that uses it, then the test), each piece gets its own hand-off and its own commit - I don't pre-build the next piece while waiting for review on the current one.

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
- **Atomic Development** - see above: code must be *written* in self-contained units in the first place, not just described that way after the fact
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

- [x] **1.3: Add lightweight DB migration system** (built as a deliberate learning/engineering exercise, per the second justification below - not because the original data-loss trigger had been hit)
  - Hand-rolled, not `pressly/goose` (matches the "learning exercise" framing, and simple enough at this scale per the original note).
  - `api/internal/db/migrations/*.sql` - numbered files (`0001_init.sql`, `0002_add_image_url.sql`), embedded via `//go:embed migrations/*.sql` into an `embed.FS`.
  - `migration.go`: `parseMigrationFilename` + `loadMigrations(fsys fs.FS)` parse and sort by version; deliberately takes an `fs.FS` parameter (not hardcoded to the real embedded files) so tests exercise sorting/error-handling via fabricated `testing/fstest.MapFS` filesystems instead of needing real files per scenario.
  - `ensureSchemaMigrationsTable`/`appliedVersions`/`applyMigration` - the `schema_migrations` table (version, name, applied_at) tracks what's run; `applyMigration` runs a migration's SQL and records it in one transaction, so a failure partway through can't apply the SQL without recording it (or vice versa) - verified via a test that feeds it deliberately-invalid SQL and confirms nothing was recorded.
  - `Migrate()` orchestrates all of the above: ensure the table, load migrations, diff against applied versions, apply what's pending, in order. Safe to call repeatedly (existing `Init`/`Migrate`/`Open` tests all still pass unchanged through this new path).
  - Built across 5 separate commits by deliberate request (relocate schema -> pure parsing logic -> bookkeeping/apply logic -> wire into `Migrate()` -> delete the now-dead `schema.sql`), even though intermediate commits didn't leave the feature fully working - prioritized commit-by-commit reviewability over each commit being independently functional.
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

- [x] **2.3: Add post images to the pipeline**
  - `image_url TEXT` added to `posts` via `migrations/0002_add_image_url.sql` (the migration system's first real second migration, per Task 1.3).
  - `models.Post.ImageURL`; `feeds.ExtractItems` populates it via a new `imageURL(item)` helper reading `item.Image.URL` (confirmed via gofeed's translator source: for plain RSS this comes from an `<enclosure type="image/...">` tag, not a channel-level `<image>`), empty string if the feed provides none.
  - `post_repo.go`'s `CreatePost`/`scanPosts` plumb the column through; scanned into a plain `string` (not a nullable wrapper), consistent with how `ContentSummary` is already handled, since `CreatePost` always writes an explicit value.
  - `PostResponse.image_url` in `posts_handler.go`, with `imageURLOrPlaceholder` substituting a constant when empty - decided (2026-07-13) the fallback belongs in the API layer, not per-client. The placeholder is a self-contained inline SVG data URI (`data:image/svg+xml;base64,...`, a generic mountain/sun "no image" glyph built from plain shapes) - no static-file route or externally-hosted dependency needed; rendered and visually confirmed correct, not just checked as well-formed XML.
  - `web/src/lib/server/api.ts`'s `PostWire`/`toPost` read the real `image_url` instead of hardcoding `''` - the now-resolved TODO comment removed. No frontend component changes needed - `PostHero`/`PostMedium`/`PostListItem` already rendered `post.imageUrl`.
  - Verified live end-to-end (real images rendering instead of broken-image icons).
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
  - **CORS middleware may no longer be needed at all** (re-evaluate before building this item specifically): the frontend's Task 6.1 architecture pivot means the browser never calls the Go API directly - all reads go through SvelteKit `load` functions and all writes through form `actions`, both server-to-server. CORS only matters for browser-originated cross-origin requests, so as long as that BFF pattern holds for every future frontend feature too, this specific bullet can likely be dropped from this task rather than implemented. Logging middleware, error formatting, and the `http.NewServeMux()` switch are unaffected and still worth doing.
  - [ ] Create cmd/api/main.go with:
    - HTTP server on port 8080 (configurable via env)
    - ~~CORS middleware (allow localhost:5173)~~ - see note above, likely unnecessary now
    - Logging middleware
    - Error response formatting
    - Health check: GET /health
  - [ ] Switch from the global `http.HandleFunc`/`DefaultServeMux` (used since Task 1.5) to an explicit `mux := http.NewServeMux()` - needed to cleanly wrap routes with logging middleware, and to stop relying on shared global state now that there are 5 routes registered (`/health`, `/api/posts`, `GET/POST /api/feeds`, `DELETE /api/feeds/{id}`)
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

- [x] **6.0: Define frontend pages and full design (ASCII wireframe -> dummy HTML)** (1-2 hours)
  - Dummy static HTML mockups live in `docs/design/issue/` (issue-v1 through v6, v6 chosen) and `docs/design/dashboard/` (dashboard-v1), plus a showcase `docs/design/index.html`. Archive/previous-issue navigation resolved as a date-nav pill bar (prev/next arrows + native date input), not a separate route - same `/` page renders whichever day's issue is loaded.
  - Note: "Feed management" ended up at `/dashboard` (not `/feeds` as originally sketched) - deliberately generic naming in case it grows beyond just feed CRUD later.

- [x] **6.1: Create API client library** (1 hour)
  - Lives at `web/src/lib/server/api.ts`, not `web/src/lib/api.js` - see the architecture note below for why. Implements `getFeeds`, `addFeed`, `deleteFeed`, `getPostsByDate`, `getPostsToday`, plus an `ApiError` class (carries HTTP status) for error handling. Wire-shape (snake_case) to domain-type (camelCase) mapping lives here too (`PostWire`/`FeedWire` + `toPost`/`toFeed`), mirroring the Go API's own `models.Post`/`PostResponse` DTO split.
  - No `stores.js` - not needed once data loading moved to SvelteKit's own `load`/`data` mechanism (see below).
  - **Architecture note (decided 2026-07-13):** originally planned as a client-side module called from `onMount`, matching the base SvelteKit env-var setup from Task 0.3 (`PUBLIC_API_URL`). Pivoted mid-build to a BFF (backend-for-frontend) pattern instead: the module moved under `web/src/lib/server/` (SvelteKit enforces this can never be imported into client-run code - a build-time guarantee, not just discipline), and the env var was renamed `PUBLIC_API_URL` -> `API_URL` (now `$env/static/private`). Reasoning: the browser should never talk to the Go API directly - all reads go through `+page.server.ts` `load` functions, all writes through form `actions` - so CORS becomes a non-issue for this app entirely (see Task 4.1's note).

- [x] **6.2: Create main page layout and navigation** (1 hour)
  - `+layout.svelte` renders a shared `Header.svelte` (masthead + Home/Dashboard nav with active-link highlighting via `$app/state`'s `page` + `$app/paths`'s `resolve()`).
  - `+page.svelte` does NOT fetch on mount - per the architecture pivot above, `web/src/routes/+page.server.ts`'s `load` function fetches server-side and returns `{ posts, error }`; the component just consumes `data` via `$props()`. No loading-state UI needed (SSR means data's already there); a `data.error` branch replaces the "render posts as simple list" fallback.
  - TODO (not urgent): `src/lib/postModel.ts`'s `formatDate` hardcodes the `hr-HR` locale. Should eventually be derived from the user's actual browser/location settings (e.g. `navigator.language`) rather than a fixed value - fine for a single-user personal project for now.

---

### PHASE 7: Frontend - Magazine Layout (~3 hours)

- [x] **7.1 + 7.2: Above/below-the-fold sections** (merged - the chosen design (issue-v6) uses a 3-tier hierarchy, not the original literal "top 10 in 2 cols, then 4 cols" split)
  - `PostHero.svelte` - the day's top 2 posts, full-bleed image with title/description overlaid (image-overlay treatment), `grid-column: span 6` (2 per row on desktop)
  - `PostMedium.svelte` - next 4 posts, plain card (image on top, content below), `grid-column: span 3` (4 per row on desktop)
  - `PostListItem.svelte` - everything after the top 6, single-column text row with a small thumbnail
  - `AboveTheFold.svelte` composes the first 2 as `PostHero` + next 4 as `PostMedium`; `BelowTheFold.svelte` composes the rest as `PostListItem` under an "Also today" label (hidden entirely if there are 6 or fewer posts total)
  - Both render with **no wrapping element** - they're fragments rendered directly into the parent's `.grid-12` container (owned by `+page.svelte`), since `PostHero`/`PostMedium`'s `grid-column: span N` requires them to be direct grid children
  - `DateNav.svelte` - the prev/next/date-jump pill bar, extracted from `+page.svelte` as its own component
  - All in `web/src/lib/components/` (not `web/src/components/` as originally sketched, to match SvelteKit's `$lib` convention)

- [x] **7.3: Bottom section** - superseded by `PostListItem.svelte`/`BelowTheFold.svelte` above (no separate `BottomNews.svelte` - one list-item component serves this role)

- [x] **7.4: Feed management UI** (1.5 hours)
  - No separate `FeedManager.svelte` - the dashboard's list/add-form/delete-buttons markup lives directly in `web/src/routes/dashboard/+page.svelte` (small enough as a single page; revisit extraction if it grows).
  - `web/src/routes/dashboard/+page.server.ts`: `load` lists feeds (same error-as-data pattern as the home page); `actions.addFeed`/`actions.deleteFeed` handle mutations - both validate defensively (required fields, well-formed URL, numeric id) before ever calling the Go API, then translate `ApiError` into `fail(status, {message})`. Neither action returns data on success - `use:enhance`'s default behavior (verified against `@sveltejs/kit`'s own source) already resets the form and calls `invalidateAll()` for any successful result.
  - Verified live end-to-end against the real Go API: add, delete (soft-delete moves a feed into "Disabled Feeds"), and validation-failure (form NOT cleared, per `enhance`'s failure-path behavior) all confirmed via screenshots.
  - TODO (not urgent): add a "Preview" button alongside "Add Feed", backed by a new `GET /api/feeds/preview?url=...` endpoint - reuses `feeds.ParseFeed`/`feeds.ExtractItems` directly, writes nothing to the DB, just shows the feed's title + a few current posts before the user commits to adding it. Motivated by discovering some feeds can silently fail (e.g. a site blocking `gofeed`'s default User-Agent with a 403) - preview surfaces that immediately instead of the user only finding out after adding a dead feed and waiting for the next fetch.
  - TODO (not urgent): the dashboard shows disabled feeds in their own section with an "Enable" button, rendered but inert (`<button type="button">`, no action wired) - this feature doesn't exist on the backend yet. Needs a `repos.EnableFeed` (clears `disabled_at`) and a corresponding form action, plus wiring the button in `dashboard/+page.svelte`.

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

**Next task:** Phases 0-4 (backend), 6, and 7 (frontend) are all done, plus Task 1.3 (DB migration system) and Task 2.3 (post images) - the full vertical slice works end-to-end with real images rendering, verified live in a browser. Remaining open items, in no particular required order: 4.1's logging middleware/error formatting (CORS itself is likely unnecessary now - see its note), Phase 5 (cron/observability polish for the fetcher), Task 8.x (polish - responsive/perf/testing docs), and the Task 7.4 TODOs (feed preview, enable-disabled-feed).

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
- [x] 0.3: Initialize SvelteKit project - scaffolded via `sv create`, Svelte 5 runes mode, TypeScript, Vitest
- [x] 6.0: Define frontend pages and full design - dummy HTML mockups in docs/design/ (issue-v6, dashboard-v1 chosen)
- [x] 6.1: Create API client library - web/src/lib/server/api.ts, pivoted to server-only (BFF pattern) mid-build - see the architecture note under 6.1 above
- [x] 6.2: Create main page layout and navigation - Header.svelte, +page.server.ts load (not onMount)
- [x] 7.1 + 7.2: Above/below-the-fold sections - PostHero, PostMedium, PostListItem, AboveTheFold, BelowTheFold, DateNav
- [x] 7.3: Bottom section - superseded by PostListItem/BelowTheFold
- [x] 7.4: Feed management UI - dashboard/+page.svelte + +page.server.ts (load + addFeed/deleteFeed actions), verified live end-to-end
- [x] 1.3: Add lightweight DB migration system - hand-rolled, numbered migrations/*.sql embedded via embed.FS, schema_migrations tracking table, built as a learning exercise ahead of its original data-loss trigger
- [x] 2.3: Add post images to the pipeline - image_url column/field/parsing/repo/API plumbing, server-side inline-SVG placeholder for posts with no image, verified live end-to-end