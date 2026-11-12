## Test

```beforeEach
rm -f .local.db
```

### Renewing a live lease succeeds and prints the id

```execute
aux4 repository write jobs1 --id job1 --data '{"task":"send-email"}' >/dev/null
aux4 repository lease jobs1 --ttl 30 >/dev/null
aux4 repository renew jobs1 --id job1 --ttl 60
```

```expect
job1
```

### Renewing extends the visibility window (not re-leasable immediately after)

```execute
aux4 repository write jobs2 --id job2 --data '{"task":"ping"}' >/dev/null
aux4 repository lease jobs2 --ttl 1 >/dev/null
aux4 repository renew jobs2 --id job2 --ttl 300 >/dev/null
aux4 repository lease jobs2 --ttl 60
```

```error:regex
^(\[\]|)$
```

### Renewing a lease that already expired fails (not resurrected)

```execute
aux4 repository write jobs3 --id job3 --data '{"task":"cleanup"}' >/dev/null
aux4 repository lease jobs3 --ttl=-10 >/dev/null
aux4 repository renew jobs3 --id job3 --ttl 60
```

```error:regex
^(\[\]|)$
```

### Renewing an id that was never leased fails

```execute
aux4 repository write jobs4 --id job4 --data '{"task":"never-leased"}' >/dev/null
aux4 repository renew jobs4 --id job4 --ttl 60
```

```error:regex
^(\[\]|)$
```

### Renewing a non-existent id fails

```execute
aux4 repository renew jobs5 --id does-not-exist --ttl 60
```

```error:regex
^(\[\]|)$
```

### A renewed lease does not disturb leaseOwner or other data fields

```execute
aux4 repository write jobs6 --id job6 --data '{"task":"keep-owner"}' >/dev/null
aux4 repository lease jobs6 --ttl 30 --owner worker-1 >/dev/null
aux4 repository renew jobs6 --id job6 --ttl 60 >/dev/null
aux4 repository read jobs6 --id job6
```

```expect:partial
{"id":"job6","task":"keep-owner","leaseExpiresAt":*?,"leaseOwner":"worker-1"}
```

### --id is required

```execute
aux4 repository renew jobs7 --ttl 60
```

```error:partial
Error: --id is required
```

### --ttl is required

```execute
aux4 repository write jobs8 --id job8 --data '{"task":"x"}' >/dev/null
aux4 repository lease jobs8 --ttl 30 >/dev/null
aux4 repository renew jobs8 --id job8
```

```error:partial
Error: --ttl is required
```
