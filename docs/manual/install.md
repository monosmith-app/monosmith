# smith install

`smith install` (alias `smith i`) does two things. With no package it installs everything
a fresh clone needs, for every language in the workspace: `npm install` for the Node
projects, the Python venv and its requirements, `go mod download`. With packages it adds
them to one project with that project's own toolchain: `npm install --workspace`, the
project's `pyproject.toml` and the venv, `go get`. Each project owns its dependencies, so a
package needs a target: `smith install -w <project> <package>`.

## When to use it

- After cloning a workspace.
- To add a dependency to a project without learning each language's package manager.

To remove a dependency use [`smith uninstall`](uninstall.md); to update, [`smith up`](up.md);
to make versions agree, [`smith sync`](sync.md).

## Synopsis

```sh
smith install
smith install -w <project> [-D] <package>...
```

## Arguments and flags

<!-- smith:begin install -->
Add a package to a project (-w), or with none install everything the workspace declares (up updates, sync aligns versions)

Usage: `smith install [packages...] [flags]`

Alias: `smith i`

| Argument | Required | Meaning |
|---|---|---|
| `packages...` | no | packages to add; none installs the whole workspace |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `-D, --save-dev` | boolean | off | add as a development dependency (npm devDependencies, the Python dev extra, a Go tool) |
| `-w, --workspace <workspace>` | string, repeatable |  | target project, by name or directory; repeat it for several |

<!-- smith:end install -->

- `-w` names the target project by name or directory, and may be repeated to add the same
  package to several.
- `-D`/`--save-dev` adds a development dependency: npm `devDependencies`, the Python `dev`
  extra, a Go tool.
- A package may carry a version (`github.com/pkg/errors@v0.9.1`) as the language's tool
  understands it.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | Installed. |
| 1 | A package was named without `-w`, the target does not exist, or the package manager failed. |

## Files and environment

Edits the project manifest (`package.json`, `pyproject.toml`, `go.mod`) and the lockfile
(`package-lock.json`); creates `node_modules`, `.venv`. The package manager's own variables
apply (for example `NPM_CONFIG_REGISTRY`), and `[workspace] registry` in `smith.toml` names
a registry to install from.

## Examples

A fresh clone:

```sh
smith install
```

```text
✔ go: installed 1 project(s)
✔ node: installed 2 project(s) (package-lock.json)
✔ python: installed 1 project(s)
...
up to date in 1s
```

Add a dependency to one project, and a dev dependency:

```sh
smith install -w web zod
```

```text
✔ web: added zod (apps/web/package.json, package-lock.json)
...
up to date in 2s
```

```sh
smith install -w web -D left-pad
```

```text
✔ web: added left-pad (apps/web/package.json, package-lock.json)
...
added 1 package in 2s
```

(The `npm warn` lines in between come from npm and are cut.)

## Common errors

| Message | Cause and remedy |
|---|---|
| `adding a package needs a target project: smith install -w <project> <package>. A bare smith install (no package) installs the whole workspace` | Add `-w <project>`. |
| `no project "nope": name it by its name or its directory (projects: api, tools, ui, web, web2)` | Check the name with `smith projects`. |

## Related

[`smith uninstall`](uninstall.md), [`smith up`](up.md), [`smith sync`](sync.md),
[`smith ci setup`](ci.md#setup) (what CI runs instead).
