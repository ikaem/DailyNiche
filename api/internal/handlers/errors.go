package handlers

import (
	"encoding/json"
	"net/http"
)

// errorResponse is the JSON shape of every error response - consistent with
// every success response in this API also being JSON.
type errorResponse struct {
	Error string `json:"error"`
}

// writeError writes a JSON error response with the given message and
// status. The JSON-equivalent of http.Error, used throughout this package
// instead of it so error responses are reliably parseable JSON (callers can
// always decode `{"error": "..."}`) rather than http.Error's plain text
// body.
func writeError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorResponse{Error: message})
}
