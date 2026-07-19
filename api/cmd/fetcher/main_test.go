package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/karlo/dailyniche/internal/db"
	"github.com/karlo/dailyniche/internal/repos"
)

// testDBFileName is used with t.TempDir() to build a real on-disk database
// path for these tests, rather than ":memory:". run() opens its own
// connection to dbPath internally, separate from whatever connection a test
// used to seed data beforehand - an in-memory database is private to the
// single connection that opened it, so a seed connection and run()'s own
// connection would each see a different, empty database. A real file on
// disk is the only way multiple separate connections actually share data,
// which is also exactly how the fetcher and API server share data for real.
const testDBFileName = "fetcher-test.db"

// testLogFileName is used with t.TempDir() the same way, so each test gets
// its own isolated log file rather than appending to a real fetcher.log.
const testLogFileName = "fetcher-test.log"

const sampleRSS = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
<channel>
  <title>Sample Blog</title>
  <link>https://example.com</link>
  <description>A sample feed for tests</description>
  <item>
    <title>First Post</title>
    <link>https://example.com/first-post</link>
    <guid>urn:uuid:first-post</guid>
    <description>The first post summary.</description>
    <pubDate>Mon, 01 Jan 2024 10:00:00 GMT</pubDate>
  </item>
  <item>
    <title>Second Post</title>
    <link>https://example.com/second-post</link>
    <guid>urn:uuid:second-post</guid>
    <description>The second post summary.</description>
    <pubDate>Tue, 02 Jan 2024 10:00:00 GMT</pubDate>
  </item>
</channel>
</rss>`

// newSampleFeedServer starts a local server serving sampleRSS.
func newSampleFeedServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(sampleRSS))
	}))
}

// countPosts returns how many rows exist in the posts table.
func countPosts(t *testing.T, dbPath string) int {
	t.Helper()
	conn, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to reopen database: %v", err)
	}
	defer conn.Close()

	var count int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM posts`).Scan(&count); err != nil {
		t.Fatalf("failed to count posts: %v", err)
	}
	return count
}

func TestParseFlags_DefaultsToAllFalse(t *testing.T) {
	// given: no flags
	// when: we parse an empty argument list
	cfg, err := parseFlags([]string{})
	if err != nil {
		t.Fatalf("parseFlags() returned error: %v", err)
	}

	// then: every option defaults to false
	if cfg.Verbose || cfg.DryRun {
		t.Errorf("expected all-false defaults, got %+v", cfg)
	}
}

func TestParseFlags_SetsFieldsFromFlags(t *testing.T) {
	// given: both flags passed
	// when: we parse them
	cfg, err := parseFlags([]string{"-verbose", "-dry-run"})
	if err != nil {
		t.Fatalf("parseFlags() returned error: %v", err)
	}

	// then: each corresponding field is true
	if !cfg.Verbose {
		t.Error("expected Verbose to be true")
	}
	if !cfg.DryRun {
		t.Error("expected DryRun to be true")
	}
}

func TestParseFlags_ReturnsErrorForUnknownFlag(t *testing.T) {
	// given: an unrecognized flag
	// when: we parse it
	_, err := parseFlags([]string{"-bogus"})

	// then: it returns an error
	if err == nil {
		t.Fatal("expected an error for an unknown flag, got nil")
	}
}

// TestRun_FetchesAndStoresPostsEndToEnd is deliberately the only test here
// exercising the actual fetch behavior. Dry-run, dedup, disabled feeds, and
// dead-feed handling are now covered exhaustively by internal/fetcher's own
// tests, which run() delegates to via fetcher.FetchAll - re-testing those
// same scenarios again through this thicker CLI wrapper would just be
// duplicate coverage of identical logic. This one test's job is narrower:
// prove run() itself - flag parsing, opening the db, calling FetchAll,
// logging to stdout and the log file, returning an exit code - is wired
// together correctly end-to-end.
func TestRun_FetchesAndStoresPostsEndToEnd(t *testing.T) {
	// given: a feed pointing at a local test server, seeded into a temp db
	server := newSampleFeedServer()
	defer server.Close()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, testDBFileName)
	logPath := filepath.Join(dir, testLogFileName)
	seedConn, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open seed database: %v", err)
	}
	if _, err := repos.CreateFeed(seedConn, "Test Feed", server.URL); err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	seedConn.Close()

	// when: we run the fetcher
	code := run(context.Background(), []string{}, dbPath, logPath)

	// then: it exits 0 and stores both posts from the sample feed
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if count := countPosts(t, dbPath); count != 2 {
		t.Fatalf("expected 2 posts stored, got %d", count)
	}

	// and: the run's log output was written to the log file, not just stdout
	logContents, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}
	if len(logContents) == 0 {
		t.Error("expected the log file to contain output, got an empty file")
	}
}

// TestRun_ReturnsInterruptedExitCodeWhenContextIsCancelled covers the branch
// a real SIGTERM can't safely exercise in a test: mapping a context.Canceled
// error from FetchAll to interruptedExitCode rather than the generic
// failure exit code 1. A pre-cancelled context deterministically triggers
// that same error without needing to send an actual OS signal.
func TestRun_ReturnsInterruptedExitCodeWhenContextIsCancelled(t *testing.T) {
	// given: a feed seeded into a temp db, but a context that's already cancelled
	dir := t.TempDir()
	dbPath := filepath.Join(dir, testDBFileName)
	logPath := filepath.Join(dir, testLogFileName)
	seedConn, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open seed database: %v", err)
	}
	if _, err := repos.CreateFeed(seedConn, "Test Feed", "https://example.com/feed.xml"); err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	seedConn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// when: we run the fetcher with the cancelled context
	code := run(ctx, []string{}, dbPath, logPath)

	// then: it reports the interrupted exit code, not the generic failure one
	if code != interruptedExitCode {
		t.Fatalf("expected exit code %d, got %d", interruptedExitCode, code)
	}
}
