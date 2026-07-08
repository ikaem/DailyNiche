package main

import (
	"flag"
	"log"
	"os"

	"github.com/karlo/dailyniche/internal/db"
	"github.com/karlo/dailyniche/internal/feeds"
	"github.com/karlo/dailyniche/internal/repos"
)

// Config holds the fetcher's command-line options.
type Config struct {
	Once    bool
	Verbose bool
	DryRun  bool
}

// parseFlags parses args (typically os.Args[1:]) into a Config.
func parseFlags(args []string) (Config, error) {
	fs := flag.NewFlagSet("fetcher", flag.ContinueOnError)
	once := fs.Bool("once", false, "run once and exit")
	verbose := fs.Bool("verbose", false, "enable verbose logging")
	dryRun := fs.Bool("dry-run", false, "parse and log without writing to the database")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	return Config{Once: *once, Verbose: *verbose, DryRun: *dryRun}, nil
}

// run executes one fetcher invocation and returns a process exit code.
// Kept separate from main() so it's callable from tests: main() calls
// os.Exit directly, which would kill the test binary if main() itself were
// invoked from a test.
func run(args []string, dbPath string) int {
	cfg, err := parseFlags(args)
	if err != nil {
		log.Printf("failed to parse flags: %v", err)
		return 2
	}

	if cfg.Verbose {
		log.Printf("starting fetcher (once=%v, dry-run=%v)", cfg.Once, cfg.DryRun)
	}

	conn, err := db.Open(dbPath)
	if err != nil {
		log.Printf("failed to open database: %v", err)
		return 1
	}
	defer conn.Close()

	log.Printf("fetcher: database ready at %s", dbPath)

	feedList, err := repos.ListFeeds(conn)
	if err != nil {
		log.Printf("failed to list feeds: %v", err)
		return 1
	}

	// newCount: posts actually inserted (CreatePost returned a positive ID)
	var newCount int
	// dupCount: posts already stored on a previous run, skipped via the
	// GUID uniqueness/dedup logic (CreatePost returned 0, no error)
	var dupCount int
	// errCount: feeds that failed to parse, or posts that failed to store -
	// each is logged individually and counted here, but doesn't abort the run
	var errCount int
	for _, feed := range feedList {
		if feed.DisabledAt != nil {
			continue
		}

		if cfg.Verbose {
			log.Printf("fetching feed %q (%s)", feed.Name, feed.URL)
		}

		parsed, err := feeds.ParseFeed(feed.URL)
		if err != nil {
			log.Printf("skipping feed %q: %v", feed.Name, err)
			errCount++
			continue
		}

		for _, post := range feeds.ExtractItems(parsed, feed.ID) {
			if cfg.DryRun {
				log.Printf("[dry-run] would store post %q (%s)", post.Title, post.GUID)
				continue
			}

			id, err := repos.CreatePost(conn, &post)
			if err != nil {
				log.Printf("failed to store post %q: %v", post.Title, err)
				errCount++
				continue
			}
			if id > 0 {
				newCount++
			} else {
				dupCount++
			}
		}
	}

	log.Printf("fetcher done: %d new, %d duplicates, %d errors", newCount, dupCount, errCount)
	return 0
}

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "dailyniche.db"
	}
	os.Exit(run(os.Args[1:], dbPath))
}
