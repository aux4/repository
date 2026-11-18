The touch command updates only a record's `ttl` and/or access-label metadata, leaving its `data` unchanged. On success it prints the record id. It is the metadata-only counterpart to `patch`.

At least one of `--ttl` / `--access` must be provided (otherwise touch errors). `--ttl <seconds>` sets `metadata.ttl` to an absolute expiry epoch (`now + seconds`), extending or shortening the record's lifetime. `--access <labels>` replaces the `metadata.access` array with the comma-separated labels. Other metadata keys are preserved; `updated_at` is bumped; `data` and `created_at` are untouched. Updates use bound parameters.

touch is **update-only**: if the id does not exist it fails and changes nothing.

**Out of scope for this version:** touch *sets* ttl/access — it cannot clear them. There is no way (yet) to remove a ttl to make a record permanent again, or drop all access labels to make a restricted record public again; both would require a metadata-key deletion that this version does not perform.

**Storage backends.** The `--db` value selects the backend by scheme: a local file path (default `.local.db`) uses an embedded SQLite database, while a `libsql://` / `wss://` / `https://` URL targets a remote Turso/libSQL database, and the literal value `cloud` resolves to the Turso database already provided by the aux4.cloud `turso:database` capability (via `TURSO_DATABASE_URL`/`TURSO_AUTH_TOKEN`). For remote databases the auth token is read from the `TURSO_AUTH_TOKEN` environment variable. Behavior is identical on all backends.

#### Usage

```bash
aux4 repository touch <repository> --id <id> [--db <path>] [--ttl <seconds>] [--access <labels>]
```

--db      Database file path or remote libsql:// URL (default: `.local.db`)
--id      The id of the existing record to touch (required)
--ttl     New time-to-live in seconds (sets metadata.ttl = now + seconds)
--access  New comma-separated access labels (replaces metadata.access)

#### Example

```bash
# extend a session's lifetime by another hour
aux4 repository touch sessions --id s1 --ttl 3600
```

```text
s1
```

```bash
# re-scope a document to a new set of teams
aux4 repository touch docs --id doc1 --access "team-a,team-c"
```

```text
doc1
```
