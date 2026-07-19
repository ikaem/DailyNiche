package fetcher

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/karlo/dailyniche/internal/feeds"
	"github.com/karlo/dailyniche/internal/repos"
)

// Options controls how FetchAll behaves.
type Options struct {
	Verbose bool
	DryRun  bool
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
func FetchAll(conn *sql.DB, opts Options) (Summary, error) {
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

		if opts.Verbose {
			log.Printf("fetching feed %q (%s)", feed.Name, feed.URL)
		}

		parsed, err := feeds.ParseFeed(feed.URL)
		if err != nil {
			log.Printf("skipping feed %q: %v", feed.Name, err)
			summary.Errors++
			continue
		}

		for _, post := range feeds.ExtractItems(parsed, feed.ID) {
			if opts.DryRun {
				log.Printf("[dry-run] would store post %q (%s)", post.Title, post.GUID)
				continue
			}

			id, err := repos.CreatePost(conn, &post)
			if err != nil {
				log.Printf("failed to store post %q: %v", post.Title, err)
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

	return summary, nil
}
