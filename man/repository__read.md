Reads records from a named aux4 repository and returns them as a JSON array. The command ensures the repository exists (initializing it if necessary) and then returns records flattened so that each item contains an "id" field plus the original JSON data fields. Record-level metadata is merged into a single "__metadata" object which always includes system timestamps (createdAt and updatedAt).

When an --id is provided, the command returns an array with the single matching record (or an empty array if none found). When no --id is supplied it returns all records in the repository as an array. If the repository has no records the command prints an empty array and exits with a non-zero status (exit code 4 in tests), making it easy to detect an empty result programmatically.

Use this command when you need to retrieve stored JSON objects from an aux4 repository and work with them as native JSON (the command outputs valid JSON suitable for piping into jq or other JSON processors). It is commonly paired with the write and find commands to populate and query repository data.

By default, records whose metadata `ttl` (an absolute expiry epoch, set via `write --ttl`) is in the past are hidden from the result — both when reading all records and when reading a single record by `--id`. Records with no `ttl` are always returned. Pass `--includeExpired true` to skip this filter and return expired records as well (useful for debugging or manual sweeps).

### Example: Read all records from a repository

This reads every record stored in the users repository and returns them as an array. Each item contains the record's id, the stored data fields, and a merged "__metadata" object with createdAt/updatedAt timestamps.

```bash
aux4 repository read users
```

```json
[
  {
    "id": "user1",
    "name": "John",
    "age": 30,
    "__metadata": {
      "role": "admin",
      "createdAt": "2023-01-01T12:00:00Z",
      "updatedAt": "2023-01-01T12:00:00Z"
    }
  },
  {
    "id": "user2",
    "name": "Jane",
    "age": 25,
    "__metadata": {
      "role": "user",
      "createdAt": "2023-01-01T12:01:00Z",
      "updatedAt": "2023-01-01T12:01:00Z"
    }
  }
]
```

### Example: Read a specific record by id

Provide --id to return only the matching record. The command still returns an array (either with a single object or empty when the id is not present).

```bash
aux4 repository read products --id prod123
```

```json
[
  {
    "id": "prod123",
    "name": "Widget",
    "price": 19.99,
    "inStock": true,
    "__metadata": {
      "category": "electronics",
      "createdAt": "2023-01-01T12:10:00Z",
      "updatedAt": "2023-01-01T12:10:00Z"
    }
  }
]
```

### Example: Empty repository or missing id

If the repository contains no matching records the command prints an empty array. In test scenarios this also results in a non-zero exit code, which can be used to detect the empty-result case in scripts.

```bash
# No records present -> prints [] and exits non-zero
aux4 repository read empty_test_repo
```

### Example: Include expired records

By default an expired record is hidden. Use `--includeExpired true` to see it anyway.

```bash
# hidden by default (record has an expired ttl)
aux4 repository read sessions --id s1

# shown with --includeExpired
aux4 repository read sessions --id s1 --includeExpired true
```

The read command is most useful together with the repository write and find commands. For writing new items, see [aux4 repository write](./write); for filtering using expressions, see [aux4 repository find](./find). Expired records can be swept from every repository with [aux4 repository clean](./clean).