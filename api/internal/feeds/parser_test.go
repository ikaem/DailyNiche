package feeds

import (
	"net/http"
	"net/http/httptest"
	"strings"
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
    <enclosure url="https://example.com/first-post.jpg" type="image/jpeg" length="12345" />
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

func TestParseFeed_SetsSelfIdentifyingUserAgent(t *testing.T) {
	// given: a local server that records the incoming User-Agent header
	var gotUserAgent string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserAgent = r.Header.Get("User-Agent")
		w.Write([]byte(sampleRSS))
	}))
	defer server.Close()

	// when: we parse a feed from that server
	if _, err := ParseFeed(server.URL); err != nil {
		t.Fatalf("ParseFeed() returned error: %v", err)
	}

	// then: the request carried our self-identifying User-Agent, not
	// gofeed's default "Gofeed/1.0" - some sites' security plugins block
	// that default outright with a 403 (confirmed live against real feeds)
	want := "DailyNiche/1.0 (personal RSS reader)"
	if gotUserAgent != want {
		t.Errorf("expected User-Agent %q, got %q", want, gotUserAgent)
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
	if first.ImageURL != "https://example.com/first-post.jpg" {
		t.Errorf("expected image url %q, got %q", "https://example.com/first-post.jpg", first.ImageURL)
	}
	wantPublished := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	if !first.PublishedAt.Equal(wantPublished) {
		t.Errorf("expected published_at %v, got %v", wantPublished, first.PublishedAt)
	}
}

func TestExtractItems_LeavesImageURLEmptyWhenFeedProvidesNone(t *testing.T) {
	// given: a parsed feed whose second item has no <enclosure>
	feed, err := gofeed.NewParser().ParseString(sampleRSS)
	if err != nil {
		t.Fatalf("failed to parse sample feed: %v", err)
	}

	// when: we extract items
	posts := ExtractItems(feed, 1)

	// then: the item without an image has an empty ImageURL, not a nil-pointer panic
	second := posts[1]
	if second.ImageURL != "" {
		t.Errorf("expected empty image url, got %q", second.ImageURL)
	}
}

func TestImageURL_PrefersGofeedsURLWhenNotADataURI(t *testing.T) {
	// given: gofeed's own extracted image is already a real http(s) URL
	item := &gofeed.Item{Image: &gofeed.Image{URL: "https://example.com/photo.jpg"}}

	// when: we resolve the image url
	got := imageURL(item)

	// then: it's used directly - no need to look at any noscript fallback
	if got != "https://example.com/photo.jpg" {
		t.Errorf("expected %q, got %q", "https://example.com/photo.jpg", got)
	}
}

func TestImageURL_FallsBackToNoscriptRealPhotoWhenPrimaryIsDataURI(t *testing.T) {
	// given: gofeed's own extracted image is a data: URI placeholder (as
	// happens with some lazy-load plugins - confirmed live,
	// rojcnet.pula.org), but the description's <noscript> fallback holds
	// the real photo URL
	item := &gofeed.Item{
		Image: &gofeed.Image{URL: "data:image/svg+xml,%3Csvg%3E%3C/svg%3E"},
		Description: `<p><img src="data:image/svg+xml,%3Csvg%3E%3C/svg%3E" />` +
			`<noscript><img src="https://example.com/real-photo.jpg" alt="" /></noscript></p>`,
	}

	// when: we resolve the image url
	got := imageURL(item)

	// then: the real photo url is used, not the useless data: URI placeholder
	if got != "https://example.com/real-photo.jpg" {
		t.Errorf("expected the real photo url, got %q", got)
	}
}

