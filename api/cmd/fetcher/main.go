package main

import (
	"flag"
	"fmt"
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
	fmt.Println("DailyNiche Feed Fetcher")
}
