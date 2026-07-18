package middleware

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// captureLog redirects the standard logger's output to a strings.Builder
// for the duration of the test, restoring it to log's real default (stderr)
// afterward.
func captureLog(t *testing.T) *strings.Builder {
	t.Helper()
	var buf strings.Builder
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
	return &buf
}

func TestLogging_LogsMethodPathAndStatus(t *testing.T) {
	// given: a handler that responds 201, wrapped in Logging
	logged := captureLog(t)
	handler := Logging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	req := httptest.NewRequest(http.MethodPost, "/api/feeds", nil)
	rec := httptest.NewRecorder()

	// when: the wrapped handler serves the request
	handler.ServeHTTP(rec, req)

	// then: the log line contains the method, path, and status code
	if got := logged.String(); !strings.Contains(got, "POST") || !strings.Contains(got, "/api/feeds") || !strings.Contains(got, "201") {
		t.Errorf("expected log to contain method, path, and status, got %q", got)
	}

	fmt.Println("hello")

}

func TestLogging_DefaultsStatusTo200WhenHandlerNeverCallsWriteHeader(t *testing.T) {
	// given: a handler that writes a body without ever calling WriteHeader
	logged := captureLog(t)
	handler := Logging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	// when: the wrapped handler serves the request
	handler.ServeHTTP(rec, req)

	// then: the log line reports the implicit 200, matching real net/http behavior
	if got := logged.String(); !strings.Contains(got, "200") {
		t.Errorf("expected log to contain status 200, got %q", got)
	}
}

func TestLogging_StillWritesResponseBodyToClient(t *testing.T) {
	// given: a handler that writes a known body
	captureLog(t)
	handler := Logging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello"))
	}))
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	// when: the wrapped handler serves the request
	handler.ServeHTTP(rec, req)

	// then: wrapping for logging doesn't interfere with the actual response
	if rec.Body.String() != "hello" {
		t.Errorf("expected body %q, got %q", "hello", rec.Body.String())
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}
