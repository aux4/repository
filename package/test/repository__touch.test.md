# repository touch

`touch <repository> --id <id>` updates only a record's `ttl` and/or `access`
metadata without changing its `data`, and prints the id. `--ttl` sets
`ttl = now + seconds`; `--access` replaces the access labels array. At least one
of `--ttl` / `--access` must be provided. It is **update-only** — touching an id
that does not exist is an error. `data` is left untouched; `updated_at` is bumped.

Clearing ttl/access (making a record permanent/public again) is out of scope for
this version — `touch` sets values, it does not unset them.

```beforeEach
rm -f .local.db
aux4 repository write users --id u1 --data '{"name":"John","total":45.50}' --metadata '{"role":"admin"}'
```

```afterAll
rm -f .local.db
```

### prints the id on success

```execute
aux4 repository touch users --id u1 --ttl 3600
```

```expect
u1
```

### --ttl sets an expiry epoch, preserving other metadata and the data

```execute
aux4 repository touch users --id u1 --ttl 3600
aux4 repository read users --id u1 --raw true --all true
```

```expect:partial
[{"id":"u1","name":"John","total":45.50,"__metadata":{"role":"admin","ttl":*?,"createdAt":"*?","updatedAt":"*?"}}]
```

### --access replaces the access labels array

```execute
aux4 repository touch users --id u1 --access "x,y"
aux4 repository read users --id u1 --raw true --all true
```

```expect:partial
[{"id":"u1","name":"John","total":45.50,"__metadata":{"role":"admin","access":["x","y"],"createdAt":"*?","updatedAt":"*?"}}]
```

### --ttl and --access together update both

```execute
aux4 repository touch users --id u1 --ttl 60 --access "z"
aux4 repository read users --id u1 --raw true --all true
```

```expect:partial
[{"id":"u1","name":"John","total":45.50,"__metadata":{"role":"admin","ttl":*?,"access":["z"],"createdAt":"*?","updatedAt":"*?"}}]
```

### leaves the record data unchanged (number literals preserved)

```execute
aux4 repository touch users --id u1 --ttl 60
aux4 repository read users --id u1
```

```expect:partial
[{"id":"u1","name":"John","total":45.50}]
```

### errors when neither --ttl nor --access is provided

```execute
aux4 repository touch users --id u1
```

```error:partial
Error: touch requires at least one of *?
```

### errors when the id does not exist (update-only)

```execute
aux4 repository touch users --id nope --ttl 60
```

```error:partial
Error: id "nope" not found *?
```
