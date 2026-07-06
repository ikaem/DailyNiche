package feeds

import "github.com/mmcdole/gofeed"

// ParseFeed fetches and parses the RSS/Atom/JSON feed at url.
func ParseFeed(url string) (*gofeed.Feed, error) {
	return gofeed.NewParser().ParseURL(url)
}
