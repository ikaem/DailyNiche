package db

import (
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
)

// migration is one numbered, named SQL file under migrations/.
type migration struct {
	version int
	name    string
	sql     string
}

// parseMigrationFilename splits a migration filename like
// "0002_add_image_url.sql" into its version (2) and name ("add_image_url").
func parseMigrationFilename(filename string) (version int, name string, err error) {
	base := strings.TrimSuffix(filename, ".sql")
	parts := strings.SplitN(base, "_", 2)
	if len(parts) != 2 || parts[1] == "" {
		return 0, "", fmt.Errorf("migration filename %q must be NNNN_description.sql", filename)
	}

	version, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", fmt.Errorf("migration filename %q must start with a numeric version: %w", filename, err)
	}

	return version, parts[1], nil
}

// loadMigrations reads every migration file under the "migrations" directory
// of fsys and returns them sorted by version. fsys is a parameter (rather
// than hardcoding the package's embedded files) so tests can exercise
// sorting/parsing with fabricated in-memory files instead of real ones.
func loadMigrations(fsys fs.FS) ([]migration, error) {
	entries, err := fs.ReadDir(fsys, "migrations")
	if err != nil {
		return nil, err
	}

	migrations := make([]migration, 0, len(entries))
	for _, entry := range entries {
		version, name, err := parseMigrationFilename(entry.Name())
		if err != nil {
			return nil, err
		}

		contents, err := fs.ReadFile(fsys, "migrations/"+entry.Name())
		if err != nil {
			return nil, err
		}

		migrations = append(migrations, migration{version: version, name: name, sql: string(contents)})
	}

	sort.Slice(migrations, func(i, j int) bool { return migrations[i].version < migrations[j].version })
	return migrations, nil
}
