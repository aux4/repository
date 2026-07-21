# repository find smoke test (cross-platform sed regression guard)

This is the explicit regression guard for the GNU/BSD `sed -i` portability bug
(fixed in API-005 via pkger multi-platform overrides). The `find` command rewrites
the user's `--expr` into a `json_extract(...)` WHERE clause using an in-place
`sed -i` edit. The correct flag form differs between platforms:

- BSD sed (macOS): `sed -i '' -E`
- GNU sed (Linux): `sed -i -E`

If the wrong variant ships for a platform, `find` fails: on Linux, the BSD form's
empty `''` argument is read by GNU sed as the SCRIPT and the real `s///` expression
as a FILENAME, producing `sed: can't read s/...: No such file or directory` and an
empty/erroring result. This test runs a REAL `find` against a seeded database and
asserts the correct filtered row — so it FAILS on Linux if the BSD variant were
wrongly installed (i.e. it would have failed before the API-005 fix).

## Test

```beforeAll
rm -f .smoke.db
```

```afterAll
rm -f .smoke.db
```

### seed a repository with owned records

Write three records owned by different owners into a dedicated database.

```execute
aux4 repository write items --db .smoke.db --id item1 --data '{"owner":"alice","label":"first"}'
aux4 repository write items --db .smoke.db --id item2 --data '{"owner":"bob","label":"second"}'
aux4 repository write items --db .smoke.db --id item3 --data '{"owner":"alice","label":"third"}'
```

```expect:partial
item1
item2
item3
```

### find by owner equality returns only matching rows

This is the exact `owner=...` expression the canonical aux4-apps/contacts app relies
on. The `sed` rewrite must turn `owner = 'alice'` into
`json_extract(data, '$.owner') = 'alice'`. On Linux with the BSD variant this step
errors and returns no rows — so this assertion is the regression guard.

```execute
aux4 repository find items --db .smoke.db --expr "owner = 'alice'"
```

```expect:partial
[{"id":"item1","owner":"alice","label":"first"},{"id":"item3","owner":"alice","label":"third"}]
```

### find by owner excludes non-matching rows

Ensure the rewritten expression actually filters — bob's record must not appear.

```execute
aux4 repository find items --db .smoke.db --expr "owner = 'bob'"
```

```expect:partial
[{"id":"item2","owner":"bob","label":"second"}]
```
