The find command queries records stored in a repository using an SQL-style filtering expression supplied with --expr. It returns a JSON array of matching records; each returned object contains the record id, the fields from the stored JSON object, and a __metadata object that merges any stored metadata with createdAt and updatedAt timestamps. When no records match the expression the command returns an empty array and exits with a non-zero status (useful for scripting).

Expressions use standard comparison operators (=, !=, <>, <, <=, >, >=) and the LIKE / NOT LIKE operators with % wildcards for pattern matching. Logical operators (and, or) and parentheses are supported so you can build compound filtering conditions. To query nested JSON fields, use json_extract(...) in the expression (for example: json_extract(data, '$.specs.wireless') = 1). Boolean values are matched using 1 (true) and 0 (false).

Use find when you need to filter records by content rather than by id. It complements repository write (to add records) and read (to fetch records by id or list all records). The command is intended for ad-hoc queries and scripting where you need flexible, SQL-like filtering over the JSON documents stored in the repository.

By default, records whose metadata `ttl` (an absolute expiry epoch, set via `write --ttl`) is in the past are excluded from the results, even when they match `--expr`. Records with no `ttl` are always eligible. Pass `--includeExpired true` to disable this filter and include expired records in the matches.

### Data restriction

Access restriction is applied on top of the `--expr` filter and cannot be bypassed through the expression. Records written with `write --access <labels>` are **restricted**; records written without `--access` are **public**.

- Pass `--access <labels>` (comma-separated) to declare the labels the caller holds.
- Public records (no `access` key) are always eligible, regardless of the caller's labels.
- A restricted record is returned only when it matches `--expr` **and** one of its labels matches one of the caller's labels.
- With no `--access`, only public records that match `--expr` are returned.
- Pass `--all true` to bypass access filtering (the `--expr` filter still applies).

How the access filter is enforced: `find` evaluates the access/visibility predicate in an **inner subquery** and applies your `--expr` only to the already-restricted rowset (`SELECT ... FROM ( SELECT ... WHERE <visibility> ) WHERE (<expr>)`). This means an `--expr` that closes a parenthesis or adds `OR 1=1` can only ever widen selection *within* the rows the caller is already authorized to see — it can never pull a restricted row into the result. As additional defense-in-depth, `find` **rejects** any `--expr` containing the SQL comment or statement-terminator sequences `;`, `--`, or `/*` with an error before building the query.

The caller's `--access` labels are safely single-quoted before being embedded in the query, so a label cannot inject SQL.

**Warning — `--expr` is a trusted, application-controlled surface.** `--expr` is a raw SQL-expression fragment. The subquery shape prevents restricted rows from appearing in the result, but a caller who controls `--expr` can still mount a **blind-extraction oracle** — e.g. `... AND (SELECT data FROM repo WHERE id='x') LIKE 'S%'` — using the presence/absence of authorized rows to infer the content of restricted ones. Access filtering protects the returned rowset, **not** against a caller crafting arbitrary SQL predicates. **Never pass raw end-user text into `--expr` when access control matters** — treat `--expr` as trusted input built by the application, and use `--access`/`--all` (never `--expr`) to express the caller's authorization.

**Note:** This is advisory scoping enforced at query time, not a hard security boundary. Anyone with direct `--db`/sqlite3 access to the database file can read every record regardless of its labels. The enforcing layer is whoever owns the database and controls the `--access`/`--all` values.

### Example: simple equality
Find records where a top-level field equals a value:

```bash
aux4 repository find users --expr "name = 'John'"
```

Expected result (one matching record):

```json
[
  {
    "id": "user1",
    "name": "John",
    "age": 30,
    "city": "NYC",
    "__metadata": {
      "role": "admin",
      "createdAt": "2021-01-01T...",
      "updatedAt": "2021-01-01T..."
    }
  }
]
```

### Example: numeric and compound expressions
Find records by numeric comparison and combined conditions:

```bash
aux4 repository find users --expr "age > 28 and city = 'NYC'"
```

This returns any records that match both clauses.

### Example: nested JSON fields
Query values nested inside the stored JSON using json_extract:

```bash
aux4 repository find devices --expr "json_extract(data, '$.specs.wireless') = 1"
```

### Example: pattern matching with LIKE
Use LIKE with % wildcards for partial matches:

```bash
aux4 repository find products --expr "name like '%Gaming%'"
```

Returns records whose name contains the substring "Gaming". 

### Example: include expired records

By default expired records are filtered out of the matches. Use `--includeExpired true` to include them:

```bash
aux4 repository find sessions --expr "role = 'admin'" --includeExpired true
```

### Example: combine an expression with access labels

Filter by expression while restricting to the caller's access labels. The restricted record is only returned when a label matches; `--all true` bypasses the access filter but still applies `--expr`.

```bash
# expr + matching label -> public and restricted team-x records
aux4 repository find docs --expr "team = 'x'" --access secret

# expr only, no label -> restricted records stay hidden
aux4 repository find docs --expr "team = 'x'"

# expr + --all true -> public and restricted matches
aux4 repository find docs --expr "team = 'x'" --all true
```

When no records match the expression the command prints an empty JSON array and exits with a non-zero code; this makes it straightforward to use find in scripts and tests that need to detect absence of results.