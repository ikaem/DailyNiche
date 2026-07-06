package feeds

import (
	"time"

	"github.com/mmcdole/gofeed"

	"github.com/karlo/dailyniche/internal/models"
)

// ParseFeed fetches and parses the RSS/Atom/JSON feed at url.
func ParseFeed(url string) (*gofeed.Feed, error) {
	return gofeed.NewParser().ParseURL(url)
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
