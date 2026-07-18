package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/karlo/dailyniche/internal/db"
	"github.com/karlo/dailyniche/internal/models"
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

// newTestPost returns a Post referencing feedID, fetched at fetchedAt.
func newTestPost(feedID int64, guid string, fetchedAt time.Time) *models.Post {
	return &models.Post{
		FeedID:         feedID,
		Title:          "Sample Post",
		URL:            "https://example.com/sample-post",
		ContentSummary: "A sample post summary.",
		PublishedAt:    fetchedAt,
		FetchedAt:      fetchedAt,
		GUID:           guid,
	}
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

func TestPosts_ReturnsPostsForToday(t *testing.T) {
	// given: a feed with one post fetched today
	conn := newTestDB(t)
	feedID, err := repos.CreateFeed(conn, "Tech Blog", "https://example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	if _, err := repos.CreatePost(conn, newTestPost(feedID, "urn:uuid:today-post", time.Now().UTC())); err != nil {
		t.Fatalf("CreatePost() returned error: %v", err)
	}

	// when: we request /api/posts with no query params
	req := httptest.NewRequest(http.MethodGet, "/api/posts", nil)
	rec := httptest.NewRecorder()
	Posts(conn)(rec, req)

	// then: it responds 200 with the post, enriched with its feed's name
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	var got []PostResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 post, got %d", len(got))
	}
	if got[0].FeedID != feedID {
		t.Errorf("expected feed_id %d, got %d", feedID, got[0].FeedID)
	}
	if got[0].FeedName != "Tech Blog" {
		t.Errorf("expected feed_name %q, got %q", "Tech Blog", got[0].FeedName)
	}
}

func TestPosts_SubstitutesPlaceholderWhenPostHasNoImage(t *testing.T) {
	// given: a post with no image url (newTestPost leaves ImageURL empty)
	conn := newTestDB(t)
	feedID, err := repos.CreateFeed(conn, "Tech Blog", "https://example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	if _, err := repos.CreatePost(conn, newTestPost(feedID, "urn:uuid:no-image-post", time.Now().UTC())); err != nil {
		t.Fatalf("CreatePost() returned error: %v", err)
	}

	// when: we request /api/posts
	req := httptest.NewRequest(http.MethodGet, "/api/posts", nil)
	rec := httptest.NewRecorder()
	Posts(conn)(rec, req)

	// then: the response substitutes the placeholder rather than an empty string
	var got []PostResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 post, got %d", len(got))
	}
	if got[0].ImageURL != placeholderImageURL {
		t.Errorf("expected placeholder image url, got %q", got[0].ImageURL)
	}
}

func TestPosts_PassesThroughRealImageURL(t *testing.T) {
	// given: a post with a real image url
	conn := newTestDB(t)
	feedID, err := repos.CreateFeed(conn, "Tech Blog", "https://example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	post := newTestPost(feedID, "urn:uuid:has-image-post", time.Now().UTC())
	post.ImageURL = "https://example.com/real-image.jpg"
	if _, err := repos.CreatePost(conn, post); err != nil {
		t.Fatalf("CreatePost() returned error: %v", err)
	}

	// when: we request /api/posts
	req := httptest.NewRequest(http.MethodGet, "/api/posts", nil)
	rec := httptest.NewRecorder()
	Posts(conn)(rec, req)

	// then: the response carries the real url unchanged, not the placeholder
	var got []PostResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 post, got %d", len(got))
	}
	if got[0].ImageURL != "https://example.com/real-image.jpg" {
		t.Errorf("expected real image url, got %q", got[0].ImageURL)
	}
}

