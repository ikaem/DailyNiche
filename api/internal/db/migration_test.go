package db

import (
	"testing"
	"testing/fstest"
)

func TestParseMigrationFilename_ValidFilename(t *testing.T) {
	// given: a well-formed migration filename
	filename := "0002_add_image_url.sql"

	// when: we parse it
	version, name, err := parseMigrationFilename(filename)

	// then: version and name are extracted correctly
	if err != nil {
		t.Fatalf("parseMigrationFilename() returned error: %v", err)
	}
	if version != 2 {
		t.Errorf("expected version 2, got %d", version)
	}
	if name != "add_image_url" {
		t.Errorf("expected name %q, got %q", "add_image_url", name)
	}
}

func TestParseMigrationFilename_MissingNumericPrefix(t *testing.T) {
	// given: a filename with no numeric prefix
	filename := "init.sql"

	// when: we parse it
	_, _, err := parseMigrationFilename(filename)

	// then: it returns an error
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}

func TestParseMigrationFilename_MissingDescription(t *testing.T) {
	// given: a filename with a version but no description
	filename := "0001.sql"

	// when: we parse it
	_, _, err := parseMigrationFilename(filename)

	// then: it returns an error
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}

func TestLoadMigrations_SortsByVersionRegardlessOfFileOrder(t *testing.T) {
	// given: migration files added to the fake filesystem out of version order
	fsys := fstest.MapFS{
		"migrations/0002_second.sql": &fstest.MapFile{Data: []byte("-- second")},
		"migrations/0001_first.sql":  &fstest.MapFile{Data: []byte("-- first")},
	}

	// when: we load migrations
	migrations, err := loadMigrations(fsys)

	// then: they come back sorted by version, not filesystem order
	if err != nil {
		t.Fatalf("loadMigrations() returned error: %v", err)
	}
	if len(migrations) != 2 {
		t.Fatalf("expected 2 migrations, got %d", len(migrations))
	}
	if migrations[0].version != 1 || migrations[1].version != 2 {
		t.Fatalf("expected versions [1, 2], got [%d, %d]", migrations[0].version, migrations[1].version)
	}
}

func TestLoadMigrations_ReadsFileContents(t *testing.T) {
	// given: a single migration file with known contents
	fsys := fstest.MapFS{
		"migrations/0001_init.sql": &fstest.MapFile{Data: []byte("CREATE TABLE example (id INTEGER);")},
	}

	// when: we load migrations
	migrations, err := loadMigrations(fsys)

	// then: the file's contents and parsed name are both present
	if err != nil {
		t.Fatalf("loadMigrations() returned error: %v", err)
	}
	if migrations[0].name != "init" {
		t.Errorf("expected name %q, got %q", "init", migrations[0].name)
	}
	if migrations[0].sql != "CREATE TABLE example (id INTEGER);" {
		t.Errorf("unexpected sql contents: %q", migrations[0].sql)
	}
}

func TestLoadMigrations_ReturnsErrorForMalformedFilename(t *testing.T) {
	// given: a migration file with a bad filename
	fsys := fstest.MapFS{
		"migrations/init.sql": &fstest.MapFile{Data: []byte("-- bad name")},
	}

	// when: we load migrations
	_, err := loadMigrations(fsys)

	// then: it returns an error rather than silently skipping the file
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}

func TestEnsureSchemaMigrationsTable_CreatesTable(t *testing.T) {
	// given: a fresh in-memory database
	conn, err := Init(":memory:")
	if err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}
	defer conn.Close()

	// when: we ensure the schema_migrations table
	if err := ensureSchemaMigrationsTable(conn); err != nil {
		t.Fatalf("ensureSchemaMigrationsTable() returned error: %v", err)
	}

	// then: the table exists
	var name string
	row := conn.QueryRow("SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?", "schema_migrations")
	if err := row.Scan(&name); err != nil {
		t.Fatalf("expected schema_migrations table to exist: %v", err)
	}
}

