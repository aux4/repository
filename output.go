package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// selectColumns is the projection shared by read and find. created_at/updated_at
// are CAST to TEXT so the driver returns the raw SQLite datetime string
// ("YYYY-MM-DD HH:MM:SS") instead of reinterpreting the DATETIME-typed column as
// an RFC3339 timestamp.
const selectColumns = "id, data, metadata, CAST(created_at AS TEXT) AS created_at, CAST(updated_at AS TEXT) AS updated_at"

// innerColumns is the projection of the inner (visibility-filtered) subquery.
const innerColumns = "id, data, metadata, created_at, updated_at"

// outputRows renders rows the same way for read and find: an empty result prints
// "[]" and exits with code 4; otherwise a pretty-printed JSON array.
func outputRows(rows []Row, raw bool) error {
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
