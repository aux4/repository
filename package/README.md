# aux4/repository

Simplified aux4 database

This package provides a lightweight JSON-based repository interface on top of SQLite. It allows you to initialize repositories (tables), write records, read records, query with expressions, delete individual records, and truncate entire tables. It uses a local SQLite database file (default `.local.db`) by default but can be customized.

## Installation

Install the package from the aux4 hub:

```bash
aux4 aux4 pkger install aux4/repository
```

## Usage

Commands are invoked via `aux4 repository <command>`. By default, the SQLite database file is `.local.db`. You can override the database file path with `--db <path>`.

### Commands

- [aux4 repository write](./commands/repository/write)  
  Write a JSON record to a repository (table). Generates a UUID if no `--id` is provided.

- [aux4 repository read](./commands/repository/read)  
  Read one or all records from a repository. Use `--id` to fetch a specific record.

- [aux4 repository find](./commands/repository/find)  
  Query a repository using an SQL-like expression on JSON fields (e.g. `age > 30 and name like "%John%"`).

- [aux4 repository delete](./commands/repository/delete)  
  Delete a record by `--id` from a repository.

- [aux4 repository truncate](./commands/repository/truncate)  
  Delete all records from a repository. Requires confirmation.

## Examples

Assuming you want to manage a repository named `users`:

1. Write a new record:

   ```bash
   aux4 repository write users --data '{"name":"Alice","age":30}'
   # Output: new record ID (UUID)
   ```

2. Read all records:

   ```bash
   aux4 repository read users
   # Output: JSON array of records, including metadata
   ```

3. Read a specific record by ID:

   ```bash
   aux4 repository read users --id 123e4567-e89b-12d3-a456-426614174000
   ```

4. Find records matching a condition:

   ```bash
   aux4 repository find users --expr "age >= 21 and name like '%A%'"
   ```

5. Delete a record:

   ```bash
   aux4 repository delete users --id 123e4567-e89b-12d3-a456-426614174000
   ```

6. Truncate the repository (delete all records):

   ```bash
   aux4 repository truncate users
   # Confirmation prompt: Are you sure you want to truncate the repository 'users'? This will delete all records.
   ```

### Custom Database File

By default, commands use `.local.db`. To specify another database file:

```bash
aux4 repository write users --db data.db --data '{"foo":"bar"}'
aux4 repository read users --db data.db
```

## How It Works

- **Initialization**: Each command automatically initializes the repository by creating a table if it doesn't exist. The table name matches the repository name, and columns include `id`, `data` (JSON), `metadata` (JSON), `created_at`, and `updated_at`.
- **Writing**: Records are stored as JSON in the `data` column. Optional metadata can also be provided via `--metadata` (default `{}`).
- **Reading & Finding**: Uses `sqlite3 -json` and `jq` to prettify and merge the JSON data with metadata timestamps.
- **Delete & Truncate**: Standard SQL operations under the hood.

## License

This project is licensed under the Apache License 2.0. See [LICENSE](./LICENSE) for details.

[![License](https://img.shields.io/badge/license-apache-blue.svg)](./LICENSE)
