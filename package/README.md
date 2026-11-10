# aux4/repository

Simplified aux4 database

aux4/repository provides a tiny JSON repository backed by SQLite. It exposes a small set of commands (write, claim, lease, read, find, delete, truncate, clean) that let you store arbitrary JSON records keyed by id, query them using SQL-like expressions, optionally expire records with a time-to-live, restrict them with access labels, and manage repository tables. It's designed to be simple to use from shell scripts and other aux4 packages, and to work with stdin-based workflows (pipe JSON into write) or explicit --data flags.

This package is a building block in the aux4 ecosystem for storing and querying structured JSON data without running a separate database server. Typical use cases include quick test fixtures, local caches, small data imports/exports, and lightweight persistence for CLI tools. It ships as a single self-contained binary with no external tools required, and can store data either in a local SQLite file (the default) or in a remote Turso/libSQL database — see [Storage backends](#storage-backends---db).

## Installation

```bash
aux4 aux4 pkger install aux4/repository
```

## Quick Start

Write a record and read it back. This uses the default database (.local.db) and the repository name "users".

```bash
aux4 repository write users --id user1 --data '{"name":"John","age":30}' --metadata '{"role":"admin"}'
```

This prints the id of the stored record (here: user1). Then read the repository:

```bash
aux4 repository read users
```

This returns a JSON array of stored records (`{ "id": ..., ...fields }`). Add `--raw true` to also include a `__metadata` object with the stored metadata plus `createdAt`/`updatedAt` timestamps.

## Storage backends (--db)

Every command accepts `--db`, which selects the storage backend by its scheme:

- **Local (default)** — a file path (or `file:` URL) uses an embedded, file-backed SQLite database. The default is `.local.db` in the current directory.

  ```bash
  aux4 repository write users --id u1 --data '{"name":"John"}'
  aux4 repository read users --db /var/data/app.db
  ```

- **Remote (Turso/libSQL)** — a `libsql://`, `wss://`, `ws://`, `https://`, or `http://` URL connects to a remote Turso/libSQL database over the network. The command surface, output, and behavior are identical to the local backend.

  ```bash
  export TURSO_AUTH_TOKEN="<your-token>"
  aux4 repository write users --db libsql://my-db.turso.io --id u1 --data '{"name":"John"}'
  aux4 repository read users --db libsql://my-db.turso.io
  ```

**Authentication.** For remote databases, the auth token is read from the `TURSO_AUTH_TOKEN` environment variable and applied automatically (unless the URL already includes an `authToken` query parameter). No token is required for a local file or an unauthenticated server.

The package is a single statically linked binary — it needs no `sqlite3`, `jq`, `sed`, or any other external tool installed, on either backend.

## Storing data (write)

The write command stores a JSON document in the named repository table. You can provide the data via --data or pipe JSON to stdin. If --id is omitted, the command will use an id field inside the JSON (if present) or generate a UUID.

The key variables:
- db (default: .local.db) — local database file path or a remote `libsql://` URL (see [Storage backends](#storage-backends---db))
- repository — repository/table name (positional)
- id — optional id to store (overrides id in JSON)
- data — JSON string to store (optional if providing via stdin)
- metadata — optional JSON metadata (defaults to {})
- ttl — optional time-to-live in seconds (see [Expiring records](#expiring-records-with-a-time-to-live-ttl))
- access — optional comma-separated access labels (see [Restricting data with access labels](#restricting-data-with-access-labels))

Example using explicit id, data, and metadata:

```bash
aux4 repository write users --id 123 --data '{"name":"John"}' --metadata '{"role":"admin"}'
```

For full command docs see [aux4 repository write](./commands/repository/write).

## Claiming an id atomically (claim)

`write` always upserts — a second `write` for the same id overwrites it. When you need a **race-free
dedupe/lock primitive** instead — "only one caller should ever win this id" — use `claim`. It writes a
record only if no *live* record with that id already exists; an existing record whose `--ttl` has expired
counts as absent and is atomically replaced. This makes it safe to call `claim` concurrently (including
from multiple processes against a shared remote database) without a separate check-then-write step: there
is no window where two callers can both believe they won.

```bash
aux4 repository claim events --id evt_12345 --data '{"source":"webhook"}' --ttl 86400
```

```text
evt_12345
```

A retried delivery of the same event id within the ttl window is rejected instead of being reprocessed:

```bash
aux4 repository claim events --id evt_12345 --data '{"source":"webhook"}' --ttl 86400
```

```text
Error: id "evt_12345" is already claimed
```

Unlike `write`, `--id` is **required** on `claim` — the id is the thing being claimed. `--ttl` uses the
same semantics as `write --ttl`; a claimed record past its ttl is hidden from `read`/`find`, removed by
`clean`, and can be claimed again (its data is replaced).

For full command docs see [aux4 repository claim](./commands/repository/claim).

## Leasing the next item (lease)

`claim` dedupes by a known id. `lease` is the complement: pick the **oldest eligible record you don't
know the id of yet**, mark it leased for a visibility timeout, and hand it back. It's the generic
primitive for building a work queue (or any other FIFO, visibility-timeout-based claim) on top of a
repository table — there is no separate lease table, no server process, and no sweeper: a lease is
just a `leaseExpiresAt` (and, with `--owner`, `leaseOwner`) field merged into the record's own `data`,
and it expires on its own once that time passes.

```bash
aux4 repository write jobs --data '{"task":"send-email"}'
aux4 repository lease jobs --ttl 30 --owner worker-1
```

```text
[{"id":"3f1a...","task":"send-email","leaseExpiresAt":1732310400,"leaseOwner":"worker-1"}]
```

A second `lease` call inside the 30 second visibility window finds nothing else eligible:

```bash
aux4 repository lease jobs --ttl 30
```

```text
[]
```

Narrow which records are eligible with `--expr` (same syntax as `find --expr`), and serialize
processing per a data field with `--group <field>` — at most one LIVE lease at a time per distinct
value of that field, so messages for the same chat/conversation/tenant are always leased one at a
time while unrelated groups still lease concurrently:

```bash
aux4 repository write chats --data '{"chatId":"c1","text":"hi"}'
aux4 repository write chats --data '{"chatId":"c1","text":"are you there?"}'
aux4 repository lease chats --ttl 60 --group chatId
```

Only the first `c1` message is leased; its sibling stays ineligible until the lease is released or
completed.

Two concurrent `lease` calls can never return the same record — the winning write re-checks the exact
eligibility predicate used to pick the candidate, so a loser's write affects 0 rows and it tries the
next candidate instead of overwriting the winner.

There is no dedicated release/nack verb: patch `leaseExpiresAt` back to `0` with `aux4 repository
patch` to make a record immediately eligible again (e.g. to retry now instead of waiting out the
visibility timeout), and `aux4 repository delete` it once the work is done.

For full command docs see [aux4 repository lease](./commands/repository/lease).

## Reading data (read)

The read command returns stored records as JSON. By default it reads all rows; use --id to select a single record. Each returned object contains an id and the fields from the stored JSON. With `--raw true`, each object also gains a `__metadata` object that merges the stored metadata with createdAt and updatedAt timestamps.

**Table vs. JSON output.** On an interactive terminal, `read` and `find` render their results as a table for readability. When the output is piped, redirected, or captured by another aux4 command it is emitted as **compact single-line JSON** (the whole array on one line, like `jq -c`) — so scripts and pipelines always receive JSON. Number literals (e.g. `45.50`) and key order are preserved. Pass `--render none` to force compact JSON even on a terminal.

```bash
# table on a terminal, raw JSON when piped into jq
aux4 repository read users
aux4 repository read users | jq '.[].name'

# force raw JSON on a terminal
aux4 repository find users --expr "age > 30" --render none
```

Example — write two records then read all:

```bash
aux4 repository write users --id user1 --data '{"name":"John","age":30}' --metadata '{"role":"admin"}'
aux4 repository write users --id user2 --data '{"name":"Jane","age":25}' --metadata '{"role":"user"}'
aux4 repository read users --raw true
```

Example expected output (excerpt):

```json
[{"id":"user1","name":"John","age":30,"__metadata":{"role":"admin","createdAt":"2026-07-21 08:31:31","updatedAt":"2026-07-21 08:31:31"}},{"id":"user2","name":"Jane","age":25,"__metadata":{"role":"user","createdAt":"2026-07-21 08:31:31","updatedAt":"2026-07-21 08:31:31"}}]
```

For full command docs see [aux4 repository read](./commands/repository/read).

## Searching with SQL expressions (find)

find lets you filter records using SQL expressions that are translated to JSON field access. The expr parameter accepts expressions like "age > 30 and name like '%John%'" and supports numeric comparisons, LIKE, logical operators, JSON path extraction, and parentheses for grouping.

Example — find by name equality:

```bash
aux4 repository find users --expr "name = 'John'" --raw true
```

Example expected output (excerpt):

```json
[{"id":"user1","name":"John","age":30,"city":"NYC","__metadata":{"role":"admin","createdAt":"2026-07-21 08:31:31","updatedAt":"2026-07-21 08:31:31"}}]
```

For full command docs see [aux4 repository find](./commands/repository/find).

## Ordering and pagination

Both `read` and `find` accept `--sort`, `--limit`, and `--offset` to order results and page through them.

`--sort` is a comma-separated list of `field [asc|desc]` clauses. The direction is **space-separated** and `asc` is the default when omitted; later clauses break ties left by earlier ones:

```bash
# order by age descending, then name ascending
aux4 repository read users --sort "age desc, name"
```

- **Data fields** sort by their JSON value and support dotted nested paths: `--sort "specs.rank desc"`.
- The **special fields** `id`, `createdAt`, and `updatedAt` sort by the record's id and its creation/update timestamps: `--sort "createdAt desc"`.
- Field names are validated (letters, digits, underscore, dot-separated segments) and directions are limited to `asc`/`desc`; an invalid field or direction is rejected with an error.

`--limit N` caps the number of returned records and `--offset N` skips the first N — both non-negative integers. `--offset` may be used without `--limit`:

```bash
# newest 10 sessions, most recent first
aux4 repository read sessions --sort "createdAt desc" --limit 10

# page 2 of 20-per-page users ordered by name
aux4 repository read users --sort "name" --limit 20 --offset 20

# ten cheapest in-stock products matching an expression
aux4 repository find products --expr "inStock = 1" --sort "price" --limit 10
```

Ordering and pagination apply on top of `--expr`, `--access`, and ttl visibility — the access/ttl filter runs first, then results are ordered and paged.

## Selecting fields (projection)

Both `read` and `find` accept `--select` to return only a subset of fields, in the order listed:

```bash
aux4 repository read users --select "first_name, last_name, age"
```

- **`id` is always included first**, then the selected data fields in the requested order (`--select "age, name"` puts `age` before `name`). Listing `id` explicitly does not duplicate it.
- **Absent fields are omitted** — this is a schema-less store, so if a document lacks a selected field it is simply left out of that row (never emitted as `null`).
- **Dotted nested paths** are supported: `--select "address.city"` navigates into the nested object and returns the value keyed by the path as written (`{"id":...,"address.city":"NYC"}`).
- Field names are validated (letters, digits, underscore, dot-separated segments); an invalid name is rejected with an error.
- Empty `--select` returns all fields (the default).

Projection is applied **in the tool after reading**, never in SQL — no field name is ever placed into a query, so `--select` has no injection surface. It composes with `--expr`, `--sort`, `--limit`/`--offset`, and `--access`; with `--raw`, the `__metadata` object is appended after the selected fields.

```bash
# project two fields from a filtered, sorted, paged result
aux4 repository find users --expr "age > 21" --select "first_name, age" --sort "age desc" --limit 10
```

## Counting records (count)

`count` returns how many records match — without returning the records themselves. It accepts the same `--expr`, `--access`, `--all`, and `--includeExpired` as `find`, and applies the same access/ttl visibility. The output is a bare integer on one line.

```bash
# total visible records
aux4 repository count users

# how many match a filter (including access-restricted records)
aux4 repository count users --expr "age >= 18" --all true
```

A count of `0` is a valid answer and exits 0 — unlike `read`/`find`, `count` does not use the empty-result exit-4 convention. There is no `--sort`/`--select`/`--limit`/`--offset` (a count is always the total of the matching set).

## Updating records in place (patch and touch)

`write` replaces a whole record (and creates it if missing). When you want to amend an existing record instead, use `patch` (data) or `touch` (ttl/access). Both are **update-only** — they error if the id does not exist — and both print the id on success.

**`patch`** shallow-merges a JSON payload into the record's `data`: listed keys set/override (existing keys keep their position, new keys are appended in payload order), unmentioned keys are preserved, and any `id` in the payload is ignored. `metadata` and `created_at` are untouched; `updated_at` is bumped.

```bash
# add/override only these fields; everything else in the record is preserved
aux4 repository patch users --id u1 --data '{"age":31,"city":"NYC"}'

# payload can come from stdin
echo '{"verified":true}' | aux4 repository patch users --id u1
```

**Note:** `patch` is a shallow top-level merge — nested objects are replaced, not deep-merged — and it cannot delete a field. These are intentional limitations of this version.

**`--increment <field,...>`** bumps one or more numeric fields by 1 server-side, applied after the `--data` merge. A missing or non-numeric field is treated as 0 before the bump. This lets a caller atomically increment a counter without computing the arithmetic itself:

```bash
# bump a retry counter and reset a lease in one call -- no shell arithmetic needed
aux4 repository patch messages --id m1 --data '{"leaseExpiresAt":0}' --increment attempts
```

`--increment` can be used with or without `--data` (if used alone, `patch` does not wait on stdin).

**`touch`** updates only metadata without changing `data`: `--ttl <seconds>` sets a new expiry epoch (`now + seconds`) and `--access "a,b"` replaces the access labels. At least one of `--ttl`/`--access` must be given.

```bash
# extend a session's ttl
aux4 repository touch sessions --id s1 --ttl 3600

# re-scope a document to new access labels
aux4 repository touch docs --id doc1 --access "team-a,team-c"
```

**Note:** `touch` only *sets* ttl/access; it cannot clear them (making a record permanent or public again is out of scope for this version).

## Expiring records with a time-to-live (TTL)

Records can be given an expiry using the `--ttl <seconds>` flag on `write`. The value is a duration in seconds; it is converted to an absolute expiry epoch using the database clock and merged into the record metadata under the `ttl` key. Any user-supplied `--metadata` is preserved:

```bash
aux4 repository write sessions --id s1 --data '{"token":"abc"}' --metadata '{"role":"admin"}' --ttl 3600
```

This stores metadata like `{"role":"admin","ttl":1783724578}`. Records with no `ttl` never expire.

Expired records (whose `ttl` is in the past) are **hidden by default** from both `read` and `find` — including reads by `--id`. Records with no `ttl` are always returned. To include expired records (for debugging or manual sweeps), pass `--includeExpired true`:

```bash
# expired records are hidden
aux4 repository read sessions

# expired records are shown
aux4 repository read sessions --includeExpired true
aux4 repository find sessions --expr "role = 'admin'" --includeExpired true
```

Hiding is a query-time filter — expired records still occupy storage until they are swept by `clean`.

**Note:** A negative value written as `--ttl -60` (with a space) is interpreted as a flag; use `--ttl=-60` to store an already-expired record, or set a past epoch directly via `--metadata '{"ttl":1}'`.

For full command docs see [aux4 repository write](./commands/repository/write).

## Restricting data with access labels

Records can be scoped to a set of access labels using the `--access <labels>` flag on `write`. The value is a comma-separated string that is split into a JSON array and stored under the `access` key in the record metadata (alongside any user-supplied `--metadata` and `ttl`):

```bash
aux4 repository write docs --id doc1 --data '{"title":"Roadmap"}' --access "team-a,team-b"
```

This stores metadata like `{"access":["team-a","team-b"]}`. A record written **without** `--access` has no `access` key and is **public** — visible to every caller. A record written **with** `--access` is **restricted** — visible only to callers that present a matching label.

`read` and `find` accept the caller's labels via `--access` and filter results accordingly:

- Public records (no `access` key) are always visible, regardless of the caller's labels.
- A restricted record is visible only when one of its labels matches one of the caller's labels.
- With no `--access`, only public records are returned.
- `--all true` bypasses access filtering and returns every record — public and restricted.

```bash
# caller holding label "team-b" -> public records + records tagged team-a/team-b
aux4 repository read docs --access team-b

# no --access -> public records only
aux4 repository read docs

# --all true -> every record, public and restricted
aux4 repository read docs --all true
```

With `find`, the access filter is applied **on top of** the `--expr` filter and cannot be bypassed through the expression:

```bash
# matches the expression AND is public or tagged "secret"
aux4 repository find docs --expr "team = 'x'" --access secret
```

`find` enforces this by evaluating the access/visibility predicate in an **inner subquery** and applying `--expr` only to the already-restricted rowset. An `--expr` that closes a parenthesis or adds `OR 1=1` can therefore only widen selection *within* the rows the caller is already authorized to see. As defense-in-depth, `find` also **rejects** any `--expr` containing the sequences `;`, `--`, or `/*`. The caller's `--access` labels are safely single-quoted, so a label cannot inject SQL.

**Warning — `--expr` is a trusted, application-controlled surface.** `--expr` is raw SQL. The subquery shape keeps restricted rows out of the result, but a caller who controls `--expr` can still build a **blind-extraction oracle** (e.g. `... AND (SELECT data FROM repo WHERE id='x') LIKE 'S%'`) to infer restricted content from which authorized rows come back. Access filtering protects the returned rowset, not against arbitrary crafted predicates. **Never pass raw end-user text into `--expr` when access control matters** — express caller authorization only via `--access`/`--all`, never by interpolating user input into `--expr`.

**Note:** Access restriction is advisory scoping enforced at query time, not a hard security boundary. Anyone with direct access to the underlying database (the local file, or the remote Turso/libSQL database and its token) can read every record regardless of its labels. The enforcing layer is whoever owns the database and controls the `--access`/`--all` values passed to `read`/`find`.

For full command docs see [aux4 repository read](./commands/repository/read) and [aux4 repository find](./commands/repository/find).

## Sweeping expired records (clean)

The `clean` command removes expired records from **every** repository in a database. It takes no repository argument — it enumerates all user tables and deletes the records whose `ttl` is in the past, printing a per-repository summary. It is intended to be run periodically (for example from [aux4/cron](/r/public/packages/aux4/cron)) as a maintenance sweep.

```bash
aux4 repository clean
```

```text
users: 1 expired records removed
sessions: 1 expired records removed
cache: 0 expired records removed
```

Use `--db` to target a specific database:

```bash
aux4 repository clean --db /var/data/app.db
```

For full command docs see [aux4 repository clean](./commands/repository/clean).

## Bulk loading (import)

`import` reads a **JSON array on stdin** and upserts every element in a **single transaction** — if any element fails, the whole import rolls back and nothing is written. This is both faster and safer than a loop of `write` calls, especially against a remote database. Each element is treated like a `write` payload (id from the element's `id` or a generated UUID; `data` is the element with `id` stripped; existing ids are overwritten). It prints one id per line, in input order.

```bash
echo '[{"id":"u1","name":"Ann"},{"name":"Bo"},{"id":"u3","name":"Cy"}]' | aux4 repository import users
```

`--ttl` and `--access` apply uniformly to every imported record. An empty array writes nothing (exit 0); a non-array or malformed payload is rejected and nothing is written.

```bash
# import a batch of short-lived, team-scoped records
echo '[{"id":"s1"},{"id":"s2"}]' | aux4 repository import sessions --ttl 3600 --access "team-a"
```

## Listing repositories (list)

`list` prints the repositories (tables) in a database as a compact JSON array of names, sorted alphabetically. It takes only `--db`. An empty database prints `[]` (exit 0).

```bash
aux4 repository list
```

```json
["accounts","sessions","users"]
```

## Deleting, truncating, and dropping

- **delete** removes one or more records by id. Supply multiple `--id` flags to remove multiple records.
- **truncate** deletes all records in a repository but keeps the (empty) table, so it can immediately accept new writes. It asks for confirmation unless you pass `--yes`.
- **drop** removes an entire repository — table, schema, and data. It also asks for confirmation unless you pass `--yes`, and is idempotent (dropping a non-existent repository is a harmless no-op). Use `truncate` to empty a repository, `drop` to remove it entirely.

```bash
# remove a single record
aux4 repository delete users --id user1

# empty a repository (schema kept)
aux4 repository truncate users --yes

# remove a repository entirely (schema + data)
aux4 repository drop sessions --yes
```

For full command docs see [aux4 repository delete](./commands/repository/delete), [aux4 repository truncate](./commands/repository/truncate), and [aux4 repository drop](./commands/repository/drop).

## Examples

### Write a record with explicit id, data, and metadata

This example writes a user record and prints the id back.

```bash
aux4 repository write users --id 123 --data '{"name":"John"}' --metadata '{"role":"admin"}'
```

The command prints:

```
123
```

### Write using STDIN (auto-generated UUID)

Pipe JSON to the write command without providing --id; the command will generate a UUID when none is provided.

```bash
echo '{"product":"Widget","price":29.99,"inStock":true}' | aux4 repository write inventory
```

This prints a generated UUID (v4). Tests expect a UUID matching the pattern:

```
^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$
```

### Read complex nested JSON

Write a record with nested JSON and read it back. The nested structure is preserved.

```bash
aux4 repository write orders --id order1 --data '{"customer":{"name":"Alice","email":"alice@example.com"},"items":[{"id":"item1","qty":2},{"id":"item2","qty":1}],"total":29.98}' --metadata '{"status":"pending","priority":"high"}'
aux4 repository read orders --id order1 --raw true
```

Expected output (excerpt):

```json
[{"id":"order1","customer":{"name":"Alice","email":"alice@example.com"},"items":[{"id":"item1","qty":2},{"id":"item2","qty":1}],"total":29.98,"__metadata":{"status":"pending","priority":"high","createdAt":"2026-07-21 08:31:31","updatedAt":"2026-07-21 08:31:31"}}]
```

### Find with numeric comparison

Write several records, then find those matching a numeric condition.

```bash
aux4 repository write users --id user1 --data '{"name":"John","age":30,"city":"NYC"}' --metadata '{"role":"admin"}'
aux4 repository write users --id user2 --data '{"name":"Jane","age":25,"city":"LA"}' --metadata '{"role":"user"}'
aux4 repository write users --id user3 --data '{"name":"Bob","age":35,"city":"NYC"}' --metadata '{"role":"user"}'
aux4 repository find users --expr "age > 28" --raw true
```

Expected output (excerpt):

```json
[{"id":"user1","name":"John","age":30,"city":"NYC","__metadata":{"role":"admin","createdAt":"2026-07-21 08:31:31","updatedAt":"2026-07-21 08:31:31"}},{"id":"user3","name":"Bob","age":35,"city":"NYC","__metadata":{"role":"user","createdAt":"2026-07-21 08:31:31","updatedAt":"2026-07-21 08:31:31"}}]
```

### Delete multiple records and verify

Write multiple records, delete several using multiple --id flags, and read to confirm remaining records.

```bash
aux4 repository write products --id prod1 --data '{"name":"Widget","price":19.99}' --metadata '{"category":"electronics"}'
aux4 repository write products --id prod2 --data '{"name":"Gadget","price":29.99}' --metadata '{"category":"electronics"}'
aux4 repository write products --id prod3 --data '{"name":"Tool","price":39.99}' --metadata '{"category":"hardware"}'
aux4 repository write products --id prod4 --data '{"name":"Device","price":49.99}' --metadata '{"category":"electronics"}'
aux4 repository delete products --id prod1 --id prod3 --id prod4
aux4 repository read products --raw true
```

Expected output (excerpt):

```json
[{"id":"prod2","name":"Gadget","price":29.99,"__metadata":{"category":"electronics","createdAt":"2026-07-21 08:31:31","updatedAt":"2026-07-21 08:31:31"}}]
```

## Configuration

This package does not require external configuration files. All runtime options are provided via command arguments and the default database file (.local.db) can be overridden with --db.

## License

This package is licensed under the Apache License 2.0.

See [LICENSE](./license) for details.
