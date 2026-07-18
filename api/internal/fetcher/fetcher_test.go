package fetcher

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/karlo/dailyniche/internal/db"
	"github.com/karlo/dailyniche/internal/repos"
)

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

// newTestDB returns a migrated in-memory database for testing.
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	conn, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

// countPosts returns how many rows exist in the posts table.
func countPosts(t *testing.T, conn *sql.DB) int {
	t.Helper()
	var count int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM posts`).Scan(&count); err != nil {
		t.Fatalf("failed to count posts: %v", err)
	}
	return count
}

func TestFetchAll_FetchesAndStoresNewPosts(t *testing.T) {
	// given: a feed pointing at a local test server
	server := newSampleFeedServer()
	defer server.Close()
	conn := newTestDB(t)
	if _, err := repos.CreateFeed(conn, "Test Feed", server.URL); err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}

	// when: we fetch all feeds
	summary, err := FetchAll(conn, Options{})

	// then: no error, both posts are stored, and the summary reports them as new
	if err != nil {
		t.Fatalf("FetchAll() returned error: %v", err)
	}
	if summary.New != 2 {
		t.Errorf("expected 2 new posts, got %d", summary.New)
	}
	if count := countPosts(t, conn); count != 2 {
		t.Fatalf("expected 2 posts stored, got %d", count)
	}
}

func TestFetchAll_CanBeCalledRepeatedlyWithoutDuplicating(t *testing.T) {
	// given: a feed seeded into a test db
	server := newSampleFeedServer()
	defer server.Close()
	conn := newTestDB(t)
	if _, err := repos.CreateFeed(conn, "Test Feed", server.URL); err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}

	// when: we fetch the same feed twice
	if _, err := FetchAll(conn, Options{}); err != nil {
		t.Fatalf("first FetchAll() returned error: %v", err)
	}
	summary, err := FetchAll(conn, Options{})
	if err != nil {
		t.Fatalf("second FetchAll() returned error: %v", err)
	}

	// then: the second run reports duplicates, not new posts, and storage
	// still holds only the original 2
	if summary.New != 0 || summary.Duplicates != 2 {
		t.Errorf("expected 0 new/2 duplicates on the second run, got %+v", summary)
	}
	if count := countPosts(t, conn); count != 2 {
		t.Fatalf("expected still 2 posts after a second run, got %d", count)
	}
}

func TestFetchAll_DryRunDoesNotStorePosts(t *testing.T) {
	// given: a feed seeded into a test db
	server := newSampleFeedServer()
	defer server.Close()
	conn := newTestDB(t)
	if _, err := repos.CreateFeed(conn, "Test Feed", server.URL); err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}

	// when: we fetch with DryRun set
	summary, err := FetchAll(conn, Options{DryRun: true})

	// then: no error, but nothing is stored
	if err != nil {
		t.Fatalf("FetchAll() returned error: %v", err)
	}
	if summary.New != 0 {
		t.Errorf("expected 0 new posts in dry-run, got %d", summary.New)
	}
	if count := countPosts(t, conn); count != 0 {
		t.Fatalf("expected 0 posts stored in dry-run, got %d", count)
	}
}

func TestFetchAll_SkipsUnreachableFeedButContinuesWithOthers(t *testing.T) {
	// given: one dead feed and one working feed
	workingServer := newSampleFeedServer()
	defer workingServer.Close()
	deadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	deadURL := deadServer.URL
	deadServer.Close()

	conn := newTestDB(t)
	if _, err := repos.CreateFeed(conn, "Dead Feed", deadURL); err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	if _, err := repos.CreateFeed(conn, "Working Feed", workingServer.URL); err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}

	// when: we fetch all feeds
	summary, err := FetchAll(conn, Options{})

	// then: no top-level error, the dead feed is counted as an error, and
	// the working feed's posts are still stored
	if err != nil {
		t.Fatalf("FetchAll() returned error: %v", err)
	}
	if summary.Errors != 1 {
		t.Errorf("expected 1 error for the dead feed, got %d", summary.Errors)
	}
	if count := countPosts(t, conn); count != 2 {
		t.Fatalf("expected 2 posts from the working feed, got %d", count)
	}
}

func TestFetchAll_SkipsDisabledFeeds(t *testing.T) {
	// given: a disabled feed
	server := newSampleFeedServer()
	defer server.Close()
	conn := newTestDB(t)
	feedID, err := repos.CreateFeed(conn, "Test Feed", server.URL)
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	if err := repos.DeleteFeed(conn, feedID); err != nil {
		t.Fatalf("DeleteFeed() returned error: %v", err)
	}

	// when: we fetch all feeds
	summary, err := FetchAll(conn, Options{})

	// then: nothing is fetched from the disabled feed
	if err != nil {
		t.Fatalf("FetchAll() returned error: %v", err)
	}
	if summary.New != 0 {
		t.Errorf("expected 0 new posts from a disabled feed, got %d", summary.New)
	}
	if count := countPosts(t, conn); count != 0 {
		t.Fatalf("expected 0 posts from a disabled feed, got %d", count)
	}
}
