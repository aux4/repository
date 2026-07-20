package main

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed sql/create.sql
var createSQL string

// Row is a single stored record as returned by the backend.
type Row struct {
	ID        string
	Data      string
	Metadata  string
	CreatedAt string
	UpdatedAt string
}

// Store is the storage backend contract. A local (modernc.org/sqlite) backend
// implements it today; a remote (Turso/libSQL) backend can be added behind the
// same interface, selected by the --db scheme in NewStore.
type Store interface {
	// Init creates the per-repository table if it does not exist.
	Init(repo string) error
	// Upsert inserts or updates a record by id.
	Upsert(repo, id, data, metadata string) error
	// Query runs a read query with bound parameters and returns the rows.
	Query(query string, args ...any) ([]Row, error)
	Close() error
}

// NewStore selects a backend based on the --db value. Local file paths use the
// pure-Go SQLite backend. Remote schemes (libsql://, http(s)://) are reserved
// for the Turso backend, which is not implemented in this milestone.
func NewStore(db string) (Store, error) {
	switch {
	case strings.HasPrefix(db, "libsql://"),
		strings.HasPrefix(db, "http://"),
		strings.HasPrefix(db, "https://"):
		return nil, fmt.Errorf("remote (Turso/libSQL) backend is not implemented yet: %s", db)
	default:
		return newLocalStore(db)
	}
}

// schemaFor returns the CREATE TABLE statement for a repository, keeping the
// embedded schema byte-identical to sql/create.sql.
func schemaFor(repo string) string {
	return strings.ReplaceAll(createSQL, "TABLE_NAME", repo)
}
