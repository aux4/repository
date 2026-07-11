# repository (multi-platform build)

The repository package ships platform-specific `.aux4` overrides under `dist/darwin` and
`dist/linux` so the `find` command uses the correct `sed -i` syntax on each OS (BSD on macOS,
GNU on Linux). These tests build the package with pkger and assert that each platform zip
carries the right `find` variant.

```beforeAll
rm -rf _build && mkdir -p _build
```

```afterAll
rm -rf _build
```

## pkger build

### emits darwin, linux and meta zips

Building the package produces one zip per platform plus the meta (hub) zip.

```execute
( cd .. && aux4 aux4 pkger build . --out test/_build ) 2>&1 | grep -E "Creating zip file (darwin|linux|aux4)_" && ls _build/darwin_aux4_repository_*.zip _build/linux_aux4_repository_*.zip _build/aux4_repository_*.zip
```

```expect:partial
Creating zip file darwin_aux4_repository_1.0.0.zip
Creating zip file linux_aux4_repository_1.0.0.zip
Creating zip file aux4_repository_1.0.0.zip
_build/aux4_repository_1.0.0.zip
_build/darwin_aux4_repository_1.0.0.zip
_build/linux_aux4_repository_1.0.0.zip
```

## find portability

### linux zip uses GNU sed (no empty-suffix argument)

The Linux override must call `sed -i -E` with no separate `''` argument, otherwise GNU sed
reads the empty string as a script and the expression as a filename (`sed: can't read s/...`).

```execute
rm -rf _build/linux && mkdir -p _build/linux && unzip -oq _build/linux_aux4_repository_1.0.0.zip -d _build/linux && jq -r '.profiles[] | select(.name=="repository").commands[] | select(.name=="find").execute[] | select(startswith("sed"))' _build/linux/aux4/repository/.aux4 | grep -oE "^sed -i( '')? -E"
```

```expect
sed -i -E
```

### darwin zip uses BSD sed (empty-suffix argument)

The macOS override keeps the native BSD form `sed -i '' -E`, which requires the separate
empty-string suffix argument.

```execute
rm -rf _build/darwin && mkdir -p _build/darwin && unzip -oq _build/darwin_aux4_repository_1.0.0.zip -d _build/darwin && jq -r '.profiles[] | select(.name=="repository").commands[] | select(.name=="find").execute[] | select(startswith("sed"))' _build/darwin/aux4/repository/.aux4 | grep -oE "^sed -i( '')? -E"
```

```expect
sed -i '' -E
```

### each platform zip carries the full repository profile

Because pkger merges at the profile level, each override must redefine the whole `repository`
profile (all commands), not just `find`.

```execute
jq -r '[.profiles[] | select(.name=="repository").commands[].name] | join(",")' _build/linux/aux4/repository/.aux4
```

```expect
init,write,read,find,delete,truncate,clean
```
