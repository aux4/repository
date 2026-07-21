# repository read/find TTY rendering

On an interactive terminal, `read` and `find` render their JSON as a table (via
`aux4 render table`); when the output is piped, redirected, or captured by another
aux4 command it stays raw JSON, unchanged. `--render none` forces raw JSON even on
a terminal. These tests focus on the raw-JSON contract that downstream consumers
depend on; the interactive table path is checked with a pseudo-terminal.

```beforeAll
rm -f .render.db
aux4 repository write --db .render.db docs --id rr1 --data '{"name":"John","total":45.50}'
aux4 repository write --db .render.db docs --id rr2 --data '{"name":"Jane","total":10.00}'
```

```afterAll
rm -f .render.db
```

## piped output stays raw JSON

### read piped is raw JSON with number literals preserved

```execute
aux4 repository read --db .render.db docs --sort id
```

```expect:partial
[
  {
    "id": "rr1",
    "name": "John",
    "total": 45.50
  },
  {
    "id": "rr2",
    "name": "Jane",
    "total": 10.00
  }
]
```

### read piped emits the JSON array exactly once

```execute
aux4 repository read --db .render.db docs --sort id | grep -c '^\['
```

```expect
1
```

### find piped is raw JSON with number literals preserved

```execute
aux4 repository find --db .render.db docs --expr "total > 0" --sort id
```

```expect:partial
[
  {
    "id": "rr1",
    "name": "John",
    "total": 45.50
  },
  {
    "id": "rr2",
    "name": "Jane",
    "total": 10.00
  }
]
```

## --render none forces raw JSON

### read --render none is raw JSON

```execute
aux4 repository read --db .render.db docs --sort id --render none
```

```expect:partial
[
  {
    "id": "rr1",
    "name": "John",
    "total": 45.50
  },
  {
    "id": "rr2",
    "name": "Jane",
    "total": 10.00
  }
]
```

### find --render none is raw JSON

```execute
aux4 repository find --db .render.db docs --expr "total > 0" --sort id --render none
```

```expect:partial
[
  {
    "id": "rr1",
    "name": "John",
    "total": 45.50
  },
  {
    "id": "rr2",
    "name": "Jane",
    "total": 10.00
  }
]
```

## interactive terminal renders a table

The interactive table path is driven by aux4 core's isatty detection and the
`render.tty` entry on the `read`/`find` commands; on a real terminal the JSON is
piped through `aux4 render table`. This is verified manually with a
pseudo-terminal (pty output is not asserted here because its exact bytes — column
widths, control characters, echo — vary by terminal/environment):

```bash
# renders a table with id / name / total columns on a terminal
script -q /dev/null aux4 repository read --db .render.db docs

#  id   name  total
#  rr1  John   45.5
#  rr2  Jane     10
```

Piping the same command (as every test above does) yields the raw JSON array
unchanged, and `--render none` forces raw JSON even on a terminal.
