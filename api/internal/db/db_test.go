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

// TestMigrate_RecordsMigrationsInSchemaMigrationsTable and
// TestMigrate_DoesNotReapplyOnSecondCall exercise the real embedded
// migrations/*.sql (not a fabricated fs.FS like migration_test.go uses), so
// they're coupled to what's actually in that directory right now: two
// migrations, 0001_init.sql and 0002_add_image_url.sql, hence "versions 1
// and 2" and "exactly two rows" below. Adding a third real migration means
// updating both again - that's expected, not a sign of flakiness.

func TestMigrate_RecordsMigrationsInSchemaMigrationsTable(t *testing.T) {
	// given: a fresh database
	conn, err := Init(":memory:")
	if err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}
	defer conn.Close()

	// when: we migrate
	if err := Migrate(conn); err != nil {
		t.Fatalf("Migrate() returned error: %v", err)
	}

	// then: both migrations are recorded as applied
	applied, err := appliedVersions(conn)
	if err != nil {
		t.Fatalf("appliedVersions() returned error: %v", err)
	}
	if !applied[1] || !applied[2] {
		t.Errorf("expected versions 1 (init) and 2 (add_image_url) to be recorded as applied, got %v", applied)
	}
}

func TestMigrate_DoesNotReapplyOnSecondCall(t *testing.T) {
	// given: a database that has already been migrated once
	conn, err := Init(":memory:")
	if err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}
	defer conn.Close()
	if err := Migrate(conn); err != nil {
		t.Fatalf("first Migrate() returned error: %v", err)
	}

	// when: we migrate again
	if err := Migrate(conn); err != nil {
		t.Fatalf("second Migrate() returned error: %v", err)
	}

	// then: schema_migrations still has exactly two rows, not duplicates.
	// (If the applied[m.version] guard in Migrate() were ever broken, this
	// second call would fail above with a primary-key constraint error
	// before ever reaching this count check.)
	var count int
	row := conn.QueryRow("SELECT COUNT(*) FROM schema_migrations")
	if err := row.Scan(&count); err != nil {
		t.Fatalf("failed to count schema_migrations rows: %v", err)
	}
	if count != 2 {
		t.Errorf("expected exactly 2 recorded migrations, got %d", count)
	}
}

func TestMigrate_AddsImageURLColumnToPosts(t *testing.T) {
	// given: a fresh database
	conn, err := Init(":memory:")
	if err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}
	defer conn.Close()

	// when: we migrate
	if err := Migrate(conn); err != nil {
		t.Fatalf("Migrate() returned error: %v", err)
	}

	// then: the posts table has an image_url column
	rows, err := conn.Query("PRAGMA table_info(posts)")
	if err != nil {
		t.Fatalf("failed to query posts table info: %v", err)
	}
	defer rows.Close()

	found := false
	for rows.Next() {
		var cid, notNull, pk int
		var name, colType string
		var dfltValue any
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
			t.Fatalf("failed to scan table info row: %v", err)
		}
		if name == "image_url" {
			found = true
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err(): %v", err)
	}
	if !found {
		t.Error("expected posts table to have an image_url column")
	}
}
