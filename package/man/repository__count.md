The count command prints how many records match an optional filter, without returning the records themselves. It applies the same access-label and ttl visibility rules as `read`/`find` (the access/ttl filter runs in an inner subquery), and reuses `find`'s SQL-expression translation and injection defenses for `--expr`.

The output is a single bare integer on one line. A count of `0` is a valid answer and exits with status 0 — unlike `read`/`find`, count does not use the empty-result exit-4 convention. There is no `--sort`, `--select`, `--limit`, or `--offset` (a count is always the total of the matching set).

**Storage backends.** The `--db` value selects the backend by scheme: a local file path (default `.local.db`) uses an embedded SQLite database, while a `libsql://` / `wss://` / `https://` URL targets a remote Turso/libSQL database (auth token read from the `TURSO_AUTH_TOKEN` environment variable). Behavior is identical on both backends.

**`--expr` filtering.** When `--expr` is empty, count returns the number of all visible records. When set, it accepts the same expression syntax as `find` (comparison operators, LIKE/NOT LIKE, and/or, parentheses, `json_extract(...)` for nested fields). Bare field references are rewritten to `json_extract(data, '$.field')`. As with `find`, an `--expr` containing the SQL comment or statement-terminator sequences `;`, `--`, or `/*` is rejected with a non-zero exit.

**Visibility.** Records written with `write --access <labels>` are counted only when the caller passes a matching `--access` label (or `--all true`); public records are always counted. Expired records (ttl in the past) are excluded unless `--includeExpired true` is passed.

#### Usage

```bash
aux4 repository count <repository> [--db <path>] [--expr <expression>] [--access <labels>] [--all <true|false>] [--includeExpired <true|false>]
```

--db              Database file path or remote libsql:// URL (default: `.local.db`)
--expr            Optional filter expression (find syntax); empty counts all visible records
--access          Comma-separated caller access labels
--all             Bypass access filtering (default: `false`)
--includeExpired  Include expired records (default: `false`)

#### Example

```bash
# total visible records
aux4 repository count users
```

```text
42
```

```bash
# how many adults, including access-restricted records
aux4 repository count users --expr "age >= 18" --all true
```

```text
37
```
