<!-- lore-master
parent: plugin.md
-->
# smith plugin pack

`smith plugin pack` turns a plugin directory into one archive, `<name>-<version>.smith-plugin.zip`,
that [`smith plugin install`](plugin-install.md) accepts, and writes the archive's SHA-256 beside it.
The archive holds the plugin's files, the program `plugin.json` names for each platform, and a
`checksums.sha256` that lists the SHA-256 of every file. Packing is deterministic: files are stored
uncompressed, so the same directory always gives the same bytes on any smith build, and anyone can
pack the same source and compare digests.

## When to use it

- You wrote a plugin and want to hand it to other people or to CI as one file.
- You are about to attach a plugin to a GitHub Release: pack it, then attach the archive **and** its
  `.sha256` file.

## Synopsis

```sh
smith plugin pack [dir] [--out <dir>]
```

## Arguments and flags

<!-- smith:begin plugin pack -->
Pack a plugin directory into an archive `smith plugin install` accepts

Usage: `smith plugin pack [dir] [flags]`

| Argument | Required | Meaning |
|---|---|---|
| `dir` | no | the plugin directory (holds a plugin.json with a name and a version); default the current directory |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `-o, --out <out>` | string | the current directory | directory to write the archive to |

<!-- smith:end plugin pack -->

## What goes into the archive

| In | Out |
|---|---|
| every file of the directory, in name order, with a fixed timestamp | `.git`, `.hg` and `.svn` folders, `.DS_Store`, earlier `*.smith-plugin.zip` files and their `.sha256`, an old `checksums.sha256`, and whatever the directory's `.smithpluginignore` lists |
| the files the first element of `command` names for any platform (`./bin/smith-plugin-widget-{os}-{arch}{exe}` matches `bin/smith-plugin-widget-linux-amd64` and `bin/smith-plugin-widget-windows-arm64.exe`), marked executable | |
| `checksums.sha256`, written by smith: `<sha256>  <path>` per file, the format of `sha256sum` | |

`.smithpluginignore` sits at the root of the plugin directory (and is not packed). One pattern per
line: `name` leaves out that name at any depth, `dir/` only a directory, `a/b` or `/a` a path from the
root, `*` and `**` are globs, `#` starts a comment. There is no negation.

```text
# not part of the plugin
src/
*.log
```

smith refuses to pack a directory that holds a symbolic link or another special file, and a
manifest whose `command` names a `./` program that no file matches. A command that is looked up on
`PATH` (`["widget-tool"]`) packs without a program. **Build the program for every platform you want to
support into the directory first**: `pack` does not build, and it does not run the plugin; use
[`smith plugin check`](plugin-check.md) for that.

Everything in the directory is packed, so keep secrets and build leftovers out of it. The command
prints what it packed.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The archive and its digest file were written. |
| 1 | No `plugin.json`, no `version` in it, an unusable name or version, a link in the directory, or a `command` program with no file. |

## Files and environment

Reads the directory; writes `<name>-<version>.smith-plugin.zip` and `<name>-<version>.smith-plugin.zip.sha256`
in `--out` (the current directory by default). It does not read `smith.toml` and needs no workspace.

## Examples

```sh
smith plugin pack ./widget --out dist
```

```text
Packed widget 0.1.0: 3 files, 14852081 bytes.
  archive  dist/widget-0.1.0.smith-plugin.zip
  sha256   db7fb4c07eeba779d2dd7dc313873cd180f8dbe133f8e4bb76c9c6956d22204d  (also in widget-0.1.0.smith-plugin.zip.sha256)
  program  bin/smith-plugin-widget-windows-amd64.exe
The plugin was not started. Publish the archive and its .sha256 file together; users install it with `smith plugin install`.
```

## Common errors

| Message | Cause and remedy |
|---|---|
| `✖ <dir> has no plugin.json` | Run it in the plugin's directory, or give the directory. |
| `✖ <dir>/plugin.json has no version` | Add `"version"` to `plugin.json`: it names the archive. |
| `plugin.json runs ./bin/..., but no file in <dir> matches it` | Build the program for at least one platform first. |
| `<path> is a link or special file` | Replace the link with a copy of the file. |

## Related

[`smith plugin install`](plugin-install.md), [`smith plugin check`](plugin-check.md),
[Packaging and distributing a plugin](../plugins/creating/packaging-a-plugin.md),
[Decision 0038](../architecture/decisions/0038-a-plugin-is-packed-as-a-checksummed-archive-and-installed-without-running-it.md).
