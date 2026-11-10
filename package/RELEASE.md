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
