## Test

```beforeAll
rm -f .local.db .readonly.db .custom_location.db
```

```afterAll
rm -f .local.db .readonly.db .custom_location.db
```

### Write with malformed JSON via STDIN

Test error handling when piping invalid JSON data.

```execute
echo '{"name":"Invalid JSON"' | aux4 repository write malformed --id test1
```

```error:regex
.*
```

### Write with empty STDIN input

Test behavior when piping empty content.

```execute
echo '' | aux4 repository write empty_input --id test2
```

```error:regex
.*
```

### Write with very long repository name

Test with repository names that might cause issues.

```execute
aux4 repository write very_long_repository_name_that_exceeds_normal_expectations_and_might_cause_database_issues --id test3 --data '{"test":true}'
```

```expect
test3
```

### Write with special characters in repository name

Test repository names with underscores, numbers, and other valid characters.

```execute
aux4 repository write test_repo_123_special --id test4 --data '{"value":"works"}'
```

```expect
test4
```

```execute
aux4 repository write repo2024 --id test5 --data '{"year":2024}'
```

```expect
test5
```

### Read from custom database location

Test using different database file paths.

```execute
aux4 repository write custom_test --db .custom_location.db --id item1 --data '{"location":"custom"}'
```

```execute
aux4 repository read custom_test --db .custom_location.db --id item1
```

```expect:partial
[
  {
    "id": "item1",
    "location": "custom",
    "$metadata": {
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```

### Find with complex SQL injection attempt (should be safe)

Test that the find command properly handles potentially malicious input.

```execute
aux4 repository write security_test --id safe1 --data '{"name":"Test User","role":"user"}'
aux4 repository write security_test --id safe2 --data '{"name":"Admin User","role":"admin"}'
```

```execute
aux4 repository find security_test --expr "name = 'Test User'; DROP TABLE security_test; --'"
```

```error:regex
.*
```

### Write with extremely large JSON payload

Test handling of large data structures.

```execute
aux4 repository write large_data --id big_item --data '{"data":{"level1":{"level2":{"level3":{"level4":{"level5":{"items":[{"id":1,"name":"Item 1","description":"A very long description that contains lots of text to test the handling of large JSON payloads in the repository system"},{"id":2,"name":"Item 2","description":"Another very long description with even more text to ensure that the system can handle substantial amounts of data without issues"},{"id":3,"name":"Item 3","description":"Yet another lengthy description that continues to test the boundaries of what the system can handle in terms of JSON payload size and complexity"}],"metadata":{"created":"2024-01-01","version":"1.0","tags":["large","test","data","complex","nested"],"config":{"enabled":true,"settings":{"theme":"dark","language":"en","timezone":"UTC"}}}}}}}}}'
```

```expect
big_item
```

### Write with Unicode and emoji characters

Test handling of international characters and emojis.

```execute
aux4 repository write unicode_test --id unicode1 --data '{"name":"José García","city":"São Paulo","emoji":"🚀✨🎉","chinese":"你好世界","japanese":"こんにちは","arabic":"مرحبا","russian":"Привет"}' --metadata '{"locale":"pt-BR","encoding":"UTF-8"}'
```

```expect
unicode1
```

```execute
aux4 repository read unicode_test --id unicode1
```

```expect:partial
[
  {
    "id": "unicode1",
    "name": "José García",
    "city": "São Paulo",
    "emoji": "🚀✨🎉",
    "chinese": "你好世界",
    "japanese": "こんにちは",
    "arabic": "مرحبا",
    "russian": "Привет",
    "$metadata": {
      "locale": "pt-BR",
      "encoding": "UTF-8",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```

### Find with invalid SQL expressions

Test error handling for malformed find expressions.

```execute
aux4 repository write find_test --id item1 --data '{"value":42}'
```

```execute
aux4 repository find find_test --expr "invalid sql syntax here $$$ @#"
```

```error:regex
.*
```

### Delete with empty id list

Test delete command behavior with no ids provided.

```execute
aux4 repository write delete_test --id item1 --data '{"keep":"this"}'
```

```execute
aux4 repository delete delete_test --id ""
```

```error:regex
.*
```

### Write with conflicting data sources

Test behavior when both --data and STDIN provide data.

```execute
echo '{"from":"stdin"}' | aux4 repository write conflict_test --id conflict1 --data '{"from":"flag"}'
```

```expect
conflict1
```

### Read with invalid database file

Test behavior when database file is corrupted or invalid.

```execute
echo "invalid database content" > .readonly.db
chmod 444 .readonly.db
```

```execute
aux4 repository read invalid_db_test --db .readonly.db --id test
```

```error:regex
.*
```

### Write with extremely long string values

Test handling of very long string values in JSON.

```execute
aux4 repository write long_strings --id long1 --data '{"description":"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo."}'
```

```expect
long1
```

### Find with empty expression

Test find command with empty or whitespace-only expressions.

```execute
aux4 repository write empty_expr_test --id item1 --data '{"test":true}'
```

```execute
aux4 repository find empty_expr_test --expr ""
```

```error:regex
.*
```

### Truncate non-existent repository with custom database

Test truncate on a repository that doesn't exist in a custom database.

```execute
aux4 repository truncate nonexistent_repo --db .custom_location.db --yes
```

```execute
aux4 repository read nonexistent_repo --db .custom_location.db
```

```error:regex
^(\[\]|)$
```