package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/karlo/dailyniche/internal/repos"
)

// FeedResponse is the JSON shape returned for each feed. Kept as its own
// type (rather than tagging models.Feed directly) for symmetry with
// PostResponse, even though nothing currently differs from the stored
// model - revisit if that ever changes.
type FeedResponse struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	URL        string     `json:"url"`
	DisabledAt *time.Time `json:"disabled_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// Feeds returns an http.HandlerFunc for GET /api/feeds, backed by conn.
func Feeds(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		feedList, err := repos.ListFeeds(conn)
		if err != nil {
			http.Error(w, "failed to list feeds", http.StatusInternalServerError)
			return
		}

		responses := make([]FeedResponse, 0, len(feedList))
		for _, f := range feedList {
			responses = append(responses, FeedResponse{
				ID:         f.ID,
				Name:       f.Name,
				URL:        f.URL,
				DisabledAt: f.DisabledAt,
				CreatedAt:  f.CreatedAt,
				UpdatedAt:  f.UpdatedAt,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(responses)
	}
}
