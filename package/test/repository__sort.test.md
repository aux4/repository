# repository sort and pagination

`read` and `find` accept `--sort`, `--limit`, and `--offset`. `--sort` is a
comma-separated list of `field [asc|desc]` clauses (`asc` is the default), where
data fields become `json_extract` lookups (dotted paths supported) and the
special fields `id`, `createdAt`, `updatedAt` map to the record's id and
timestamps. `--limit` / `--offset` are non-negative integers applied as bound
parameters on the outer query, so ordering and pagination compose with `--expr`,
`--access`, and ttl visibility.

## read ordering by data fields

```beforeEach
rm -f .local.db
aux4 repository write nums --id a --data '{"n":30}'
aux4 repository write nums --id b --data '{"n":25}'
aux4 repository write nums --id c --data '{"n":35}'
```

```afterAll
rm -f .local.db
```

### sorts ascending by default

```execute
aux4 repository read nums --sort "n"
```

```expect:partial
[
  {
    "id": "b",
    "n": 25
  },
  {
    "id": "a",
    "n": 30
  },
  {
    "id": "c",
    "n": 35
  }
]
```

### sorts descending with an explicit direction

```execute
aux4 repository read nums --sort "n desc"
```

```expect:partial
[
  {
    "id": "c",
    "n": 35
  },
  {
    "id": "a",
    "n": 30
  },
  {
    "id": "b",
    "n": 25
  }
]
```

## read ordering by multiple fields

```beforeEach
rm -f .local.db
aux4 repository write people --id p1 --data '{"age":30,"name":"Bob"}'
aux4 repository write people --id p2 --data '{"age":30,"name":"Alice"}'
aux4 repository write people --id p3 --data '{"age":25,"name":"Zoe"}'
```

```afterAll
rm -f .local.db
```

### applies each clause in order (age desc, then name asc)

```execute
aux4 repository read people --sort "age desc, name"
```

```expect:partial
[
  {
    "id": "p2",
    "age": 30,
    "name": "Alice"
  },
  {
    "id": "p1",
    "age": 30,
    "name": "Bob"
  },
  {
    "id": "p3",
    "age": 25,
    "name": "Zoe"
  }
]
```

## read ordering by a nested field

```beforeEach
rm -f .local.db
aux4 repository write dev --id d1 --data '{"specs":{"rank":2}}'
aux4 repository write dev --id d2 --data '{"specs":{"rank":1}}'
aux4 repository write dev --id d3 --data '{"specs":{"rank":3}}'
```

```afterAll
rm -f .local.db
```

### sorts by a dotted path

```execute
aux4 repository read dev --sort "specs.rank"
```

```expect:partial
[
  {
    "id": "d2",
    "specs": {
      "rank": 1
    }
  },
  {
    "id": "d1",
    "specs": {
      "rank": 2
    }
  },
  {
    "id": "d3",
    "specs": {
      "rank": 3
    }
  }
]
```

## read ordering by special fields

```beforeEach
rm -f .local.db
aux4 repository write ord --id c --data '{"v":3}'
aux4 repository write ord --id a --data '{"v":1}'
aux4 repository write ord --id b --data '{"v":2}'
```

```afterAll
rm -f .local.db
```

### sorts by the id column

```execute
aux4 repository read ord --sort "id"
```

```expect:partial
[
  {
    "id": "a",
    "v": 1
  },
  {
    "id": "b",
    "v": 2
  },
  {
    "id": "c",
    "v": 3
  }
]
```

## read ordering by createdAt

```beforeEach
rm -f .local.db
aux4 repository write ev --id first --data '{"v":1}'
sleep 1
aux4 repository write ev --id second --data '{"v":2}'
```

```afterAll
rm -f .local.db
```

### newest-first with createdAt desc

```execute
aux4 repository read ev --sort "createdAt desc"
```

```expect:partial
[
  {
    "id": "second",
    "v": 2
  },
  {
    "id": "first",
    "v": 1
  }
]
```

## read ordering by updatedAt

```beforeEach
rm -f .local.db
aux4 repository write up --id x --data '{"v":1}'
aux4 repository write up --id y --data '{"v":2}'
sleep 1
aux4 repository write up --id x --data '{"v":1}'
```

```afterAll
rm -f .local.db
```

### most-recently-updated first with updatedAt desc

```execute
aux4 repository read up --sort "updatedAt desc"
```

```expect:partial
[
  {
    "id": "x",
    "v": 1
  },
  {
    "id": "y",
    "v": 2
  }
]
```

## read pagination

