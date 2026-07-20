package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// whereRe rewrites `identifier <op> value` fragments into json_extract() lookups.
// It mirrors the ERE and operator set of the pure-aux4 (sed) implementation:
// case-insensitive, global, leaving and/or, parentheses, and existing
// json_extract(...) expressions untouched.
var whereRe = regexp.MustCompile(`(?i)([a-zA-Z_][a-zA-Z0-9_]*)[[:space:]]*(=|!=|<>|<=|>=|<|>|like|not like)[[:space:]]*('[^']*'|[^[:space:]]+)`)

func translateWhere(expr string) string {
	return whereRe.ReplaceAllString(expr, `json_extract(data, '$$.${1}') ${2} ${3}`)
}

// cmdFind queries records matching a --expr, applying access/ttl visibility in an
// inner subquery so the user expression can only narrow the authorized rowset.
// args: db, repository, expr, raw, includeExpired, access, all
func cmdFind(args []string) error {
	db := arg(args, 0)
	repo := arg(args, 1)
	expr := arg(args, 2)
	raw := arg(args, 3) == "true"
	includeExpired := arg(args, 4) == "true"
	accessArg := arg(args, 5)
	all := arg(args, 6) == "true"

	// Reject SQL comment / statement-terminator sequences before building SQL.
	if strings.Contains(expr, ";") || strings.Contains(expr, "--") || strings.Contains(expr, "/*") {
		fmt.Fprintln(os.Stderr, "Error: --expr may not contain the SQL comment or statement-terminator sequences (; -- or /*)")
		os.Exit(exitInjection)
	}

	translated := translateWhere(expr)
	visClause, queryArgs := visibilityClause(includeExpired, all, accessArg)

	inner := "SELECT " + innerColumns + " FROM " + repo + " WHERE 1=1" + visClause
	query := "SELECT " + selectColumns + " FROM ( " + inner + " ) WHERE (" + translated + ")"

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
		// Mirror the sqlite3 CLI's "Parse error near line N: ..." wording so the
		// message stays recognizable to callers (and tests) of the shell version.
		fmt.Fprintln(os.Stderr, "Parse error near line 1: "+err.Error())
		os.Exit(exitInjection)
	}

	return outputRows(rows, raw)
}
