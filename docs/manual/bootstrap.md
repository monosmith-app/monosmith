# smith bootstrap

`smith bootstrap` writes a small launcher you commit: `tools/smith/smith` (sh) and
`tools/smith/smith.cmd` (Windows). Run `tools/smith/smith build` instead of `smith build` and a
laptop and CI use the same smith, the one the workspace pins in `[ci] smith_version`.

## When to use it

- The team has people on different smith versions, and a newer smith refuses an older
  `smith.toml` (or the other way round).
- You want a first `git clone`, then one command, with no install step to document.

## Synopsis

```sh
smith bootstrap [--check] [--yes]
```

## Arguments and flags

<!-- smith:begin bootstrap -->
Write the launcher that runs the smith version the workspace pins

Usage: `smith bootstrap [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--check` | boolean | off | change nothing; fail when the launcher is missing or out of date |
| `--commit` | boolean | off | commit the change as one operation so smith undo can revert it (nothing is committed without it) |
| `-y, --yes` | boolean | off | do not ask for confirmation |

<!-- smith:end bootstrap -->

## What the launcher does

It reads `smith_version` from `smith.toml` each time it runs, so changing the pin needs no new
file. Then:

1. No pin, or `latest`: it runs the smith on `PATH`.
2. The smith on `PATH` has the pinned version: it runs that one. A smith that is newer or older
   than the pin is not used.
3. A copy in `.smith/bin/<version>` exists: it runs that copy.
4. `tools/smith/install.sh` exists (`smith ci generate` writes it for workspaces with no Node,
   Python or .NET toolchain): it installs the pinned version into `.smith/bin/<version>`.
5. `npx` exists: it runs `npx monosmith@<version>`.
6. Otherwise it stops with the command to install that version.

The Windows launcher does steps 1, 2, 3 and 5 (the installer is a shell script). A smith on `PATH`
that is the launcher itself is skipped, so putting `tools/smith` on `PATH` is safe. `.smith/` is
already ignored by git.

On Windows, git does not keep the executable bit of the sh launcher: after `git add`, run
`git update-index --chmod=+x tools/smith/smith`.

## Notes

- CI pipelines that smith generates already install the pinned version
  ([ADR 0055](../architecture/decisions/0055-a-committed-launcher-runs-the-pinned-smith.md)); the
  launcher is for laptops and scripts.
- A pin must be an exact version. smith has no range syntax for it.
- With `commit_operations = true` under `[workspace]` the files are one commit.
