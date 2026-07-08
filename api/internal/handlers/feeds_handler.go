package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/karlo/dailyniche/internal/models"
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

// toFeedResponse converts a stored Feed into its API response shape.
func toFeedResponse(f *models.Feed) FeedResponse {
	return FeedResponse{
		ID:         f.ID,
		Name:       f.Name,
		URL:        f.URL,
		DisabledAt: f.DisabledAt,
		CreatedAt:  f.CreatedAt,
		UpdatedAt:  f.UpdatedAt,
	}
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
			responses = append(responses, toFeedResponse(&f))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(responses)
	}
}

// createFeedRequest is the expected JSON body for POST /api/feeds.
type createFeedRequest struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// CreateFeed returns an http.HandlerFunc for POST /api/feeds, backed by conn.
func CreateFeed(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createFeedRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}

		name := strings.TrimSpace(req.Name)
		if name == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}

		rawURL := strings.TrimSpace(req.URL)
		if rawURL == "" {
			http.Error(w, "url is required", http.StatusBadRequest)
			return
		}
		parsedURL, err := url.ParseRequestURI(rawURL)
		if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
			http.Error(w, "url must be a valid absolute URL", http.StatusBadRequest)
			return
		}

		id, err := repos.CreateFeed(conn, name, rawURL)
		if err != nil {
			if errors.Is(err, repos.ErrDuplicateURL) {
				http.Error(w, "a feed with this url already exists", http.StatusConflict)
				return
			}
			http.Error(w, "failed to create feed", http.StatusInternalServerError)
			return
		}

		feed, err := repos.GetFeed(conn, id)
		if err != nil {
			http.Error(w, "failed to load created feed", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(toFeedResponse(feed))
	}
}