func TestPosts_FiltersByDateParam(t *testing.T) {
	// given: posts fetched on two different days
	conn := newTestDB(t)
	feedID, err := repos.CreateFeed(conn, "Tech Blog", "https://example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	onTargetID, err := repos.CreatePost(conn, newTestPost(feedID, "urn:uuid:on-target", time.Date(2024, 3, 15, 9, 0, 0, 0, time.UTC)))
	if err != nil {
		t.Fatalf("CreatePost() returned error: %v", err)
	}
	if _, err := repos.CreatePost(conn, newTestPost(feedID, "urn:uuid:other-day", time.Date(2024, 3, 16, 9, 0, 0, 0, time.UTC))); err != nil {
		t.Fatalf("CreatePost() returned error: %v", err)
	}

	// when: we request posts for the target date
	req := httptest.NewRequest(http.MethodGet, "/api/posts?date=2024-03-15", nil)
	rec := httptest.NewRecorder()
	Posts(conn)(rec, req)

	// then: only the matching post is returned
	var got []PostResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 post, got %d", len(got))
	}
	if got[0].ID != onTargetID {
		t.Errorf("expected post id %d, got %d", onTargetID, got[0].ID)
	}
}

func TestPosts_FiltersByFeedIDParam(t *testing.T) {
	// given: posts from two different feeds, fetched today
	conn := newTestDB(t)
	feedA, err := repos.CreateFeed(conn, "Feed A", "https://a.example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	feedB, err := repos.CreateFeed(conn, "Feed B", "https://b.example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	now := time.Now().UTC()
	if _, err := repos.CreatePost(conn, newTestPost(feedA, "urn:uuid:feed-a-post", now)); err != nil {
		t.Fatalf("CreatePost() returned error: %v", err)
	}
	if _, err := repos.CreatePost(conn, newTestPost(feedB, "urn:uuid:feed-b-post", now)); err != nil {
		t.Fatalf("CreatePost() returned error: %v", err)
	}

	// when: we request posts filtered to feed A
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/posts?feed_id=%d", feedA), nil)
	rec := httptest.NewRecorder()
	Posts(conn)(rec, req)

	// then: only feed A's post is returned
	var got []PostResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 post, got %d", len(got))
	}
	if got[0].FeedID != feedA {
		t.Errorf("expected feed_id %d, got %d", feedA, got[0].FeedID)
	}
}

func TestPosts_ReturnsBadRequestForInvalidDate(t *testing.T) {
	// given: an empty database
	conn := newTestDB(t)

	// when: we request with a malformed date
	req := httptest.NewRequest(http.MethodGet, "/api/posts?date=not-a-date", nil)
	rec := httptest.NewRecorder()
	Posts(conn)(rec, req)

	// then: it responds 400
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestPosts_ReturnsBadRequestForInvalidFeedID(t *testing.T) {
	// given: an empty database
	conn := newTestDB(t)

	// when: we request with a non-numeric feed_id
	req := httptest.NewRequest(http.MethodGet, "/api/posts?feed_id=abc", nil)
	rec := httptest.NewRecorder()
	Posts(conn)(rec, req)

	// then: it responds 400
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestPosts_ReturnsEmptyArrayWhenNoPosts(t *testing.T) {
	// given: an empty database
	conn := newTestDB(t)

	// when: we request posts
	req := httptest.NewRequest(http.MethodGet, "/api/posts", nil)
	rec := httptest.NewRecorder()
	Posts(conn)(rec, req)

	// then: it responds 200 with an empty JSON array, not null
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	// Checking the raw string here on purpose, not decoding into
	// []PostResponse: this test exists specifically to prove the JSON
	// literally renders as [], not null. If we decoded it first and just
	// checked len(got) == 0, that would pass identically whether the wire
	// format was [] or null - json.Unmarshal treats both as "empty slice"
	// when decoding into a slice type. Checking the raw string is the only
	// way to prove the wire format itself is [], which is exactly the
	// guarantee the responses := make([]PostResponse, 0, len(posts))
	// pre-allocation in posts_handler.go is meant to provide.
	if body := strings.TrimSpace(rec.Body.String()); body != "[]" {
		t.Errorf("expected empty JSON array \"[]\", got %q", body)
	}
}
