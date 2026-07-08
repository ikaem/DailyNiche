package main

import (
	"net/http"
	"net/http/httptest"
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
	if cfg.Once || cfg.Verbose || cfg.DryRun {
		t.Errorf("expected all-false defaults, got %+v", cfg)
	}
}

func TestParseFlags_SetsFieldsFromFlags(t *testing.T) {
	// given: all three flags passed
	// when: we parse them
	cfg, err := parseFlags([]string{"-once", "-verbose", "-dry-run"})
	if err != nil {
		t.Fatalf("parseFlags() returned error: %v", err)
	}

	// then: each corresponding field is true
	if !cfg.Once {
		t.Error("expected Once to be true")
	}
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

func TestRun_FetchesAndStoresNewPosts(t *testing.T) {
	// given: a feed pointing at a local test server, seeded into a temp db
	server := newSampleFeedServer()
	defer server.Close()

	dbPath := filepath.Join(t.TempDir(), testDBFileName)
	seedConn, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open seed database: %v", err)
	}
	if _, err := repos.CreateFeed(seedConn, "Test Feed", server.URL); err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	seedConn.Close()

	// when: we run the fetcher once
	code := run([]string{"-once"}, dbPath)

	// then: it exits 0 and stores both posts from the sample feed
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if count := countPosts(t, dbPath); count != 2 {
		t.Fatalf("expected 2 posts stored, got %d", count)
	}
}

func TestRun_CanBeCalledRepeatedlyWithoutDuplicating(t *testing.T) {
	// given: a feed seeded into a temp db
	server := newSampleFeedServer()
	defer server.Close()

	dbPath := filepath.Join(t.TempDir(), testDBFileName)
	seedConn, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open seed database: %v", err)
	}
	if _, err := repos.CreateFeed(seedConn, "Test Feed", server.URL); err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	seedConn.Close()

	// when: we run the fetcher twice against the same feed
	if code := run([]string{"-once"}, dbPath); code != 0 {
		t.Fatalf("expected first run to exit 0, got %d", code)
	}
	if code := run([]string{"-once"}, dbPath); code != 0 {
		t.Fatalf("expected second run to exit 0, got %d", code)
	}

	// then: posts are not duplicated
	if count := countPosts(t, dbPath); count != 2 {
		t.Fatalf("expected still 2 posts after a second run, got %d", count)
	}
}

func TestRun_DryRunDoesNotStorePosts(t *testing.T) {
	// given: a feed seeded into a temp db
	server := newSampleFeedServer()
	defer server.Close()

	dbPath := filepath.Join(t.TempDir(), testDBFileName)
	seedConn, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open seed database: %v", err)
	}
	if _, err := repos.CreateFeed(seedConn, "Test Feed", server.URL); err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	seedConn.Close()

	// when: we run the fetcher with -dry-run
	code := run([]string{"-once", "-dry-run"}, dbPath)

	// then: it exits 0 but stores nothing
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if count := countPosts(t, dbPath); count != 0 {
		t.Fatalf("expected 0 posts stored in dry-run, got %d", count)
	}
}

func TestRun_SkipsUnreachableFeedButContinuesWithOthers(t *testing.T) {
	// given: one dead feed and one working feed, both seeded into a temp db
	workingServer := newSampleFeedServer()
	defer workingServer.Close()

	deadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	deadURL := deadServer.URL
	deadServer.Close()

	dbPath := filepath.Join(t.TempDir(), testDBFileName)
	seedConn, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open seed database: %v", err)
	}
	if _, err := repos.CreateFeed(seedConn, "Dead Feed", deadURL); err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	if _, err := repos.CreateFeed(seedConn, "Working Feed", workingServer.URL); err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	seedConn.Close()

	// when: we run the fetcher
	code := run([]string{"-once"}, dbPath)

	// then: it still exits 0, and the working feed's posts got stored
	if code != 0 {
		t.Fatalf("expected exit code 0 despite one dead feed, got %d", code)
	}
	if count := countPosts(t, dbPath); count != 2 {
		t.Fatalf("expected 2 posts from the working feed, got %d", count)
	}
}

func TestRun_SkipsDisabledFeeds(t *testing.T) {
	// given: a disabled feed seeded into a temp db
	server := newSampleFeedServer()
	defer server.Close()

	dbPath := filepath.Join(t.TempDir(), testDBFileName)
	seedConn, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open seed database: %v", err)
	}
	feedID, err := repos.CreateFeed(seedConn, "Test Feed", server.URL)
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	if err := repos.DeleteFeed(seedConn, feedID); err != nil {
		t.Fatalf("DeleteFeed() returned error: %v", err)
	}
	seedConn.Close()

	// when: we run the fetcher
	code := run([]string{"-once"}, dbPath)

	// then: it exits 0 but fetches nothing from the disabled feed
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if count := countPosts(t, dbPath); count != 0 {
		t.Fatalf("expected 0 posts from a disabled feed, got %d", count)
	}
}
