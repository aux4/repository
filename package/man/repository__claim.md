The claim command atomically writes a JSON record into the named repository table **only if no live
record with that id already exists** — it is the race-free primitive for dedupe (e.g. "have I already
processed this webhook event id?"). Unlike `write`, which always upserts (insert-or-update), `claim`
never overwrites a live record: concurrent callers racing to claim the same id will have exactly one
succeed.

An existing record whose `--ttl` has expired counts as **absent**. Claiming an id held by an expired
record succeeds and atomically replaces its data, metadata, `createdAt`, and `updatedAt` in the same
database statement — there is no separate read-then-write step, so this is safe under concurrency on
both the local SQLite backend and a remote Turso/libSQL database.

The command accepts JSON either via the `--data` parameter or from STDIN, same as `write`; any top-level
`id` field in the input is stripped before storing (the id is the claim key, not part of the stored data).
Before claiming, the repository is initialized (its table created) if it does not already exist.

**On success:** exits 0 and prints the claimed id to stdout.

**When the id is already held by a live record:** exits non-zero and prints `Error: id "<id>" is already
claimed` to stderr — the existing record is left untouched.

**`--id` is required.** Unlike `write`, `claim` never generates an id for you — the id is the thing being
claimed, so the caller must supply it (e.g. a webhook's event id).

**Storage backends.** The `--db` value selects the backend by its scheme: a local file path (default
`.local.db`) uses an embedded SQLite database, while a `libsql://` / `wss://` / `https://` URL targets a
remote Turso/libSQL database. For remote databases the auth token is read from the `TURSO_AUTH_TOKEN`
environment variable. Behavior — including the atomicity guarantee — is identical on both backends.

The optional `--ttl <seconds>` flag sets a time-to-live for the claim, same semantics as `write --ttl`:
a duration in seconds converted to an absolute expiry epoch (`strftime('%s','now') + <ttl>`), merged into
the record metadata under the `ttl` key. When `--ttl` is omitted the claim never expires on its own (it
still disappears if explicitly deleted/truncated). A claimed record past its `ttl` is hidden from `read`
and `find` by default (same visibility rules as `write`), is removed by `clean`, and — as above — counts
as absent for a later `claim` of the same id.

**Note:** Because a negative value that follows a space (for example `--ttl -60`) is interpreted as a
flag, use the `--ttl=-60` form if you need to claim an already-expired record for testing.

### Example: Dedupe a webhook event id

Claim the event's id before processing it; a duplicate delivery of the same id is rejected instead of
being processed twice.

```bash
aux4 repository claim events --id evt_12345 --data '{"source":"webhook"}' --ttl 86400
```

```text
evt_12345
```

A retried delivery of the same event id within the ttl window fails instead of reprocessing:

```bash
aux4 repository claim events --id evt_12345 --data '{"source":"webhook"}' --ttl 86400
```

```text
Error: id "evt_12345" is already claimed
```

### Example: Re-claiming after expiry

Once the ttl window has passed, the same id can be claimed again — the prior (expired) record is
atomically replaced:

```bash
aux4 repository claim locks --id resource-1 --data '{"owner":"job-a"}' --ttl=-1
aux4 repository claim locks --id resource-1 --data '{"owner":"job-b"}' --ttl 300
```

Both calls print `resource-1`; the second call's data (`job-b`) is what gets stored, because the first
claim had already expired by the time the second ran.

### Example: Concurrent claims — exactly one wins

```bash
for i in 1 2 3; do
  aux4 repository claim locks --id resource-2 --data "{\"owner\":\"job-$i\"}" --ttl 300 &
done
wait
```

Exactly one of the three background calls prints `resource-2` and exits 0; the other two print the
"already claimed" error and exit non-zero — there is no window where two callers can both believe they
hold the claim.

Related commands: use `aux4 repository write` for ordinary insert-or-update records, `aux4 repository
read`/`find` to inspect claimed records, and `aux4 repository clean` to remove expired ones.
