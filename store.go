package main

import (
	_ "embed"
	"strings"
)

//go:embed package/sql/create.sql
var createSQL string

// Row is a single stored record as returned by the backend.
type Row struct {
	ID        string
	Data      string
	Metadata  string
	CreatedAt string
	UpdatedAt string
}

// UpsertRecord is one record to write during a bulk import.
type UpsertRecord struct {
	ID       string
	Data     string
	Metadata string
}

// Store is the storage backend contract. A local (modernc.org/sqlite) backend
// implements it today; a remote (Turso/libSQL) backend can be added behind the
// same interface, selected by the --db scheme in NewStore.
type Store interface {
	// Init creates the per-repository table if it does not exist.
	Init(repo string) error
	// Upsert inserts or updates a record by id.
	Upsert(repo, id, data, metadata string) error
	// BulkUpsert upserts many records in a single transaction (rolled back on any
	// error, so an import either fully applies or writes nothing).
	BulkUpsert(repo string, records []UpsertRecord) error
	// Query runs a read query with bound parameters and returns the rows.
	Query(query string, args ...any) ([]Row, error)
	// QueryInt runs a query returning a single integer (e.g. count(*)).
	QueryInt(query string, args ...any) (int64, error)
	// Exec runs a write query with bound parameters and returns rows affected.
	Exec(query string, args ...any) (int64, error)
	// ListTables returns the user tables (excluding sqlite_* internal tables) in
	// sqlite_master enumeration order.
	ListTables() ([]string, error)
	Close() error
}

// NewStore selects a backend based on the --db value. Remote schemes
// (libsql://, ws(s)://, http(s)://) use the Turso/libSQL backend; anything else
// (a file path or file: URL) uses the local pure-Go SQLite backend.
func NewStore(db string) (Store, error) {
	switch {
	case strings.HasPrefix(db, "libsql://"),
		strings.HasPrefix(db, "wss://"),
		strings.HasPrefix(db, "ws://"),
		strings.HasPrefix(db, "https://"),
		strings.HasPrefix(db, "http://"):
		return newRemoteStore(db)
	default:
		return newLocalStore(db)
	}
}

// schemaFor returns the CREATE TABLE statement for a repository, keeping the
// embedded schema byte-identical to sql/create.sql.
func schemaFor(repo string) string {
	return strings.ReplaceAll(createSQL, "TABLE_NAME", repo)
}
