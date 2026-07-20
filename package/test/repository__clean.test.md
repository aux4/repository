## Test

```beforeEach
rm -f .local.db
```

```afterAll
rm -f .local.db
```

### Clean removes expired records across every repository

Write expired records (past ttl) and live records (future ttl or no ttl) into two repositories, then run clean. Every expired record is removed and a per-repository summary is printed. This suite resets the database between steps, so each scenario writes and cleans in a single step.

```execute
aux4 repository write users --id u_dead --data '{"n":"d"}' --metadata '{"ttl":1}'
aux4 repository write users --id u_live --data '{"n":"l"}' --ttl 3600
aux4 repository write sessions --id s_dead --data '{"n":"d"}' --metadata '{"ttl":1}'
aux4 repository write sessions --id s_live --data '{"n":"l"}'
aux4 repository clean
```

```expect:partial
users: 1 expired records removed
sessions: 1 expired records removed
```

### Live records survive a clean

After cleaning, the live records remain and the expired ones are gone.

```execute
aux4 repository write users --id u_dead --data '{"n":"d"}' --metadata '{"ttl":1}'
aux4 repository write users --id u_live --data '{"n":"l"}' --ttl 3600
aux4 repository clean
aux4 repository read users --includeExpired true
```

```expect:partial
[
  {
    "id": "u_live",
    "n": "l"
  }
]
```

### Clean reports zero when nothing has expired

A repository whose records have no ttl or a future ttl is reported with a count of zero and left untouched.

```execute
aux4 repository write cache --id c1 --data '{"n":1}' --ttl 3600
aux4 repository write cache --id c2 --data '{"n":2}'
aux4 repository clean
aux4 repository read cache
```

```expect:partial
cache: 0 expired records removed
[
  {
    "id": "c1",
    "n": 1
  },
  {
    "id": "c2",
    "n": 2
  }
]
```
