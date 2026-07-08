package handlers

import "time"

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
