package main

// cmdInit creates the repository table if it does not already exist.
// args: db, repository
func cmdInit(args []string) error {
	db := arg(args, 0)
	repo := arg(args, 1)

	store, err := NewStore(db)
	if err != nil {
		return err
	}
	defer store.Close()

	return store.Init(repo)
}
