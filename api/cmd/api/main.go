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

	// HandleFunc registers into net/http's global DefaultServeMux, not a mux
	// we control. Fine for this one learning-exercise route; Phase 4.1 will
	// switch to an explicit http.NewServeMux() once middleware/more routes
	// are added, to avoid relying on shared global state.
	http.HandleFunc("/health", handlers.Health)
	http.HandleFunc("/api/posts", handlers.Posts(conn))
	// GET and POST share the same literal path, so both need an explicit
	// method prefix - registering the same bare "/api/feeds" pattern twice
	// would panic at startup.
	http.HandleFunc("GET /api/feeds", handlers.Feeds(conn))
	http.HandleFunc("POST /api/feeds", handlers.CreateFeed(conn))
	http.HandleFunc("DELETE /api/feeds/{id}", handlers.DeleteFeed(conn))

	log.Printf("DailyNiche API server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
