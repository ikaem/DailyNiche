package handlers

import (
	"encoding/json"
	"net/http"
)

// Health reports that the server is up and reachable.
// Signature must match http.HandlerFunc: func(http.ResponseWriter, *http.Request).
func Health(w http.ResponseWriter, r *http.Request) {
	// Set a response header. Only mutates an in-memory map on w - nothing is
	// sent to the client yet. Must happen before WriteHeader/Write, since
	// headers are flushed to the wire at that point and can't be changed after.
	w.Header().Set("Content-Type", "application/json")

	// Send the status line (e.g. "HTTP/1.1 200 OK") plus the headers set above.
	// Optional in this case (Write/Encode would default to 200 if we skipped
	// this), but explicit here for clarity.
	w.WriteHeader(http.StatusOK)

	// Stream a JSON-encoded response body directly to the client via w.
	// w satisfies io.Writer, which is why the encoder can target it directly.
	// Encode adds a trailing newline after the JSON - expected, not a bug.
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
