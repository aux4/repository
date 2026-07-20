package main

import "fmt"

// cmdClean removes expired records (ttl in the past) from every user table in the
// database and prints "<table>: <n> expired records removed" per table, in
// sqlite_master enumeration order.
// args: db
func cmdClean(args []string) error {
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

	for _, table := range tables {
		query := "DELETE FROM " + table + " WHERE json_extract(metadata, '$.ttl') IS NOT NULL AND json_extract(metadata, '$.ttl') <= strftime('%s', 'now') + 0"
		n, err := store.Exec(query)
		if err != nil {
			return err
		}
		fmt.Printf("%s: %d expired records removed\n", table, n)
	}

	return nil
}
