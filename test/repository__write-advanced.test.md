## Test

```beforeAll
rm -f .local.db
```

```afterAll
rm -f .local.db
```

### Write using STDIN without --data flag

Test piping JSON data directly to the write command without using --data.

```execute
echo '{"name":"Alice","email":"alice@example.com","age":28}' | aux4 repository write users --id user_stdin
```

```expect
user_stdin
```

```execute
aux4 repository read users --id user_stdin
```

```expect:partial
[
  {
    "id": "user_stdin",
    "name": "Alice",
    "email": "alice@example.com",
    "age": 28,
    "$metadata": {
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```

### Write using STDIN with auto-generated UUID

Test piping JSON data without specifying an ID, should generate a UUID.

```execute
echo '{"product":"Widget","price":29.99,"inStock":true}' | aux4 repository write inventory
```

```expect:regex
^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$
```

### Write using STDIN with complex nested JSON

Test piping complex nested JSON structures.

```execute
echo '{"order":{"id":"ORD-123","customer":{"name":"Bob","address":{"street":"456 Oak St","city":"LA","zip":"90210"}},"items":[{"sku":"WIDGET-1","qty":2,"price":15.99},{"sku":"GADGET-1","qty":1,"price":25.99}],"totals":{"subtotal":57.97,"tax":5.22,"total":63.19}}}' | aux4 repository write orders --id complex_order
```

```expect
complex_order
```

```execute
aux4 repository read orders --id complex_order
```

```expect:partial
[
  {
    "id": "complex_order",
    "order": {
      "id": "ORD-123",
      "customer": {
        "name": "Bob",
        "address": {
          "street": "456 Oak St",
          "city": "LA",
          "zip": "90210"
        }
      },
      "items": [
        {
          "sku": "WIDGET-1",
          "qty": 2,
          "price": 15.99
        },
        {
          "sku": "GADGET-1",
          "qty": 1,
          "price": 25.99
        }
      ],
      "totals": {
        "subtotal": 57.97,
        "tax": 5.22,
        "total": 63.19
      }
    },
    "$metadata": {
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```

### Write using STDIN with metadata

Test piping JSON data while also providing metadata.

```execute
echo '{"title":"Important Document","content":"This is a test document","version":"1.2"}' | aux4 repository write documents --id doc1 --metadata '{"author":"John Doe","department":"Engineering","classification":"internal"}'
```

```expect
doc1
```

```execute
aux4 repository read documents --id doc1
```

```expect:partial
[
  {
    "id": "doc1",
    "title": "Important Document",
    "content": "This is a test document",
    "version": "1.2",
    "$metadata": {
      "author": "John Doe",
      "department": "Engineering",
      "classification": "internal",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```

### Write using STDIN with JSON containing id field

Test that when JSON contains an "id" field and no --id is specified, it uses the id from JSON.

```execute
echo '{"id":"from_json","name":"Test Item","description":"Item with id in JSON"}' | aux4 repository write items
```

```expect
from_json
```

```execute
aux4 repository read items --id from_json
```

```expect:partial
[
  {
    "id": "from_json",
    "name": "Test Item",
    "description": "Item with id in JSON",
    "$metadata": {
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```

### Write using STDIN with JSON containing id field but --id flag overrides

Test that --id flag takes precedence over id field in JSON.

```execute
echo '{"id":"ignored_id","name":"Override Test","value":42}' | aux4 repository write tests --id override_id
```

```expect
override_id
```

```execute
aux4 repository read tests --id override_id
```

```expect:partial
[
  {
    "id": "override_id",
    "name": "Override Test",
    "value": 42,
    "$metadata": {
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```

### Write using STDIN with arrays and objects

Test piping JSON with arrays and nested objects.

```execute
echo '{"text":"Sample text","numbers":[1,2,3],"config":{"enabled":true}}' | aux4 repository write special_data --id special_chars_unique
```

```expect
special_chars_unique
```

```execute
aux4 repository read special_data --id special_chars_unique
```

```expect:partial
[
  {
    "id": "special_chars_unique",
    "text": "Sample text",
    "numbers": [
      1,
      2,
      3
    ],
    "config": {
      "enabled": true
    },
    "$metadata": {
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```

### Write large JSON payload via STDIN

Test handling of larger JSON payloads.

```execute
echo '{"users":[{"id":1,"name":"User1","email":"user1@test.com","roles":["admin","user"]},{"id":2,"name":"User2","email":"user2@test.com","roles":["user"]},{"id":3,"name":"User3","email":"user3@test.com","roles":["guest"]}],"metadata":{"total":3,"created":"2024-01-01","source":"import"},"config":{"theme":"dark","language":"en","notifications":{"email":true,"sms":false,"push":true}}}' | aux4 repository write bulk --id bulk_data
```

```expect
bulk_data
```

```execute
aux4 repository read bulk --id bulk_data
```

```expect:partial
[
  {
    "id": "bulk_data",
    "users": [
      {
        "id": 1,
        "name": "User1",
        "email": "user1@test.com",
        "roles": [
          "admin",
          "user"
        ]
      },
      {
        "id": 2,
        "name": "User2",
        "email": "user2@test.com",
        "roles": [
          "user"
        ]
      },
      {
        "id": 3,
        "name": "User3",
        "email": "user3@test.com",
        "roles": [
          "guest"
        ]
      }
    ],
    "metadata": {
      "total": 3,
      "created": "2024-01-01",
      "source": "import"
    },
    "config": {
      "theme": "dark",
      "language": "en",
      "notifications": {
        "email": true,
        "sms": false,
        "push": true
      }
    },
    "$metadata": {
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```