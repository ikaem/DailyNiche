package repos

import (
	"database/sql"
	"errors"
	"time"

	"modernc.org/sqlite"

	"github.com/karlo/dailyniche/internal/models"
)

// ErrDuplicateURL is returned by CreateFeed when a feed with the same URL
// already exists (feeds.url is UNIQUE).
var ErrDuplicateURL = errors.New("feed url already exists")

// sqliteConstraintUniqueCode is SQLite's stable, public extended result
// code for a UNIQUE constraint violation specifically (SQLITE_CONSTRAINT_
// UNIQUE) - see sqlite.org/rescode.html. SQLite errors carry a primary code
// (e.g. generic SQLITE_CONSTRAINT = 19) and often a more specific extended
// code (SQLITE_CONSTRAINT_UNIQUE = 2067); the driver surfaces the extended
// one, which is what a UNIQUE violation actually reports here. Referenced
// by value rather than importing modernc.org/sqlite/lib (an internal
// implementation detail of the driver) just for one constant.
const sqliteConstraintUniqueCode = 2067

// CreateFeed inserts a new feed and returns its assigned ID. Returns
// ErrDuplicateURL, not a raw driver error, if a feed with this URL already
// exists - callers check with errors.Is(err, ErrDuplicateURL), never
// needing to know a SQLite driver is involved at all.
func CreateFeed(conn *sql.DB, name, url string) (int64, error) {
	now := time.Now().UTC()
	result, err := conn.Exec(
		`INSERT INTO feeds (name, url, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		name, url, now, now,
	)
	if err != nil {
		// sqliteErr: destination pointer for errors.As to fill in below.
		var sqliteErr *sqlite.Error
		// errors.As checks whether err *or anything it wraps* is a
		// *sqlite.Error, unlike a raw type assertion which only matches err
		// itself. Code() still must be checked - this type covers every
		// SQLite failure kind, not just constraint violations.
		if errors.As(err, &sqliteErr) && sqliteErr.Code() == sqliteConstraintUniqueCode {
			return 0, ErrDuplicateURL
		}
		return 0, err
	}
	return result.LastInsertId()
}

// ListFeeds returns every feed, active and disabled, ordered by name.
func ListFeeds(conn *sql.DB) ([]models.Feed, error) {
	rows, err := conn.Query(`SELECT id, name, url, disabled_at, created_at, updated_at FROM feeds ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	feeds := []models.Feed{}
	for rows.Next() {
		var f models.Feed
		// Scan args are purely positional - they map to the SELECT columns
		// above left to right, not by column/field name. Keep this order in
		// sync with the SELECT whenever either one changes.
		if err := rows.Scan(&f.ID, &f.Name, &f.URL, &f.DisabledAt, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		feeds = append(feeds, f)
	}
	// rows.Next() returning false is ambiguous: ran out of rows (fine) vs a
	// real error mid-stream (not fine). rows.Err() disambiguates - always
	// check it after a Next() loop ends, easy to forget since it never
	// triggers in normal local testing.
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return feeds, nil
}

// GetFeed returns the feed with the given ID.
func GetFeed(conn *sql.DB, id int64) (*models.Feed, error) {
	var f models.Feed
	// QueryRow itself can't fail synchronously - unlike Query, it has no
	// error return here. Any error (including "no row matched", surfaced as
	// sql.ErrNoRows) only appears once we call .Scan() below.
	err := conn.QueryRow(
		`SELECT id, name, url, disabled_at, created_at, updated_at FROM feeds WHERE id = ?`,
		id,
	).Scan(&f.ID, &f.Name, &f.URL, &f.DisabledAt, &f.CreatedAt, &f.UpdatedAt)
	// If no feed matches id, Scan returns sql.ErrNoRows here - that's the
	// mechanism by which a nonexistent ID surfaces as an error.
	if err != nil {
		return nil, err
	}
	return &f, nil
}

// UpdateFeed updates a feed's name and url, and refreshes its updated_at.
func UpdateFeed(conn *sql.DB, feed *models.Feed) error {
	feed.UpdatedAt = time.Now().UTC()
	_, err := conn.Exec(
		`UPDATE feeds SET name = ?, url = ?, updated_at = ? WHERE id = ?`,
		feed.Name, feed.URL, feed.UpdatedAt, feed.ID,
	)
	return err
}

// DeleteFeed soft-deletes a feed by setting disabled_at rather than removing
// the row - past posts must keep resolving feed_id so archived issues never
// change (see CLAUDE.md: "Feed Deletion is a Soft Delete").
func DeleteFeed(conn *sql.DB, id int64) error {
	now := time.Now().UTC()
	_, err := conn.Exec(
		`UPDATE feeds SET disabled_at = ?, updated_at = ? WHERE id = ?`,
		now, now, id,
	)
	return err
}
