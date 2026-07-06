package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth_RespondsOKWithStatusJSON(t *testing.T) {
	// given: a request to the health endpoint
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	// when: the handler processes the request
	Health(rec, req)

	// then: it responds with 200 and a JSON status body
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	wantBody := "{\"status\":\"ok\"}\n"
	if rec.Body.String() != wantBody {
		t.Fatalf("expected body %q, got %q", wantBody, rec.Body.String())
	}
	wantContentType := "application/json"
	if got := rec.Header().Get("Content-Type"); got != wantContentType {
		t.Fatalf("expected Content-Type %q, got %q", wantContentType, got)
	}
}
