package main

import "strings"

// visibilityClause builds the hidden WHERE predicate (ttl expiry + access
// restriction) applied in the inner subquery of read and find. Access labels
// are returned as bound-parameter args (never string-interpolated) so a crafted
// label cannot break out of the SQL string.
//
//   - includeExpired true  -> omit the ttl predicate (show expired records)
//   - all true             -> omit the access predicate (public + restricted)
//   - access ""            -> public rows only (access IS NULL)
//   - access "a,b"         -> public OR rows whose access array overlaps a/b
func visibilityClause(includeExpired, all bool, accessArg string) (string, []any) {
	var clause strings.Builder
	var args []any

	if !includeExpired {
		clause.WriteString(" AND (json_extract(metadata,'$.ttl') IS NULL OR json_extract(metadata,'$.ttl') > strftime('%s','now') + 0)")
	}

	switch {
	case all:
		// no access predicate
	case accessArg == "":
		clause.WriteString(" AND json_extract(metadata,'$.access') IS NULL")
	default:
		labels := strings.Split(accessArg, ",")
		placeholders := make([]string, len(labels))
		for i, l := range labels {
			placeholders[i] = "?"
			args = append(args, l)
		}
		clause.WriteString(" AND (json_extract(metadata,'$.access') IS NULL OR EXISTS (SELECT 1 FROM json_each(metadata,'$.access') WHERE value IN (")
		clause.WriteString(strings.Join(placeholders, ","))
		clause.WriteString(")))")
	}

	return clause.String(), args
}
