package repos

import (
	"database/sql"
	"time"

	"github.com/karlo/dailyniche/internal/models"
)

// CreateFeed inserts a new feed and returns its assigned ID.
func CreateFeed(conn *sql.DB, name, url string) (int64, error) {
	now := time.Now().UTC()
	result, err := conn.Exec(
		`INSERT INTO feeds (name, url, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		name, url, now, now,
	)
	if err != nil {
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