```beforeEach
rm -f .local.db
aux4 repository write pg --id r1 --data '{"n":1}'
aux4 repository write pg --id r2 --data '{"n":2}'
aux4 repository write pg --id r3 --data '{"n":3}'
aux4 repository write pg --id r4 --data '{"n":4}'
aux4 repository write pg --id r5 --data '{"n":5}'
```

```afterAll
rm -f .local.db
```

### --limit returns the first N in sort order

```execute
aux4 repository read pg --sort "n" --limit 2
```

```expect:partial
[
  {
    "id": "r1",
    "n": 1
  },
  {
    "id": "r2",
    "n": 2
  }
]
```

### --limit with --offset paginates

```execute
aux4 repository read pg --sort "n" --limit 2 --offset 2
```

```expect:partial
[
  {
    "id": "r3",
    "n": 3
  },
  {
    "id": "r4",
    "n": 4
  }
]
```

### --offset without --limit skips leading records

```execute
aux4 repository read pg --sort "n" --offset 3
```

```expect:partial
[
  {
    "id": "r4",
    "n": 4
  },
  {
    "id": "r5",
    "n": 5
  }
]
```

## find ordering and pagination

```beforeEach
rm -f .local.db
aux4 repository write items --id i1 --data '{"cat":"a","n":3}'
aux4 repository write items --id i2 --data '{"cat":"a","n":1}'
aux4 repository write items --id i3 --data '{"cat":"b","n":2}'
```

```afterAll
rm -f .local.db
```

### sorts matches of an --expr

```execute
aux4 repository find items --expr "cat = 'a'" --sort "n"
```

```expect:partial
[
  {
    "id": "i2",
    "cat": "a",
    "n": 1
  },
  {
    "id": "i1",
    "cat": "a",
    "n": 3
  }
]
```

### applies --sort and --limit together

```execute
aux4 repository find items --expr "n > 0" --sort "n" --limit 2
```

```expect:partial
[
  {
    "id": "i2",
    "cat": "a",
    "n": 1
  },
  {
    "id": "i3",
    "cat": "b",
    "n": 2
  }
]
```

### applies --sort desc with --offset

```execute
aux4 repository find items --expr "n > 0" --sort "n desc" --offset 1
```

```expect:partial
[
  {
    "id": "i3",
    "cat": "b",
    "n": 2
  },
  {
    "id": "i2",
    "cat": "a",
    "n": 1
  }
]
```

## ordering composes with access restriction

```beforeEach
rm -f .local.db
aux4 repository write docs --id pub --data '{"n":1}'
aux4 repository write docs --id s1 --data '{"n":2}' --access secret
aux4 repository write docs --id s2 --data '{"n":3}' --access secret
```

```afterAll
rm -f .local.db
```

### find with --access sorts only the authorized rowset

```execute
aux4 repository find docs --expr "n > 0" --access secret --sort "n desc"
```

```expect:partial
[
  {
    "id": "s2",
    "n": 3
  },
  {
    "id": "s1",
    "n": 2
  },
  {
    "id": "pub",
    "n": 1
  }
]
```

### read with --access, --sort and --limit

```execute
aux4 repository read docs --access secret --sort "n" --limit 2
```

```expect:partial
[
  {
    "id": "pub",
    "n": 1
  },
  {
    "id": "s1",
    "n": 2
  }
]
```

## rejection cases

```beforeEach
rm -f .local.db
aux4 repository write nums --id a --data '{"n":1}'
```

```afterAll
rm -f .local.db
```

### read rejects an invalid sort field

```execute
aux4 repository read nums --sort "bad-field"
```

```error:partial
Error: invalid --sort field *?
```

### read rejects an invalid sort direction

```execute
aux4 repository read nums --sort "n up"
```

```error:partial
Error: invalid --sort direction *?
```

### read rejects a non-integer --limit

```execute
aux4 repository read nums --limit abc
```

```error:partial
Error: invalid --limit value *?
```

### read rejects a negative --limit

```execute
aux4 repository read nums --limit=-5
```

```error:partial
Error: invalid --limit value *?
```

### read rejects a non-integer --offset

```execute
aux4 repository read nums --offset xyz
```

```error:partial
Error: invalid --offset value *?
```

### find rejects an invalid sort field

```execute
aux4 repository find nums --expr "n > 0" --sort "n;"
```

```error:partial
Error: invalid --sort field *?
```

### find rejects a non-integer --limit

```execute
aux4 repository find nums --expr "n > 0" --limit abc
```

```error:partial
Error: invalid --limit value *?
```
