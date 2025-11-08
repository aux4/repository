## Test

```beforeAll
rm -f .local.db
```

```afterAll
rm -f .local.db
```

### Find from empty repository returns empty array

Find in a repository that doesn't exist yet should return an empty array and exit with code 4.

```execute
aux4 repository find empty_test_repo --expr "name = 'John'"
```

```error:regex
^(\[\]|)$
```

### Find records by simple equality

First write multiple records, then find specific ones using equality expressions.

```execute
aux4 repository write users --id user1 --data '{"name":"John","age":30,"city":"NYC"}' --metadata '{"role":"admin"}'
aux4 repository write users --id user2 --data '{"name":"Jane","age":25,"city":"LA"}' --metadata '{"role":"user"}'
aux4 repository write users --id user3 --data '{"name":"Bob","age":35,"city":"NYC"}' --metadata '{"role":"user"}'
```

```execute
aux4 repository find users --expr "name = 'John'"
```

```expect:partial
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

### Find records by numeric comparison

Find records using numeric comparison operators.

```execute
aux4 repository find users --expr "age > 28"
```

```expect:partial
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

### Find records using LIKE operator

Find records using the LIKE operator for partial string matching.

```execute
aux4 repository find users --expr "name like '%o%'"
```

```expect:partial
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

### Find records using compound expressions

Find records using AND/OR logical operators.

```execute
aux4 repository find users --expr "city = 'NYC' and age < 35"
```

```expect:partial
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

### Find records with inequality operators

Test various inequality operators.

```execute
aux4 repository find users --expr "age != 30"
```

```expect:partial
[
  {
    "id": "user2",
    "name": "Jane",
    "age": 25,
    "city": "LA",
    "$metadata": {
      "role": "user",
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

### Find records with complex nested JSON data

Write and find records with nested JSON structures.

```execute
aux4 repository write orders --id order1 --data '{"customer":{"name":"Alice","email":"alice@example.com"},"total":29.98,"status":"pending"}' --metadata '{"priority":"high"}'
aux4 repository write orders --id order2 --data '{"customer":{"name":"Bob","email":"bob@example.com"},"total":45.50,"status":"completed"}' --metadata '{"priority":"low"}'
```

```execute
aux4 repository find orders --expr "total > 30"
```

```expect:partial
[
  {
    "id": "order2",
    "customer": {
      "name": "Bob",
      "email": "bob@example.com"
    },
    "total": 45.50,
    "status": "completed",
    "$metadata": {
      "priority": "low",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```

### Find with no matching results returns empty array

Try to find records that don't match any criteria.

```execute
aux4 repository find users --expr "age > 100"
```

```error:regex
^(\[\]|)$
```

### Find records with boolean values

Write and find records with boolean fields.

```execute
aux4 repository write products --id prod1 --data '{"name":"Widget","price":19.99,"inStock":true,"featured":false}' --metadata '{"category":"electronics"}'
aux4 repository write products --id prod2 --data '{"name":"Gadget","price":29.99,"inStock":false,"featured":true}' --metadata '{"category":"electronics"}'
```

```execute
aux4 repository find products --expr "inStock = 1"
```

```expect:partial
[
  {
    "id": "prod1",
    "name": "Widget",
    "price": 19.99,
    "inStock": true,
    "featured": false,
    "$metadata": {
      "category": "electronics",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```