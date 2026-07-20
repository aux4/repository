package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

// cmdWrite stores a record and prints its id.
// args: db, repository, id, data, metadata, ttl, access
func cmdWrite(args []string) error {
	db := arg(args, 0)
	repo := arg(args, 1)
	idFlag := arg(args, 2)
	dataArg := arg(args, 3)
	metaArg := arg(args, 4)
	ttlArg := arg(args, 5)
	accessArg := arg(args, 6)

	// Payload comes from --data, or from stdin when --data is empty.
	input := dataArg
	if input == "" {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		input = string(b)
	}
	input = strings.TrimSpace(input)

	kvs, err := parseObject([]byte(input))
	if err != nil {
		return err
	}

	// id resolution priority: --id -> payload .id -> generated UUID v4.
	id := idFlag
	if id == "" {
		for _, e := range kvs {
			if e.Key == "id" {
				id = rawToID(e.Val)
				break
			}
		}
	}
	if id == "" {
		id = generateUUID()
	}

	// Stored data is always the payload with any top-level id removed.
	dataBytes, err := emitObject(removeKV(kvs, "id"))
	if err != nil {
		return err
	}

	// metadata = user metadata (default {}) plus computed keys, in insertion
	// order {...userMeta, ttl, access}.
	if strings.TrimSpace(metaArg) == "" {
		metaArg = "{}"
	}
	metaKVs, err := parseObject([]byte(metaArg))
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

	store, err := NewStore(db)
	if err != nil {
		return err
	}
	defer store.Close()

	if err := store.Init(repo); err != nil {
		return err
	}
	if err := store.Upsert(repo, id, string(dataBytes), string(metaBytes)); err != nil {
		return err
	}

	fmt.Println(id)
	return nil
}

// rawToID extracts a usable id from a payload .id value. JSON strings are
// unquoted; JSON null/false are treated as absent (matching jq's `.id // empty`).
func rawToID(raw json.RawMessage) string {
	s := strings.TrimSpace(string(raw))
	if s == "null" || s == "false" {
		return ""
	}
	if len(s) > 0 && s[0] == '"' {
		var str string
		if err := json.Unmarshal(raw, &str); err == nil {
			return str
		}
	}
	return s
}

// generateUUID returns a canonical lowercase UUID v4.
func generateUUID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand should never fail; fall back to a time-based value.
		return fmt.Sprintf("%016x-%016x", time.Now().UnixNano(), time.Now().Unix())
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
