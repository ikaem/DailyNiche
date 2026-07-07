package repos

import (
	"database/sql"
	"testing"

	"github.com/karlo/dailyniche/internal/db"
)

// newTestDB returns a migrated in-memory database for testing.
func newTestDB(t *testing.T) *sql.DB {
	// Marks this as a test helper: failures inside here are reported at the
	// caller's line number instead of pointing into this function.
	t.Helper()
	conn, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	// Registers cleanup on the calling test itself, so it runs when that
	// test finishes (pass/fail/panic) - unlike a plain defer here, which
	// would close conn immediately when newTestDB returns, before any test
	// got to use it.
	t.Cleanup(func() { conn.Close() })
	return conn
}

func TestCreateFeed_InsertsAndReturnsID(t *testing.T) {
	// given: a fresh migrated database
	conn := newTestDB(t)

	// when: we create a feed
	id, err := CreateFeed(conn, "Sample Blog", "https://example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}

	// then: it returns a positive, assigned ID
	if id <= 0 {
		t.Errorf("expected a positive ID, got %d", id)
	}
}

func TestListFeeds_ReturnsCreatedFeed(t *testing.T) {
	// given: a database with one feed created
	conn := newTestDB(t)
	id, err := CreateFeed(conn, "Sample Blog", "https://example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}

	// when: we list feeds
	feeds, err := ListFeeds(conn)
	if err != nil {
		t.Fatalf("ListFeeds() returned error: %v", err)
	}

	// then: the created feed is present with expected fields
	if len(feeds) != 1 {
		t.Fatalf("expected 1 feed, got %d", len(feeds))
	}
	f := feeds[0]
	if f.ID != id {
		t.Errorf("expected ID %d, got %d", id, f.ID)
	}
	if f.Name != "Sample Blog" {
		t.Errorf("expected name %q, got %q", "Sample Blog", f.Name)
	}
	if f.URL != "https://example.com/feed.xml" {
		t.Errorf("expected url %q, got %q", "https://example.com/feed.xml", f.URL)
	}
	if f.DisabledAt != nil {
		t.Errorf("expected DisabledAt to be nil for a new feed, got %v", f.DisabledAt)
	}
	if f.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	if f.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestListFeeds_ReturnsEmptySliceWhenNoFeeds(t *testing.T) {
	// given: a fresh database with no feeds
	conn := newTestDB(t)

	// when: we list feeds
	feeds, err := ListFeeds(conn)
	if err != nil {
		t.Fatalf("ListFeeds() returned error: %v", err)
	}

	// then: it returns an empty slice, not an error
	if len(feeds) != 0 {
		t.Errorf("expected 0 feeds, got %d", len(feeds))
	}
}

func TestListFeeds_OrdersByName(t *testing.T) {
	// given: feeds created in non-alphabetical order
	conn := newTestDB(t)
	if _, err := CreateFeed(conn, "Zebra Blog", "https://zebra.example.com/feed.xml"); err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	if _, err := CreateFeed(conn, "Alpha Blog", "https://alpha.example.com/feed.xml"); err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}

	// when: we list feeds
	feeds, err := ListFeeds(conn)
	if err != nil {
		t.Fatalf("ListFeeds() returned error: %v", err)
	}

	// then: they come back ordered alphabetically by name
	if len(feeds) != 2 {
		t.Fatalf("expected 2 feeds, got %d", len(feeds))
	}
	if feeds[0].Name != "Alpha Blog" || feeds[1].Name != "Zebra Blog" {
		t.Errorf("expected [Alpha Blog, Zebra Blog], got [%s, %s]", feeds[0].Name, feeds[1].Name)
	}
}
