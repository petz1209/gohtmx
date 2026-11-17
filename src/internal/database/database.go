package database

import (
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"

	_ "modernc.org/sqlite"
)

//go:embed sql/ddl.sql
var ddl string

//go:embed sql/init.sql
var seed string

// Open opens the SQLite database at path. If the file does not exist yet,
// it is created and loaded with the schema (ddl.sql) and sample data (init.sql).
// An existing file is used as is.
func Open(path string) (*sql.DB, error) {
	exists, err := fileExists(path)
	if err != nil {
		return nil, err
	}

	// foreign_keys has to be enabled per connection, the DSN pragma does that for every pooled connection.
	db, err := sql.Open("sqlite", "file:"+path+"?_pragma=foreign_keys(1)")
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}

	if exists {
		return db, nil
	}

	if err := initialize(db); err != nil {
		// Don't leave a half-initialized file behind, otherwise the next start would skip initialization.
		db.Close()
		os.Remove(path)
		return nil, err
	}

	return db, nil
}

func initialize(db *sql.DB) error {
	if _, err := db.Exec(ddl); err != nil {
		return fmt.Errorf("create schema: %w", err)
	}
	if _, err := db.Exec(seed); err != nil {
		return fmt.Errorf("load sample data: %w", err)
	}
	return nil
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, fmt.Errorf("stat %s: %w", path, err)
}
