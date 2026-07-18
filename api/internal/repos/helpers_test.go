package repos

import (
	"database/sql"
	"testing"
	"time"

	"github.com/karlo/dailyniche/internal/db"
	"github.com/karlo/dailyniche/internal/models"
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

// newTestPost returns a Post referencing feedID, ready to insert.
func newTestPost(feedID int64, guid string) *models.Post {
	return &models.Post{
		FeedID:         feedID,
		Title:          "Sample Post",
		URL:            "https://example.com/sample-post",
		ContentSummary: "A sample post summary.",
		ImageURL:       "https://example.com/sample-image.jpg",
		PublishedAt:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		FetchedAt:      time.Date(2024, 1, 1, 10, 5, 0, 0, time.UTC),
		GUID:           guid,
	}
}
