#### Description

The `renew` command extends a **still-live** lease on a specific record by `--id`, resetting its
`leaseExpiresAt` to `now + --ttl` seconds. It is a heartbeat for a caller that is still working on a
record it already leased via `aux4 repository lease` and needs to keep its hold on it before the
visibility timeout elapses — without re-leasing (which could hand the record to a different caller).

- **Only renews a live lease.** The update re-checks, at write time, that `data.leaseExpiresAt` is
  still in the future. If the lease already expired — and especially if it has since been picked up
  by another `lease` call — `renew` does **not** resurrect it; it fails instead, so a slow caller can
  never silently steal back a record another worker is now processing.
- **No separate lease token.** Like `lease`/`patch`, this writes directly into the record's own
  `data` field — there is no separate lease table and no server process.
- **Does not touch any other field.** Only `leaseExpiresAt` is updated; `leaseOwner` and the rest of
  the record's `data` are left exactly as they were.

**On success:** exits 0 and prints the `id`.

**When the id does not exist, or its lease is not currently live:** prints `[]` and exits with code
`4` — the same convention `lease`/`find`/`read` use for an empty result.

**`--id` and `--ttl` are both required.**

#### Usage

```bash
aux4 repository renew <repository> --id <id> --ttl <seconds> [--db <path|libsql://...>]
```

--db          The database holding the leased record (default `.local.db`)
--repository  The name of the repository holding the leased record (positional)
--id          The id of the leased record to renew (required)
--ttl         New visibility timeout in seconds, counted from now (required)

#### Example

```bash
aux4 repository write jobs --id job1 --data '{"task":"send-email"}'
aux4 repository lease jobs --ttl 30
aux4 repository renew jobs --id job1 --ttl 30
```

```text
job1
```

Renewing after the lease has already expired (and was not re-leased) still fails, because the
eligibility check is on the lease, not the id:

```bash
aux4 repository renew jobs --id job1 --ttl 30
```

```text
[]
```

Related commands: `aux4 repository lease` to acquire the lease in the first place, `aux4 repository
patch` to release it early (reset `leaseExpiresAt` to `0`), `aux4 repository delete` to complete it.
