# The vertical-slices plugin

The `vertical-slices` plugin keeps code organised by what it does: one folder per outcome, flat,
its files named for the role they play. For Go, Python, Dart, Rust, Ruby, C++, .NET, JVM and PHP
projects it adds a `slice-check` task that runs before `lint`; for Node projects it checks, in
`smith doctor`, that the root ESLint config loads the rules that enforce the same layout.

## At a glance

| | |
|---|---|
| Kind | `convention` |
| Roles | `structure` |
| Applies to | the languages of the checkers it knows, plus `node` for doctor |
| Capabilities | `contribute`, `doctor` |
| Stacks it publishes | none; turn it on with `[conventions.vertical-slices]` (`smith new` does) or name it in a stack's `structure` |
| Project kinds | any |

## What it does

For every project of a supported language the plugin contributes a task and makes the project's
`lint` (when it has one) wait for it:

```text
util:slice-check — checks the vertical-slice layout: role-named files, no junk-drawer folders
  command    smith slices check --lang go .
  cwd        libs/util
  dependsOn  —
  inputs     **/*.go
  outputs    —
  env        —
  cache      true
```

The inputs are the project's source files, so the result is reused until a file changes, appears or
disappears. The rules per language (file naming, barrels, test placement, which folders are skipped)
are in [Conventions](../../../reference/conventions.md#vertical-slices); the checker is
[`smith slices check`](../../../manual/slices-check.md).

A violation fails the task, and with it `lint`:

```text
util:slice-check │ helpers/ - "helpers" is a junk drawer; name the folder for the outcome it serves
util:slice-check │ smith: 1 file(s) or folder(s) break the vertical-slice layout; an exception is written in the project README
util:slice-check ✖ exited with code 1
```

For Node there is no task. Node's rules live in `@monosmith/eslint-config` (run by the `lint` task
of the [eslint](eslint.md) plugin), and the plugin's `doctor` method fails when the root ESLint
config does not load that package.

## Settings

```toml
[conventions.vertical-slices]
# languages = ["go"]     # default: every language the convention supports
# exclude = ["legacy"]   # projects it leaves alone
```

A workspace can instead name the plugin in a stack: `structure = ["vertical-slices"]`, so only
those projects are checked.

## Releasing and CI

No `publish` and no `ci` method. `slice-check` is a task like any other, so `smith ci verify` runs
it for the affected projects.

## Verified

smith 1.62.0, Windows 11. In a Go workspace: `smith explain util:slice-check`, `smith run
slice-check -p util` (`vertical-slice layout OK`), a failing run with a `helpers/` folder (above), and
`smith plugin check vertical-slices`. The per-language rules were not re-verified here: they belong
to `smith slices check`.

## See also

- [Conventions](../../../reference/conventions.md), [`smith slices`](../../../manual/slices.md)
- [Creating a structure plugin](../../creating/structure-plugin.md): write your own rules
