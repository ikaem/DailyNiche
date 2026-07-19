package feeds

import (
	"time"

	"github.com/mmcdole/gofeed"

	"github.com/karlo/dailyniche/internal/models"
)

// parser is a single, package-level gofeed.Parser configured with a
// self-identifying User-Agent, reused across every ParseFeed call rather
// than constructing a fresh default one each time. Two things motivated
// this: gofeed's default UA is the literal string "Gofeed/1.0", which some
// sites' security plugins block outright with a 403 - confirmed live
// against multiple real feeds (a WordPress site, and rojcnet.pula.org/rss)
// while a browser-like UA passes; and building a new *gofeed.Parser per
// call was always wasteful once fetching many feeds in one run (see the
// fetcher's FetchAll loop). Safe to share: gofeed.Parser's only mutable
// state is its embedded *http.Client, which is itself documented as safe
// for concurrent use - not that it matters yet, since FetchAll's loop is
// sequential today anyway.
var parser = newParser()

func newParser() *gofeed.Parser {
	p := gofeed.NewParser()
	p.UserAgent = "DailyNiche/1.0 (personal RSS reader)"
	return p
}

// ParseFeed fetches and parses the RSS/Atom/JSON feed at url.
func ParseFeed(url string) (*gofeed.Feed, error) {
	return parser.ParseURL(url)
}

// ExtractItems converts a parsed feed's items into Posts associated with
// feedID, ready for storage.
func ExtractItems(feed *gofeed.Feed, feedID int64) []models.Post {
	posts := make([]models.Post, 0, len(feed.Items))
	for _, item := range feed.Items {
		posts = append(posts, models.Post{
			FeedID:         feedID,
			Title:          item.Title,
			URL:            item.Link,
			ContentSummary: contentSummary(item),
			ImageURL:       imageURL(item),
			PublishedAt:    publishedAt(item),
			FetchedAt:      time.Now().UTC(),
			GUID:           guid(item),
		})
	}
	return posts
}

// guid returns the item's GUID, falling back to its link when the feed
// didn't provide one (not every feed sets <guid> correctly).
func guid(item *gofeed.Item) string {
	if item.GUID != "" {
		return item.GUID
	}
	return item.Link
}

// contentSummary prefers the item's description, falling back to its full
// content when no description is present.
func contentSummary(item *gofeed.Item) string {
	if item.Description != "" {
		return item.Description
	}
	return item.Content
}

// imageURL returns the item's image URL, or an empty string if the feed
// provided none - not every feed includes one.
func imageURL(item *gofeed.Item) string {
	if item.Image != nil {
		return item.Image.URL
	}
	return ""
}

// publishedAt prefers the item's parsed publish date, falling back to its
// updated date, and finally to now if the feed provided neither.
func publishedAt(item *gofeed.Item) time.Time {
	if item.PublishedParsed != nil {
		return item.PublishedParsed.UTC()
	}
	if item.UpdatedParsed != nil {
		return item.UpdatedParsed.UTC()
	}
	return time.Now().UTC()
}
