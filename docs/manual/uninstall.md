# smith uninstall

`smith uninstall` (alias `smith un`) removes packages from one project, the counterpart of
[`smith install`](install.md). A dependency belongs to the project that uses it, so it needs
a target: `smith uninstall -p <project> <package>`.

The `package.json` at the top of the workspace is not a project: name it with `-p .`
(`smith deps up` lists its packages as `(root)`). A development dependency is removed like any
other: `smith uninstall -p . @swc/core`.

A name that is a library of the workspace (a project, its folder, or the package it
exports) has its link removed: `smith uninstall -p web @acme/core`. When the project did not
depend on it, smith says so and changes nothing. `workspace:<name>` names the library
explicitly, and `--registry` removes a registry package that has the name of a project.

## When to use it

- A dependency is no longer used by a project.

## Asking first

`smith uninstall` shows the project manifests it changes. A terminal asks `Go ahead? [y/N]`; without a terminal, smith refuses and names `--yes`.

## Synopsis

```sh
smith uninstall -p <project> <package>...
```

## Arguments and flags

<!-- smith:begin uninstall -->
Remove packages from a project (the opposite of install)

Usage: `smith uninstall <packages...> [flags]`

Alias: `smith un`

| Argument | Required | Meaning |
|---|---|---|
| `packages...` | yes | packages to remove |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--commit` | boolean | off | commit the change as one operation so smith undo can revert it (nothing is committed without it) |
| `-p, --project <project>` | string, repeatable |  | target project, by name or directory; repeat it or separate them with commas (also -w, --workspace, --projects) |
| `--registry` | boolean | off | remove the names as registry packages, even when a name is also a project of this workspace |
| `-y, --yes` | boolean | off | do not ask; without a terminal the packages are removed only with --yes |

<!-- smith:end uninstall -->

`-p` is repeatable. The packages must be declared by the project; the language's own tool
removes them and the lockfile is refreshed.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | Removed. |
| 1 | No target was given, the project does not exist, or the package manager failed. |

## Files and environment

Edits the project manifest and the lockfile. Same environment as `smith install`.

## Examples

```sh
smith uninstall -p web zod
```

```text
✔ web: removed zod (apps/web/package.json, package-lock.json)
...
up to date in 1s
```

## Common errors

| Message | Cause and remedy |
|---|---|
| `a dependency belongs to the project that uses it: smith uninstall -p <project> <package>` | Add `-p`. |

## Related

[`smith install`](install.md), [`smith sync`](sync.md), [`smith doctor`](doctor.md) (reports a dev
tool declared outside the root).
