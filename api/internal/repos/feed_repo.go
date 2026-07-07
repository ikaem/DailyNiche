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
