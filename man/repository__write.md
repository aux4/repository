The write command inserts a JSON record into the named repository table, ensuring the repository is initialized first. Before inserting it will call the repository init flow which creates the table if needed. The command accepts JSON either via the --data parameter or from STDIN; it removes any id field from the stored JSON payload so the id is stored as a top-level key and the remaining JSON is saved in the data column. Metadata is stored alongside the record and defaults to an empty object when not provided.

ID selection follows a clear priority: an explicit --id flag takes precedence; if no --id is provided the command will use an id field present in the input JSON; if neither is present the command generates a UUID. After a successful insert the command prints the id to stdout.

The implementation uses temporary files and escapes quotes to avoid SQL injection issues when embedding JSON into the insert statement; these implementation details are handled for you and do not require additional steps. The command returns only the record id on success and cleans up temporary files it created.

The optional `--ttl <seconds>` flag sets a time-to-live for the record. It is a duration in seconds that is converted to an absolute expiry epoch using the database clock (`strftime('%s','now') + <ttl>`), and the resulting value is merged into the record metadata under the `ttl` key. Any user-supplied `--metadata` is preserved: `--metadata '{"role":"admin"}' --ttl 3600` produces metadata like `{"role":"admin","ttl":1783724578}`. When `--ttl` is omitted, no `ttl` key is added and metadata behaves exactly as before. Records whose `ttl` is in the past are hidden from `read` and `find` by default and are removed by `clean`.

**Note:** Because a negative value that follows a space (for example `--ttl -60`) is interpreted as a flag, use the `--ttl=-60` form (or write a past epoch directly via `--metadata`) if you need to store an already-expired record for testing.

### Example: Write with explicit id, data and metadata
Write a record by passing the JSON on the command line along with an explicit id and optional metadata.

```bash
aux4 repository write users --id 123 --data '{"name":"John"}' --metadata '{"role":"admin"}'
```

This prints the id (123) on success and stores the JSON (without the id field) in the repository.

### Example: Write from STDIN and auto-generate an id
Pipe JSON into the command and let the command generate a UUID when no --id is provided.

```bash
echo '{"price":100}' | aux4 repository write payments
```

The command prints the generated UUID. You can then read the record back with:

```bash
aux4 repository read payments --id <generated-uuid>
```

### Example: Use id from JSON or override it with --id
If the input JSON contains an id field and no --id flag is provided, that id will be used. Supplying --id always overrides an id present in the JSON.

```bash
# uses id from JSON
echo '{"id":"from_json","name":"Item"}' | aux4 repository write items

# overrides id in JSON
echo '{"id":"ignored","name":"Item"}' | aux4 repository write items --id explicit_id
```

Both cases print the final id used and store the remaining JSON (without id) in the repository.

### Example: Write a record with a time-to-live

Store a session that should expire one hour from now. The expiry epoch is merged into metadata under `ttl`.

```bash
aux4 repository write sessions --id s1 --data '{"token":"abc"}' --metadata '{"role":"admin"}' --ttl 3600
```

Reading it back with `--raw true` shows the merged metadata:

```json
[
  {
    "id": "s1",
    "token": "abc",
    "__metadata": {
      "role": "admin",
      "ttl": 1783724578,
      "createdAt": "2026-07-10 22:02:58",
      "updatedAt": "2026-07-10 22:02:58"
    }
  }
]
```

Related commands: use aux4 repository read to fetch inserted records, aux4 repository find to filter records by expressions, and aux4 repository clean to remove expired records.