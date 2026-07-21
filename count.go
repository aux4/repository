package main

import (
	"fmt"
	"os"
	"strings"
)

// cmdCount prints the number of records matching an optional --expr, applying the
// same access/ttl visibility and expression translation as find. It reuses the
// injection rejection (exit 4) and prints a bare integer; a count of 0 is a valid
// answer and exits 0 (the empty-result exit-4 convention does not apply here).
// args: db, repository, expr, access, all, includeExpired
func cmdCount(args []string) error {
	db := arg(args, 0)
	repo := arg(args, 1)
	expr := arg(args, 2)
	accessArg := arg(args, 3)
	all := arg(args, 4) == "true"
	includeExpired := arg(args, 5) == "true"

	// Reject SQL comment / statement-terminator sequences before building SQL.
	if strings.Contains(expr, ";") || strings.Contains(expr, "--") || strings.Contains(expr, "/*") {
		fmt.Fprintln(os.Stderr, "Error: --expr may not contain the SQL comment or statement-terminator sequences (; -- or /*)")
		os.Exit(exitInjection)
	}

	where := "1=1"
	if strings.TrimSpace(expr) != "" {
		where = translateWhere(expr)
	}

	visClause, queryArgs := visibilityClause(includeExpired, all, accessArg)
	inner := "SELECT " + innerColumns + " FROM " + repo + " WHERE 1=1" + visClause
	query := "SELECT count(*) FROM ( " + inner + " ) WHERE (" + where + ")"

	store, err := NewStore(db)
	if err != nil {
		return err
	}
	defer store.Close()

	if err := store.Init(repo); err != nil {
		return err
	}

	n, err := store.QueryInt(query, queryArgs...)
	if err != nil {
		// Mirror find's SQL-error wording so a malformed --expr is recognizable.
		fmt.Fprintln(os.Stderr, "Parse error near line 1: "+err.Error())
		os.Exit(exitInjection)
	}

	fmt.Println(n)
	return nil
}
