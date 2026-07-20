package main

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

// newLocalStore opens the pure-Go, CGo-free SQLite backend (modernc.org/sqlite)
// against a local file path.
func newLocalStore(path string) (*sqlStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	return &sqlStore{db: db}, nil
}
