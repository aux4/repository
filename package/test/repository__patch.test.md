# repository patch

`patch <repository> --id <id> --data <json>` shallow-merges a JSON payload into an
existing record's `data`: top-level keys set/override existing values (kept in
their original position), brand-new keys are appended in payload order, and keys
not mentioned are preserved. It is **update-only** — patching an id that does not
exist is an error (unlike `write`, which upserts). `metadata` (ttl/access) and
`created_at` are preserved; `updated_at` is bumped. It prints the id.

Deep-merge and field deletion are intentionally out of scope for this version.

```beforeEach
rm -f .local.db
aux4 repository write users --id u1 --data '{"name":"John","age":30,"total":45.50}' --metadata '{"role":"admin"}' --ttl 3600 --access "a,b"
```

```afterAll
rm -f .local.db
```

### prints the id on success

```execute
aux4 repository patch users --id u1 --data '{"age":31}'
```

```expect
u1
```

### overrides existing keys in place and appends new keys in payload order

`age` is updated where it already sits; `city` is appended after the existing
fields. Number literals are preserved.

```execute
aux4 repository patch users --id u1 --data '{"age":31,"city":"NYC"}'
aux4 repository read users --id u1 --all true
```

```expect:partial
[{"id":"u1","name":"John","age":31,"total":45.50,"city":"NYC"}]
```

### preserves unmentioned keys and metadata (ttl/access untouched)

```execute
aux4 repository patch users --id u1 --data '{"city":"LA"}'
aux4 repository read users --id u1 --raw true --all true
```

```expect:partial
[{"id":"u1","name":"John","age":30,"total":45.50,"city":"LA","__metadata":{"role":"admin","ttl":*?,"access":["a","b"],"createdAt":"*?","updatedAt":"*?"}}]
```

### ignores an id field in the patch payload

The `id` comes from `--id`; any `id` in the payload is stripped (not stored in data).

```execute
aux4 repository patch users --id u1 --data '{"id":"HACKED","note":"ok"}'
aux4 repository read users --id u1 --all true
```

```expect:partial
[{"id":"u1","name":"John","age":30,"total":45.50,"note":"ok"}]
```

### preserves number literals in patched values

```execute
aux4 repository patch users --id u1 --data '{"bonus":12.00}'
aux4 repository read users --id u1 --all true
```

```expect:partial
[{"id":"u1","name":"John","age":30,"total":45.50,"bonus":12.00}]
```

### reads the payload from stdin

```execute
echo '{"via":"stdin"}' | aux4 repository patch users --id u1
aux4 repository read users --id u1 --all true
```

```expect:partial
[{"id":"u1","name":"John","age":30,"total":45.50,"via":"stdin"}]
```

### errors when the id does not exist (update-only, not an upsert)

```execute
aux4 repository patch users --id nope --data '{"x":1}'
```

```error:partial
Error: id "nope" not found *?
```

### the failed patch does not create the record

```execute
aux4 repository read users --id nope
```

```error:regex
^(\[\]|)$
```
