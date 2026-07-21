package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// selectColumns is the projection shared by read and find. created_at/updated_at
// are CAST to TEXT so the driver returns the raw SQLite datetime string
// ("YYYY-MM-DD HH:MM:SS") instead of reinterpreting the DATETIME-typed column as
// an RFC3339 timestamp.
const selectColumns = "id, data, metadata, CAST(created_at AS TEXT) AS created_at, CAST(updated_at AS TEXT) AS updated_at"

// innerColumns is the projection of the inner (visibility-filtered) subquery.
const innerColumns = "id, data, metadata, created_at, updated_at"

// outputRows renders rows the same way for read and find: an empty result prints
// "[]" and exits with code 4; otherwise a compact single-line JSON array. When
// selectFields is non-empty, only those data fields are projected (id first).
func outputRows(rows []Row, raw bool, selectFields []string) error {
	if len(rows) == 0 {
		fmt.Println("[]")
		os.Exit(exitInjection)
	}

	objs := make([][]byte, 0, len(rows))
	for _, r := range rows {
		obj, err := composeRow(r, raw, selectFields)
		if err != nil {
			return err
		}
		objs = append(objs, obj)
	}

	out, err := compactArray(objs)
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

// composeRow builds the output object for a row: { "id": <id>, ...data } and,
// when raw is set, appends "__metadata": { ...metadata, createdAt, updatedAt }.
// When selectFields is non-empty, only those data fields are included (in the
// requested order); id is always first, absent fields are omitted, and dotted
// paths are navigated and emitted under the field expression as the (flat) key.
func composeRow(r Row, raw bool, selectFields []string) ([]byte, error) {
	dataKVs, err := parseObject([]byte(r.Data))
	if err != nil {
		return nil, err
	}

	idBytes, err := json.Marshal(r.ID)
	if err != nil {
		return nil, err
	}

	result := []kv{{Key: "id", Val: json.RawMessage(idBytes)}}
	if len(selectFields) == 0 {
		result = append(result, dataKVs...)
	} else {
		seen := map[string]bool{"id": true}
		for _, field := range selectFields {
			if seen[field] {
				continue
			}
			seen[field] = true
			if val, ok := lookupPath(dataKVs, field); ok {
				result = append(result, kv{Key: field, Val: val})
			}
		}
	}

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

// lookupPath navigates a dot-separated field path through ordered data fields and
// returns the raw value at that path. It reports false when any segment is
// missing or a non-final segment is not an object (the field is then omitted).
func lookupPath(dataKVs []kv, path string) (json.RawMessage, bool) {
	segments := strings.Split(path, ".")
	cur := dataKVs
	for i, seg := range segments {
		var val json.RawMessage
		found := false
		for _, e := range cur {
			if e.Key == seg {
				val = e.Val
				found = true
				break
			}
		}
		if !found {
			return nil, false
		}
		if i == len(segments)-1 {
			return val, true
		}
		sub, err := parseObject(val)
		if err != nil {
			// A non-final segment that is not an object cannot be navigated.
			return nil, false
		}
		cur = sub
	}
	return nil, false
}
