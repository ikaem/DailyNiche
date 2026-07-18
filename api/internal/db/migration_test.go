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
