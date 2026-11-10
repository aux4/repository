# aux4/repository 1.3.7

## Added

- `patch` command: new `--increment <field,...>` flag bumps one or more
  numeric fields by 1 server-side, applied after the `--data` merge (missing
  or non-numeric fields treated as 0). Lets a caller atomically increment a
  counter (e.g. a message's retry/attempts count) without computing the
  arithmetic itself, so packages composing a durable primitive on top of
  `repository` (like aux4/queue's `--storage repository` `nack`) never need
  shell `$((...))` math. Using `--increment` without `--data` no longer waits
  on stdin.

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
