package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// cmdTouch updates only a record's metadata (ttl and/or access) without changing
// its data, and prints the id. It is update-only: the id must already exist. At
// least one of --ttl / --access must be provided. --ttl sets ttl = now + seconds;
// --access replaces the access labels array. data is left untouched; updated_at
// is bumped.
//
// Clearing ttl/access (making a record permanent/public again) is out of scope
// for v1 — touch sets values, it does not unset them.
// args: db, repository, id, ttl, access
func cmdTouch(args []string) error {
	db := arg(args, 0)
	repo := arg(args, 1)
	id := arg(args, 2)
	ttlArg := strings.TrimSpace(arg(args, 3))
	accessArg := arg(args, 4)

	if id == "" {
		return fmt.Errorf("--id is required for touch")
	}
	if ttlArg == "" && accessArg == "" {
		return fmt.Errorf("touch requires at least one of --ttl or --access")
	}

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

	metaKVs, err := parseObject([]byte(rows[0].Metadata))
	if err != nil {
		return err
	}
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
	newMeta, err := emitObject(metaKVs)
	if err != nil {
		return err
	}

	if _, err := store.Exec("UPDATE "+repo+" SET metadata = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", string(newMeta), id); err != nil {
		return err
	}

	fmt.Println(id)
	return nil
}
