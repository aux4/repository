# repository access restriction

The `write` command can stamp one or more access labels onto a record under `metadata.access`.
`read` and `find` then restrict which records a caller sees based on the labels passed with
`--access`, with `--all true` bypassing the filter entirely. This is advisory scoping enforced
at query time, not a hard security boundary.

```beforeEach
rm -f .local.db
```

```afterAll
rm -f .local.db
```

## write stamps access labels

### single label is stored as a JSON array under metadata.access

Writing with `--access a` stores `access: ["a"]` in the record metadata alongside any other
system fields.

```execute
aux4 repository write docs --id ra --data '{"n":"ra"}' --access a
aux4 repository read docs --id ra --raw true --all true
```

```expect:partial
[
  {
    "id": "ra",
    "n": "ra",
    "__metadata": {
      "access": [
        "a"
      ],
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```

### comma-separated labels are split into a multi-element array

Writing with `--access "a,b"` stores `access: ["a","b"]`.

```execute
aux4 repository write docs --id rab --data '{"n":"rab"}' --access "a,b"
aux4 repository read docs --id rab --raw true --all true
```

```expect:partial
[
  {
    "id": "rab",
    "n": "rab",
    "__metadata": {
      "access": [
        "a",
        "b"
      ],
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```

### writing without --access leaves the record public (no access key)

When `--access` is omitted, no `access` key is added and the record is public.

```execute
aux4 repository write docs --id pub --data '{"n":"pub"}'
aux4 repository read docs --id pub --raw true
```

```expect:partial
[
  {
    "id": "pub",
    "n": "pub",
    "__metadata": {
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```

## read honors access labels

```beforeEach
rm -f .local.db
aux4 repository write docs --id pub --data '{"n":"pub"}'
aux4 repository write docs --id ra --data '{"n":"ra"}' --access a
aux4 repository write docs --id rab --data '{"n":"rab"}' --access "a,b"
aux4 repository write docs --id rc --data '{"n":"rc"}' --access c
```

### a matching label reveals the restricted record

Reading as a caller holding label `a` returns the public record plus the records tagged with
`a`.

```execute
aux4 repository read docs --access a
```

```expect:partial
[
  {
    "id": "pub",
    "n": "pub"
  },
  {
    "id": "ra",
    "n": "ra"
  },
  {
    "id": "rab",
    "n": "rab"
  }
]
```

### a non-matching label hides the restricted record

A caller holding only label `z` sees just the public record; `ra`, `rab` and `rc` are hidden.

```execute
aux4 repository read docs --access z
```

```expect:partial
[
  {
    "id": "pub",
    "n": "pub"
  }
]
```

### no --access returns only public records

With no `--access`, restricted records are all hidden and only the public record is returned.

```execute
aux4 repository read docs
```

```expect:partial
[
  {
    "id": "pub",
    "n": "pub"
  }
]
```

### a public record is visible to everyone including no-access reads

The public record `pub` appears for a labeled caller, an unmatched caller, and a caller with no
`--access` at all.

```execute
aux4 repository read docs --access c
```

```expect:partial
[
  {
    "id": "pub",
    "n": "pub"
  },
  {
    "id": "rc",
    "n": "rc"
  }
]
```

### a multi-group record is visible to a caller holding only one of its labels

The record `rab` is tagged `["a","b"]`; a caller holding only `b` can see it.

```execute
aux4 repository read docs --access b
```

```expect:partial
[
  {
    "id": "pub",
    "n": "pub"
  },
  {
    "id": "rab",
    "n": "rab"
  }
]
```

### --all true returns every record including restricted ones

Passing `--all true` bypasses access filtering and returns the public record plus all
restricted records regardless of labels.

```execute
aux4 repository read docs --all true
```

```expect:partial
[
  {
    "id": "pub",
    "n": "pub"
  },
  {
    "id": "ra",
    "n": "ra"
  },
  {
    "id": "rab",
    "n": "rab"
  },
  {
    "id": "rc",
    "n": "rc"
  }
]
```

### reading a restricted record by id without a matching label returns empty

Reading `rc` by id without the `c` label returns an empty array.

```execute
aux4 repository read docs --id rc
```

```error:regex
^(\[\]|)$
```

### reading a restricted record by id with a matching label returns it

```execute
aux4 repository read docs --id rc --access c
```

```expect:partial
[
  {
    "id": "rc",
    "n": "rc"
  }
]
```

## find honors access labels alongside --expr

