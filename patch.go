package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// cmdPatch shallow-merges a JSON payload into an existing record's data and
// prints the id. It is update-only (NOT an upsert): the id must already exist.
// Top-level keys in the payload set/override existing keys (updated in place,
// keeping their position); brand-new keys are appended in payload order. Any
// .id in the payload is ignored (the id comes from --id). created_at and
// metadata (ttl/access) are preserved; updated_at is bumped.
//
// An optional comma-separated --increment list names numeric fields that get
// bumped by 1 server-side (missing/non-numeric treated as 0), applied after
// the --data merge so a caller can patch other fields and atomically bump a
// counter (e.g. a message's attempts) in the same call, without computing
// arithmetic itself. This is the generic primitive behind aux4/queue's
// --storage repository nack: Go does the +1, the package DSL never shells
// out to `$((...))`.
//
// Deep-merge and field deletion (unset) are intentionally out of scope for v1.
// args: db, repository, id, data, increment
func cmdPatch(args []string) error {
	db := arg(args, 0)
	repo := arg(args, 1)
	id := arg(args, 2)
	dataArg := arg(args, 3)
	incrementArg := arg(args, 4)

	if id == "" {
		return fmt.Errorf("--id is required for patch")
	}

	input := dataArg
	if input == "" && incrementArg == "" {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		input = string(b)
	}
	input = strings.TrimSpace(input)
	if input == "" {
		input = "{}"
	}

	patchKVs, err := parseObject([]byte(input))
	if err != nil {
		return err
	}
	patchKVs = removeKV(patchKVs, "id")

	store, err := NewStore(db)
	if err != nil {
		return err
	}
	defer store.Close()

	if err := store.Init(repo); err != nil {
		return err
	}

	rows, err := store.Query("SELECT id, data, metadata, CAST(created_at AS TEXT) AS created_at, CAST(updated_at AS TEXT) AS updated_at FROM "+repo+" WHERE id = ?", id)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return fmt.Errorf("id %q not found in repository %q", id, repo)
	}

	merged, err := parseObject([]byte(rows[0].Data))
	if err != nil {
		return err
	}
	for _, e := range patchKVs {
		merged = setKV(merged, e.Key, e.Val)
	}

	for _, field := range splitNonEmpty(incrementArg, ",") {
		current := 0.0
		if raw, ok := getKV(merged, field); ok {
			if n, err := strconv.ParseFloat(strings.TrimSpace(string(raw)), 64); err == nil {
				current = n
			}
		}
		next := current + 1
		var nextRaw []byte
		if next == float64(int64(next)) {
			nextRaw = []byte(strconv.FormatInt(int64(next), 10))
		} else {
			nextRaw = []byte(strconv.FormatFloat(next, 'f', -1, 64))
		}
		merged = setKV(merged, field, nextRaw)
	}

	newData, err := emitObject(merged)
	if err != nil {
		return err
	}

	if _, err := store.Exec("UPDATE "+repo+" SET data = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", string(newData), id); err != nil {
		return err
	}

	fmt.Println(id)
	return nil
}

// splitNonEmpty splits s by sep, trims each part, and drops empty results.
func splitNonEmpty(s, sep string) []string {
	if s == "" {
		return nil
	}
	var out []string
	for _, p := range strings.Split(s, sep) {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
