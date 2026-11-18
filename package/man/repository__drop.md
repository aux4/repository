The drop command removes an entire repository — its table, its schema, and all of its records. It prompts for confirmation before doing so; pass `--yes` to skip the prompt in scripts.

drop contrasts with `truncate`: `truncate` deletes every record but keeps the (empty) table so it can immediately accept new writes, whereas `drop` removes the table itself. After a drop, the repository no longer appears in `list`; the next command that writes to or reads that repository will recreate it as an empty table (repositories are created on demand).

drop is **idempotent** — dropping a repository that does not exist is a no-op and exits 0 (no error). Dropping one repository never affects the others in the database.

**Storage backends.** The `--db` value selects the backend by scheme: a local file path (default `.local.db`) uses an embedded SQLite database, while a `libsql://` / `wss://` / `https://` URL targets a remote Turso/libSQL database, and the literal value `cloud` resolves to the Turso database already provided by the aux4.cloud `turso:database` capability (via `TURSO_DATABASE_URL`/`TURSO_AUTH_TOKEN`). For remote databases the auth token is read from the `TURSO_AUTH_TOKEN` environment variable. Behavior is identical on all backends.

#### Usage

```bash
aux4 repository drop <repository> [--db <path>] [--yes]
```

--db   Database file path or remote libsql:// URL (default: `.local.db`)
--yes  Skip the confirmation prompt

#### Example

```bash
# remove a repository entirely, without a prompt
aux4 repository drop sessions --yes
```

```bash
# dropping a repository that does not exist is a harmless no-op
aux4 repository drop already-gone --yes
```
