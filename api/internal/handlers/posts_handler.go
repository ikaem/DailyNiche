package handlers

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/karlo/dailyniche/internal/repos"
)

// PostResponse is the JSON shape returned for each post, enriched with the
// owning feed's name so API consumers don't need a separate lookup. FeedID
// is kept alongside FeedName, not replaced by it - callers still need the
// numeric ID for filtering (feed_id query param) or any future per-feed
// features.
type PostResponse struct {
	ID             int64     `json:"id"`
	FeedID         int64     `json:"feed_id"`
	FeedName       string    `json:"feed_name"`
	Title          string    `json:"title"`
	URL            string    `json:"url"`
	ContentSummary string    `json:"content_summary"`
	ImageURL       string    `json:"image_url"`
	PublishedAt    time.Time `json:"published_at"`
	FetchedAt      time.Time `json:"fetched_at"`
}

// placeholderImageSVG is a generic "no image" glyph (a photo icon: a sun and
// mountains on a neutral ground) built from plain shapes, no external asset.
const placeholderImageSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 300">` +
	`<rect width="400" height="300" fill="#e5e5e5"/>` +
	`<circle cx="150" cy="110" r="25" fill="#bdbdbd"/>` +
	`<polygon points="60,240 160,140 220,200 280,120 340,240" fill="#bdbdbd"/>` +
	`</svg>`

// placeholderImageURL substitutes for a post's image_url when a feed
// provided none. Decided (2026-07-13, see CLAUDE.md Task 2.3) that this
// fallback belongs here, in the API layer, not in each client - so any
// future client (web, mobile, ...) gets consistent placeholder behavior for
// free. An inline SVG data URI keeps it self-contained: no static-file
// route or externally-hosted dependency needed.
var placeholderImageURL = "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(placeholderImageSVG))

// imageURLOrPlaceholder returns url, or placeholderImageURL if url is empty.
func imageURLOrPlaceholder(url string) string {
	if url != "" {
		return url
	}
	return placeholderImageURL
}

// Posts returns an http.HandlerFunc for GET /api/posts, backed by conn.
// Query params: date (YYYY-MM-DD, defaults to today in UTC), feed_id
// (optional).
func Posts(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		date, err := parseDateParam(r.URL.Query().Get("date"))
		if err != nil {
			http.Error(w, "invalid date, expected YYYY-MM-DD", http.StatusBadRequest)
			return
		}

		// *int64, not int64: a plain int64 has no way to represent "no filter
		// requested" - its zero value (0) would be indistinguishable from
		// "filter by feed ID 0". nil means "no filter"; a non-nil pointer
		// means "filter by this ID".
		var feedIDFilter *int64
		if raw := r.URL.Query().Get("feed_id"); raw != "" {
			id, err := strconv.ParseInt(raw, 10, 64)
			if err != nil {
				http.Error(w, "invalid feed_id", http.StatusBadRequest)
				return
			}
			feedIDFilter = &id
		}

		// TODO: when feed_id is given, this fetches every post for the date
		// and filters in Go below, rather than pushing the feed_id filter
		// down into the SQL query. Fine at our scale (a day's posts is a
		// small number); revisit if per-day post volume ever grows enough
		// for this to matter (e.g. add a ListPostsByDateAndFeed repo func).
		posts, err := repos.ListPostsByDate(conn, date)
		if err != nil {
			http.Error(w, "failed to list posts", http.StatusInternalServerError)
			return
		}

		feedNames, err := feedNameLookup(conn)
		if err != nil {
			http.Error(w, "failed to list feeds", http.StatusInternalServerError)
			return
		}

		responses := make([]PostResponse, 0, len(posts))
		for _, p := range posts {
			if feedIDFilter != nil && p.FeedID != *feedIDFilter {
				continue
			}
			responses = append(responses, PostResponse{
				ID:             p.ID,
				FeedID:         p.FeedID,
				FeedName:       feedNames[p.FeedID],
				Title:          p.Title,
				URL:            p.URL,
				ContentSummary: p.ContentSummary,
				ImageURL:       imageURLOrPlaceholder(p.ImageURL),
				PublishedAt:    p.PublishedAt,
				FetchedAt:      p.FetchedAt,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(responses)
	}
}

// parseDateParam parses raw (the "date" query param) as YYYY-MM-DD.
//
// An empty raw defaults to today (UTC) rather than erroring - GET /api/posts
// with no query params at all is the primary use case this whole app is
// built around ("ping it, get today's news"), so "no date given" must mean
// "today", not "bad request".
//
// time.Parse returns a UTC time here since the "2006-01-02" layout has no
// timezone token to parse from the input - nothing further to convert.
func parseDateParam(raw string) (time.Time, error) {
	if raw == "" {
		return time.Now().UTC(), nil
	}
	return time.Parse("2006-01-02", raw)
}

// feedNameLookup builds a feed ID -> name map for enriching post responses,
// e.g. given feeds (1, "Tech Blog") and (2, "Cooking Blog"), it returns:
//
//	map[int64]string{1: "Tech Blog", 2: "Cooking Blog"}
//
// Includes disabled feeds too (repos.ListFeeds already does), so posts under
// a since-removed feed still resolve a name - past issues must display
// exactly as before.
func feedNameLookup(conn *sql.DB) (map[int64]string, error) {
	feedList, err := repos.ListFeeds(conn)
	if err != nil {
		return nil, err
	}
	names := make(map[int64]string, len(feedList))
	for _, f := range feedList {
		names[f.ID] = f.Name
	}
	return names, nil
}
