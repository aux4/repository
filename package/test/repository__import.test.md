# repository import

`import <repository>` reads a JSON array on stdin and upserts every element in a
single transaction (rolled back on any error, so an import either fully applies
or writes nothing). Per element the id is resolved like `write` (the element's
`id` if present, otherwise a generated lowercase UUID v4) and `data` is the
element with `id` stripped. Optional `--ttl` / `--access` apply uniformly to
every imported record. It prints one id per line, in input order.

Per-element metadata is out of scope for this version.

```beforeEach
rm -f .local.db
```

```afterAll
rm -f .local.db
```

### upserts every element and prints ids in input order

```execute
echo '[{"id":"a","n":1},{"id":"b","n":2},{"id":"c","n":3}]' | aux4 repository import users
```

```expect
a
b
c
```

### stored records round-trip with number literals preserved

```execute
echo '[{"id":"p1","price":45.50},{"id":"p2","price":12.00}]' | aux4 repository import prices
aux4 repository read prices --sort id
```

```expect:partial
[{"id":"p1","price":45.50},{"id":"p2","price":12.00}]
```

### generates a UUID for an element without an id

```execute
echo '[{"n":1}]' | aux4 repository import gen
```

```expect:regex
^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$
```

### an id inside an element is used and stripped from data

```execute
echo '[{"id":"e1","name":"Ann"}]' | aux4 repository import users
aux4 repository read users --id e1
```

```expect:partial
[{"id":"e1","name":"Ann"}]
```

### re-importing the same id overwrites (upsert)

```execute
echo '[{"id":"a","n":1}]' | aux4 repository import users
echo '[{"id":"a","n":99,"extra":true}]' | aux4 repository import users
aux4 repository read users --id a
```

```expect:partial
[{"id":"a","n":99,"extra":true}]
```

### --ttl and --access apply uniformly to every imported record

```execute
echo '[{"id":"s1","v":1},{"id":"s2","v":2}]' | aux4 repository import sessions --ttl 3600 --access "team"
aux4 repository read sessions --id s1 --raw true --all true
```

```expect:partial
[{"id":"s1","v":1,"__metadata":{"ttl":*?,"access":["team"],"createdAt":"*?","updatedAt":"*?"}}]
```

### an empty array writes nothing and exits 0

```execute
echo '[]' | aux4 repository import users
```

```expect

```

### an empty array leaves the repository empty

```execute
echo '[]' | aux4 repository import empty_repo
aux4 repository read empty_repo
```

```error:regex
^(\[\]|)$
```

### a non-array payload is rejected

```execute
echo '{"not":"an array"}' | aux4 repository import users
```

```error:partial
Error: import expects a JSON array *?
```

### malformed JSON is rejected and nothing is written

```execute
echo '[{"id":"a"' | aux4 repository import users
```

```error:partial
Error: import expects a JSON array *?
```

### the rejected import wrote no records

```execute
aux4 repository read users
```

```error:regex
^(\[\]|)$
```
