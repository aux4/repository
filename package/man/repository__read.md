Reads records from a named aux4 repository and returns them as a JSON array. The command ensures the repository exists (initializing it if necessary) and then returns records flattened so that each item contains an "id" field plus the original JSON data fields. Record-level metadata is merged into a single "__metadata" object which always includes system timestamps (createdAt and updatedAt).

When an --id is provided, the command returns an array with the single matching record (or an empty array if none found). When no --id is supplied it returns all records in the repository as an array. If the repository has no records the command prints an empty array and exits with a non-zero status (exit code 4 in tests), making it easy to detect an empty result programmatically.

Use this command when you need to retrieve stored JSON objects from an aux4 repository and work with them as native JSON (the command outputs valid JSON suitable for piping into jq or other JSON processors). It is commonly paired with the write and find commands to populate and query repository data.

**Storage backends.** The `--db` value selects the backend by scheme: a local file path (default `.local.db`) uses an embedded SQLite database, while a `libsql://` / `wss://` / `https://` URL targets a remote Turso/libSQL database (auth token read from the `TURSO_AUTH_TOKEN` environment variable). Behavior is identical on both backends.

**Output rendering.** On an interactive terminal the JSON result is rendered as a table for readability; when the output is piped, redirected, or captured by another command it stays raw JSON, unchanged. Pass `--render none` to force raw JSON even on a terminal.

By default, records whose metadata `ttl` (an absolute expiry epoch, set via `write --ttl`) is in the past are hidden from the result — both when reading all records and when reading a single record by `--id`. Records with no `ttl` are always returned. Pass `--includeExpired true` to skip this filter and return expired records as well (useful for debugging or manual sweeps).

### Data restriction

Records written with `write --access <labels>` carry one or more access labels under `metadata.access` and are **restricted**; records written without `--access` are **public**. The `read` command filters results by the caller's labels:

- Pass `--access <labels>` (comma-separated) to declare the labels the caller holds.
- Public records (no `access` key) are always visible, regardless of the caller's labels.
- A restricted record is visible only when one of its labels matches one of the caller's labels.
- With no `--access`, only public records are returned.
- Pass `--all true` to bypass access filtering entirely and return every record — public and restricted — regardless of labels.

The label filter applies both when listing all records and when reading a single record by `--id`: reading a restricted record by id without a matching label returns an empty array. The caller's `--access` labels are safely single-quoted before being embedded in the query, so a label cannot inject SQL.

**Note:** This is advisory scoping enforced at query time, not a hard security boundary. Anyone with direct access to the underlying database (the local file, or the remote database and its token) can read every record regardless of its labels. The enforcing layer is whoever owns the database and controls the `--access`/`--all` values.

### Example: Read all records from a repository

This reads every record stored in the users repository and returns them as an array. Each item contains the record's id, the stored data fields, and a merged "__metadata" object with createdAt/updatedAt timestamps.

```bash
aux4 repository read users
```

```json
[
  {
    "id": "user1",
    "name": "John",
    "age": 30,
    "__metadata": {
      "role": "admin",
      "createdAt": "2023-01-01T12:00:00Z",
      "updatedAt": "2023-01-01T12:00:00Z"
    }
  },
  {
    "id": "user2",
    "name": "Jane",
    "age": 25,
    "__metadata": {
      "role": "user",
      "createdAt": "2023-01-01T12:01:00Z",
      "updatedAt": "2023-01-01T12:01:00Z"
    }
  }
]
```

### Example: Read a specific record by id

Provide --id to return only the matching record. The command still returns an array (either with a single object or empty when the id is not present).

```bash
aux4 repository read products --id prod123
```

```json
[
  {
    "id": "prod123",
    "name": "Widget",
    "price": 19.99,
    "inStock": true,
    "__metadata": {
      "category": "electronics",
      "createdAt": "2023-01-01T12:10:00Z",
      "updatedAt": "2023-01-01T12:10:00Z"
    }
  }
]
```

### Example: Empty repository or missing id

If the repository contains no matching records the command prints an empty array. In test scenarios this also results in a non-zero exit code, which can be used to detect the empty-result case in scripts.

```bash
# No records present -> prints [] and exits non-zero
aux4 repository read empty_test_repo
```

### Example: Include expired records

By default an expired record is hidden. Use `--includeExpired true` to see it anyway.

```bash
# hidden by default (record has an expired ttl)
aux4 repository read sessions --id s1

# shown with --includeExpired
aux4 repository read sessions --id s1 --includeExpired true
```

### Example: Read restricted records by access label

Given a public record and records tagged with access labels, a caller sees only public records plus the ones matching its labels; `--all true` returns everything.

```bash
# caller holding label "a" -> public records + records tagged "a"
aux4 repository read docs --access a

# no --access -> public records only
aux4 repository read docs

# --all true -> every record, public and restricted
aux4 repository read docs --all true
```

### Ordering and pagination

Results can be ordered with `--sort` and paged with `--limit` / `--offset`. These apply after the access/ttl visibility filter and the `--id` filter.

- **`--sort`** is a comma-separated list of `field [asc|desc]` clauses (space-separated direction; `asc` is the default when omitted). Later clauses break ties left by earlier ones. Example: `--sort "age desc, name"` orders by `age` descending, then `name` ascending.
  - Data fields sort by their JSON value and support dotted nested paths, e.g. `--sort "specs.rank"`.
  - The special fields `id`, `createdAt`, and `updatedAt` sort by the record's id and its creation/update timestamps.
  - Field names are validated (letters, digits, underscore, dot-separated) and directions are restricted to `asc`/`desc`; an invalid field or direction is rejected with an error.
- **`--limit N`** caps the number of returned records; **`--offset N`** skips the first N (both non-negative integers, applied as bound parameters). `--offset` without `--limit` is allowed. Empty values leave the result unbounded and in its natural order.

```bash
# newest 10 sessions, most recent first
aux4 repository read sessions --sort "createdAt desc" --limit 10

# page 2 of 20-per-page users ordered by name
aux4 repository read users --sort "name" --limit 20 --offset 20

# order by a nested field
aux4 repository read products --sort "specs.rank desc"
```

The read command is most useful together with the repository write and find commands. For writing new items, see [aux4 repository write](./write); for filtering using expressions, see [aux4 repository find](./find). Expired records can be swept from every repository with [aux4 repository clean](./clean).