package main

import (
	"flag"
	"log"
	"os"

	"github.com/karlo/dailyniche/internal/db"
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

func main() {
	cfg, err := parseFlags(os.Args[1:])
	if err != nil {
		log.Printf("failed to parse flags: %v", err)
		os.Exit(2)
	}

	if cfg.Verbose {
		log.Printf("starting fetcher (once=%v, dry-run=%v)", cfg.Once, cfg.DryRun)
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "dailyniche.db"
	}

	conn, err := db.Open(dbPath)
	if err != nil {
		log.Printf("failed to open database: %v", err)
		os.Exit(1)
	}
	defer conn.Close()

	log.Printf("fetcher: database ready at %s", dbPath)
	// Feed fetching wired in Task 3.3, once feed/post repos exist.
}
