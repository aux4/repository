# repository select (field projection)

`read` and `find` accept `--select "a, b, c"` to project a subset of fields, in
the order listed. `id` is always included first; fields absent from a document
are omitted (not `null`); dotted paths (`address.city`) navigate nested objects
and are keyed by the path as written. Projection happens in Go — no field name
ever reaches the query, so there is zero injection surface. `--select` composes
with `--expr`, `--sort`, `--limit`/`--offset`, `--access`, and `--raw`. Output
stays compact single-line JSON with number literals preserved.

## read projection

```beforeEach
rm -f .local.db
aux4 repository write users --id u1 --data '{"first_name":"John","last_name":"Doe","age":30,"total":45.50,"address":{"city":"NYC","zip":"10001"}}'
aux4 repository write users --id u2 --data '{"first_name":"Jane","last_name":"Smith","age":25,"address":{"city":"LA"}}'
aux4 repository write users --id u3 --data '{"first_name":"Bob","last_name":"Jones","age":35,"total":12.00}'
```

```afterAll
rm -f .local.db
```

### projects a single field (id first)

```execute
aux4 repository read users --select "first_name" --sort id
```

```expect:partial
[{"id":"u1","first_name":"John"},{"id":"u2","first_name":"Jane"},{"id":"u3","first_name":"Bob"}]
```

### projects multiple fields in the requested order

Output field order matches `--select` order (age before first_name), not storage order.

```execute
aux4 repository read users --select "age, first_name" --sort id
```

```expect:partial
[{"id":"u1","age":30,"first_name":"John"},{"id":"u2","age":25,"first_name":"Jane"},{"id":"u3","age":35,"first_name":"Bob"}]
```

### omits fields absent from a document

u2 has no `total`, so its row simply lacks the key (no `null`). Number literals are preserved.

```execute
aux4 repository read users --select "total" --sort id
```

```expect:partial
[{"id":"u1","total":45.50},{"id":"u2"},{"id":"u3","total":12.00}]
```

### navigates a dotted nested path (keyed by the path as written)

u3 has no `address`, so `address.city` is omitted for it.

```execute
aux4 repository read users --select "first_name, address.city" --sort id
```

```expect:partial
[{"id":"u1","first_name":"John","address.city":"NYC"},{"id":"u2","first_name":"Jane","address.city":"LA"},{"id":"u3","first_name":"Bob"}]
```

### includes id first even when not listed

```execute
aux4 repository read users --id u1 --select "last_name"
```

```expect:partial
[{"id":"u1","last_name":"Doe"}]
```

### does not duplicate id when it is explicitly listed

```execute
aux4 repository read users --id u1 --select "id, first_name"
```

```expect:partial
[{"id":"u1","first_name":"John"}]
```

### empty --select returns all fields

```execute
aux4 repository read users --id u2
```

```expect:partial
[{"id":"u2","first_name":"Jane","last_name":"Smith","age":25,"address":{"city":"LA"}}]
```

### composes with --sort

```execute
aux4 repository read users --select "first_name" --sort "age desc"
```

```expect:partial
[{"id":"u3","first_name":"Bob"},{"id":"u1","first_name":"John"},{"id":"u2","first_name":"Jane"}]
```

### composes with --sort and --limit

```execute
aux4 repository read users --select "first_name" --sort "age" --limit 2
```

```expect:partial
[{"id":"u2","first_name":"Jane"},{"id":"u1","first_name":"John"}]
```

### composes with --raw (metadata appended after selected fields)

```execute
aux4 repository read users --id u1 --select "first_name" --raw true
```

```expect:partial
[{"id":"u1","first_name":"John","__metadata":{"createdAt":"*?","updatedAt":"*?"}}]
```

### rejects an invalid field name

```execute
aux4 repository read users --select "bad-field"
```

```error:partial
Error: invalid --select field *?
```

## find projection

```beforeEach
rm -f .local.db
aux4 repository write users --id u1 --data '{"first_name":"John","last_name":"Doe","age":30,"total":45.50}'
aux4 repository write users --id u2 --data '{"first_name":"Jane","last_name":"Smith","age":25}'
aux4 repository write users --id u3 --data '{"first_name":"Bob","last_name":"Jones","age":35,"total":12.00}'
```

```afterAll
rm -f .local.db
```

### projects fields of matched records

```execute
aux4 repository find users --expr "age > 0" --select "last_name" --sort id
```

```expect:partial
[{"id":"u1","last_name":"Doe"},{"id":"u2","last_name":"Smith"},{"id":"u3","last_name":"Jones"}]
```

### composes with --expr and preserves order and number literals

```execute
aux4 repository find users --expr "age > 26" --select "total, first_name" --sort id
```

```expect:partial
[{"id":"u1","total":45.50,"first_name":"John"},{"id":"u3","total":12.00,"first_name":"Bob"}]
```

### rejects an invalid field name

```execute
aux4 repository find users --expr "age > 0" --select "first.name!"
```

```error:partial
Error: invalid --select field *?
```

## projection composes with access restriction

```beforeEach
rm -f .local.db
aux4 repository write docs --id pub --data '{"name":"Public","n":1}'
aux4 repository write docs --id sec --data '{"name":"Secret","n":2}' --access secret
```

```afterAll
rm -f .local.db
```

### find with a matching label projects both authorized rows

```execute
aux4 repository find docs --expr "n > 0" --access secret --select "name" --sort id
```

```expect:partial
[{"id":"pub","name":"Public"},{"id":"sec","name":"Secret"}]
```

### read without the label projects only the public row

```execute
aux4 repository read docs --select "name" --sort id
```

```expect:partial
[{"id":"pub","name":"Public"}]
```