```beforeEach
rm -f .local.db
aux4 repository write docs --id pub1 --data '{"team":"x","n":1}'
aux4 repository write docs --id pub2 --data '{"team":"y","n":2}'
aux4 repository write docs --id sec1 --data '{"team":"x","n":3}' --access secret
aux4 repository write docs --id sec2 --data '{"team":"y","n":4}' --access secret
```

### expr and access filters both apply

Finding `team = "x"` as a caller with the `secret` label returns both the public and the
restricted record for team x, but never team y records.

```execute
aux4 repository find docs --expr 'team = "x"' --access secret
```

```expect:partial
[
  {
    "id": "pub1",
    "team": "x",
    "n": 1
  },
  {
    "id": "sec1",
    "team": "x",
    "n": 3
  }
]
```

### access is not bypassable through expr

Finding `team = "x"` without the `secret` label returns only the public team-x record; the
restricted record stays hidden even though it matches the expression.

```execute
aux4 repository find docs --expr 'team = "x"'
```

```expect:partial
[
  {
    "id": "pub1",
    "team": "x",
    "n": 1
  }
]
```

### find --all true applies expr but bypasses access

Finding `team = "y"` with `--all true` returns both the public and restricted team-y records.

```execute
aux4 repository find docs --expr 'team = "y"' --all true
```

```expect:partial
[
  {
    "id": "pub2",
    "team": "y",
    "n": 2
  },
  {
    "id": "sec2",
    "team": "y",
    "n": 4
  }
]
```

## access labels are safely quoted

### a label containing SQL metacharacters cannot inject

A caller passing a label crafted to break out of the SQL string (`x') OR ('1'='1`) is treated
as a single opaque label. It matches nothing beyond public records, proving the label is safely
quoted rather than executed.

```execute
aux4 repository write docs --id pub --data '{"n":"pub"}'
aux4 repository write docs --id ra --data '{"n":"ra"}' --access a
aux4 repository read docs --access "x') OR ('1'='1"
```

```expect:partial
[
  {
    "id": "pub",
    "n": "pub"
  }
]
```

## find access filter cannot be bypassed via --expr injection

The `--expr` fragment is composed into SQL. To keep access filtering enforceable, `find`
evaluates the access/visibility predicate in an inner subquery and applies `--expr` only to the
already-restricted rowset, and additionally rejects `--expr` values containing SQL
comment/statement-terminator sequences. These tests confirm a no-access caller can never coax a
restricted row out through a crafted `--expr`.

```beforeEach
rm -f .local.db
aux4 repository write docs --id pub --data '{"t":"doc-public"}'
aux4 repository write docs --id sec --data '{"t":"doc-secret"}' --access secret
```

### a balanced OR 1=1 payload cannot leak restricted rows

`t like '%doc%' OR 1=1` is a syntactically balanced predicate (not caught by the sequence
rejection). Because the access filter runs in the inner subquery, `OR 1=1` can only widen the
selection *within* the already-authorized rowset — the restricted `sec` row is never present, so
a no-access caller still sees only the public row.

```execute
aux4 repository find docs --expr "t like '%doc%' OR 1=1"
```

```expect:partial
[
  {
    "id": "pub",
    "t": "doc-public"
  }
]
```

### an unbalanced parenthesis payload cannot leak restricted rows

`t like '%doc%') OR (1=1` attempts to rebalance parentheses around the old clause layout. With
the inner-subquery shape it still only filters the authorized rowset, so only the public row is
returned.

```execute
aux4 repository find docs --expr "t like '%doc%') OR (1=1"
```

```expect:partial
[
  {
    "id": "pub",
    "t": "doc-public"
  }
]
```

### a paren-close plus line comment is rejected

The original exploit `t like '%doc%') OR 1=1 --` closes the paren and comments out the rest.
The `--` sequence is rejected before any SQL is built.

```execute
aux4 repository find docs --expr "t like '%doc%') OR 1=1 --"
```

```error:partial
Error: --expr may not contain *?
```

### a trailing statement terminator is rejected

```execute
aux4 repository find docs --expr "t like '%doc%'; DROP TABLE docs"
```

```error:partial
Error: --expr may not contain *?
```

### a block comment is rejected

```execute
aux4 repository find docs --expr "t like '%doc%' /* leak */"
```

```error:partial
Error: --expr may not contain *?
```

### a UNION injection fails without leaking

A balanced `UNION SELECT` is not a valid predicate in the outer WHERE, so the query fails with a
parse error and returns no rows rather than appending the restricted record.

```execute
aux4 repository find docs --expr "t like '%doc%' UNION SELECT id,data,metadata,created_at,updated_at FROM docs"
```

```error:partial
Parse error *?
```
