package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

// cmdImport bulk-upserts a JSON array read from stdin, applying all elements in a
// single transaction (rolled back on any error, so nothing is partially written).
// Per element, id resolution matches write (element .id if present, else a
// generated lowercase UUID v4); data is the element with .id stripped. Optional
// --ttl / --access apply uniformly to every imported record. It prints one id per
// line in input order. An empty array writes nothing and exits 0.
//
// Per-element metadata is out of scope for this version — ttl/access are uniform.
// args: db, repository, ttl, access
func cmdImport(args []string) error {
	db := arg(args, 0)
	repo := arg(args, 1)
	ttlArg := strings.TrimSpace(arg(args, 2))
	accessArg := arg(args, 3)

	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	var elements []json.RawMessage
	if err := json.Unmarshal(raw, &elements); err != nil {
		return fmt.Errorf("import expects a JSON array on stdin: %v", err)
	}

	// Build the uniform metadata once (same absolute ttl epoch for every record).
	var metaKVs []kv
	if ttlArg != "" {
		seconds, err := strconv.ParseInt(ttlArg, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid --ttl value %q: %v", ttlArg, err)
		}
		epoch := time.Now().Unix() + seconds
		metaKVs = setKV(metaKVs, "ttl", json.RawMessage(strconv.FormatInt(epoch, 10)))
	}
	if accessArg != "" {
		labels := strings.Split(accessArg, ",")
		arr, err := json.Marshal(labels)
		if err != nil {
			return err
		}
		metaKVs = setKV(metaKVs, "access", json.RawMessage(arr))
	}
	metaBytes, err := emitObject(metaKVs)
	if err != nil {
		return err
	}
	metadata := string(metaBytes)

	// Parse and validate every element up front, so a malformed element aborts the
	// import before any write happens.
	records := make([]UpsertRecord, 0, len(elements))
	ids := make([]string, 0, len(elements))
	for i, el := range elements {
		kvs, err := parseObject(el)
		if err != nil {
			return fmt.Errorf("import element %d is not a JSON object: %v", i, err)
		}
		id := ""
		for _, e := range kvs {
			if e.Key == "id" {
				id = rawToID(e.Val)
				break
			}
		}
		if id == "" {
			id = generateUUID()
		}
		dataBytes, err := emitObject(removeKV(kvs, "id"))
		if err != nil {
			return err
		}
		records = append(records, UpsertRecord{ID: id, Data: string(dataBytes), Metadata: metadata})
		ids = append(ids, id)
	}

	store, err := NewStore(db)
	if err != nil {
		return err
	}
	defer store.Close()

	if err := store.Init(repo); err != nil {
		return err
	}

	if len(records) == 0 {
		return nil
	}
	if err := store.BulkUpsert(repo, records); err != nil {
		return err
	}

	for _, id := range ids {
		fmt.Println(id)
	}
	return nil
}
