package handlers

import (
	"database/sql"
	"time"

	"github.com/karlo/dailyniche/internal/repos"
)

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
