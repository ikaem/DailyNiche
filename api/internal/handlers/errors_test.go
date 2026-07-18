package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteError_WritesJSONErrorBodyAndStatus(t *testing.T) {
	// given: a response recorder
	rec := httptest.NewRecorder()

	// when: we write an error
	writeError(rec, "name is required", http.StatusBadRequest)

	// then: the status code is set
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	// and: the Content-Type is JSON, not http.Error's plain text
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("expected Content-Type %q, got %q", "application/json", got)
	}

	// and: the body is valid JSON matching {"error": "<message>"}
	var got errorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode error response as JSON: %v", err)
	}
	if got.Error != "name is required" {
		t.Errorf("expected error message %q, got %q", "name is required", got.Error)
	}
}
