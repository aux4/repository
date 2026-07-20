# repository (multi-platform build)

`aux4/repository` is a self-contained, statically linked Go binary — no `sqlite3`,
`jq`, or `sed` at runtime, and no platform-specific `.aux4` overrides. The build
cross-compiles the binary for all eight platform/arch combinations into
`package/dist/<os>/<arch>/`, and `pkger build` produces one zip per combination
(plus the hub meta zip), each carrying the matching binary at the package root.

These tests build the package and assert that every platform zip ships the right
binary and the single, unified `repository` command surface (no leftover
`visibility-clause` / `translate-where` helpers from the old shell version).

```beforeAll
rm -rf _mp && mkdir -p _mp/pkg/repository _mp/out
( cd ../.. && aux4 build ) >/dev/null 2>&1
cp -R ../.aux4 ../sql ../man ../LICENSE ../README.md ../dist _mp/pkg/repository/
aux4 aux4 pkger build _mp/pkg/repository --out _mp/out >/dev/null 2>&1
```

```afterAll
rm -rf _mp
```

## pkger build

### emits one zip per platform/arch plus the meta zip

```execute
ls _mp/out/ | grep -E "_aux4_repository_1.3.0.zip$|^aux4_repository_1.3.0.zip$" | sort
```

```expect
aux4_repository_1.3.0.zip
darwin_amd64_aux4_repository_1.3.0.zip
darwin_arm64_aux4_repository_1.3.0.zip
linux_386_aux4_repository_1.3.0.zip
linux_amd64_aux4_repository_1.3.0.zip
linux_arm64_aux4_repository_1.3.0.zip
windows_386_aux4_repository_1.3.0.zip
windows_amd64_aux4_repository_1.3.0.zip
windows_arm64_aux4_repository_1.3.0.zip
```

## per-platform binary

### the linux/amd64 zip carries the linux/amd64 binary

The binary inside the zip is byte-identical to `dist/linux/amd64/aux4-repository`.

```execute
rm -rf _mp/x_linux && mkdir -p _mp/x_linux && unzip -oq _mp/out/linux_amd64_aux4_repository_1.3.0.zip -d _mp/x_linux && cmp -s _mp/x_linux/aux4/repository/aux4-repository ../dist/linux/amd64/aux4-repository && echo "linux/amd64 binary matches"
```

```expect
linux/amd64 binary matches
```

### the darwin/arm64 zip carries the darwin/arm64 binary

```execute
rm -rf _mp/x_darwin && mkdir -p _mp/x_darwin && unzip -oq _mp/out/darwin_arm64_aux4_repository_1.3.0.zip -d _mp/x_darwin && cmp -s _mp/x_darwin/aux4/repository/aux4-repository ../dist/darwin/arm64/aux4-repository && echo "darwin/arm64 binary matches"
```

```expect
darwin/arm64 binary matches
```

### the windows/amd64 zip carries the .exe binary

```execute
unzip -l _mp/out/windows_amd64_aux4_repository_1.3.0.zip | grep -oE "aux4/repository/aux4-repository\.exe" | head -1
```

```expect
aux4/repository/aux4-repository.exe
```

## command surface

### the repository profile exposes exactly the Go-backed commands

The packaged manifest carries only the seven public commands routed to the
binary — the retired `visibility-clause` and `translate-where` helpers are gone.

```execute
rm -rf _mp/x_man && mkdir -p _mp/x_man && unzip -oq _mp/out/linux_amd64_aux4_repository_1.3.0.zip -d _mp/x_man && jq -r '[.profiles[] | select(.name=="repository").commands[].name] | join(",")' _mp/x_man/aux4/repository/.aux4
```

```expect
init,write,read,find,delete,truncate,clean
```

### the retired translate-where profile is absent

```execute
jq -r '[.profiles[] | select(.name=="repository:translate-where")] | length' _mp/x_man/aux4/repository/.aux4
```

```expect
0
```

### the manifest declares no system dependencies

The binary is self-contained, so no `sqlite3` / `jq` / `sed` system dependencies
are declared.

```execute
jq -r '.system // [] | length' _mp/x_man/aux4/repository/.aux4
```

```expect
0
```
