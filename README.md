# aux4/repository

Simplified aux4 database

aux4/repository provides a tiny, file-backed JSON repository built on SQLite. It exposes a small set of commands (write, read, find, delete, truncate) that let you store arbitrary JSON records keyed by id, query them using SQL-like expressions, and manage repository tables inside a local SQLite database. It's designed to be simple to use from shell scripts and other aux4 packages, and to work with stdin-based workflows (pipe JSON into write) or explicit --data flags.

This package is a building block in the aux4 ecosystem for storing and querying structured JSON data without running a separate database server. Typical use cases include quick test fixtures, local caches, small data imports/exports, and lightweight persistence for CLI tools. It keeps behavior predictable by using sqlite3 and jq as the underlying tools and follows aux4 conventions for variables and profiles.

## Installation

```bash
aux4 aux4 pkger install aux4/repository
```

## System Dependencies

This package requires system dependencies. You need to have one of the following system installers:

- [brew](/r/public/packages/aux4/system-installer-brew)
- [apt](/r/public/packages/aux4/system-installer-apt)
- [linux](/r/public/packages/aux4/system-installer-linux)

For more details, see [system-installer](/r/public/packages/aux4/pkger/commands/aux4/pkger/system).

## Quick Start

Write a record and read it back. This uses the default database (.local.db) and the repository name "users".

```bash
aux4 repository write users --id user1 --data '{"name":"John","age":30}' --metadata '{"role":"admin"}'
```

This prints the id of the stored record (here: user1). Then read the repository:

```bash
aux4 repository read users
```

This returns JSON array(s) of stored records with merged metadata and timestamps.

## Storing data (write)

The write command stores a JSON document in the named repository table. You can provide the data via --data or pipe JSON to stdin. If --id is omitted, the command will use an id field inside the JSON (if present) or generate a UUID.

The key variables:
- db (default: .local.db) — path to the SQLite database
- repository — repository/table name (positional)
- id — optional id to store (overrides id in JSON)
- data — JSON string to store (optional if providing via stdin)
- metadata — optional JSON metadata (defaults to {})

Example using explicit id, data, and metadata:

```bash
aux4 repository write users --id 123 --data '{"name":"John"}' --metadata '{"role":"admin"}'
```

For full command docs see [aux4 repository write](./commands/repository/write).

## Reading data (read)

The read command returns stored records as JSON. By default it reads all rows; use --id to select a single record. Each returned object contains an id and the fields from the stored JSON, plus a $metadata object that merges stored metadata with createdAt and updatedAt timestamps.

Example — write two records then read all:

```bash
aux4 repository write users --id user1 --data '{"name":"John","age":30}' --metadata '{"role":"admin"}'
aux4 repository write users --id user2 --data '{"name":"Jane","age":25}' --metadata '{"role":"user"}'
aux4 repository read users
```

Example expected output (excerpt):

```json
[
  {
    "id": "user1",
    "name": "John",
    "age": 30,
    "$metadata": {
      "role": "admin",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  },
  {
    "id": "user2",
    "name": "Jane",
    "age": 25,
    "$metadata": {
      "role": "user",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```

For full command docs see [aux4 repository read](./commands/repository/read).

## Searching with SQL expressions (find)

find lets you filter records using SQL expressions that are translated to JSON field access. The expr parameter accepts expressions like "age > 30 and name like '%John%'" and supports numeric comparisons, LIKE, logical operators, JSON path extraction, and parentheses for grouping.

Example — find by name equality:

```bash
aux4 repository find users --expr "name = 'John'"
```

Example expected output (excerpt):

```json
[
  {
    "id": "user1",
    "name": "John",
    "age": 30,
    "city": "NYC",
    "$metadata": {
      "role": "admin",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```

For full command docs see [aux4 repository find](./commands/repository/find).

## Deleting and truncating

- delete removes one or more records by id. Supply multiple --id flags to remove multiple records.
- truncate deletes all records in a repository (it asks for confirmation unless you pass --yes).

Example — delete a single record:

```bash
aux4 repository delete users --id user1
```

Example — truncate a repository:

```bash
aux4 repository truncate users --yes
```

For full command docs see [aux4 repository delete](./commands/repository/delete) and [aux4 repository truncate](./commands/repository/truncate).

## Examples

### Write a record with explicit id, data, and metadata

This example writes a user record and prints the id back.

```bash
aux4 repository write users --id 123 --data '{"name":"John"}' --metadata '{"role":"admin"}'
```

The command prints:

```
123
```

### Write using STDIN (auto-generated UUID)

Pipe JSON to the write command without providing --id; the command will generate a UUID when none is provided.

```bash
echo '{"product":"Widget","price":29.99,"inStock":true}' | aux4 repository write inventory
```

This prints a generated UUID (v4). Tests expect a UUID matching the pattern:

```
^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$
```

### Read complex nested JSON

Write a record with nested JSON and read it back. The nested structure is preserved.

```bash
aux4 repository write orders --id order1 --data '{"customer":{"name":"Alice","email":"alice@example.com"},"items":[{"id":"item1","qty":2},{"id":"item2","qty":1}],"total":29.98}' --metadata '{"status":"pending","priority":"high"}'
aux4 repository read orders --id order1
```

Expected output (excerpt):

```json
[
  {
    "id": "order1",
    "customer": {
      "name": "Alice",
      "email": "alice@example.com"
    },
    "items": [
      {
        "id": "item1",
        "qty": 2
      },
      {
        "id": "item2",
        "qty": 1
      }
    ],
    "total": 29.98,
    "$metadata": {
      "status": "pending",
      "priority": "high",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```

### Find with numeric comparison

Write several records, then find those matching a numeric condition.

```bash
aux4 repository write users --id user1 --data '{"name":"John","age":30,"city":"NYC"}' --metadata '{"role":"admin"}'
aux4 repository write users --id user2 --data '{"name":"Jane","age":25,"city":"LA"}' --metadata '{"role":"user"}'
aux4 repository write users --id user3 --data '{"name":"Bob","age":35,"city":"NYC"}' --metadata '{"role":"user"}'
aux4 repository find users --expr "age > 28"
```

Expected output (excerpt):

```json
[
  {
    "id": "user1",
    "name": "John",
    "age": 30,
    "city": "NYC",
    "$metadata": {
      "role": "admin",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  },
  {
    "id": "user3",
    "name": "Bob",
    "age": 35,
    "city": "NYC",
    "$metadata": {
      "role": "user",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```

### Delete multiple records and verify

Write multiple records, delete several using multiple --id flags, and read to confirm remaining records.

```bash
aux4 repository write products --id prod1 --data '{"name":"Widget","price":19.99}' --metadata '{"category":"electronics"}'
aux4 repository write products --id prod2 --data '{"name":"Gadget","price":29.99}' --metadata '{"category":"electronics"}'
aux4 repository write products --id prod3 --data '{"name":"Tool","price":39.99}' --metadata '{"category":"hardware"}'
aux4 repository write products --id prod4 --data '{"name":"Device","price":49.99}' --metadata '{"category":"electronics"}'
aux4 repository delete products --id prod1 --id prod3 --id prod4
aux4 repository read products
```

Expected output (excerpt):

```json
[
  {
    "id": "prod2",
    "name": "Gadget",
    "price": 29.99,
    "$metadata": {
      "category": "electronics",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```

## Configuration

This package does not require external configuration files. All runtime options are provided via command arguments and the default database file (.local.db) can be overridden with --db.

## License

This package is licensed under the Apache License 2.0.

See [LICENSE](./license) for details.
