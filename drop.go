package main

// cmdDrop removes an entire repository (its table, schema and data). It is
// idempotent: dropping a repository that does not exist is a no-op with exit 0.
// Unlike truncate (which keeps the empty table), drop removes the schema too.
// The confirmation prompt lives in the .aux4 command (bypassed by --yes).
// args: db, repository
func cmdDrop(args []string) error {
	db := arg(args, 0)
	repo := arg(args, 1)

	store, err := NewStore(db)
	if err != nil {
		return err
	}
	defer store.Close()

	_, err = store.Exec("DROP TABLE IF EXISTS " + repo)
	return err
}
