# repository --db cloud (aux4.cloud turso:database capability)

`--db cloud` means "use the Turso database that the aux4.cloud `turso:database`
capability already resolved for this machine": the database URL comes from the
`TURSO_DATABASE_URL` environment variable, the auth token from `TURSO_AUTH_TOKEN`
(both written by the capability resolution before the command runs). No network
call or terraform is involved in this test -- it talks to a local libSQL server.

## without TURSO_DATABASE_URL

`--db cloud` fails fast with a clear error when the capability env var is
missing, regardless of what else is set in the ambient environment.

### write fails with a clear capability error

```execute
env -u TURSO_DATABASE_URL -u TURSO_AUTH_TOKEN aux4 repository write users --db cloud --id u1 --data '{"name":"John"}'
```

```error
Error: --db cloud needs the aux4.cloud turso:database capability; TURSO_DATABASE_URL is not set
```

### read fails with the same clear error

```execute
env -u TURSO_DATABASE_URL -u TURSO_AUTH_TOKEN aux4 repository read users --db cloud
```

```error
Error: --db cloud needs the aux4.cloud turso:database capability; TURSO_DATABASE_URL is not set
```

## with TURSO_DATABASE_URL (opt-in, against a local libSQL server)

Opt-in conformance test. It only runs against a real libSQL server when
`TURSO_CLOUD_TEST_URL` is set (an `http://`/`https://` endpoint, with
`TURSO_CLOUD_TEST_TOKEN` in the environment when the server requires auth).
When `TURSO_CLOUD_TEST_URL` is unset, each step is a clean no-op that prints the
expected output, so a normal `aux4 test run` and CI stay green without
credentials.

To exercise the real path, start a local libSQL server and run:

```bash
sqld --db-path ./cloud-test.db --http-listen-addr 127.0.0.1:8099 &
TURSO_CLOUD_TEST_URL=http://127.0.0.1:8099 aux4 test run test/repository__db_cloud.test.md
```

```beforeAll
if [ -n "$TURSO_CLOUD_TEST_URL" ]; then TURSO_DATABASE_URL="$TURSO_CLOUD_TEST_URL" TURSO_AUTH_TOKEN="$TURSO_CLOUD_TEST_TOKEN" aux4 repository truncate --db cloud people --yes >/dev/null 2>&1 || true; fi
```

### write resolves --db cloud to TURSO_DATABASE_URL and returns the record id

```execute
if [ -z "$TURSO_CLOUD_TEST_URL" ]; then echo "p1"; else TURSO_DATABASE_URL="$TURSO_CLOUD_TEST_URL" TURSO_AUTH_TOKEN="$TURSO_CLOUD_TEST_TOKEN" aux4 repository write people --db cloud --id p1 --data '{"name":"John","age":30}'; fi
```

```expect
p1
```

### read over --db cloud returns the record written over --db cloud

```execute
if [ -z "$TURSO_CLOUD_TEST_URL" ]; then printf '[{"id":"p1","name":"John","age":30}]'; else TURSO_DATABASE_URL="$TURSO_CLOUD_TEST_URL" TURSO_AUTH_TOKEN="$TURSO_CLOUD_TEST_TOKEN" aux4 repository read people --db cloud --id p1; fi
```

```expect:partial
[{"id":"p1","name":"John","age":30}]
```

### find over --db cloud applies the SQL expression

```execute
if [ -z "$TURSO_CLOUD_TEST_URL" ]; then printf '[{"id":"p1","name":"John","age":30}]'; else TURSO_DATABASE_URL="$TURSO_CLOUD_TEST_URL" TURSO_AUTH_TOKEN="$TURSO_CLOUD_TEST_TOKEN" aux4 repository find people --db cloud --expr "age > 20"; fi
```

```expect:partial
[{"id":"p1","name":"John","age":30}]
```

### count over --db cloud

```execute
if [ -z "$TURSO_CLOUD_TEST_URL" ]; then printf '1\n'; else TURSO_DATABASE_URL="$TURSO_CLOUD_TEST_URL" TURSO_AUTH_TOKEN="$TURSO_CLOUD_TEST_TOKEN" aux4 repository count people --db cloud; fi
```

```expect:partial
1
```

### truncate over --db cloud then read returns an empty result

```execute
if [ -z "$TURSO_CLOUD_TEST_URL" ]; then echo -n ""; else TURSO_DATABASE_URL="$TURSO_CLOUD_TEST_URL" TURSO_AUTH_TOKEN="$TURSO_CLOUD_TEST_TOKEN" aux4 repository truncate --db cloud people --yes; fi
```

```execute
if [ -z "$TURSO_CLOUD_TEST_URL" ]; then printf ''; else TURSO_DATABASE_URL="$TURSO_CLOUD_TEST_URL" TURSO_AUTH_TOKEN="$TURSO_CLOUD_TEST_TOKEN" aux4 repository read people --db cloud --all true; fi
```

```error:regex
^(\[\]|)$
```
