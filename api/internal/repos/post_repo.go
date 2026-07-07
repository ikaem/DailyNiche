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
		`INSERT INTO posts (feed_id, title, url, content_summary, published_at, fetched_at, guid, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(guid) DO NOTHING`,
		post.FeedID, post.Title, post.URL, post.ContentSummary, post.PublishedAt, post.FetchedAt, post.GUID, time.Now().UTC(),
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
