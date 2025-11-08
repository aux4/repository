## Test

```beforeAll
rm -f .local.db
```

```afterAll
rm -f .local.db
```

### Truncate empty repository

Test truncating a repository that doesn't exist yet or has no records.

```execute
aux4 repository truncate empty_repo --yes
```

```execute
aux4 repository read empty_repo
```

```error:regex
^(\[\]|)$
```

### Truncate repository with single record

Write a record and then truncate the repository.

```execute
aux4 repository write users --id user1 --data '{"name":"John","age":30}' --metadata '{"role":"admin"}'
```

```execute
aux4 repository truncate users --yes
```

```execute
aux4 repository read users
```

```error:regex
^(\[\]|)$
```

### Truncate repository with multiple records

Write multiple records and verify they are all deleted after truncate.

```execute
aux4 repository write products --id prod1 --data '{"name":"Widget","price":19.99}' --metadata '{"category":"electronics"}'
aux4 repository write products --id prod2 --data '{"name":"Gadget","price":29.99}' --metadata '{"category":"electronics"}'
aux4 repository write products --id prod3 --data '{"name":"Tool","price":39.99}' --metadata '{"category":"hardware"}'
```

```execute
aux4 repository read products
```

```expect:partial
[
  {
    "id": "prod1",
    "name": "Widget",
    "price": 19.99,
    "$metadata": {
      "category": "electronics",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  },
  {
    "id": "prod2",
    "name": "Gadget",
    "price": 29.99,
    "$metadata": {
      "category": "electronics",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  },
  {
    "id": "prod3",
    "name": "Tool",
    "price": 39.99,
    "$metadata": {
      "category": "hardware",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```

```execute
aux4 repository truncate products --yes
```

```execute
aux4 repository read products
```

```error:regex
^(\[\]|)$
```

### Truncate repository preserves table structure

Verify that after truncate, the repository can still accept new records.

```execute
aux4 repository write orders --id order1 --data '{"total":100}' --metadata '{"status":"pending"}'
aux4 repository truncate orders --yes
aux4 repository write orders --id order2 --data '{"total":200}' --metadata '{"status":"completed"}'
```

```execute
aux4 repository read orders
```

```expect:partial
[
  {
    "id": "order2",
    "total": 200,
    "$metadata": {
      "status": "completed",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```

### Truncate one repository doesn't affect others

Create records in multiple repositories and truncate only one.

```execute
aux4 repository write users --id user1 --data '{"name":"John"}' --metadata '{}'
aux4 repository write products --id prod1 --data '{"name":"Widget"}' --metadata '{}'
```

```execute
aux4 repository truncate users --yes
```

```execute
aux4 repository read users
```

```error:regex
^(\[\]|)$
```

```execute
aux4 repository read products
```

```expect:partial
[
  {
    "id": "prod1",
    "name": "Widget",
    "$metadata": {
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```

### Truncate repository with complex nested JSON data

Verify truncate works with complex data structures.

```execute
aux4 repository write complex --id item1 --data '{"customer":{"name":"Alice","address":{"street":"123 Main St","city":"NYC"}},"items":[{"id":"a","qty":2},{"id":"b","qty":1}]}' --metadata '{"version":"1.0","tags":["important","processed"]}'
```

```execute
aux4 repository truncate complex --yes
```

```execute
aux4 repository read complex
```

```error:regex
^(\[\]|)$
```