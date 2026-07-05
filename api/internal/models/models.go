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
	PublishedAt    time.Time
	FetchedAt      time.Time
	GUID           string
	CreatedAt      time.Time
}
