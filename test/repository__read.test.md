## Test

```beforeAll
rm -f .local.db
```

```afterAll
rm -f .local.db
```

### Read from empty repository returns empty array

Read from a repository that doesn't exist yet should return an empty array and exit with code 4.

```execute
aux4 repository read empty_test_repo
```

```error:regex
^(\[\]|)$
```

### Read all records from repository with multiple entries

First write multiple records, then read all of them.

```execute
aux4 repository write users --id user1 --data '{"name":"John","age":30}' --metadata '{"role":"admin"}'
aux4 repository write users --id user2 --data '{"name":"Jane","age":25}' --metadata '{"role":"user"}'
```

```execute
aux4 repository read users
```

```expect:partial
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

### Read specific record by id

Write a record and then read it by its specific id.

```execute
aux4 repository write products --id prod123 --data '{"name":"Widget","price":19.99,"inStock":true}' --metadata '{"category":"electronics"}'
```

```execute
aux4 repository read products --id prod123
```

```expect:partial
[
  {
    "id": "prod123",
    "name": "Widget",
    "price": 19.99,
    "inStock": true,
    "$metadata": {
      "category": "electronics",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```

### Read non-existent record by id returns empty array

Try to read a record that doesn't exist.

```execute
aux4 repository read nonexistent_repo --id nonexistent
```

```error:regex
^(\[\]|)$
```

### Read preserves complex nested JSON data

Write and read back a record with nested JSON structures.

```execute
aux4 repository write orders --id order1 --data '{"customer":{"name":"Alice","email":"alice@example.com"},"items":[{"id":"item1","qty":2},{"id":"item2","qty":1}],"total":29.98}' --metadata '{"status":"pending","priority":"high"}'
```

```execute
aux4 repository read orders --id order1
```

```expect:partial
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

### Read merges metadata with timestamps

Verify that custom metadata is properly merged with system-generated timestamps.

```execute
aux4 repository write settings --id config1 --data '{"theme":"dark","notifications":true}' --metadata '{"environment":"production","version":"1.0"}'
```

```execute
aux4 repository read settings --id config1
```

```expect:partial
[
  {
    "id": "config1",
    "theme": "dark",
    "notifications": true,
    "$metadata": {
      "environment": "production",
      "version": "1.0",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```

### Read with empty metadata defaults to empty object

Write a record without metadata and verify it defaults to empty object with timestamps.

```execute
aux4 repository write logs --id log1 --data '{"level":"info","message":"System started"}'
```

```execute
aux4 repository read logs --id log1
```

```expect:partial
[
  {
    "id": "log1",
    "level": "info",
    "message": "System started",
    "$metadata": {
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```