func TestEnsureSchemaMigrationsTable_IsIdempotent(t *testing.T) {
	// given: a database where the table has already been ensured once
	conn, err := Init(":memory:")
	if err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}
	defer conn.Close()
	if err := ensureSchemaMigrationsTable(conn); err != nil {
		t.Fatalf("first ensureSchemaMigrationsTable() returned error: %v", err)
	}

	// when: we ensure it again
	// then: it does not error
	if err := ensureSchemaMigrationsTable(conn); err != nil {
		t.Fatalf("second ensureSchemaMigrationsTable() returned error: %v", err)
	}
}

func TestAppliedVersions_ReturnsEmptyWhenNoneRecorded(t *testing.T) {
	// given: a fresh schema_migrations table with no rows
	conn, err := Init(":memory:")
	if err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}
	defer conn.Close()
	if err := ensureSchemaMigrationsTable(conn); err != nil {
		t.Fatalf("ensureSchemaMigrationsTable() returned error: %v", err)
	}

	// when: we read applied versions
	applied, err := appliedVersions(conn)

	// then: it's empty, not an error
	if err != nil {
		t.Fatalf("appliedVersions() returned error: %v", err)
	}
	if len(applied) != 0 {
		t.Errorf("expected no applied versions, got %v", applied)
	}
}

func TestAppliedVersions_ReturnsRecordedVersions(t *testing.T) {
	// given: a schema_migrations table with two recorded versions
	conn, err := Init(":memory:")
	if err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}
	defer conn.Close()
	if err := ensureSchemaMigrationsTable(conn); err != nil {
		t.Fatalf("ensureSchemaMigrationsTable() returned error: %v", err)
	}
	if _, err := conn.Exec("INSERT INTO schema_migrations (version, name) VALUES (1, 'init'), (2, 'add_image_url')"); err != nil {
		t.Fatalf("failed to seed schema_migrations: %v", err)
	}

	// when: we read applied versions
	applied, err := appliedVersions(conn)

	// then: both versions are present
	if err != nil {
		t.Fatalf("appliedVersions() returned error: %v", err)
	}
	if !applied[1] || !applied[2] {
		t.Errorf("expected versions 1 and 2 to be applied, got %v", applied)
	}
}

func TestApplyMigration_RunsSQLAndRecordsVersion(t *testing.T) {
	// given: a fresh database with the bookkeeping table ready, and a
	// migration that creates a table
	conn, err := Init(":memory:")
	if err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}
	defer conn.Close()
	if err := ensureSchemaMigrationsTable(conn); err != nil {
		t.Fatalf("ensureSchemaMigrationsTable() returned error: %v", err)
	}
	m := migration{version: 1, name: "create_example", sql: "CREATE TABLE example (id INTEGER)"}

	// when: we apply it
	if err := applyMigration(conn, m); err != nil {
		t.Fatalf("applyMigration() returned error: %v", err)
	}

	// then: the migration's table was created
	var name string
	row := conn.QueryRow("SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?", "example")
	if err := row.Scan(&name); err != nil {
		t.Errorf("expected migration's table to exist: %v", err)
	}

	// and: it was recorded in schema_migrations
	applied, err := appliedVersions(conn)
	if err != nil {
		t.Fatalf("appliedVersions() returned error: %v", err)
	}
	if !applied[1] {
		t.Errorf("expected version 1 to be recorded as applied, got %v", applied)
	}
}

func TestApplyMigration_RollsBackOnFailure(t *testing.T) {
	// given: a migration with invalid SQL
	conn, err := Init(":memory:")
	if err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}
	defer conn.Close()
	if err := ensureSchemaMigrationsTable(conn); err != nil {
		t.Fatalf("ensureSchemaMigrationsTable() returned error: %v", err)
	}
	m := migration{version: 1, name: "broken", sql: "NOT VALID SQL"}

	// when: we attempt to apply it
	err = applyMigration(conn, m)

	// then: it returns an error, and nothing was recorded as applied
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	applied, appliedErr := appliedVersions(conn)
	if appliedErr != nil {
		t.Fatalf("appliedVersions() returned error: %v", appliedErr)
	}
	if len(applied) != 0 {
		t.Errorf("expected no versions recorded after a failed migration, got %v", applied)
	}
}
