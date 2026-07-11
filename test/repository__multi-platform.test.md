# repository (multi-platform build)

The repository package ships platform-specific `.aux4` overrides under `dist/darwin` and
`dist/linux`. The `find` command delegates its where-clause translation to a private,
namespaced `repository translate-where run` command that lives in a nested
`repository:translate-where` profile. Only that tiny nested profile is overridden per platform
so it uses the correct `sed -i` syntax on each OS (BSD on macOS, GNU on Linux). The full
`repository` profile (including the private `translate-where` routing command) is defined once
in the root `.aux4` and merged into every platform zip. Because pkger merges whole profiles by
name, the dist overrides swap only the `repository:translate-where` profile and never touch the
main `repository` profile. These tests build the package with pkger and assert that each
platform zip carries the full `repository` profile plus the right `translate-where` variant.

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
Creating zip file darwin_aux4_repository_1.1.0.zip
Creating zip file linux_aux4_repository_1.1.0.zip
Creating zip file aux4_repository_1.1.0.zip
_build/aux4_repository_1.1.0.zip
_build/darwin_aux4_repository_1.1.0.zip
_build/linux_aux4_repository_1.1.0.zip
```

## find portability

### linux zip uses GNU sed (no empty-suffix argument)

The Linux override must call `sed -i -E` with no separate `''` argument, otherwise GNU sed
reads the empty string as a script and the expression as a filename (`sed: can't read s/...`).
The sed now lives in the private `repository:translate-where` profile's `run` command.

```execute
rm -rf _build/linux && mkdir -p _build/linux && unzip -oq _build/linux_aux4_repository_1.1.0.zip -d _build/linux && jq -r '.profiles[] | select(.name=="repository:translate-where").commands[] | select(.name=="run").execute[] | select(startswith("sed"))' _build/linux/aux4/repository/.aux4 | grep -oE "^sed -i( '')? -E"
```

```expect
sed -i -E
```

### darwin zip uses BSD sed (empty-suffix argument)

The macOS override keeps the native BSD form `sed -i '' -E`, which requires the separate
empty-string suffix argument.

```execute
rm -rf _build/darwin && mkdir -p _build/darwin && unzip -oq _build/darwin_aux4_repository_1.1.0.zip -d _build/darwin && jq -r '.profiles[] | select(.name=="repository:translate-where").commands[] | select(.name=="run").execute[] | select(startswith("sed"))' _build/darwin/aux4/repository/.aux4 | grep -oE "^sed -i( '')? -E"
```

```expect
sed -i '' -E
```

### each platform zip carries the full repository profile

Because pkger merges at the profile level, the whole `repository` profile is defined once in
the root `.aux4` and merged unchanged into every platform zip (the dist overrides carry only
the tiny nested `repository:translate-where` profile). The profile keeps its public commands
plus the private `visibility-clause` and `translate-where` helper commands.

```execute
jq -r '[.profiles[] | select(.name=="repository").commands[].name] | join(",")' _build/linux/aux4/repository/.aux4
```

```expect
init,write,read,find,delete,truncate,clean,visibility-clause,translate-where
```

### the repository profile appears exactly once (not duplicated by dist)

The dist override must not duplicate the main `repository` profile; only the nested
`repository:translate-where` profile is platform-specific.

```execute
jq -r '[.profiles[] | select(.name=="repository")] | length' _build/linux/aux4/repository/.aux4
```

```expect
1
```

### the nested translate-where profile is present once per platform

```execute
jq -r '[.profiles[] | select(.name=="repository:translate-where")] | length' _build/linux/aux4/repository/.aux4
```

```expect
1
```

### find delegates to translate-where (no inline sed)

The `find` command no longer contains an inline `sed`; it calls the private, namespaced
`repository translate-where run` helper so the platform-specific part is isolated.

```execute
jq -r '.profiles[] | select(.name=="repository").commands[] | select(.name=="find").execute[] | select(contains("translate-where"))' _build/linux/aux4/repository/.aux4
```

```expect
aux4 repository translate-where run --file ${tmpWhere}
```

### find contains no sed instruction

```execute
jq -r '[.profiles[] | select(.name=="repository").commands[] | select(.name=="find").execute[] | select(startswith("sed"))] | length' _build/linux/aux4/repository/.aux4
```

```expect
0
```
