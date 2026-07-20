The clean command sweeps expired records from every repository in a database. Unlike the other commands, it takes no repository argument — it enumerates all user tables in the database (every table in `sqlite_master` of type `table`, excluding SQLite internal tables whose names start with `sqlite_`) and, for each one, deletes the records whose metadata `ttl` is an absolute expiry epoch in the past (`json_extract(metadata,'$.ttl') <= strftime('%s','now')`). Records with no `ttl` are never touched.

A `ttl` is written by `aux4 repository write --ttl <seconds>`, which stores `current epoch + ttl` under the metadata `ttl` key. Expired records are already hidden from `read` and `find` by default; `clean` is what actually removes them from storage, reclaiming space. It is designed to be run periodically — for example from [aux4/cron](/r/public/packages/aux4/cron) — as a background maintenance sweep.

For each repository, the command prints a one-line summary reporting how many expired records were removed. Repositories with no expired records are reported with a count of zero and left untouched.

#### Usage

```bash
aux4 repository clean [--db <path>]
```

--db   The database to clean — a local file path (default: `.local.db`) or a remote `libsql://` URL. For remote Turso/libSQL databases the auth token is read from the `TURSO_AUTH_TOKEN` environment variable.

#### Example

Remove expired records from every repository in the default database:

```bash
aux4 repository clean
```

```text
users: 1 expired records removed
sessions: 1 expired records removed
cache: 0 expired records removed
```

#### Example: scheduled cleanup

Run the sweep against a specific database on a schedule (for instance via aux4/cron):

```bash
aux4 repository clean --db /var/data/app.db
```

Related commands: use [aux4 repository write](./write) with `--ttl` to give records an expiry, and [aux4 repository read](./read) / [aux4 repository find](./find) with `--includeExpired true` to inspect records that have already expired but not yet been swept.
