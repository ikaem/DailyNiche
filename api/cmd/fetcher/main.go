package main

import (
	"context"
	"errors"
	"flag"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

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

// configureLogging installs a slog default logger writing to out - typically
// io.MultiWriter(os.Stdout, logFile), so a run's output lands in both places
// at once rather than only wherever stdout happens to be redirected. Level
// is what -verbose now controls: Debug surfaces FetchAll's per-feed/per-post
// detail, Info shows only the start/completion summary lines - replacing the
// old approach of threading a Verbose flag through Options and manually
// gating individual log calls with it.
func configureLogging(out io.Writer, verbose bool) {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{Level: level})))
}

// interruptedExitCode is returned when a run stops early due to a shutdown
// signal rather than a real failure - 130 matches the conventional Unix
// exit code for a SIGINT-terminated process (128+2); reused here for
// SIGTERM too rather than picking a second code, since both signals mean
// the same thing to this CLI: stop cleanly, don't treat it as an error.
//
// Only observable when running the compiled binary directly. `go run`
// wraps the built binary in its own process and does not forward this exit
// code faithfully - confirmed live, a SIGTERM sent to a `go run` process
// surfaced as exit code 143 (the default, unhandled SIGTERM code) instead.
const interruptedExitCode = 130

// run executes one fetcher invocation and returns a process exit code.
// Kept separate from main() so it's callable from tests: main() calls
// os.Exit directly, which would kill the test binary if main() itself were
// invoked from a test. The actual fetch loop lives in internal/fetcher, so
// it can also be called from the API's on-demand fetch endpoint - this is
// just the CLI wrapper around it (flags, db lifecycle, exit codes).
func run(ctx context.Context, args []string, dbPath string, logPath string) int {
	cfg, err := parseFlags(args)
	if err != nil {
		slog.Error("failed to parse flags", "error", err)
		return 2
	}

	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		slog.Error("failed to open log file", "error", err)
		return 1
	}
	defer logFile.Close()

	configureLogging(io.MultiWriter(os.Stdout, logFile), cfg.Verbose)

	conn, err := db.Open(dbPath)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		return 1
	}
	defer conn.Close()

	slog.Debug("database ready", "db_path", dbPath)

	summary, err := fetcher.FetchAll(ctx, conn, fetcher.Options{DryRun: cfg.DryRun})
	if err != nil {
		if errors.Is(err, context.Canceled) {
			slog.Warn("fetcher stopped early due to shutdown signal")
			return interruptedExitCode
		}
		slog.Error("failed to fetch feeds", "error", err)
		return 1
	}

	if summary.Errors > 0 {
		slog.Warn("fetcher completed with errors", "errors", summary.Errors)
	}
	return 0
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "dailyniche.db"
	}
	logPath := os.Getenv("LOG_PATH")
	if logPath == "" {
		logPath = "fetcher.log"
	}

	code := run(ctx, os.Args[1:], dbPath, logPath)
	stop()
	os.Exit(code)
}
