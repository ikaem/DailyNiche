package db

import (
	"database/sql"
	_ "embed"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

// Init opens (creating if necessary) the SQLite database at path.
func Init(path string) (*sql.DB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

// Migrate applies schema.sql. Safe to call repeatedly - every statement in
// schema.sql uses IF NOT EXISTS.
func Migrate(conn *sql.DB) error {
	_, err := conn.Exec(schemaSQL)
	return err
}
