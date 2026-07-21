package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// sortFieldRe validates a sort field: dot-separated identifier segments only
// (letters/digits/underscore, each segment starting with a letter or underscore).
// This matches the identifier shape the find translator accepts and guarantees no
// raw user text is ever spliced into an ORDER BY expression.
var sortFieldRe = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*(\.[a-zA-Z_][a-zA-Z0-9_]*)*$`)

// orderAndPaginate builds the trailing ORDER BY / LIMIT / OFFSET clause for the
// outer query and returns any bound parameters (for LIMIT/OFFSET). ORDER BY
// expressions cannot be bound, so every field is validated against sortFieldRe
// and the direction is whitelisted to asc/desc; LIMIT and OFFSET are always bound
// parameters. Empty inputs produce no clause (preserving the default order).
func orderAndPaginate(sortArg, limitArg, offsetArg string) (string, []any, error) {
	var sb strings.Builder
	var args []any

	orderBy, err := buildOrderBy(sortArg)
	if err != nil {
		return "", nil, err
	}
	sb.WriteString(orderBy)

	limStr := strings.TrimSpace(limitArg)
	offStr := strings.TrimSpace(offsetArg)

	var limit, offset *int64
	if limStr != "" {
		n, err := parseNonNegInt(limStr, "--limit")
		if err != nil {
			return "", nil, err
		}
		limit = &n
	}
	if offStr != "" {
		n, err := parseNonNegInt(offStr, "--offset")
		if err != nil {
			return "", nil, err
		}
		offset = &n
	}

	switch {
	case limit != nil && offset != nil:
		sb.WriteString(" LIMIT ? OFFSET ?")
		args = append(args, *limit, *offset)
	case limit != nil:
		sb.WriteString(" LIMIT ?")
		args = append(args, *limit)
	case offset != nil:
		// SQLite requires a LIMIT for OFFSET; -1 means unlimited.
		sb.WriteString(" LIMIT -1 OFFSET ?")
		args = append(args, *offset)
	}

	return sb.String(), args, nil
}

// buildOrderBy translates a comma-separated "field [asc|desc]" list into an
// ORDER BY clause. Data fields become json_extract lookups (dotted paths
// supported); the special fields id/createdAt/updatedAt map to real columns.
func buildOrderBy(sortArg string) (string, error) {
	s := strings.TrimSpace(sortArg)
	if s == "" {
		return "", nil
	}

	var terms []string
	for _, clause := range strings.Split(s, ",") {
		clause = strings.TrimSpace(clause)
		if clause == "" {
			continue
		}

		parts := strings.Fields(clause)
		field := parts[0]
		dir := "asc"
		if len(parts) >= 2 {
			dir = strings.ToLower(parts[1])
		}
		if len(parts) > 2 {
			return "", fmt.Errorf("invalid --sort clause %q (expected 'field [asc|desc]')", clause)
		}
		if dir != "asc" && dir != "desc" {
			return "", fmt.Errorf("invalid --sort direction %q (expected asc or desc)", parts[1])
		}
		if !sortFieldRe.MatchString(field) {
			return "", fmt.Errorf("invalid --sort field %q", field)
		}

		var col string
		switch field {
		case "id":
			col = "id"
		case "createdAt":
			col = "created_at"
		case "updatedAt":
			col = "updated_at"
		default:
			col = "json_extract(data,'$." + field + "')"
		}
		terms = append(terms, col+" "+strings.ToUpper(dir))
	}

	if len(terms) == 0 {
		return "", nil
	}
	return " ORDER BY " + strings.Join(terms, ", "), nil
}

func parseNonNegInt(s, flag string) (int64, error) {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("invalid %s value %q (expected a non-negative integer)", flag, s)
	}
	return n, nil
}
