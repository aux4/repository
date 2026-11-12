package main

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"time"
)

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

// InsertOnly performs a race-free "claim" of an id: a single INSERT ... ON
// CONFLICT(id) DO UPDATE ... WHERE statement, so the decide-then-write is
// atomic at the database level (no separate read-then-write round trip).
//
//   - no existing row         -> plain INSERT applies, 1 row affected -> claimed
//   - existing row, still live (metadata.ttl absent or in the future)
//     -> the conflict WHERE is false, the UPDATE branch does not apply,
//     0 rows affected -> NOT claimed (already held)
//   - existing row, EXPIRED (metadata.ttl present and <= now)
//     -> the conflict WHERE is true, the row is replaced (data, metadata,
//     created_at, updated_at all reset) in the same statement, 1 row
//     affected -> claimed
//
// Unqualified column names in the conflict WHERE/SET refer to the existing
// (conflicting) row; `excluded.*` refers to the incoming candidate row — this
// is standard SQLite upsert semantics, not a second query.
func (s *sqlStore) InsertOnly(repo, id, data, metadata string) (bool, error) {
	query := "INSERT INTO " + repo + " (id, data, metadata) VALUES (?, ?, ?) " +
		"ON CONFLICT(id) DO UPDATE SET data = excluded.data, metadata = excluded.metadata, " +
		"created_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP " +
		"WHERE json_extract(metadata, '$.ttl') IS NOT NULL AND json_extract(metadata, '$.ttl') <= strftime('%s','now') + 0"
	res, err := s.db.Exec(query, id, data, metadata)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (s *sqlStore) BulkUpsert(repo string, records []UpsertRecord) error {
	query := "INSERT INTO " + repo + " (id, data, metadata) VALUES (?, ?, ?) " +
		"ON CONFLICT(id) DO UPDATE SET data = excluded.data, metadata = excluded.metadata, updated_at = CURRENT_TIMESTAMP"

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	for _, r := range records {
		if _, err := tx.Exec(query, r.ID, r.Data, r.Metadata); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
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

// Lease implements the generic lease-next primitive described on the Store
// interface. It is a plain SQLite backend operation (no app-level locking
// beyond the database's own writer serialization): candidates are read, then
// the winning row is claimed with a single UPDATE whose WHERE clause re-checks
// the exact same eligibility predicate used to select it. If a concurrent
// caller already leased that row (or, with --group, another row in the same
// group) between the read and the write, RowsAffected is 0 and the next
// candidate is tried -- so two concurrent callers can never walk away with the
// same row, and a losing caller never silently overwrites a winner's lease.
func (s *sqlStore) Lease(repo, whereClause string, whereArgs []any, groupField string, ttlSeconds int64, owner string) (*Row, error) {
	predicate := "(json_extract(data,'$.leaseExpiresAt') IS NULL OR json_extract(data,'$.leaseExpiresAt') <= strftime('%s','now') + 0)"
	if whereClause != "" {
		predicate += " AND (" + whereClause + ")"
	}
	if groupField != "" {
		predicate += " AND (json_extract(data,'$." + groupField + "') IS NULL OR NOT EXISTS (" +
			"SELECT 1 FROM " + repo + " g WHERE g.id <> " + repo + ".id" +
			" AND json_extract(g.data,'$." + groupField + "') = json_extract(" + repo + ".data,'$." + groupField + "')" +
			" AND json_extract(g.data,'$.leaseExpiresAt') > strftime('%s','now') + 0))"
	}

	selectQuery := "SELECT " + selectColumns + " FROM " + repo + " WHERE " + predicate +
		" ORDER BY created_at ASC, id ASC LIMIT 20"
	candidates, err := s.Query(selectQuery, whereArgs...)
	if err != nil {
		return nil, err
	}

	exp := time.Now().Unix() + ttlSeconds
	for _, cand := range candidates {
		dataKVs, err := parseObject([]byte(cand.Data))
		if err != nil {
			return nil, err
		}
		dataKVs = setKV(dataKVs, "leaseExpiresAt", json.RawMessage(strconv.FormatInt(exp, 10)))
		if owner != "" {
			ownerBytes, err := json.Marshal(owner)
			if err != nil {
				return nil, err
			}
			dataKVs = setKV(dataKVs, "leaseOwner", json.RawMessage(ownerBytes))
		}
		newData, err := emitObject(dataKVs)
		if err != nil {
			return nil, err
		}

		updateQuery := "UPDATE " + repo + " SET data = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND " + predicate
		args := append([]any{string(newData), cand.ID}, whereArgs...)
		n, err := s.Exec(updateQuery, args...)
		if err != nil {
			return nil, err
		}
		if n == 1 {
			cand.Data = string(newData)
			return &cand, nil
		}
		// Lost the race on this candidate (leased or group-claimed by someone
		// else between the read and the write) -- try the next one.
	}

	return nil, nil
}

// Renew extends a still-live lease on a specific record by id (a heartbeat
// for a caller that is still processing it), to now + ttlSeconds. It is the
// same read-then-conditional-write race protection as Lease: the UPDATE
// re-checks that leaseExpiresAt is still in the future at write time, so a
// lease that already expired -- and may have been picked up by a different
// Lease call in the meantime -- can never be silently resurrected by a late
// renew. Returns false (no error) when the id does not exist or its lease is
// not currently live.
func (s *sqlStore) Renew(repo, id string, ttlSeconds int64) (bool, error) {
	rows, err := s.Query("SELECT "+selectColumns+" FROM "+repo+" WHERE id = ?", id)
	if err != nil {
		return false, err
	}
	if len(rows) == 0 {
		return false, nil
	}

	dataKVs, err := parseObject([]byte(rows[0].Data))
	if err != nil {
		return false, err
	}
	exp := time.Now().Unix() + ttlSeconds
	dataKVs = setKV(dataKVs, "leaseExpiresAt", json.RawMessage(strconv.FormatInt(exp, 10)))
	newData, err := emitObject(dataKVs)
	if err != nil {
		return false, err
	}

	updateQuery := "UPDATE " + repo + " SET data = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND json_extract(data,'$.leaseExpiresAt') > strftime('%s','now') + 0"
	n, err := s.Exec(updateQuery, string(newData), id)
	if err != nil {
		return false, err
	}
	return n == 1, nil
}
