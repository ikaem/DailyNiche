package main

import (
	"log"
	"os"

	"github.com/karlo/dailyniche/internal/db"
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

	log.Printf("DailyNiche API server: database ready at %s", dbPath)
}
