package handlers

import (
	"database/sql"
	"testing"
	"time"

	"github.com/karlo/dailyniche/internal/db"
	"github.com/karlo/dailyniche/internal/repos"
)

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

func TestParseDateParam_DefaultsToTodayWhenEmpty(t *testing.T) {
	// given: no date string
	// when: we parse an empty string
	before := time.Now().UTC()
	got, err := parseDateParam("")
	after := time.Now().UTC()

	// then: it returns "now", in UTC
	if err != nil {
		t.Fatalf("parseDateParam() returned error: %v", err)
	}
	if got.Before(before) || got.After(after) {
		t.Errorf("expected a time between %v and %v, got %v", before, after, got)
	}
	if got.Location() != time.UTC {
		t.Errorf("expected UTC location, got %v", got.Location())
	}
}

func TestParseDateParam_ParsesValidDate(t *testing.T) {
	// given: a valid YYYY-MM-DD string
	// when: we parse it
	got, err := parseDateParam("2026-07-08")
	if err != nil {
		t.Fatalf("parseDateParam() returned error: %v", err)
	}

	// then: it matches the expected date, in UTC
	want := time.Date(2026, 7, 8, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}

func TestParseDateParam_ReturnsErrorForInvalidDate(t *testing.T) {
	// given: a malformed date string
	// when: we parse it
	_, err := parseDateParam("not-a-date")

	// then: it returns an error
	if err == nil {
		t.Fatal("expected an error for an invalid date, got nil")
	}
}

func TestFeedNameLookup_ReturnsNamesByID(t *testing.T) {
	// given: two feeds, one of them later disabled
	conn := newTestDB(t)
	activeID, err := repos.CreateFeed(conn, "Active Blog", "https://active.example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	disabledID, err := repos.CreateFeed(conn, "Disabled Blog", "https://disabled.example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	if err := repos.DeleteFeed(conn, disabledID); err != nil {
		t.Fatalf("DeleteFeed() returned error: %v", err)
	}

	// when: we build the lookup
	names, err := feedNameLookup(conn)
	if err != nil {
		t.Fatalf("feedNameLookup() returned error: %v", err)
	}

	// then: both feeds resolve, including the disabled one - past issues
	// must still show correct feed names even after a feed is removed
	if names[activeID] != "Active Blog" {
		t.Errorf("expected %q, got %q", "Active Blog", names[activeID])
	}
	if names[disabledID] != "Disabled Blog" {
		t.Errorf("expected disabled feed's name to still resolve, got %q", names[disabledID])
	}
}
