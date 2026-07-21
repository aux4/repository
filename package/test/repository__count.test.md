# repository count

`count <repository>` prints the number of records matching an optional `--expr`,
applying the same access/ttl visibility as `find`. It reuses find's expression
translation and `;`/`--`/`/*` injection rejection. The output is a bare integer;
a count of `0` is a valid answer and exits 0 (unlike `read`/`find`, which exit 4
on an empty result).

```beforeEach
rm -f .local.db
aux4 repository write items --id pub1 --data '{"age":30}'
aux4 repository write items --id pub2 --data '{"age":25}'
aux4 repository write items --id sec --data '{"age":40}' --access secret
aux4 repository write items --id exp --data '{"age":50}' --metadata '{"ttl":1}'
```

```afterAll
rm -f .local.db
```

### counts visible (public, non-expired) records by default

The restricted and expired records are excluded, leaving the two public ones.

```execute
aux4 repository count items
```

```expect
2
```

### --all --includeExpired counts every record

```execute
aux4 repository count items --all true --includeExpired true
```

```expect
4
```

### counts records matching an --expr (visibility still applies)

`age > 26` matches pub1 (30) among the visible public records; sec and exp are hidden.

```execute
aux4 repository count items --expr "age > 26"
```

```expect
1
```

### --access includes matching restricted records

```execute
aux4 repository count items --access secret
```

```expect
3
```

### --includeExpired includes expired records

```execute
aux4 repository count items --includeExpired true
```

```expect
3
```

### a count of zero exits 0

```execute
aux4 repository count items --expr "age > 999"
```

```expect
0
```

### rejects SQL comment / terminator sequences in --expr

```execute
aux4 repository count items --expr "age > 0; DROP TABLE items"
```

```error:partial
Error: --expr may not contain *?
```

### counts all records with a bare true expression

```execute
aux4 repository count items --expr "1=1" --all true --includeExpired true
```

```expect
4
```
