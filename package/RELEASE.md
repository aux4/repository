# aux4/repository 1.3.5

## Added

- `lease` command: the generic "lease the next item" primitive used to build a
  durable, visibility-timeout-based work queue on top of any repository table.
  Atomically picks the oldest eligible record (optionally narrowed by `--expr`,
  optionally serialized per a `--group <field>` value so only one record per
  group is ever leased at a time), marks it leased by merging `leaseExpiresAt`
  (and, with `--owner`, `leaseOwner`) into its own `data`, and returns it.
  Concurrent callers never win the same record: the leasing write re-checks the
  same eligibility predicate used to pick the candidate, so a losing caller's
  write affects 0 rows instead of silently overwriting the winner. No separate
  release/nack verb is needed — patch `leaseExpiresAt` back to `0` to make a
  record immediately eligible again, or `delete` it once the work is done.

# aux4/repository 1.3.4

## Added

- `claim` command: atomic insert-only write. Creates the record only if no live
  record with that id exists (an expired record counts as absent and is replaced).
  Prints the id and exits 0 when claimed; exits non-zero when the id is already
  held. Safe under concurrent callers on both local SQLite and Turso/libSQL, so it
  can back webhook deduplication.

## Fixed

- Release binaries are now built with `-ldflags="-s -w"` (debug symbols
  stripped). The 8-platform package had grown past the hub's 50 MB compressed
  limit, so publishes reported success but never appeared on the hub. The
  package is now about 32 MB.
