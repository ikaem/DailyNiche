package feeds

import (
	"net/http"
	"net/http/httptest"
	"testing"
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
