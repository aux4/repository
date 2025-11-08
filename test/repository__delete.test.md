## Test

```beforeAll
rm -f .local.db
```

```afterAll
rm -f .local.db
```

### Delete single record by id

Write a record and then delete it by its id.

```execute
aux4 repository write users --id user1 --data '{"name":"John","age":30}' --metadata '{"role":"admin"}'
aux4 repository write users --id user2 --data '{"name":"Jane","age":25}' --metadata '{"role":"user"}'
```

```execute
aux4 repository delete users --id user1
```

```execute
aux4 repository read users
```

```expect:partial
[
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

### Delete multiple records with multiple id flags

Write multiple records and delete several of them using multiple --id flags.

```execute
aux4 repository write products --id prod1 --data '{"name":"Widget","price":19.99}' --metadata '{"category":"electronics"}'
aux4 repository write products --id prod2 --data '{"name":"Gadget","price":29.99}' --metadata '{"category":"electronics"}'
aux4 repository write products --id prod3 --data '{"name":"Tool","price":39.99}' --metadata '{"category":"hardware"}'
aux4 repository write products --id prod4 --data '{"name":"Device","price":49.99}' --metadata '{"category":"electronics"}'
```

```execute
aux4 repository delete products --id prod1 --id prod3 --id prod4
```

```execute
aux4 repository read products
```

```expect:partial
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

### Delete non-existent record

Try to delete a record that doesn't exist - should not cause an error.

```execute
aux4 repository write items --id item1 --data '{"name":"Test Item"}' --metadata '{}'
```

```execute
aux4 repository delete items --id nonexistent
```

```execute
aux4 repository read items
```

```expect:partial
[
  {
    "id": "item1",
    "name": "Test Item",
    "$metadata": {
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```

### Delete all records using multiple ids

Write several records and delete all of them using multiple --id flags.

```execute
aux4 repository write orders --id order1 --data '{"total":100}' --metadata '{"status":"pending"}'
aux4 repository write orders --id order2 --data '{"total":200}' --metadata '{"status":"completed"}'
aux4 repository write orders --id order3 --data '{"total":300}' --metadata '{"status":"pending"}'
```

```execute
aux4 repository delete orders --id order1 --id order2 --id order3
```

```execute
aux4 repository read orders
```

```error:regex
^(\[\]|)$
```

### Delete from empty repository

Try to delete from a repository that doesn't exist yet.

```execute
aux4 repository delete empty_repo --id someid
```

```execute
aux4 repository read empty_repo
```

```error:regex
^(\[\]|)$
```

### Delete mix of existing and non-existent records

Delete a combination of records that exist and don't exist using multiple --id flags.

```execute
aux4 repository write settings --id config1 --data '{"theme":"dark"}' --metadata '{"env":"prod"}'
aux4 repository write settings --id config2 --data '{"theme":"light"}' --metadata '{"env":"dev"}'
```

```execute
aux4 repository delete settings --id config1 --id nonexistent --id config2
```

```execute
aux4 repository read settings
```

```error:regex
^(\[\]|)$
```

### Delete single record from repository with multiple records

Ensure only the specified record is deleted when multiple exist.

```execute
aux4 repository write logs --id log1 --data '{"level":"info","message":"Started"}' --metadata '{}'
aux4 repository write logs --id log2 --data '{"level":"error","message":"Failed"}' --metadata '{}'
aux4 repository write logs --id log3 --data '{"level":"warn","message":"Warning"}' --metadata '{}'
```

```execute
aux4 repository delete logs --id log2
```

```execute
aux4 repository read logs
```

```expect:partial
[
  {
    "id": "log1",
    "level": "info",
    "message": "Started",
    "$metadata": {
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  },
  {
    "id": "log3",
    "level": "warn",
    "message": "Warning",
    "$metadata": {
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```