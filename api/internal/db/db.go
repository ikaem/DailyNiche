package db

import (
	"database/sql"
	"embed"
	"fmt"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

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

// Migrate applies any migrations under migrations/ not yet recorded in
// schema_migrations, in version order. Safe to call repeatedly.
func Migrate(conn *sql.DB) error {
	if err := ensureSchemaMigrationsTable(conn); err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	migrations, err := loadMigrations(migrationsFS)
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	applied, err := appliedVersions(conn)
	if err != nil {
		return fmt.Errorf("failed to read applied migrations: %w", err)
	}

	for _, m := range migrations {
		if applied[m.version] {
			continue
		}
		if err := applyMigration(conn, m); err != nil {
			return fmt.Errorf("failed to apply migration %04d_%s: %w", m.version, m.name, err)
		}
	}
	return nil
}

// Open initializes and migrates the database in one call - the common path
// entry points (cmd/api, cmd/fetcher) use to get a ready-to-use connection.
func Open(path string) (*sql.DB, error) {
	conn, err := Init(path)
	if err != nil {
		return nil, err
	}
	if err := Migrate(conn); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}
