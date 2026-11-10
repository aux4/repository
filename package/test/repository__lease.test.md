## Test

```beforeEach
rm -f .local.db
rm -f lease-winners.txt
```

```afterEach
rm -f lease-winners.txt
```

### Leasing the only eligible record succeeds

```execute
aux4 repository write jobs --id job1 --data '{"task":"send-email"}'
aux4 repository lease jobs --ttl 60
```

```expect:partial
{"id":"job1","task":"send-email","leaseExpiresAt":*?}
```

### A record under an active lease is not eligible again

A second lease attempt within the visibility window finds nothing else to lease.

```execute
aux4 repository write jobs --id job2 --data '{"task":"ping"}'
aux4 repository lease jobs --ttl 60 --expr "id = 'job2'" >/dev/null
aux4 repository lease jobs --ttl 60 --expr "id = 'job2'"
```

```error:regex
^(\[\]|)$
```

### An expired lease is eligible again and gets re-leased

A negative `--ttl` sets `leaseExpiresAt` already in the past, so the record is immediately eligible
for the next lease call.

```execute
aux4 repository write jobs --id job3 --data '{"task":"cleanup"}'
aux4 repository lease jobs --ttl=-10 --expr "id = 'job3'"
aux4 repository lease jobs --ttl 60 --expr "id = 'job3'" --owner worker-2
```

```expect:partial
{"id":"job3","task":"cleanup","leaseExpiresAt":*?}
{"id":"job3","task":"cleanup","leaseExpiresAt":*?,"leaseOwner":"worker-2"}
```

### Releasing a lease early (patch leaseExpiresAt back to 0) makes it eligible again

```execute
aux4 repository write jobs --id job4 --data '{"task":"retry-me"}'
aux4 repository lease jobs --ttl 300 --expr "id = 'job4'" >/dev/null
aux4 repository patch jobs --id job4 --data '{"leaseExpiresAt":0}' >/dev/null
aux4 repository lease jobs --ttl 60 --expr "id = 'job4'"
```

```expect:partial
{"id":"job4","task":"retry-me","leaseExpiresAt":*?}
```

### Completing a lease (delete) removes it from future leases

```execute
aux4 repository write jobs --id job5 --data '{"task":"once"}'
aux4 repository lease jobs --ttl 60 --expr "id = 'job5'" >/dev/null
aux4 repository delete jobs --id job5
aux4 repository lease jobs --ttl 60 --expr "id = 'job5'"
```

```error:regex
^(\[\]|)$
```

### FIFO ordering: the oldest eligible record is leased first

```execute
aux4 repository write fifo --id first --data '{"n":1}'
aux4 repository write fifo --id second --data '{"n":2}'
aux4 repository lease fifo --ttl 60
```

```expect:partial
{"id":"first","n":1,"leaseExpiresAt":*?}
```

### No eligible record prints [] and exits non-zero

```execute
aux4 repository lease empty-repo --ttl 60
```

```error:regex
^(\[\]|)$
```

### --ttl is required

```execute
aux4 repository lease jobs
```

```error:partial
Error: --ttl is required
```

### Group serialization: only one record per group value is leased at a time

Two messages share the same `chatId`; only the first is eligible until its lease is released.

```execute
aux4 repository write chats --id m1 --data '{"chatId":"c1","text":"hi"}'
aux4 repository write chats --id m2 --data '{"chatId":"c1","text":"are you there?"}'
aux4 repository write chats --id m3 --data '{"chatId":"c2","text":"unrelated"}'
aux4 repository lease chats --ttl 300 --group chatId --owner w1
aux4 repository lease chats --ttl 300 --group chatId --owner w2
```

```expect:partial
{"id":"m1","chatId":"c1","text":"hi","leaseExpiresAt":*?,"leaseOwner":"w1"}
{"id":"m3","chatId":"c2","text":"unrelated","leaseExpiresAt":*?,"leaseOwner":"w2"}
```

The second lease call skips `m2` (same group `c1` as the already-leased `m1`) and picks the next
eligible candidate from a different group (`m3`) instead.

### Group serialization: releasing the active lease frees the group for its sibling

```execute
aux4 repository write chats2 --id g1 --data '{"chatId":"c1","text":"first"}'
aux4 repository write chats2 --id g2 --data '{"chatId":"c1","text":"second"}'
aux4 repository lease chats2 --ttl 300 --group chatId >/dev/null
aux4 repository patch chats2 --id g1 --data '{"leaseExpiresAt":0}' >/dev/null
aux4 repository lease chats2 --ttl 300 --group chatId
```

```expect:partial
{"id":"g2","chatId":"c1","text":"second","leaseExpiresAt":*?}
```

### Concurrency: N parallel leases of the same single record, exactly one wins

```execute
aux4 repository write race --id R1 --data '{"n":1}'
for i in 1 2 3 4 5 6 7 8 9 10; do
  (aux4 repository lease race --ttl 300 --expr "id = 'R1'" 2>/dev/null | grep -q '"id":"R1"' && echo "$i" >> lease-winners.txt) &
done
wait
wc -l < lease-winners.txt | tr -d ' '
```

```expect
1
```