func TestImageURL_ReturnsEmptyWhenOnlyDataURIAvailableAnywhere(t *testing.T) {
	// given: gofeed's own image is a data: URI, and there's no noscript
	// fallback anywhere to recover a real photo from
	item := &gofeed.Item{
		Image:       &gofeed.Image{URL: "data:image/svg+xml,%3Csvg%3E%3C/svg%3E"},
		Description: "<p>Just some text, no image markup at all.</p>",
	}

	// when: we resolve the image url
	got := imageURL(item)

	// then: it's empty, not the useless data: URI
	if got != "" {
		t.Errorf("expected empty image url, got %q", got)
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

func TestContentSummary_StripsHTMLTagsAndDecodesEntities(t *testing.T) {
	// given: an item whose description is raw HTML with an entity, as some
	// WordPress feeds provide (confirmed live against rojcnet.pula.org)
	// instead of a clean plain-text excerpt
	item := &gofeed.Item{
		Description: `<p>Kada govorimo o rodnoj ravnopravnosti&#8230;</p><img src="https://example.com/photo.jpg" alt="" />`,
	}

	// when: we build the content summary
	summary := contentSummary(item)

	// then: no tags remain, and the entity is decoded rather than left literal
	if strings.ContainsAny(summary, "<>") {
		t.Errorf("expected no HTML tags in summary, got %q", summary)
	}
	want := "Kada govorimo o rodnoj ravnopravnosti…"
	if summary != want {
		t.Errorf("expected %q, got %q", want, summary)
	}
}

func TestContentSummary_FallsBackToContentWhenDescriptionEmpty(t *testing.T) {
	// given: an item with no description but real content
	item := &gofeed.Item{
		Content: "<p>Full article content.</p>",
	}

	// when: we build the content summary
	summary := contentSummary(item)

	// then: it falls back to the (stripped) content
	if summary != "Full article content." {
		t.Errorf("expected %q, got %q", "Full article content.", summary)
	}
}

func TestContentSummary_TruncatesLongPlainText(t *testing.T) {
	// given: a description far longer than the summary length limit
	item := &gofeed.Item{Description: strings.Repeat("word ", 100)}

	// when: we build the content summary
	summary := contentSummary(item)

	// then: it's truncated with a trailing ellipsis, and not left with
	// trailing whitespace from a mid-word cut
	if len([]rune(summary)) > maxContentSummaryLength+1 {
		t.Errorf("expected summary truncated to ~%d runes, got %d: %q", maxContentSummaryLength, len([]rune(summary)), summary)
	}
	if !strings.HasSuffix(summary, "…") {
		t.Errorf("expected truncated summary to end with an ellipsis, got %q", summary)
	}
	if strings.HasSuffix(strings.TrimSuffix(summary, "…"), " ") {
		t.Errorf("expected no trailing whitespace before the ellipsis, got %q", summary)
	}
}

func TestContentSummary_SkipsNoscriptLazyLoadFallback(t *testing.T) {
	// given: a description shaped like the real bug (rojcnet.pula.org) - a
	// lazy-load placeholder <img>, a <noscript> fallback whose content the
	// HTML tokenizer treats as one raw, unparsed text blob (confirmed via
	// x/net/html's own rawTag handling), and finally the real paragraph text
	item := &gofeed.Item{
		Description: `<p><img src="data:image/svg+xml,%3Csvg%3E%3C/svg%3E" data-tf-src="https://example.com/real.jpg" /><noscript><img width="960" height="503" src="https://example.com/real.jpg" class="attachment-full" alt="" /></noscript></p><p>Kada govorimo o rodnoj ravnopravnosti.</p>`,
	}

	// when: we build the content summary
	summary := contentSummary(item)

	// then: neither the noscript's raw markup nor any tag leaks into the
	// summary - only the real paragraph text does, and it's not buried
	// behind (or cut off by truncation before reaching) the raw <img> text
	if strings.ContainsAny(summary, "<>") {
		t.Errorf("expected no HTML tags in summary, got %q", summary)
	}
	want := "Kada govorimo o rodnoj ravnopravnosti."
	if summary != want {
		t.Errorf("expected %q, got %q", want, summary)
	}
}

func TestExtractItems_ProducesReadableSummaryFromRawHTMLDescription(t *testing.T) {
	// given: a feed whose item description is raw WordPress post HTML -
	// lazy-load image markup, an ellipsis entity, and a syndication
	// footer - reproducing the real bug reported live against
	// rojcnet.pula.org, instead of a clean plain-text excerpt
	const wordPressStyleRSS = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
<channel>
  <title>Sample Blog</title>
  <link>https://example.com</link>
  <description>A sample feed for tests</description>
  <item>
    <title>Geopoetika solidarnosti</title>
    <link>https://example.com/geopoetika</link>
    <guid>urn:uuid:geopoetika</guid>
    <description>&lt;p&gt;&lt;img src="https://example.com/photo.jpg" class="attachment-full" alt="" /&gt;&lt;/p&gt;&lt;p&gt;Kada se na jednom mjestu okupe akteri nezavisne kulturne scene&#8230;&lt;/p&gt;&lt;p&gt;The post &lt;a href="https://example.com/geopoetika"&gt;Geopoetika solidarnosti&lt;/a&gt; first appeared on &lt;a href="https://example.com"&gt;Rojcnet&lt;/a&gt;.&lt;/p&gt;</description>
    <pubDate>Mon, 01 Jan 2024 10:00:00 GMT</pubDate>
  </item>
</channel>
</rss>`
	feed, err := gofeed.NewParser().ParseString(wordPressStyleRSS)
	if err != nil {
		t.Fatalf("failed to parse feed: %v", err)
	}

	// when: we extract items
	posts := ExtractItems(feed, 1)

	// then: the summary is clean, readable text - no tags, entity decoded
	summary := posts[0].ContentSummary
	if strings.ContainsAny(summary, "<>") {
		t.Errorf("expected no HTML tags in summary, got %q", summary)
	}
	if !strings.Contains(summary, "Kada se na jednom mjestu okupe akteri nezavisne kulturne scene…") {
		t.Errorf("expected readable Croatian text in summary, got %q", summary)
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
