The drop command removes an entire repository — its table, its schema, and all of its records. It prompts for confirmation before doing so; pass `--yes` to skip the prompt in scripts.

drop contrasts with `truncate`: `truncate` deletes every record but keeps the (empty) table so it can immediately accept new writes, whereas `drop` removes the table itself. After a drop, the repository no longer appears in `list`; the next command that writes to or reads that repository will recreate it as an empty table (repositories are created on demand).

drop is **idempotent** — dropping a repository that does not exist is a no-op and exits 0 (no error). Dropping one repository never affects the others in the database.

**Storage backends.** The `--db` value selects the backend by scheme: a local file path (default `.local.db`) uses an embedded SQLite database, while a `libsql://` / `wss://` / `https://` URL targets a remote Turso/libSQL database (auth token read from the `TURSO_AUTH_TOKEN` environment variable). Behavior is identical on both backends.

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
