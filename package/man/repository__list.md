The list command prints the names of the repositories (tables) stored in a database as a compact JSON array, sorted alphabetically. Unlike the other commands it takes no repository argument — it operates on the whole database given by `--db`.

Only user repositories are listed; SQLite's internal tables (those named `sqlite_*`) are excluded. An empty database prints `[]` and exits 0 — an empty list is a valid result, not an error (list does not use the empty-result exit-4 convention that `read`/`find` use). The output is machine-readable JSON, so it pipes cleanly into `jq` or another repository command.

**Storage backends.** The `--db` value selects the backend by scheme: a local file path (default `.local.db`) uses an embedded SQLite database, while a `libsql://` / `wss://` / `https://` URL targets a remote Turso/libSQL database (auth token read from the `TURSO_AUTH_TOKEN` environment variable). Behavior is identical on both backends.

#### Usage

```bash
aux4 repository list [--db <path>]
```

--db  Database file path or remote libsql:// URL (default: `.local.db`)

#### Example

```bash
aux4 repository list
```

```text
["accounts","sessions","users"]
```

```bash
# iterate over every repository, counting records in each
aux4 repository list | jq -r '.[]' | while read repo; do
  echo "$repo: $(aux4 repository count "$repo")"
done
```
