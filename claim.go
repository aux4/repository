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

// cmdClaim atomically "claims" an id: it writes the record only if no LIVE
// record with that id already exists (an existing but expired record counts
// as absent and is replaced). This is the race-free primitive for webhook/event
// dedupe — unlike `write` (always an upsert), two concurrent callers racing to
// claim the same id will have exactly one succeed.
//
// Exit 0 and print the id when claimed; a non-zero exit with a clear message
// when the id is already held by a live record.
//
// args: db, repository, id, data, ttl
func cmdClaim(args []string) error {
	db := arg(args, 0)
	repo := arg(args, 1)
	id := strings.TrimSpace(arg(args, 2))
	dataArg := arg(args, 3)
	ttlArg := arg(args, 4)

	if id == "" {
		return fmt.Errorf("--id is required")
	}

	// Payload comes from --data, or from stdin when --data is empty (same
	// convention as `write`).
	input := dataArg
	if input == "" {
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

	kvs, err := parseObject([]byte(input))
	if err != nil {
		return err
	}
	dataBytes, err := emitObject(removeKV(kvs, "id"))
	if err != nil {
		return err
	}

	metaKVs, err := parseObject([]byte("{}"))
	if err != nil {
		return err
	}
	if strings.TrimSpace(ttlArg) != "" {
		seconds, err := strconv.ParseInt(strings.TrimSpace(ttlArg), 10, 64)
		if err != nil {
			return fmt.Errorf("invalid --ttl value %q: %v", ttlArg, err)
		}
		epoch := time.Now().Unix() + seconds
		metaKVs = setKV(metaKVs, "ttl", json.RawMessage(strconv.FormatInt(epoch, 10)))
	}
	metaBytes, err := emitObject(metaKVs)
	if err != nil {
		return err
	}

	store, err := NewStore(db)
	if err != nil {
		return err
	}
	defer store.Close()

	if err := store.Init(repo); err != nil {
		return err
	}

	claimed, err := store.InsertOnly(repo, id, string(dataBytes), string(metaBytes))
	if err != nil {
		return err
	}
	if !claimed {
		return fmt.Errorf("id %q is already claimed", id)
	}

	fmt.Println(id)
	return nil
}
