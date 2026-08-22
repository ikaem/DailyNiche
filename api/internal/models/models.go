package models

import "time"

type Feed struct {
	ID         int64
	Name       string
	URL        string
	DisabledAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Post struct {
	ID             int64
	FeedID         int64
	Title          string
	URL            string
	ContentSummary string
	ImageURL       string
	PublishedAt    time.Time
	FetchedAt      time.Time
	GUID           string
	CreatedAt      time.Time
	// FavoritedAt/ReadLaterAt mirror SavedPost's columns, populated via a
	// LEFT JOIN against saved_posts (see ListPostsByDate) - nil for the
	// vast majority of posts, which have no saved_posts row at all.
	FavoritedAt *time.Time
	ReadLaterAt *time.Time
}

// SavedPost is a post's favorite/read-later state - a 1:1 extension of
// Post, not an independent entity, hence PostID doubles as its own primary
// key rather than having a separate ID (see saved_posts migration notes).
type SavedPost struct {
	PostID      int64
	FavoritedAt *time.Time
	ReadLaterAt *time.Time
}
