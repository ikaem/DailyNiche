package repos

import (
	"database/sql"
	"errors"
	"time"

	"github.com/karlo/dailyniche/internal/models"
)

// GetSavedPost returns postID's favorite/read-later state. If no row
// exists yet, it returns a zero-value SavedPost (both timestamps nil), not
// an error - "never saved" is the default, expected state for almost
// every post, not an exceptional one.
func GetSavedPost(conn *sql.DB, postID int64) (*models.SavedPost, error) {
	sp := &models.SavedPost{PostID: postID}
	err := conn.QueryRow(
		`SELECT favorited_at, read_later_at FROM saved_posts WHERE post_id = ?`,
		postID,
	).Scan(&sp.FavoritedAt, &sp.ReadLaterAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sp, nil
		}
		return nil, err
	}
	return sp, nil
}

// FavoritePost marks postID as favorited. Upserts rather than a plain
// UPDATE, since a post may have no saved_posts row at all yet - read_later_at,
// if already set, is left untouched, since the two states are independent
// (a post can be both favorited and read-later at once).
func FavoritePost(conn *sql.DB, postID int64) error {
	_, err := conn.Exec(
		`INSERT INTO saved_posts (post_id, favorited_at) VALUES (?, ?)
		 ON CONFLICT (post_id) DO UPDATE SET favorited_at = excluded.favorited_at`,
		postID, time.Now().UTC(),
	)
	return err
}

// UnfavoritePost clears postID's favorited_at. A plain UPDATE, unlike
// FavoritePost - unfavoriting a post with no saved_posts row at all (never
// favorited) is a harmless no-op, not an error.
func UnfavoritePost(conn *sql.DB, postID int64) error {
	_, err := conn.Exec(`UPDATE saved_posts SET favorited_at = NULL WHERE post_id = ?`, postID)
	return err
}

// MarkReadLater marks postID as read-later - upserts for the same reason
// FavoritePost does.
func MarkReadLater(conn *sql.DB, postID int64) error {
	_, err := conn.Exec(
		`INSERT INTO saved_posts (post_id, read_later_at) VALUES (?, ?)
		 ON CONFLICT (post_id) DO UPDATE SET read_later_at = excluded.read_later_at`,
		postID, time.Now().UTC(),
	)
	return err
}

// UnmarkReadLater clears postID's read_later_at - same no-op-if-missing
// reasoning as UnfavoritePost.
func UnmarkReadLater(conn *sql.DB, postID int64) error {
	_, err := conn.Exec(`UPDATE saved_posts SET read_later_at = NULL WHERE post_id = ?`, postID)
	return err
}
