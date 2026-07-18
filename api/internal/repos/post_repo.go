package repos

import (
	"database/sql"
	"time"

	"github.com/karlo/dailyniche/internal/models"
)

// CreatePost inserts post if its GUID isn't already present. Returns the
// newly assigned ID if a row was inserted, or 0 if the GUID already existed
// - a duplicate GUID is not an error, it's silently skipped, since feeds
// commonly re-list posts we've already stored on later fetches.
func CreatePost(conn *sql.DB, post *models.Post) (int64, error) {
	result, err := conn.Exec(
		`INSERT INTO posts (feed_id, title, url, content_summary, image_url, published_at, fetched_at, guid, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(guid) DO NOTHING`,
		post.FeedID, post.Title, post.URL, post.ContentSummary, post.ImageURL, post.PublishedAt, post.FetchedAt, post.GUID, time.Now().UTC(),
	)
	if err != nil {
		return 0, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	if affected == 0 {
		return 0, nil
	}

	return result.LastInsertId()
}

// ListPostsByDate returns posts fetched during the UTC calendar day that
// date falls on, newest published first. This is what a daily magazine
// issue is built from - fetched_at is when we discovered the post, not
// when the feed originally published it.
func ListPostsByDate(conn *sql.DB, date time.Time) ([]models.Post, error) {
	// Example: date passed in as 2024-03-15 22:00:00 -05:00 (i.e. 10pm in a
	// UTC-5 zone).
	date = date.UTC() // 2024-03-16 03:00:00 UTC - crossed into the next day
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	// start = 2024-03-16 00:00:00 UTC
	end := start.Add(24 * time.Hour)
	// end   = 2024-03-17 00:00:00 UTC

	rows, err := conn.Query(
		`SELECT id, feed_id, title, url, content_summary, image_url, published_at, fetched_at, guid, created_at
		 FROM posts
		 WHERE fetched_at >= ? AND fetched_at < ?
		 ORDER BY published_at DESC`,
		start, end,
	)
	if err != nil {
		return nil, err
	}
	return scanPosts(rows)
}

// ListPostsByFeed returns every post for feedID, newest published first.
func ListPostsByFeed(conn *sql.DB, feedID int64) ([]models.Post, error) {
	rows, err := conn.Query(
		`SELECT id, feed_id, title, url, content_summary, image_url, published_at, fetched_at, guid, created_at
		 FROM posts
		 WHERE feed_id = ?
		 ORDER BY published_at DESC`,
		feedID,
	)
	if err != nil {
		return nil, err
	}
	return scanPosts(rows)
}

// DeletePostsByDate deletes posts fetched before the given time. Intended
// for deliberate, manually-triggered retention cleanup - never called as
// part of the normal fetch/serve flow, since silently removing past posts
// would violate "past issues never change".
func DeletePostsByDate(conn *sql.DB, before time.Time) error {
	_, err := conn.Exec(`DELETE FROM posts WHERE fetched_at < ?`, before.UTC())
	return err
}

// scanPosts reads every remaining row from rows into a []models.Post,
// closing rows before returning.
func scanPosts(rows *sql.Rows) ([]models.Post, error) {
	defer rows.Close()

	posts := []models.Post{}
	for rows.Next() {
		var p models.Post
		if err := rows.Scan(&p.ID, &p.FeedID, &p.Title, &p.URL, &p.ContentSummary, &p.ImageURL, &p.PublishedAt, &p.FetchedAt, &p.GUID, &p.CreatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}
