package feeds

import (
	"regexp"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"golang.org/x/net/html"

	"github.com/karlo/dailyniche/internal/models"
)

// maxContentSummaryLength caps how long a stored summary can be - some
// feeds' <description>/<content:encoded> contain the entire raw post
// (not a short excerpt), which would otherwise dump an entire article's
// worth of text into what's meant to be a one- or two-line teaser.
const maxContentSummaryLength = 300

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
// content when no description is present. Some feeds' description/content
// is full raw post HTML (images, WordPress lazy-load scaffolding, "first
// appeared on" footers) rather than a clean plain-text excerpt - confirmed
// live against a real feed (rojcnet.pula.org) whose posts rendered as a
// wall of literal markup instead of readable text. stripHTML+truncate turn
// that into an actual short, readable summary regardless of the source.
func contentSummary(item *gofeed.Item) string {
	raw := item.Description
	if raw == "" {
		raw = item.Content
	}
	return truncate(stripHTML(raw), maxContentSummaryLength)
}

// rawTextTags are elements whose content the HTML tokenizer treats as one
// opaque, unparsed blob rather than nested markup to descend into (see
// x/net/html's own rawTag handling) - matching HTML5's "raw text"/
// "escapable raw text" elements. None of these ever hold genuine visible
// article text in an RSS description, but confirmed live
// (rojcnet.pula.org) that WordPress lazy-load plugins commonly wrap a
// <noscript><img .../></noscript> fallback in feed content - without
// skipping it, that fallback's raw markup leaks into the summary as
// literal text instead of being stripped like everything else.
var rawTextTags = map[string]bool{
	"iframe":    true,
	"noembed":   true,
	"noframes":  true,
	"noscript":  true,
	"plaintext": true,
	"script":    true,
	"style":     true,
	"title":     true,
	"textarea":  true,
	"xmp":       true,
}

// stripHTML returns just the text content of s, discarding every tag and
// attribute and decoding HTML entities (e.g. "&#8230;" becomes "…") along
// the way. Uses a real HTML tokenizer rather than a regex - regexes are
// notoriously unreliable against real-world (often malformed) HTML.
func stripHTML(s string) string {
	var sb strings.Builder
	tokenizer := html.NewTokenizer(strings.NewReader(s))
	skipping := false
	for {
		switch tokenizer.Next() {
		case html.ErrorToken:
			// io.EOF (or a genuine parse error, treated the same way here -
			// return whatever text was gathered before the tokenizer gave up).
			// This return is what ends this otherwise-infinite for {} loop.
			return strings.Join(strings.Fields(sb.String()), " ")
		case html.StartTagToken:
			// rawTextTags[...] is a set-membership check: the map's value is
			// never used, only whether the key is present at all (true if
			// present, the zero value false if the tag name isn't a key).
			if name, _ := tokenizer.TagName(); rawTextTags[string(name)] {
				skipping = true
			}
		case html.EndTagToken:
			if name, _ := tokenizer.TagName(); rawTextTags[string(name)] {
				skipping = false
			}
		case html.TextToken:
			if skipping {
				continue
			}
			// Text()'s returned slice may be overwritten by the next Next()
			// call, so it must be copied out via string(...) immediately.
			sb.WriteString(string(tokenizer.Text()))
			sb.WriteString(" ")
		}
	}
}

// truncate shortens s to at most maxRunes runes (not bytes - Croatian and
// other non-ASCII text is multi-byte UTF-8, and slicing by byte count risks
// cutting a rune in half), breaking at the last word boundary rather than
// mid-word, and appending an ellipsis if anything was actually cut.
func truncate(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}

	cut := string(runes[:maxRunes])
	// finds intext of last space
	// if that index is bigger than 0, we cut the string at that index
	if idx := strings.LastIndex(cut, " "); idx > 0 {
		cut = cut[:idx]
	}
	return strings.TrimSpace(cut) + "…"
}

// noscriptImgSrcPattern matches the src of the first <img> tag inside an
// already-isolated <noscript> raw-text blob. Narrow, deliberate use of a
// regex here - unlike stripHTML's general "strip all markup" job, this
// only ever runs against a small fragment already known to be exactly one
// <noscript> block's contents, extracting one attribute from it.
var noscriptImgSrcPattern = regexp.MustCompile(`<img[^>]*\ssrc="([^"]+)"`)

// imageURL returns the item's image URL, or an empty string if the feed
// provided none - not every feed includes one. gofeed's own extracted
// image is trusted unless it's a data: URI - confirmed live
// (rojcnet.pula.org) that some lazy-load plugins put a fake inline SVG
// placeholder directly in <img src>, with the real photo only reachable
// via a <noscript> fallback (or a plugin-specific data-* attribute gofeed
// doesn't know about) - a data: URI is never a genuine content photo, so
// in that case we look for a real one in the raw description/content's
// <noscript> fallback instead of showing the placeholder as if it were
// the post's image.
func imageURL(item *gofeed.Item) string {
	if item.Image != nil && !strings.HasPrefix(item.Image.URL, "data:") {
		return item.Image.URL
	}
	if src := firstNoscriptImageSrc(item.Description); src != "" {
		return src
	}
	if src := firstNoscriptImageSrc(item.Content); src != "" {
		return src
	}
	return ""
}

// firstNoscriptImageSrc looks for a <noscript>...<img src="...">...</noscript>
// block within raw HTML and returns that img's src, or "" if none is found.
func firstNoscriptImageSrc(raw string) string {
	tokenizer := html.NewTokenizer(strings.NewReader(raw))
	inNoscript := false
	for {
		switch tokenizer.Next() {
		case html.ErrorToken:
			// reached the end (or a parse error) with no match found - this
			// return is what ends this otherwise-infinite for {} loop.
			return ""
		case html.StartTagToken:
			if name, _ := tokenizer.TagName(); string(name) == "noscript" {
				inNoscript = true
			}
		case html.EndTagToken:
			if name, _ := tokenizer.TagName(); string(name) == "noscript" {
				inNoscript = false
			}
		case html.TextToken:
			if !inNoscript {
				continue
			}
			if m := noscriptImgSrcPattern.FindStringSubmatch(string(tokenizer.Text())); m != nil {
				return m[1]
			}
		}
	}
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
