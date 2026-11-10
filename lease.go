package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// cmdLease atomically picks the oldest eligible record in a repository and
// marks it leased by merging leaseExpiresAt (and, when --owner is given,
// leaseOwner) into its data field. This is the generic "lease the next item"
// primitive: a visibility-timeout lease over any repository, optionally
// scoped by a --where expression (same syntax as `find`) and optionally
// serialized per a --group data field (only one LIVE lease at a time per
// distinct group value).
//
// A lease is just a leaseExpiresAt (unix seconds) written into the record's
// own `data` -- there is no separate lease table and no server process. It
// expires on its own: once leaseExpiresAt is in the past, the record is
// eligible again for anyone's next `lease` call. There is no explicit
// "release" command; callers release early by patching leaseExpiresAt back
// to 0 (immediate redelivery) via `repository patch`, or remove the record
// entirely via `repository delete` (done).
//
// Exit 0 and print the leased record (same shape as `find --raw`) on success;
// exit with exitInjection and print "[]" when nothing is eligible to lease,
// mirroring find/read's empty-result convention.
// args: db, repository, expr, group, ttl, owner, raw
func cmdLease(args []string) error {
	db := arg(args, 0)
	repo := arg(args, 1)
	expr := strings.TrimSpace(arg(args, 2))
	group := strings.TrimSpace(arg(args, 3))
	ttlArg := strings.TrimSpace(arg(args, 4))
	owner := strings.TrimSpace(arg(args, 5))
	raw := arg(args, 6) == "true"

	if ttlArg == "" {
		return fmt.Errorf("--ttl is required")
	}
	ttlSeconds, err := strconv.ParseInt(ttlArg, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid --ttl value %q: %v", ttlArg, err)
	}

	var whereClause string
	var whereArgs []any
	if expr != "" {
		if strings.Contains(expr, ";") || strings.Contains(expr, "--") || strings.Contains(expr, "/*") {
			fmt.Fprintln(os.Stderr, "Error: --expr may not contain the SQL comment or statement-terminator sequences (; -- or /*)")
			os.Exit(exitInjection)
		}
		whereClause = translateWhere(expr)
	}

	if group != "" {
		if !isSimpleIdentifier(group) {
			return fmt.Errorf("invalid --group field name %q", group)
		}
	}

	store, err := NewStore(db)
	if err != nil {
		return err
	}
	defer store.Close()

	if err := store.Init(repo); err != nil {
		return err
	}

	row, err := store.Lease(repo, whereClause, whereArgs, group, ttlSeconds, owner)
	if err != nil {
		return err
	}
	if row == nil {
		fmt.Println("[]")
		os.Exit(exitInjection)
	}

	return outputRows([]Row{*row}, raw, nil)
}

// isSimpleIdentifier guards --group (interpolated directly into SQL as a JSON
// path segment) against anything but a plain field name.
func isSimpleIdentifier(s string) bool {
	for i, r := range s {
		isAlpha := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_'
		isDigit := r >= '0' && r <= '9'
		if i == 0 && !isAlpha {
			return false
		}
		if !isAlpha && !isDigit {
			return false
		}
	}
	return s != ""
}
