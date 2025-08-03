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

```expect:regex:ignoreCase
^"?[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}"?$
```
