package main

// cmdTruncate deletes every record from a repository, preserving the table so it
// can accept new records afterward. The confirm prompt lives in the .aux4
// command (bypassed by --yes); the binary only performs the delete.
// args: db, repository
func cmdTruncate(args []string) error {
	db := arg(args, 0)
	repo := arg(args, 1)

	store, err := NewStore(db)
	if err != nil {
		return err
	}
	defer store.Close()

	if err := store.Init(repo); err != nil {
		return err
	}

	_, err = store.Exec("DELETE FROM " + repo)
	return err
}
