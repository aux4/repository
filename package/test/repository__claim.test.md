## Test

```beforeEach
rm -f .local.db
rm -f race-winners.txt
```

```afterEach
rm -f race-winners.txt
```

### Claim a fresh id

Claiming an id that does not exist yet succeeds and prints the id.

```execute
aux4 repository claim events --id evt1 --data '{"source":"webhook"}' --ttl 3600
```

```expect
evt1
```

### Claiming an already-held id fails, and the original data is preserved

A second claim of the same (still live) id is rejected with a non-zero exit and a clear message —
the id was already claimed and its ttl has not expired. The rejected claim must not have overwritten
the original data.

```execute
aux4 repository claim events --id evt1 --data '{"source":"webhook"}' --ttl 3600
aux4 repository claim events --id evt1 --data '{"source":"retry"}' --ttl 3600 2>claim-retry.err; cat claim-retry.err; rm -f claim-retry.err
aux4 repository read events --id evt1 --raw true
```

```expect:partial
Error: id "evt1" is already claimed**[{"id":"evt1","source":"webhook","__metadata":{"ttl":*?,"createdAt":"*?","updatedAt":"*?"}}]
```

### An expired claim counts as absent and can be re-claimed, replacing its data

A negative `--ttl` computes an expiry epoch already in the past, so the record is immediately
expired. A later claim of the same id succeeds (does NOT get the "already claimed" error) and
atomically replaces the stored data.

```execute
aux4 repository claim events --id evt2 --data '{"source":"first"}' --ttl=-10
aux4 repository claim events --id evt2 --data '{"source":"second"}' --ttl 3600
aux4 repository read events --id evt2 --raw true
```

```expect:partial
evt2**evt2**[{"id":"evt2","source":"second","__metadata":{"ttl":*?,"createdAt":"*?","updatedAt":"*?"}}]
```

### Claim without --ttl never expires

```execute
aux4 repository claim events --id evt3 --data '{"source":"no-ttl"}'
aux4 repository read events --id evt3 --raw true
```

```expect:partial
[{"id":"evt3","source":"no-ttl","__metadata":{"createdAt":"*?","updatedAt":"*?"}}]
```

### Claim requires an id

```execute
aux4 repository claim events --data '{"source":"no-id"}'
```

```error:partial
Error: --id is required
```

### Concurrency: N parallel claims of the same id, exactly one wins

Fire 10 concurrent claims of the same id. Exactly one process must exit 0 (and print the id); the
rest must fail with the "already claimed" error — there is no read-then-write race window since
`InsertOnly` is a single `INSERT ... ON CONFLICT DO NOTHING`-style statement.

```execute
for i in 1 2 3 4 5 6 7 8 9 10; do
  (aux4 repository claim race --id EVT-RACE --data "{\"worker\":$i}" --ttl 3600 >/dev/null 2>/dev/null && echo "$i" >> race-winners.txt) &
done
wait
wc -l < race-winners.txt | tr -d ' '
```

```expect
1
```
