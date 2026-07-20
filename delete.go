package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// cmdDelete removes records by id. Ids arrive as a JSON array (aux4 value(id*)).
// Deleting non-existent ids or from an empty repository is a no-op, not an error.
// An empty id list is rejected.
// args: db, repository, idsJSON
func cmdDelete(args []string) error {
	db := arg(args, 0)
	repo := arg(args, 1)
	idsJSON := strings.TrimSpace(arg(args, 2))

	var ids []string
	if idsJSON != "" {
		var raw []json.RawMessage
		if err := json.Unmarshal([]byte(idsJSON), &raw); err != nil {
			return fmt.Errorf("invalid id list %q: %v", idsJSON, err)
		}
		for _, r := range raw {
			ids = append(ids, rawToID(r))
		}
	}

	// Drop empty ids; an all-empty / missing id list is an error.
	kept := ids[:0]
	for _, id := range ids {
		if id != "" {
			kept = append(kept, id)
		}
	}
	ids = kept
	if len(ids) == 0 {
		return fmt.Errorf("no id provided to delete")
	}

	store, err := NewStore(db)
	if err != nil {
		return err
	}
	defer store.Close()

	if err := store.Init(repo); err != nil {
		return err
	}

	placeholders := make([]string, len(ids))
	bound := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		bound[i] = id
	}
	query := "DELETE FROM " + repo + " WHERE id IN (" + strings.Join(placeholders, ",") + ")"
	_, err = store.Exec(query, bound...)
	return err
}
