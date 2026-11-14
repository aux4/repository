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
aux4 repository write jobs1 --id job1 --data '{"task":"send-email"}' >/dev/null
aux4 repository lease jobs1 --ttl 60
```

```expect:partial
[{"id":"job1","task":"send-email","leaseExpiresAt":*?}]
```

### A record under an active lease is not eligible again

A second lease attempt within the visibility window finds nothing else to lease.

```execute
aux4 repository write jobs2 --id job2 --data '{"task":"ping"}' >/dev/null
aux4 repository lease jobs2 --ttl 60 >/dev/null
aux4 repository lease jobs2 --ttl 60
```

```error:regex
^(\[\]|)$
```

### An expired lease is eligible again and gets re-leased

A negative `--ttl` sets `leaseExpiresAt` already in the past, so the record is immediately eligible
for the next lease call.

```execute
aux4 repository write jobs3 --id job3 --data '{"task":"cleanup"}' >/dev/null
aux4 repository lease jobs3 --ttl=-10
aux4 repository lease jobs3 --ttl 60 --owner worker-2
```

```expect:partial
[{"id":"job3","task":"cleanup","leaseExpiresAt":*?}]
[{"id":"job3","task":"cleanup","leaseExpiresAt":*?,"leaseOwner":"worker-2"}]
```

### Releasing a lease early (patch leaseExpiresAt back to 0) makes it eligible again

```execute
aux4 repository write jobs4 --id job4 --data '{"task":"retry-me"}' >/dev/null
aux4 repository lease jobs4 --ttl 300 >/dev/null
aux4 repository patch jobs4 --id job4 --data '{"leaseExpiresAt":0}' >/dev/null
aux4 repository lease jobs4 --ttl 60
```

```expect:partial
[{"id":"job4","task":"retry-me","leaseExpiresAt":*?}]
```

### Completing a lease (delete) removes it from future leases

```execute
aux4 repository write jobs5 --id job5 --data '{"task":"once"}' >/dev/null
aux4 repository lease jobs5 --ttl 60 >/dev/null
aux4 repository delete jobs5 --id job5
aux4 repository lease jobs5 --ttl 60
```

```error:regex
^(\[\]|)$
```

### FIFO ordering: the oldest eligible record is leased first

```execute
aux4 repository write fifo --id first --data '{"n":1}' >/dev/null
aux4 repository write fifo --id second --data '{"n":2}' >/dev/null
aux4 repository lease fifo --ttl 60
```

```expect:partial
[{"id":"first","n":1,"leaseExpiresAt":*?}]
```

### No eligible record prints [] and exits non-zero

```execute
aux4 repository lease emptyrepo --ttl 60
```

```error:regex
^(\[\]|)$
```

### --ttl is required

```execute
aux4 repository lease jobs6
```

```error:partial
Error: --ttl is required
```

### Group serialization: only one record per group value is leased at a time

Two messages share the same `chatId`; only the first is eligible until its lease is released.

```execute
aux4 repository write chats --id m1 --data '{"chatId":"c1","text":"hi"}' >/dev/null
aux4 repository write chats --id m2 --data '{"chatId":"c1","text":"are you there?"}' >/dev/null
aux4 repository write chats --id m3 --data '{"chatId":"c2","text":"unrelated"}' >/dev/null
aux4 repository lease chats --ttl 300 --group chatId --owner w1
aux4 repository lease chats --ttl 300 --group chatId --owner w2
```

```expect:partial
[{"id":"m1","chatId":"c1","text":"hi","leaseExpiresAt":*?,"leaseOwner":"w1"}]
[{"id":"m3","chatId":"c2","text":"unrelated","leaseExpiresAt":*?,"leaseOwner":"w2"}]
```

The second lease call skips `m2` (same group `c1` as the already-leased `m1`) and picks the next
eligible candidate from a different group (`m3`) instead.

### Group serialization: completing the active lease frees the group for its sibling

Releasing `g1`'s own lease early (patch `leaseExpiresAt` back to `0`) makes `g1` itself eligible
again too, and being the older record it would win the next lease again (FIFO). To prove the group
is genuinely freed for its sibling, complete `g1` (delete it) instead of merely releasing it.

```execute
aux4 repository write chats2 --id g1 --data '{"chatId":"c1","text":"first"}' >/dev/null
aux4 repository write chats2 --id g2 --data '{"chatId":"c1","text":"second"}' >/dev/null
aux4 repository lease chats2 --ttl 300 --group chatId >/dev/null
aux4 repository delete chats2 --id g1 >/dev/null
aux4 repository lease chats2 --ttl 300 --group chatId
```

```expect:partial
[{"id":"g2","chatId":"c1","text":"second","leaseExpiresAt":*?}]
```

### Concurrency: N parallel leases of the same single record, exactly one wins

```execute
aux4 repository write race --id R1 --data '{"n":1}' >/dev/null
for i in 1 2 3 4 5 6 7 8 9 10; do
  (aux4 repository lease race --ttl 300 >/dev/null 2>/dev/null && echo "$i" >> lease-winners.txt) &
done
wait
wc -l < lease-winners.txt | tr -d ' '
```

```expect
1
```
