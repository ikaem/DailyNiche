package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInit_CreatesDatabaseFile(t *testing.T) {
	// given: a path to a database file that doesn't exist yet
	path := filepath.Join(t.TempDir(), "test.db")

	// when: we initialize and migrate a database at that path
	conn, err := Init(path)
	if err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}
	defer conn.Close()

	if err := Migrate(conn); err != nil {
		t.Fatalf("Migrate() returned error: %v", err)
	}

	// then: the database file exists on disk
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected database file to exist at %s: %v", path, err)
	}
}

func TestInit_SubsequentRunsDontError(t *testing.T) {
	// given: a database that has already been initialized and migrated once
	path := filepath.Join(t.TempDir(), "test.db")

	conn1, err := Init(path)
	if err != nil {
		t.Fatalf("first Init() returned error: %v", err)
	}
	if err := Migrate(conn1); err != nil {
		t.Fatalf("first Migrate() returned error: %v", err)
	}
	conn1.Close()

	// when: we initialize and migrate the same path again
	conn2, err := Init(path)
	if err != nil {
		t.Fatalf("second Init() returned error: %v", err)
	}
	defer conn2.Close()

	// then: neither call errors
	if err := Migrate(conn2); err != nil {
		t.Fatalf("second Migrate() returned error: %v", err)
	}
}

func TestMigrate_CreatesExpectedTables(t *testing.T) {
	// given: a fresh in-memory database
	conn, err := Init(":memory:")
	if err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}
	defer conn.Close()

	// when: we run the migration
	if err := Migrate(conn); err != nil {
		t.Fatalf("Migrate() returned error: %v", err)
	}

	// then: both the feeds and posts tables exist
	for _, table := range []string{"feeds", "posts"} {
		var name string
		row := conn.QueryRow("SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?", table)
		if err := row.Scan(&name); err != nil {
			t.Errorf("expected table %q to exist: %v", table, err)
		}
	}
}

func TestOpen_InitializesAndMigratesInOneCall(t *testing.T) {
	// given: a path to a database file that doesn't exist yet
	path := filepath.Join(t.TempDir(), "test.db")

	// when: we call Open
	conn, err := Open(path)
	if err != nil {
		t.Fatalf("Open() returned error: %v", err)
	}
	defer conn.Close()

	// then: the database file exists and the schema has been applied
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected database file to exist at %s: %v", path, err)
	}
	for _, table := range []string{"feeds", "posts"} {
		var name string
		row := conn.QueryRow("SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?", table)
		if err := row.Scan(&name); err != nil {
			t.Errorf("expected table %q to exist: %v", table, err)
		}
	}
}

func TestMigrate_IsIdempotent(t *testing.T) {
	// given: a database that has already been migrated once
	conn, err := Init(":memory:")
	if err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}
	defer conn.Close()

	if err := Migrate(conn); err != nil {
		t.Fatalf("first Migrate() returned error: %v", err)
	}

	// when: we run the migration again on the same connection
	// then: it does not error
	if err := Migrate(conn); err != nil {
		t.Fatalf("second Migrate() returned error: %v", err)
	}
}
