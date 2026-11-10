#### Description

The `lease` command atomically picks the **oldest eligible record** in a repository and marks it
leased for a visibility timeout, returning it. It is the generic primitive for building a work-queue
(or any other visibility-timeout-based claim) on top of an `aux4/repository` table: lease a record,
do the work, then either delete it (done), patch it back to immediately-available (retry), or leave
it alone to expire and be re-leased automatically.

A lease is nothing more than a `leaseExpiresAt` (and, when `--owner` is given, `leaseOwner`) field
merged into the record's own `data` — there is no separate lease table, no server process, and no
background sweeper. A lease "expires on its own": once `leaseExpiresAt` is in the past, the record is
eligible again for the very next `lease` call, exactly as an expired `claim` counts as absent.

- **Eligibility.** A record is eligible when it has no `leaseExpiresAt` yet, or `leaseExpiresAt` has
  already passed. Optionally narrow eligibility further with `--expr` (same syntax as `find --expr`).
- **Ordering.** Among eligible records, the one with the oldest `createdAt` (ties broken by `id`) is
  leased first — first-in-first-out.
- **Group serialization (`--group`).** When `--group <field>` is given, a record is skipped if
  another record with the *same value* of `data.<field>` currently holds a LIVE lease — so at most
  one record per distinct group value is ever leased at a time, while unrelated groups still lease
  concurrently. Records with no value for `<field>` are never group-restricted.
- **Concurrency.** Two concurrent `lease` calls can never return the same record: the winning write is
  a single `UPDATE ... WHERE id = ? AND <the same eligibility predicate>` — if another caller already
  leased that record (or, under `--group`, a sibling in the same group) between the read and the
  write, the update affects 0 rows and the next eligible candidate is tried instead. This holds on
  both the local SQLite backend and a remote Turso/libSQL database.
- **Releasing early / retrying now.** There is no dedicated "release" or "nack" verb — reset
  `data.leaseExpiresAt` to `0` (or any past value) with `aux4 repository patch` to make a record
  immediately eligible again before its visibility timeout naturally elapses.
- **Completing.** Remove the record with `aux4 repository delete` once the work it represents is
  done; a deleted record can never be leased again.

**On success:** exits 0 and prints a single-element JSON array containing the leased record (same
shape as `find --raw`): `[{"id": ..., ...data, "__metadata": {...}}]` (the `__metadata` block only
appears with `--raw true`).

**When nothing is eligible to lease:** prints `[]` and exits with code `4` — the same convention
`find` and `read` use for an empty result.

**`--ttl` is required.** It is the visibility timeout in seconds: how long the lease is held before it
expires on its own and becomes eligible again. There is no default — every lease must say how long it
is allowed to hold the record.

#### Usage

```bash
aux4 repository lease <repository> --ttl <seconds> [--expr <expression>] [--group <field>] \
  [--owner <id>] [--raw <true|false>] [--db <path|libsql://...>]
```

--db        The database to lease from (default `.local.db`)
--repository The name of the repository to lease from (positional)
--expr      SQL expression narrowing which records are eligible (same syntax as `find --expr`)
--group     Data field name to serialize on: only one LIVE lease at a time per distinct value
--ttl       Visibility timeout in seconds (required)
--owner     Opaque identifier of the leasing caller, stored as `data.leaseOwner`
--raw       Include `__metadata` (ttl/access/createdAt/updatedAt) in the output (default `false`)

#### Example

```bash
aux4 repository write jobs --data '{"task":"send-email"}'
aux4 repository lease jobs --ttl 30 --owner worker-1
```

```text
[{"id":"3f1a...","task":"send-email","leaseExpiresAt":1732310400,"leaseOwner":"worker-1"}]
```

A second `lease` call within the 30 second visibility window finds nothing else eligible:

```bash
aux4 repository lease jobs --ttl 30
```

```text
[]
```

#### Example: serializing per group

```bash
aux4 repository write chats --data '{"chatId":"c1","text":"hi"}'
aux4 repository write chats --data '{"chatId":"c1","text":"are you there?"}'
aux4 repository lease chats --ttl 60 --group chatId
```

Only the first `c1` message is leased; the second stays ineligible (even though it is also eligible
by timeout) until the first lease is released (patched back to `leaseExpiresAt: 0`) or completed
(deleted).

Related commands: `aux4 repository patch` to release a lease early or bump retry counters, `aux4
repository delete` to complete the work, `aux4 repository claim` for id-based (not FIFO) dedupe.
