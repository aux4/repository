Deletes one or more records from a named repository in the configured database. Provide the repository name as the positional argument and one or more --id flags to specify which records to remove. The command performs an idempotent operation: attempting to remove an id that does not exist (or removing from an empty/non-existent repository) does not cause an error.

When multiple --id flags are provided the command deletes all matching records in a single operation. The command does not return the deleted records — use the read command to verify the repository contents after the operation.

This command is commonly used together with the write and read commands in workflows that insert, remove, and then validate records. See the read command for inspecting repository contents: [aux4 repository read](./read).

### Example: Delete a single record
Write a record, delete it by id, then confirm only the remaining records are present.

```bash
aux4 repository write users --id user1 --data '{"name":"John","age":30}' --metadata '{"role":"admin"}'
aux4 repository write users --id user2 --data '{"name":"Jane","age":25}' --metadata '{"role":"user"}'

aux4 repository delete users --id user1

aux4 repository read users
```

Expected result (user2 remains):

```text
[
  {
    "id": "user2",
    "name": "Jane",
    "age": 25,
    "$metadata": {
      "role": "user",
      "createdAt": "...",
      "updatedAt": "..."
    }
  }
]
```

### Example: Delete multiple records
Provide multiple --id flags to remove several records in one call.

```bash
aux4 repository delete products --id prod1 --id prod3 --id prod4
aux4 repository read products
```

Result: only records not listed in the --id flags remain.

### Example: Deleting non-existent ids
Deleting an id that doesn't exist is a no-op and does not return an error. This makes batch delete operations safe when some targets may already be absent.

```bash
aux4 repository delete items --id nonexistent
aux4 repository read items
```

Result: repository contents are unchanged if the id was not present.
