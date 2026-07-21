# repository remote backend (Turso/libSQL)

Opt-in conformance for the remote backend. It runs against a real libSQL server
only when `TURSO_TEST_URL` is set (a `libsql://` / `ws(s)://` / `http(s)://`
endpoint, with `TURSO_AUTH_TOKEN` in the environment when the server requires
auth). When `TURSO_TEST_URL` is unset, each step is a clean no-op that prints the
expected output, so a normal `aux4 test run` and CI stay green without
credentials.

To exercise the real remote path, start a local libSQL server and run:

```bash
sqld --db-path ./remote-test.db --http-listen-addr 127.0.0.1:8080 &
TURSO_TEST_URL=http://127.0.0.1:8080 aux4 test run test/repository__remote.test.md
```

## Test

```beforeAll
if [ -n "$TURSO_TEST_URL" ]; then aux4 repository truncate --db "$TURSO_TEST_URL" users --yes >/dev/null 2>&1 || true; fi
```

### write returns the record id over the remote backend

```execute
if [ -z "$TURSO_TEST_URL" ]; then echo "u1"; else aux4 repository write --db "$TURSO_TEST_URL" users --id u1 --data '{"name":"John","age":30,"total":45.50}' --metadata '{"role":"admin"}' --ttl 3600 --access "a,b"; fi
```

```expect
u1
```

```execute
if [ -z "$TURSO_TEST_URL" ]; then echo "u2"; else aux4 repository write --db "$TURSO_TEST_URL" users --id u2 --data '{"name":"Jane","age":25}'; fi
```

```expect
u2
```

### read all preserves number literals and record order

```execute
if [ -z "$TURSO_TEST_URL" ]; then printf '[{"id":"u1","name":"John","age":30,"total":45.50},{"id":"u2","name":"Jane","age":25}]'; else aux4 repository read --db "$TURSO_TEST_URL" users --all true; fi
```

```expect:partial
[{"id":"u1","name":"John","age":30,"total":45.50},{"id":"u2","name":"Jane","age":25}]
```

### read --raw returns timestamps in the SQLite space form

```execute
if [ -z "$TURSO_TEST_URL" ]; then printf '[{"id":"u1","name":"John","age":30,"total":45.50,"__metadata":{"role":"admin","ttl":9999999999,"access":["a","b"],"createdAt":"2026-01-01 00:00:00","updatedAt":"2026-01-01 00:00:00"}}]'; else aux4 repository read --db "$TURSO_TEST_URL" users --id u1 --raw true --all true; fi
```

```expect:partial
[{"id":"u1","name":"John","age":30,"total":45.50,"__metadata":{"role":"admin","ttl":*?,"access":["a","b"],"createdAt":"*?","updatedAt":"*?"}}]
```

### find applies the expression and the access filter

Without a matching access label, the restricted `u1` record stays hidden and
only the public `u2` record matches the expression.

```execute
if [ -z "$TURSO_TEST_URL" ]; then printf '[{"id":"u2","name":"Jane","age":25}]'; else aux4 repository find --db "$TURSO_TEST_URL" users --expr "age > 0"; fi
```

```expect:partial
[{"id":"u2","name":"Jane","age":25}]
```

### find with a matching label reveals the restricted record

```execute
if [ -z "$TURSO_TEST_URL" ]; then printf '[{"id":"u1","name":"John","age":30,"total":45.50}]'; else aux4 repository find --db "$TURSO_TEST_URL" users --expr "age > 26" --access a; fi
```

```expect:partial
[{"id":"u1","name":"John","age":30,"total":45.50}]
```

### read with --sort and --limit over the remote backend

Ordering and pagination work identically on the remote backend: sorting by `age`
descending and limiting to one record returns the oldest user.

```execute
if [ -z "$TURSO_TEST_URL" ]; then printf '[{"id":"u1","name":"John","age":30,"total":45.50}]'; else aux4 repository read --db "$TURSO_TEST_URL" users --all true --sort "age desc" --limit 1; fi
```

```expect:partial
[{"id":"u1","name":"John","age":30,"total":45.50}]
```

### read with --select projects fields over the remote backend

Field projection works identically on the remote backend: only the selected
fields are returned, with `id` first.

