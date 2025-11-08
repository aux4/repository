The find command queries records stored in a repository using an SQL-style filtering expression supplied with --expr. It returns a JSON array of matching records; each returned object contains the record id, the fields from the stored JSON object, and a $metadata object that merges any stored metadata with createdAt and updatedAt timestamps. When no records match the expression the command returns an empty array and exits with a non-zero status (useful for scripting).

Expressions use standard comparison operators (=, !=, <>, <, <=, >, >=) and the LIKE / NOT LIKE operators with % wildcards for pattern matching. Logical operators (and, or) and parentheses are supported so you can build compound filtering conditions. To query nested JSON fields, use json_extract(...) in the expression (for example: json_extract(data, '$.specs.wireless') = 1). Boolean values are matched using 1 (true) and 0 (false).

Use find when you need to filter records by content rather than by id. It complements repository write (to add records) and read (to fetch records by id or list all records). The command is intended for ad-hoc queries and scripting where you need flexible, SQL-like filtering over the JSON documents stored in the repository.

### Example: simple equality
Find records where a top-level field equals a value:

```bash
aux4 repository find users --expr "name = 'John'"
```

Expected result (one matching record):

```json
[
  {
    "id": "user1",
    "name": "John",
    "age": 30,
    "city": "NYC",
    "$metadata": {
      "role": "admin",
      "createdAt": "2021-01-01T...",
      "updatedAt": "2021-01-01T..."
    }
  }
]
```

### Example: numeric and compound expressions
Find records by numeric comparison and combined conditions:

```bash
aux4 repository find users --expr "age > 28 and city = 'NYC'"
```

This returns any records that match both clauses.

### Example: nested JSON fields
Query values nested inside the stored JSON using json_extract:

```bash
aux4 repository find devices --expr "json_extract(data, '$.specs.wireless') = 1"
```

### Example: pattern matching with LIKE
Use LIKE with % wildcards for partial matches:

```bash
aux4 repository find products --expr "name like '%Gaming%'"
```

Returns records whose name contains the substring "Gaming". 

When no records match the expression the command prints an empty JSON array and exits with a non-zero code; this makes it straightforward to use find in scripts and tests that need to detect absence of results.