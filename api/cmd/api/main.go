package main

import (
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/karlo/dailyniche/internal/db"
	"github.com/karlo/dailyniche/internal/handlers"
	"github.com/karlo/dailyniche/internal/middleware"
)

// configureLogging installs a slog default logger writing to out - mirrors
// cmd/fetcher's own configureLogging. Needed so fetcher.FetchAll's log
// output (Warn on a per-feed failure, Info on start/completion) lands
// somewhere durable when triggered via the on-demand POST /api/fetch
// endpoint, instead of only the container's stdout, which is lost the next
// time the container gets recreated (e.g. on every deploy).
func configureLogging(out io.Writer) {
	slog.SetDefault(slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{Level: slog.LevelInfo})))
}

func main() {
	logPath := os.Getenv("LOG_PATH")
	if logPath == "" {
		logPath = "api.log"
	}
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	defer logFile.Close()
	configureLogging(io.MultiWriter(os.Stdout, logFile))

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "dailyniche.db"
	}

	conn, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer conn.Close()

	log.Printf("database ready at %s", dbPath)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// An explicit mux, not net/http's global DefaultServeMux - needed to
	// cleanly wrap routes with middleware (logging, added next) without
	// relying on shared global state.
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handlers.Health)
	mux.HandleFunc("/api/posts", handlers.Posts(conn))
	// GET and POST share the same literal path, so both need an explicit
	// method prefix - registering the same bare "/api/feeds" pattern twice
	// would panic at startup.
	mux.HandleFunc("GET /api/feeds", handlers.Feeds(conn))
	mux.HandleFunc("POST /api/feeds", handlers.CreateFeed(conn))
	mux.HandleFunc("PUT /api/feeds/{id}", handlers.UpdateFeed(conn))
	mux.HandleFunc("DELETE /api/feeds/{id}", handlers.DeleteFeed(conn))
	mux.HandleFunc("POST /api/feeds/{id}/enable", handlers.EnableFeed(conn))
	mux.HandleFunc("POST /api/fetch", handlers.Fetch(conn))
	mux.HandleFunc("POST /api/posts/{id}/favorite", handlers.FavoritePost(conn))
	mux.HandleFunc("DELETE /api/posts/{id}/favorite", handlers.UnfavoritePost(conn))
	mux.HandleFunc("POST /api/posts/{id}/read-later", handlers.MarkReadLater(conn))
	mux.HandleFunc("DELETE /api/posts/{id}/read-later", handlers.UnmarkReadLater(conn))

	log.Printf("DailyNiche API server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, middleware.Logging(mux)); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