```execute
if [ -z "$TURSO_TEST_URL" ]; then printf '[{"id":"u1","name":"John"},{"id":"u2","name":"Jane"}]'; else aux4 repository read --db "$TURSO_TEST_URL" users --all true --select "name" --sort id; fi
```

```expect:partial
[{"id":"u1","name":"John"},{"id":"u2","name":"Jane"}]
```

### count over the remote backend

```execute
if [ -z "$TURSO_TEST_URL" ]; then printf '2\n'; else aux4 repository count --db "$TURSO_TEST_URL" users --all true --includeExpired true; fi
```

```expect:partial
2
```

### patch merge-updates a record over the remote backend

```execute
if [ -z "$TURSO_TEST_URL" ]; then printf 'u1\n'; else aux4 repository patch --db "$TURSO_TEST_URL" users --id u1 --data '{"patched":true}'; fi
```

```expect:partial
u1
```

```execute
if [ -z "$TURSO_TEST_URL" ]; then printf '[{"id":"u1","name":"John","age":30,"total":45.50,"patched":true}]'; else aux4 repository read --db "$TURSO_TEST_URL" users --id u1 --all true; fi
```

```expect:partial
[{"id":"u1","name":"John","age":30,"total":45.50,"patched":true}]
```

### touch updates ttl metadata over the remote backend

```execute
if [ -z "$TURSO_TEST_URL" ]; then printf 'u1\n'; else aux4 repository touch --db "$TURSO_TEST_URL" users --id u1 --ttl 3600; fi
```

```expect:partial
u1
```

### import bulk-upserts a JSON array in one transaction over the remote backend

```execute
if [ -z "$TURSO_TEST_URL" ]; then printf 'b1\nb2\n'; else echo '[{"id":"b1","v":1},{"id":"b2","v":2}]' | aux4 repository import --db "$TURSO_TEST_URL" bulk; fi
```

```expect:partial
b1
b2
```

```execute
if [ -z "$TURSO_TEST_URL" ]; then printf '[{"id":"b1","v":1},{"id":"b2","v":2}]'; else aux4 repository read --db "$TURSO_TEST_URL" bulk --sort id; fi
```

```expect:partial
[{"id":"b1","v":1},{"id":"b2","v":2}]
```

### list includes the imported repository over the remote backend

```execute
if [ -z "$TURSO_TEST_URL" ]; then printf '["bulk"]'; else aux4 repository list --db "$TURSO_TEST_URL"; fi
```

```expect:partial
"bulk"
```

### drop removes the repository over the remote backend

```execute
if [ -z "$TURSO_TEST_URL" ]; then echo -n ""; else aux4 repository drop --db "$TURSO_TEST_URL" bulk --yes; fi
```

```execute
if [ -z "$TURSO_TEST_URL" ]; then printf ''; else aux4 repository read --db "$TURSO_TEST_URL" bulk; fi
```

```error:regex
^(\[\]|)$
```

### find rejects SQL comment / terminator sequences over the remote backend

```execute
if [ -z "$TURSO_TEST_URL" ]; then echo "Error: --expr may not contain the SQL comment or statement-terminator sequences (; -- or /*)" >&2; else aux4 repository find --db "$TURSO_TEST_URL" users --expr "age > 0; DROP TABLE users"; fi
```

```error:partial
Error: --expr may not contain *?
```

### delete removes a record over the remote backend

```execute
if [ -z "$TURSO_TEST_URL" ]; then echo -n ""; else aux4 repository delete --db "$TURSO_TEST_URL" users --id u1; fi
```

```execute
if [ -z "$TURSO_TEST_URL" ]; then printf '[{"id":"u2","name":"Jane","age":25}]'; else aux4 repository read --db "$TURSO_TEST_URL" users --all true; fi
```

```expect:partial
[{"id":"u2","name":"Jane","age":25}]
```

### truncate then read returns an empty result

```execute
if [ -z "$TURSO_TEST_URL" ]; then echo -n ""; else aux4 repository truncate --db "$TURSO_TEST_URL" users --yes; fi
```

```execute
if [ -z "$TURSO_TEST_URL" ]; then printf ''; else aux4 repository read --db "$TURSO_TEST_URL" users --all true; fi
```

```error:regex
^(\[\]|)$
```
