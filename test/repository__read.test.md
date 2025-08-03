### Reading from an empty repository returns an empty array

# Ensure the database is reset before the test
beforeEach
rm -f test.db

# Execute the read command on an empty repository
```execute
aux4 repository read testrepo --db test.db
```

# Expect an empty JSON array when no records exist
```expect
[]
```

### Reading a specific record by id returns the correct JSON object with metadata keys

# Reset and seed the database with a single record
beforeEach
rm -f test.db
sqlite3 test.db "CREATE TABLE testrepo (id text primary key, data text, metadata text, created_at text, updated_at text); INSERT INTO testrepo VALUES('123','{\"foo\":\"bar\"}','{}','2021-01-01','2021-01-02');"

# Execute the read command for the seeded record
```execute
aux4 repository read testrepo --id 123 --db test.db
```

# Expect the JSON output including the data field and merged metadata timestamps
```expect
[
  {
    "id": "123",
    "foo": "bar",
    "$metadata": {
      "createdAt": "2021-01-01",
      "updatedAt": "2021-01-02"
    }
  }
]
```