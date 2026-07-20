package main

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

// localStore is the pure-Go, CGo-free SQLite backend (modernc.org/sqlite).
type localStore struct {
	db *sql.DB
}

func newLocalStore(path string) (*localStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	return &localStore{db: db}, nil
}

func (s *localStore) Init(repo string) error {
	_, err := s.db.Exec(schemaFor(repo))
	return err
}

func (s *localStore) Upsert(repo, id, data, metadata string) error {
	query := "INSERT INTO " + repo + " (id, data, metadata) VALUES (?, ?, ?) " +
		"ON CONFLICT(id) DO UPDATE SET data = excluded.data, metadata = excluded.metadata, updated_at = CURRENT_TIMESTAMP"
	_, err := s.db.Exec(query, id, data, metadata)
	return err
}

func (s *localStore) Query(query string, args ...any) ([]Row, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Row
	for rows.Next() {
		var r Row
		if err := rows.Scan(&r.ID, &r.Data, &r.Metadata, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *localStore) Exec(query string, args ...any) (int64, error) {
	res, err := s.db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (s *localStore) ListTables() ([]string, error) {
	rows, err := s.db.Query("SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}
	return tables, rows.Err()
}

func (s *localStore) Close() error {
	return s.db.Close()
}
