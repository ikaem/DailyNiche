package fetcher

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/karlo/dailyniche/internal/feeds"
	"github.com/karlo/dailyniche/internal/repos"
)

// Options controls how FetchAll behaves.
type Options struct {
	DryRun bool
}

// Summary reports what happened during a FetchAll run.
type Summary struct {
	FeedsProcessed int
	New            int
	Duplicates     int
	Errors         int
}

// FetchAll fetches every enabled feed in conn, storing new posts. A
// per-feed or per-post failure is logged and counted in Summary.Errors but
// never aborts the run - one broken feed must not prevent every other feed
// from being fetched. Callable from both the CLI (cmd/fetcher) and the API
// (an on-demand fetch endpoint) - the same loop, reused from two entry
// points rather than duplicated or shelled out to as a separate process.
//
// Logging goes through the global slog default logger rather than an
// injected instance, matching this codebase's existing plain-global-logger
// convention (see middleware.Logging's use of the standard log package).
// Verbosity is therefore not an Options field - slog's own level filtering
// replaces the old manual "if opts.Verbose" gating; whoever configures the
// default logger's minimum level (cmd/fetcher/main.go, from its -verbose
// flag) controls whether the Debug-level lines below are shown.
func FetchAll(conn *sql.DB, opts Options) (Summary, error) {
	start := time.Now()
	slog.Info("fetch started", "dry_run", opts.DryRun)

	feedList, err := repos.ListFeeds(conn)
	if err != nil {
		return Summary{}, fmt.Errorf("failed to list feeds: %w", err)
	}

	var summary Summary
	for _, feed := range feedList {
		if feed.DisabledAt != nil {
			continue
		}
		summary.FeedsProcessed++

		slog.Debug("fetching feed", "feed_name", feed.Name, "feed_url", feed.URL)

		parsed, err := feeds.ParseFeed(feed.URL)
		if err != nil {
			slog.Warn("skipping feed", "feed_name", feed.Name, "feed_url", feed.URL, "error", err)
			summary.Errors++
			continue
		}

		for _, post := range feeds.ExtractItems(parsed, feed.ID) {
			if opts.DryRun {
				slog.Debug("dry-run: would store post", "title", post.Title, "guid", post.GUID)
				continue
			}

			id, err := repos.CreatePost(conn, &post)
			if err != nil {
				slog.Warn("failed to store post", "title", post.Title, "error", err)
				summary.Errors++
				continue
			}
			if id > 0 {
				summary.New++
			} else {
				summary.Duplicates++
			}
		}
	}

	slog.Info("fetch completed",
		"duration", time.Since(start),
		"feeds_processed", summary.FeedsProcessed,
		"new", summary.New,
		"duplicates", summary.Duplicates,
		"errors", summary.Errors,
	)

	return summary, nil
}
