The import command bulk-loads records into a repository from a JSON array supplied on stdin. All elements are upserted in a **single transaction** — if any element fails, the whole import is rolled back and nothing is written. This makes import both faster and safer than a loop of individual `write` calls, especially against a remote (Turso/libSQL) database where each round trip has latency.

Each array element is treated like a `write` payload: the record id is the element's `id` field when present and non-empty, otherwise a generated lowercase UUID v4; the stored `data` is the element with its `id` removed; and the same `INSERT ... ON CONFLICT(id) DO UPDATE` upsert is used, so re-importing an existing id overwrites it. import prints one id per line, in input order, so callers can map results back to inputs.

- **`--ttl` / `--access`** apply uniformly to every imported record (ttl is stored as an absolute expiry epoch, access as a labels array). Per-element metadata is out of scope for this version — every record in one import shares the same ttl/access.
- An **empty array** (`[]`) writes nothing and exits 0.
- A **non-array or malformed** JSON payload is rejected with a non-zero exit and nothing is written.

**Storage backends.** The `--db` value selects the backend by scheme: a local file path (default `.local.db`) uses an embedded SQLite database, while a `libsql://` / `wss://` / `https://` URL targets a remote Turso/libSQL database (auth token read from the `TURSO_AUTH_TOKEN` environment variable). The single-transaction guarantee holds on both backends.

#### Usage

```bash
echo '<json-array>' | aux4 repository import <repository> [--db <path>] [--ttl <seconds>] [--access <labels>]
```

--db      Database file path or remote libsql:// URL (default: `.local.db`)
--ttl     Optional ttl in seconds applied to every imported record
--access  Optional comma-separated access labels applied to every imported record

#### Example

```bash
# bulk-load three users; the middle one gets a generated id
echo '[{"id":"u1","name":"Ann"},{"name":"Bo"},{"id":"u3","name":"Cy"}]' | aux4 repository import users
```

```text
u1
7850318a-8958-4b46-bfea-8f62afec6043
u3
```

```bash
# import a batch of short-lived, team-scoped records
echo '[{"id":"s1"},{"id":"s2"}]' | aux4 repository import sessions --ttl 3600 --access "team-a"
```
