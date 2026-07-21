package main

import "database/sql"

// sqlStore implements Store over any database/sql driver. Both the local
// (modernc.org/sqlite) and remote (Turso/libSQL) backends wrap it, so they run
// byte-identical SQL and share the same output shaping — libSQL is a SQLite
// superset, so json_extract/json_each/strftime/ON CONFLICT all port unchanged.
type sqlStore struct {
	db *sql.DB
}

func (s *sqlStore) Init(repo string) error {
	_, err := s.db.Exec(schemaFor(repo))
	return err
}

func (s *sqlStore) Upsert(repo, id, data, metadata string) error {
	query := "INSERT INTO " + repo + " (id, data, metadata) VALUES (?, ?, ?) " +
		"ON CONFLICT(id) DO UPDATE SET data = excluded.data, metadata = excluded.metadata, updated_at = CURRENT_TIMESTAMP"
	_, err := s.db.Exec(query, id, data, metadata)
	return err
}

func (s *sqlStore) Query(query string, args ...any) ([]Row, error) {
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

func (s *sqlStore) QueryInt(query string, args ...any) (int64, error) {
	var n int64
	err := s.db.QueryRow(query, args...).Scan(&n)
	return n, err
}

func (s *sqlStore) Exec(query string, args ...any) (int64, error) {
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

func (s *sqlStore) ListTables() ([]string, error) {
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

func (s *sqlStore) Close() error {
	return s.db.Close()
}
