package main

import (
	"log"
	"net/http"
	"os"

	"github.com/karlo/dailyniche/internal/db"
	"github.com/karlo/dailyniche/internal/handlers"
)

func main() {
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
	mux.HandleFunc("DELETE /api/feeds/{id}", handlers.DeleteFeed(conn))

	log.Printf("DailyNiche API server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
