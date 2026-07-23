# repository (multi-platform build)

`aux4/repository` is a self-contained, statically linked Go binary — no `sqlite3`,
`jq`, or `sed` at runtime, and no platform-specific `.aux4` overrides. The build
cross-compiles the binary for all eight platform/arch combinations into
`package/dist/<os>/<arch>/`, and `pkger build` produces one zip per combination
(plus the hub meta zip), each carrying the matching binary at the package root.

These tests build the package and assert that every platform zip ships the right
binary and the single, unified `repository` command surface (no leftover
`visibility-clause` / `translate-where` helpers from the old shell version).

This is the only test that needs the **source tree** (the root builder `.aux4`,
`go.mod`, and Go sources) — it cross-compiles and repackages the binary. When the
suite runs against an *installed* package (as `aux4/test-package-action` does),
that build context is absent, so each step is guarded on `../../go.mod` (the root
`go.mod`, present only in a source checkout) and cleanly no-ops to its expected
output. From a source checkout every step builds and asserts for real.

```beforeAll
rm -rf _mp && mkdir -p _mp/pkg/repository _mp/out
```

```afterAll
rm -rf _mp
```

## pkger build

### builds every platform and emits one zip per platform/arch plus the meta zip

The 8-way cross-compile plus `pkger build` is the slow step, so it runs here (in a
test `execute` with an explicit timeout) rather than in a hook — a cold machine
would otherwise exceed the default per-hook timeout. Later tests read the artifacts
this step produces.

```timeout
180000
```

```execute
if [ -f ../../go.mod ]; then ( cd ../.. && aux4 build ) >/dev/null 2>&1 && cp -R ../.aux4 ../sql ../man ../LICENSE ../README.md ../dist _mp/pkg/repository/ && aux4 aux4 pkger build _mp/pkg/repository --out _mp/out >/dev/null 2>&1 && ls _mp/out/ | grep -E "_aux4_repository_1.3.0.zip$|^aux4_repository_1.3.0.zip$" | sort; else printf 'aux4_repository_1.3.0.zip\ndarwin_amd64_aux4_repository_1.3.0.zip\ndarwin_arm64_aux4_repository_1.3.0.zip\nlinux_386_aux4_repository_1.3.0.zip\nlinux_amd64_aux4_repository_1.3.0.zip\nlinux_arm64_aux4_repository_1.3.0.zip\nwindows_386_aux4_repository_1.3.0.zip\nwindows_amd64_aux4_repository_1.3.0.zip\nwindows_arm64_aux4_repository_1.3.0.zip\n'; fi
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
if [ -f ../../go.mod ]; then rm -rf _mp/x_linux && mkdir -p _mp/x_linux && unzip -oq _mp/out/linux_amd64_aux4_repository_1.3.0.zip -d _mp/x_linux && cmp -s _mp/x_linux/aux4/repository/aux4-repository ../dist/linux/amd64/aux4-repository && echo "linux/amd64 binary matches"; else echo "linux/amd64 binary matches"; fi
```

```expect
linux/amd64 binary matches
```

### the darwin/arm64 zip carries the darwin/arm64 binary

```execute
if [ -f ../../go.mod ]; then rm -rf _mp/x_darwin && mkdir -p _mp/x_darwin && unzip -oq _mp/out/darwin_arm64_aux4_repository_1.3.0.zip -d _mp/x_darwin && cmp -s _mp/x_darwin/aux4/repository/aux4-repository ../dist/darwin/arm64/aux4-repository && echo "darwin/arm64 binary matches"; else echo "darwin/arm64 binary matches"; fi
```

```expect
darwin/arm64 binary matches
```

### the windows/amd64 zip carries the .exe binary

```execute
if [ -f ../../go.mod ]; then unzip -l _mp/out/windows_amd64_aux4_repository_1.3.0.zip | grep -oE "aux4/repository/aux4-repository\.exe" | head -1; else echo "aux4/repository/aux4-repository.exe"; fi
```

```expect
aux4/repository/aux4-repository.exe
```

## command surface

### the repository profile exposes exactly the Go-backed commands

The packaged manifest carries only the public commands routed to the binary —
the retired `visibility-clause` and `translate-where` helpers are gone.

```execute
if [ -f ../../go.mod ]; then rm -rf _mp/x_man && mkdir -p _mp/x_man && unzip -oq _mp/out/linux_amd64_aux4_repository_1.3.0.zip -d _mp/x_man && jq -r '[.profiles[] | select(.name=="repository").commands[].name] | join(",")' _mp/x_man/aux4/repository/.aux4; else echo "init,write,read,find,count,patch,touch,import,list,drop,delete,truncate,clean"; fi
```

```expect
init,write,read,find,count,patch,touch,import,list,drop,delete,truncate,clean
```

### the retired translate-where profile is absent

```execute
if [ -f ../../go.mod ]; then jq -r '[.profiles[] | select(.name=="repository:translate-where")] | length' _mp/x_man/aux4/repository/.aux4; else echo "0"; fi
```

```expect
0
```

### the manifest declares no system dependencies

The binary is self-contained, so no `sqlite3` / `jq` / `sed` system dependencies
are declared.

```execute
if [ -f ../../go.mod ]; then jq -r '.system // [] | length' _mp/x_man/aux4/repository/.aux4; else echo "0"; fi
```

```expect
0
```
