package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/karlo/dailyniche/internal/repos"
)

// SavedPostResponse is the JSON shape returned after toggling a post's
// favorite/read-later state - just the two saved-state fields, not a full
// PostResponse, since the caller already has the rest of the post
// rendered and only needs the updated state back.
type SavedPostResponse struct {
	FavoritedAt *time.Time `json:"favorited_at"`
	ReadLaterAt *time.Time `json:"read_later_at"`
}

// savedPostToggleHandler builds a handler that toggles one saved-state
// column for the post at {id}: checks the post exists (404 otherwise),
// runs mutate, then responds with the post's current saved state. Shared
// by FavoritePost/UnfavoritePost/MarkReadLater/UnmarkReadLater, which
// differ only in which repos function they call.
func savedPostToggleHandler(conn *sql.DB, mutate func(*sql.DB, int64) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeError(w, "invalid post id", http.StatusBadRequest)
			return
		}

		if _, err := repos.GetPost(conn, id); err != nil {
			writeError(w, "post not found", http.StatusNotFound)
			return
		}

		if err := mutate(conn, id); err != nil {
			writeError(w, "failed to update saved state", http.StatusInternalServerError)
			return
		}

		sp, err := repos.GetSavedPost(conn, id)
		if err != nil {
			writeError(w, "failed to load saved state", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(SavedPostResponse{
			FavoritedAt: sp.FavoritedAt,
			ReadLaterAt: sp.ReadLaterAt,
		})
	}
}

// FavoritePost returns an http.HandlerFunc for POST /api/posts/{id}/favorite.
func FavoritePost(conn *sql.DB) http.HandlerFunc {
	return savedPostToggleHandler(conn, repos.FavoritePost)
}

// UnfavoritePost returns an http.HandlerFunc for DELETE /api/posts/{id}/favorite.
func UnfavoritePost(conn *sql.DB) http.HandlerFunc {
	return savedPostToggleHandler(conn, repos.UnfavoritePost)
}

// MarkReadLater returns an http.HandlerFunc for POST /api/posts/{id}/read-later.
func MarkReadLater(conn *sql.DB) http.HandlerFunc {
	return savedPostToggleHandler(conn, repos.MarkReadLater)
}

// UnmarkReadLater returns an http.HandlerFunc for DELETE /api/posts/{id}/read-later.
func UnmarkReadLater(conn *sql.DB) http.HandlerFunc {
	return savedPostToggleHandler(conn, repos.UnmarkReadLater)
}
