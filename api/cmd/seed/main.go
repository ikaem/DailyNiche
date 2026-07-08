// Command seed populates the database with sample feeds and posts spread
// across today and yesterday, for local manual testing of the API/frontend
// without needing real RSS feeds. Not part of the production fetch/serve
// flow - dev convenience only.
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/karlo/dailyniche/internal/db"
	"github.com/karlo/dailyniche/internal/models"
	"github.com/karlo/dailyniche/internal/repos"
)

type seedPost struct {
	title   string
	url     string
	summary string
	daysAgo int // 0 = today, 1 = yesterday
}

type seedFeed struct {
	name  string
	url   string
	posts []seedPost
}

var seedFeeds = []seedFeed{
	{
		name: "Tech Blog",
		url:  "https://example.com/tech/feed.xml",
		posts: []seedPost{
			{title: "Go 2.0 Announced", url: "https://example.com/tech/go-2", summary: "The Go team announces the next major version.", daysAgo: 0},
			{title: "Why SQLite Is Enough", url: "https://example.com/tech/sqlite-enough", summary: "A case for boring, reliable databases.", daysAgo: 0},
			{title: "Static Sites Are Back", url: "https://example.com/tech/static-sites", summary: "The pendulum swings again.", daysAgo: 1},
		},
	},
	{
		name: "Cooking Blog",
		url:  "https://example.com/cooking/feed.xml",
		posts: []seedPost{
			{title: "Perfect Sourdough Starter", url: "https://example.com/cooking/sourdough", summary: "A no-fuss guide to your first starter.", daysAgo: 0},
			{title: "Weeknight Pasta Ideas", url: "https://example.com/cooking/weeknight-pasta", summary: "Five pastas in under 20 minutes.", daysAgo: 1},
		},
	},
	{
		name: "Travel Blog",
		url:  "https://example.com/travel/feed.xml",
		posts: []seedPost{
			{title: "A Weekend in Ljubljana", url: "https://example.com/travel/ljubljana", summary: "Small city, big charm.", daysAgo: 0},
			{title: "Packing Light for Winter", url: "https://example.com/travel/packing-light", summary: "One bag, three weeks, no regrets.", daysAgo: 1},
		},
	},
}

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "dailyniche.db"
	}

	conn, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer conn.Close()

	// CreateFeed has no ON CONFLICT handling (unlike CreatePost) - url is
	// UNIQUE, so calling it twice for the same URL errors. Look up existing
	// feeds first and reuse their IDs, so this command is actually safe to
	// run repeatedly.
	existingFeeds, err := repos.ListFeeds(conn)
	if err != nil {
		log.Fatalf("failed to list existing feeds: %v", err)
	}
	feedIDByURL := make(map[string]int64, len(existingFeeds))
	for _, f := range existingFeeds {
		feedIDByURL[f.URL] = f.ID
	}

	now := time.Now().UTC()

	var feedCount, newPostCount int
	for feedIdx, sf := range seedFeeds {
		feedID, ok := feedIDByURL[sf.url]
		if !ok {
			feedID, err = repos.CreateFeed(conn, sf.name, sf.url)
			if err != nil {
				log.Fatalf("failed to create feed %q: %v", sf.name, err)
			}
			feedCount++
		}

		for postIdx, sp := range sf.posts {
			fetchedAt := now.AddDate(0, 0, -sp.daysAgo)
			// Fixed, index-based GUIDs make re-running this command
			// idempotent - CreatePost silently skips ones already stored.
			guid := fmt.Sprintf("urn:uuid:seed-feed%d-post%d", feedIdx, postIdx)

			id, err := repos.CreatePost(conn, &models.Post{
				FeedID:         feedID,
				Title:          sp.title,
				URL:            sp.url,
				ContentSummary: sp.summary,
				PublishedAt:    fetchedAt,
				FetchedAt:      fetchedAt,
				GUID:           guid,
			})
			if err != nil {
				log.Fatalf("failed to create post %q: %v", sp.title, err)
			}
			if id > 0 {
				newPostCount++
			}
		}
	}

	log.Printf("seed complete: %d feeds ensured, %d new posts added to %s", feedCount, newPostCount, dbPath)
}
