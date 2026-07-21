package main

// cmdRead returns records, optionally filtered by id, applying access/ttl
// visibility in an inner subquery (the id filter is applied in the outer query
// so it can only narrow the authorized rowset). Ordering and pagination
// (--sort/--limit/--offset) are applied on the outer query. Empty result prints
// "[]" and exits with code 4.
// args: db, repository, id, raw, includeExpired, access, all, sort, limit, offset, select
func cmdRead(args []string) error {
	db := arg(args, 0)
	repo := arg(args, 1)
	id := arg(args, 2)
	raw := arg(args, 3) == "true"
	includeExpired := arg(args, 4) == "true"
	accessArg := arg(args, 5)
	all := arg(args, 6) == "true"
	sortArg := arg(args, 7)
	limitArg := arg(args, 8)
	offsetArg := arg(args, 9)
	selectArg := arg(args, 10)

	// Validate ordering/pagination and the select projection before building SQL
	// or touching the store.
	pagination, pagArgs, err := orderAndPaginate(sortArg, limitArg, offsetArg)
	if err != nil {
		return err
	}
	selectFields, err := parseSelect(selectArg)
	if err != nil {
		return err
	}

	visClause, queryArgs := visibilityClause(includeExpired, all, accessArg)

	idClause := "1=1"
	if id != "" {
		idClause = "id = ?"
		queryArgs = append(queryArgs, id)
	}
	queryArgs = append(queryArgs, pagArgs...)

	inner := "SELECT " + innerColumns + " FROM " + repo + " WHERE 1=1" + visClause
	query := "SELECT " + selectColumns + " FROM ( " + inner + " ) WHERE " + idClause + pagination

	store, err := NewStore(db)
	if err != nil {
		return err
	}
	defer store.Close()

	if err := store.Init(repo); err != nil {
		return err
	}

	rows, err := store.Query(query, queryArgs...)
	if err != nil {
		return err
	}

	return outputRows(rows, raw, selectFields)
}
