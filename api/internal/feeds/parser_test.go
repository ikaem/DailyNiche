package feeds

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
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
    <description>The second post summary.</description>
    <pubDate>Tue, 02 Jan 2024 10:00:00 GMT</pubDate>
  </item>
</channel>
</rss>`

func TestParseFeed_FetchesAndParsesOverHTTP(t *testing.T) {
	// given: a local HTTP server serving a sample RSS feed
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(sampleRSS))
	}))
	defer server.Close()

	// when: we parse the feed at that URL
	feed, err := ParseFeed(server.URL)
	if err != nil {
		t.Fatalf("ParseFeed() returned error: %v", err)
	}

	// then: the feed and its items are parsed correctly
	if feed.Title != "Sample Blog" {
		t.Errorf("expected feed title %q, got %q", "Sample Blog", feed.Title)
	}
	if len(feed.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(feed.Items))
	}
}

func TestParseFeed_ReturnsErrorForUnreachableURL(t *testing.T) {
	// given: a URL with nothing listening on it
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	badURL := server.URL
	server.Close()

	// when: we try to parse a feed from it
	_, err := ParseFeed(badURL)

	// then: it returns an error rather than panicking
	if err == nil {
		t.Fatal("expected an error for an unreachable URL, got nil")
	}
}

func TestExtractItems_ConvertsItemsToPosts(t *testing.T) {
	// given: a parsed feed with two items
	feed, err := gofeed.NewParser().ParseString(sampleRSS)
	if err != nil {
		t.Fatalf("failed to parse sample feed: %v", err)
	}

	// when: we extract items for feed ID 42
	posts := ExtractItems(feed, 42)

	// then: the first post is converted with all expected fields
	if len(posts) != 2 {
		t.Fatalf("expected 2 posts, got %d", len(posts))
	}
	first := posts[0]
	if first.FeedID != 42 {
		t.Errorf("expected FeedID 42, got %d", first.FeedID)
	}
	if first.Title != "First Post" {
		t.Errorf("expected title %q, got %q", "First Post", first.Title)
	}
	if first.URL != "https://example.com/first-post" {
		t.Errorf("expected url %q, got %q", "https://example.com/first-post", first.URL)
	}
	if first.GUID != "urn:uuid:first-post" {
		t.Errorf("expected guid %q, got %q", "urn:uuid:first-post", first.GUID)
	}
	if first.ContentSummary != "The first post summary." {
		t.Errorf("expected content summary %q, got %q", "The first post summary.", first.ContentSummary)
	}
	wantPublished := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	if !first.PublishedAt.Equal(wantPublished) {
		t.Errorf("expected published_at %v, got %v", wantPublished, first.PublishedAt)
	}
}

func TestExtractItems_FallsBackToLinkWhenGUIDMissing(t *testing.T) {
	// given: a parsed feed whose second item has no <guid>
	feed, err := gofeed.NewParser().ParseString(sampleRSS)
	if err != nil {
		t.Fatalf("failed to parse sample feed: %v", err)
	}

	// when: we extract items
	posts := ExtractItems(feed, 1)

	// then: the item without a guid falls back to using its link
	second := posts[1]
	if second.GUID != "https://example.com/second-post" {
		t.Errorf("expected guid to fall back to link %q, got %q", "https://example.com/second-post", second.GUID)
	}
}

func TestExtractItems_SetsFetchedAtToNowInUTC(t *testing.T) {
	// given: a parsed feed
	feed, err := gofeed.NewParser().ParseString(sampleRSS)
	if err != nil {
		t.Fatalf("failed to parse sample feed: %v", err)
	}

	// when: we extract items
	before := time.Now().UTC()
	posts := ExtractItems(feed, 1)
	after := time.Now().UTC()

	// then: FetchedAt falls between before/after, and is in UTC
	for _, p := range posts {
		if p.FetchedAt.Before(before) || p.FetchedAt.After(after) {
			t.Errorf("expected FetchedAt between %v and %v, got %v", before, after, p.FetchedAt)
		}
		if p.FetchedAt.Location() != time.UTC {
			t.Errorf("expected FetchedAt in UTC, got location %v", p.FetchedAt.Location())
		}
	}
}
