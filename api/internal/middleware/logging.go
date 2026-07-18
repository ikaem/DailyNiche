package middleware

import (
	"log"
	"net/http"
	"time"
)

// statusRecorder wraps http.ResponseWriter to capture the status code
// written to it - the standard library's ResponseWriter has no way to read
// back what was already written, so wrapping it is the standard way to
// observe it after the fact.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

// WriteHeader records status before delegating to the embedded
// ResponseWriter, so Logging can log it once the handler returns.
func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// Logging wraps next, logging each request's method, path, status code, and
// duration once the request completes.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// Defaults to 200: if a handler writes a body without ever calling
		// WriteHeader explicitly, net/http implicitly sends 200 OK on the
		// first Write - this default mirrors that real behavior.
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		log.Printf("%s %s %d %s", r.Method, r.URL.Path, rec.status, time.Since(start))
	})
}
