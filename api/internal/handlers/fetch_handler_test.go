package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/karlo/dailyniche/internal/repos"
)

const fetchSampleRSS = `<?xml version="1.0" encoding="UTF-8"?>
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

func newFetchSampleFeedServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fetchSampleRSS))
	}))
}

func TestFetch_FetchesAndReturnsSummary(t *testing.T) {
	// given: a feed pointing at a local test server
	server := newFetchSampleFeedServer()
	defer server.Close()
	conn := newTestDB(t)
	if _, err := repos.CreateFeed(conn, "Test Feed", server.URL); err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}

	// when: we request POST /api/fetch
	req := httptest.NewRequest(http.MethodPost, "/api/fetch", nil)
	rec := httptest.NewRecorder()
	Fetch(conn)(rec, req)

	// then: it responds 200 with a summary reflecting the 2 new posts
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	var got FetchSummaryResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got.New != 2 {
		t.Errorf("expected 2 new posts, got %d", got.New)
	}
	if got.Duplicates != 0 || got.Errors != 0 {
		t.Errorf("expected 0 duplicates/errors, got %+v", got)
	}
}

func TestFetch_ReportsDuplicatesOnSecondCall(t *testing.T) {
	// given: a feed already fetched once
	server := newFetchSampleFeedServer()
	defer server.Close()
	conn := newTestDB(t)
	if _, err := repos.CreateFeed(conn, "Test Feed", server.URL); err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	Fetch(conn)(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/api/fetch", nil))

	// when: we request POST /api/fetch again
	req := httptest.NewRequest(http.MethodPost, "/api/fetch", nil)
	rec := httptest.NewRecorder()
	Fetch(conn)(rec, req)

	// then: the second fetch reports duplicates, not new posts
	var got FetchSummaryResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got.New != 0 || got.Duplicates != 2 {
		t.Errorf("expected 0 new/2 duplicates on second fetch, got %+v", got)
	}
}
