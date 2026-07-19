package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/karlo/dailyniche/internal/fetcher"
)

// FetchSummaryResponse is the JSON shape returned after an on-demand fetch.
type FetchSummaryResponse struct {
	New        int `json:"new"`
	Duplicates int `json:"duplicates"`
	Errors     int `json:"errors"`
}

// Fetch returns an http.HandlerFunc for POST /api/fetch, backed by conn.
// Runs fetcher.FetchAll synchronously and returns its summary once the
// whole fetch completes - the simple, blocking approach decided in
// CLAUDE.md's Task 5.3, not async job-tracking/polling. A personal project
// with a handful of feeds likely fetches in well under a few seconds;
// revisit only if real usage shows otherwise.
func Fetch(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		summary, err := fetcher.FetchAll(conn, fetcher.Options{})
		if err != nil {
			writeError(w, "failed to fetch feeds", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(FetchSummaryResponse{
			New:        summary.New,
			Duplicates: summary.Duplicates,
			Errors:     summary.Errors,
		})
	}
}
