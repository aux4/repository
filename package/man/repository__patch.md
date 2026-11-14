The patch command merge-updates the `data` of an existing record. It performs a **shallow top-level merge**: keys present in the patch payload set or override the corresponding keys in the stored record, keys not mentioned are preserved, and any `id` field in the payload is ignored (the id comes from `--id`). On success it prints the record id.

patch is **update-only** — it is not an upsert. If the id does not exist, patch fails with an error and creates nothing. This is the key distinction from `write`, which inserts or replaces. Use `write` to create records and `patch` to amend existing ones.

- **Key order:** existing keys keep their original position (their value is updated in place); brand-new keys are appended in the order they appear in the payload.
- **Preserved:** the record's `metadata` (including ttl and access labels) and its `created_at` timestamp are left untouched. `updated_at` is bumped to the current time. Number literals in the patched values are preserved.
- The payload is supplied with `--data` or piped via stdin. Updates use bound parameters.
- **`--increment`:** a comma-separated list of numeric fields to bump by 1 server-side, applied after the `--data` merge. A missing or non-numeric field is treated as 0 before the bump. This lets a caller atomically increment a counter (e.g. a retry/attempts count) without computing the arithmetic itself — useful for composing durable primitives (like a work queue's `nack`) purely in the aux4 DSL, with no shell `$((...))` math. When `--increment` is used without `--data`, patch does not wait on stdin.

**Out of scope for this version:** patch does a shallow merge only — nested objects are replaced wholesale, not deep-merged — and it cannot delete/unset a field (a key can only be set or overridden). To change ttl/access without touching data, use `touch`.

**Storage backends.** The `--db` value selects the backend by scheme: a local file path (default `.local.db`) uses an embedded SQLite database, while a `libsql://` / `wss://` / `https://` URL targets a remote Turso/libSQL database (auth token read from the `TURSO_AUTH_TOKEN` environment variable). Behavior is identical on both backends.

#### Usage

```bash
aux4 repository patch <repository> --id <id> [--db <path>] [--data <json>] [--increment <field,...>]
```

--db         Database file path or remote libsql:// URL (default: `.local.db`)
--id         The id of the existing record to patch (required)
--data       JSON patch payload (or piped via stdin)
--increment  Comma-separated numeric fields to bump by 1, applied after `--data`

#### Example

```bash
# add a phone and correct the city on an existing contact
aux4 repository patch contacts --id c123 --data '{"phone":"555-0100","city":"NYC"}'
```

```text
c123
```

```bash
# atomically bump a retry counter without computing attempts + 1 yourself
aux4 repository patch messages --id m1 --data '{"leaseExpiresAt":0}' --increment attempts
```

```text
m1
```

```bash
# patching an unknown id fails (patch does not create records)
aux4 repository patch contacts --id missing --data '{"x":1}'
```

```text
Error: id "missing" not found in repository "contacts"
```
