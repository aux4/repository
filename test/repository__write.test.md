## Test

```beforeEach
rm -f .local.db
```

### Write a record with explicit id, data, and metadata

Ensure that when an `id`, `data`, and `metadata` are provided, the command returns the exact `id`.

```execute
aux4 repository write users --id 123 --data '{"name":"John"}' --metadata '{"role":"admin"}'
```

```expect
123
```

### Write a record using --data only, relying on default metadata

Write a record by specifying `--data` and a fixed `--id`, omitting `--metadata` to use its default. The command should return the provided `id`.

```execute
aux4 repository write items --id item1 --data '{"price":9.99,"inStock":true}'
```

```expect
item1
```

### Generate a UUID when no id is provided and input via STDIN

Pipe JSON data without an `id` into the command and omit the `--id` flag. The command should generate and return a UUID in the standard v4 format, optionally wrapped in double quotes.

```execute
echo '{"price":100}' | aux4 repository write payments
```

```expect:partial
*?-*?-*?-*?-*?
```

### Write with --ttl merges an expiry epoch into metadata

Providing `--ttl <seconds>` computes an absolute expiry epoch (current epoch + seconds) and merges it into the record metadata under the `ttl` key, preserving any user-supplied metadata. Writing and reading back in a single step (this suite resets the database between steps).

```execute
aux4 repository write sessions --id s1 --data '{"token":"abc"}' --metadata '{"role":"admin"}' --ttl 3600
aux4 repository read sessions --id s1 --raw true
```

```expect:partial
[
  {
    "id": "s1",
    "token": "abc",
    "__metadata": {
      "role": "admin",
      "ttl": *?,
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```

### Write without --ttl leaves metadata unchanged

When `--ttl` is omitted, no `ttl` key is added and metadata behaves exactly as before.

```execute
aux4 repository write sessions --id s2 --data '{"token":"xyz"}' --metadata '{"role":"user"}'
aux4 repository read sessions --id s2 --raw true
```

```expect:partial
[
  {
    "id": "s2",
    "token": "xyz",
    "__metadata": {
      "role": "user",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```
