package main

import (
	"flag"
	"log"
	"os"

	"github.com/karlo/dailyniche/internal/db"
	"github.com/karlo/dailyniche/internal/fetcher"
)

// Config holds the fetcher's command-line options.
type Config struct {
	Verbose bool
	DryRun  bool
}

// parseFlags parses args (typically os.Args[1:]) into a Config.
func parseFlags(args []string) (Config, error) {
	fs := flag.NewFlagSet("fetcher", flag.ContinueOnError)
	verbose := fs.Bool("verbose", false, "enable verbose logging")
	dryRun := fs.Bool("dry-run", false, "parse and log without writing to the database")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	return Config{Verbose: *verbose, DryRun: *dryRun}, nil
}

// run executes one fetcher invocation and returns a process exit code.
// Kept separate from main() so it's callable from tests: main() calls
// os.Exit directly, which would kill the test binary if main() itself were
// invoked from a test. The actual fetch loop lives in internal/fetcher, so
// it can also be called from the API's on-demand fetch endpoint - this is
// just the CLI wrapper around it (flags, db lifecycle, exit codes).
func run(args []string, dbPath string) int {
	cfg, err := parseFlags(args)
	if err != nil {
		log.Printf("failed to parse flags: %v", err)
		return 2
	}

	if cfg.Verbose {
		log.Printf("starting fetcher (dry-run=%v)", cfg.DryRun)
	}

	conn, err := db.Open(dbPath)
	if err != nil {
		log.Printf("failed to open database: %v", err)
		return 1
	}
	defer conn.Close()

	log.Printf("fetcher: database ready at %s", dbPath)

	summary, err := fetcher.FetchAll(conn, fetcher.Options{Verbose: cfg.Verbose, DryRun: cfg.DryRun})
	if err != nil {
		log.Printf("failed to fetch feeds: %v", err)
		return 1
	}

	log.Printf("fetcher done: %d new, %d duplicates, %d errors", summary.New, summary.Duplicates, summary.Errors)
	return 0
}

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "dailyniche.db"
	}
	os.Exit(run(os.Args[1:], dbPath))
}
