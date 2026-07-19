package main

import (
	"flag"
	"log/slog"
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

// configureLogging installs a slog default logger writing to out. This is
// what -verbose now controls: Debug level surfaces FetchAll's per-feed/
// per-post detail, Info level shows only the start/completion summary lines
// - replacing the old approach of threading a Verbose flag through Options
// and manually gating individual log calls with it.
func configureLogging(out *os.File, verbose bool) {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{Level: level})))
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
		slog.Error("failed to parse flags", "error", err)
		return 2
	}

	configureLogging(os.Stdout, cfg.Verbose)

	conn, err := db.Open(dbPath)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		return 1
	}
	defer conn.Close()

	slog.Debug("database ready", "db_path", dbPath)

	summary, err := fetcher.FetchAll(conn, fetcher.Options{DryRun: cfg.DryRun})
	if err != nil {
		slog.Error("failed to fetch feeds", "error", err)
		return 1
	}

	if summary.Errors > 0 {
		slog.Warn("fetcher completed with errors", "errors", summary.Errors)
	}
	return 0
}

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "dailyniche.db"
	}
	os.Exit(run(os.Args[1:], dbPath))
}
