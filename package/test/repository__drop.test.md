# repository drop

`drop <repository>` removes an entire repository — its table, schema and all its
records. It prompts for confirmation (bypassed with `--yes`). Unlike `truncate`,
which empties a repository but keeps its schema, `drop` removes the whole table.
Dropping a repository that does not exist is an idempotent no-op (exit 0).

```beforeEach
rm -f .local.db
aux4 repository write users --id u1 --data '{"n":1}'
aux4 repository write sessions --id s1 --data '{"n":1}'
```

```afterAll
rm -f .local.db
```

### removes the repository from the list (--yes bypasses the prompt)

```execute
aux4 repository drop users --yes
aux4 repository list
```

```expect:partial
["sessions"]
```

### a dropped repository has no records

Reading a dropped repository returns an empty result (the data is gone; the table
is recreated empty on read).

```execute
aux4 repository drop users --yes
aux4 repository read users
```

```error:regex
^(\[\]|)$
```

### dropping a non-existent repository is a no-op

```execute
aux4 repository drop nonexistent --yes
aux4 repository list
```

```expect:partial
["sessions","users"]
```

### dropping one repository leaves the others intact

```execute
aux4 repository drop sessions --yes
aux4 repository read users --id u1
```

```expect:partial
[{"id":"u1","n":1}]
```
