## Test

```beforeAll
rm -f .local.db
```

```afterAll
rm -f .local.db
```

### Main repository command shows help when no subcommand

Test that the main repository command shows help information when called without arguments.

```execute
aux4 repository
```

```expect:partial
write
```

### Main repository command with --help flag

Test that the main repository command shows help with the --help flag.

```execute
aux4 repository --help
```

```expect:partial
Simplified aux4 database
```

### Main repository command lists available subcommands

Verify that help output includes the available subcommands.

```execute
aux4 repository --help
```

```expect:partial
write
```

```expect:partial
read
```

```expect:partial
find
```

```expect:partial
delete
```

```expect:partial
truncate
```