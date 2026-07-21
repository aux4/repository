# repository list

`list` prints the repositories (tables) in a database as a compact JSON array of
names, sorted alphabetically. It takes only `--db` (no repository argument). An
empty database prints `[]` and exits 0 — an empty list is a valid answer.

```beforeEach
rm -f .local.db
```

```afterAll
rm -f .local.db .custom.db
```

### an empty database lists no repositories

```execute
aux4 repository list
```

```expect
[]
```

### lists the repositories sorted by name

```execute
aux4 repository write users --id u1 --data '{"n":1}'
aux4 repository write sessions --id s1 --data '{"n":1}'
aux4 repository write accounts --id a1 --data '{"n":1}'
aux4 repository list
```

```expect:partial
["accounts","sessions","users"]
```

### lists from a custom database

```execute
aux4 repository write things --db .custom.db --id t1 --data '{"n":1}'
aux4 repository list --db .custom.db
```

```expect:partial
["things"]
```
