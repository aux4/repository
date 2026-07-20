package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// whereRe rewrites `identifier <op> value` fragments into json_extract() lookups.
// It mirrors the ERE and operator set of the pure-aux4 (sed) implementation:
// case-insensitive, global, leaving and/or, parentheses, and existing
// json_extract(...) expressions untouched.
var whereRe = regexp.MustCompile(`(?i)([a-zA-Z_][a-zA-Z0-9_]*)[[:space:]]*(=|!=|<>|<=|>=|<|>|like|not like)[[:space:]]*('[^']*'|[^[:space:]]+)`)

func translateWhere(expr string) string {
	return whereRe.ReplaceAllString(expr, `json_extract(data, '$$.${1}') ${2} ${3}`)
}

// cmdFind queries records matching a --expr, applying access/ttl visibility in an
// inner subquery so the user expression can only narrow the authorized rowset.
// args: db, repository, expr, raw, includeExpired, access, all
func cmdFind(args []string) error {
	db := arg(args, 0)
	repo := arg(args, 1)
	expr := arg(args, 2)
	raw := arg(args, 3) == "true"
	includeExpired := arg(args, 4) == "true"
	accessArg := arg(args, 5)
	all := arg(args, 6) == "true"

	// Reject SQL comment / statement-terminator sequences before building SQL.
	if strings.Contains(expr, ";") || strings.Contains(expr, "--") || strings.Contains(expr, "/*") {
		fmt.Fprintln(os.Stderr, "Error: --expr may not contain the SQL comment or statement-terminator sequences (; -- or /*)")
		os.Exit(exitInjection)
	}

	translated := translateWhere(expr)

	// Visibility clause (ttl expiry + access restriction) for the inner subquery.
	var accessLabels []any
	ttlClause := ""
	if !includeExpired {
		ttlClause = " AND (json_extract(metadata,'$.ttl') IS NULL OR json_extract(metadata,'$.ttl') > strftime('%s','now') + 0)"
	}
	accessClause := ""
	if all {
		// no access predicate: return public and restricted records
	} else if accessArg == "" {
		accessClause = " AND json_extract(metadata,'$.access') IS NULL"
	} else {
		labels := strings.Split(accessArg, ",")
		placeholders := make([]string, len(labels))
		for i, l := range labels {
			placeholders[i] = "?"
			accessLabels = append(accessLabels, l)
		}
		accessClause = " AND (json_extract(metadata,'$.access') IS NULL OR EXISTS (SELECT 1 FROM json_each(metadata,'$.access') WHERE value IN (" + strings.Join(placeholders, ",") + ")))"
	}

	inner := "SELECT id, data, metadata, created_at, updated_at FROM " + repo + " WHERE 1=1" + ttlClause + accessClause
	// CAST created_at/updated_at to TEXT so the driver returns the raw SQLite
	// datetime string ("YYYY-MM-DD HH:MM:SS") rather than reinterpreting the
	// DATETIME-typed column as an RFC3339 timestamp.
	query := "SELECT id, data, metadata, CAST(created_at AS TEXT) AS created_at, CAST(updated_at AS TEXT) AS updated_at FROM ( " + inner + " ) WHERE (" + translated + ")"

	store, err := NewStore(db)
	if err != nil {
		return err
	}
	defer store.Close()

	if err := store.Init(repo); err != nil {
		return err
	}

	rows, err := store.Query(query, accessLabels...)
	if err != nil {
		// Mirror the sqlite3 CLI's "Parse error near line N: ..." wording so the
		// message stays recognizable to callers (and tests) of the shell version.
		fmt.Fprintln(os.Stderr, "Parse error near line 1: "+err.Error())
		os.Exit(exitInjection)
	}

	if len(rows) == 0 {
		fmt.Println("[]")
		os.Exit(exitInjection)
	}

	objs := make([][]byte, 0, len(rows))
	for _, r := range rows {
		obj, err := composeRow(r, raw)
		if err != nil {
			return err
		}
		objs = append(objs, obj)
	}

	out, err := prettyArray(objs)
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

// composeRow builds the output object for a row: { "id": <id>, ...data } and,
// when raw is set, appends "__metadata": { ...metadata, createdAt, updatedAt }.
func composeRow(r Row, raw bool) ([]byte, error) {
	dataKVs, err := parseObject([]byte(r.Data))
	if err != nil {
		return nil, err
	}

	idBytes, err := json.Marshal(r.ID)
	if err != nil {
		return nil, err
	}

	result := []kv{{Key: "id", Val: json.RawMessage(idBytes)}}
	result = append(result, dataKVs...)

	if raw {
		metaKVs, err := parseObject([]byte(r.Metadata))
		if err != nil {
			return nil, err
		}
		createdBytes, err := json.Marshal(r.CreatedAt)
		if err != nil {
			return nil, err
		}
		updatedBytes, err := json.Marshal(r.UpdatedAt)
		if err != nil {
			return nil, err
		}
		metaKVs = setKV(metaKVs, "createdAt", json.RawMessage(createdBytes))
		metaKVs = setKV(metaKVs, "updatedAt", json.RawMessage(updatedBytes))

		metaBytes, err := emitObject(metaKVs)
		if err != nil {
			return nil, err
		}
		result = append(result, kv{Key: "__metadata", Val: json.RawMessage(metaBytes)})
	}

	return emitObject(result)
}
