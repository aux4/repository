package main

import (
	"encoding/json"
	"fmt"
	"sort"
)

// cmdList prints a compact JSON array of the repository (table) names in the
// database, sorted alphabetically for a stable result. An empty database prints
// "[]" and exits 0 (an empty list is a valid answer, not the read/find empty
// exit-4 case).
// args: db
func cmdList(args []string) error {
	db := arg(args, 0)

	store, err := NewStore(db)
	if err != nil {
		return err
	}
	defer store.Close()

	tables, err := store.ListTables()
	if err != nil {
		return err
	}
	sort.Strings(tables)
	if tables == nil {
		tables = []string{}
	}

	out, err := json.Marshal(tables)
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}
