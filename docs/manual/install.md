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
smith install -w <project> <library of this workspace>... [--range <range>]
smith install -w <project> --registry <name>...
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
| `--range <range>` | string |  | the version range written for a library of the workspace (for example '*' or '^1.2.0'); without it smith follows the workspace or asks |
| `--registry` | boolean | off | install the names from the registry, even when a name is also a project of this workspace |
| `-D, --save-dev` | boolean | off | add as a development dependency (npm devDependencies, the Python dev extra, a Go tool) |
| `--verbose` | boolean | off | show the package manager's own output and warnings (otherwise warnings fold into one line) |
| `-w, --workspace <workspace>` | string, repeatable |  | target project, by name or directory; repeat it for several |
| `-y, --yes` | boolean | off | never ask: accept the defaults, and refuse a name that could be a project or a registry package |

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

When an install fails, smith puts the manifest and the lock
files back as they were before the command. A wrong name does not stay in the
manifest and does not break the next install.

## An internal library

A project can depend on a lib or a package of the same workspace. Give its name to
`smith install -w <project>` like any package:

```sh
smith install -w web @acme/core        # Node: "@acme/core" goes into package.json
smith install -w api shop-core-py      # Python: the distribution name goes into pyproject.toml
smith install -w svc core-go           # Go: a require of the module path, go.work finds the folder
```

smith links the library when the name is the name of a project of the workspace, its folder,
or the package or module name the project exports (`@acme/core`, `shop-core-py`,
`example.com/shop/libs/core-go`). Nothing is downloaded: the dependency points at the
folder. Each language writes it in its own form and installs or locks as it needs:

| Language | What smith writes |
|---|---|
| Node | the package name in `dependencies` of `package.json`, then `npm install` |
| Python | the distribution name in `[project] dependencies`; the library is installed in the shared venv first, so pip never asks PyPI |
| Go | `require <module> v0.0.0` in `go.mod`, and the library in `go.work` (a `replace` line when the workspace has no `go.work`) |
| Rust | a path dependency in `Cargo.toml`, by `cargo add --path` |
| .NET | a `ProjectReference` in the project file, by `dotnet add reference` |

For any other language smith says which manifest to edit by hand.
`smith plugin list` shows in the `LINK` column which plugins can do this.

smith shows the line it will write and the file it goes in before it writes. A failed link
puts the manifest and the lock files back, like a failed install. The command also makes the
dependency an edge of the project graph: `smith graph` and the up and down columns of the
Projects view show it at once, before any source file imports the library.

### When a name could be two things

A name that is a project and also a package on the language registry (npm has a package
`core`) could mean either. smith never picks for you:

- `workspace:core` always means the project. This is the prefix pnpm and yarn use. It works
  with a folder or a package name too (`workspace:libs/core`).
- `--registry` always means the registry, for every name in the command.
- A plain `core` is checked: smith asks the registry (it waits 8 seconds and goes on if the
  registry does not answer). If the registry has no such package, smith links the project and
  prints one line that says so. If it has one, or it could not be asked, smith asks which
  you mean and explains both answers. Without a terminal, or with `--yes`, it refuses and
  prints both fixes.

### The version range

Node writes a range, and there is a real choice. The workspace's habit is followed when it
has one (every library link already uses the same range). When it has none, smith asks:
`*` (always the workspace copy, nothing to edit when the library changes) or the current
version `^1.2.0` (the workspace copy while it fits, the registry copy when it does not). Pass
`--range '*'` to decide in advance. Without a terminal smith writes `*` and says so. The
other languages have no range to choose.

### What smith refuses

| Case | Message names |
|---|---|
| a project linked to itself | the project and its folder |
| a library that is an app | the app, and that only a lib or a package can be depended upon |
| another language | both projects and both languages |
| a cycle | the cycle, such as `core → ui → core` |
| a name that is no project (`workspace:cor`) | the near names: `Did you mean core?` |

Remove the link with `smith uninstall -w <project> <library>`. If the project did not
depend on it, smith says so and changes nothing. A registry package with the name of a
project is removed with `smith uninstall -w <project> --registry <name>`.

The TUI and the VS Code extension ask first: "A library of this workspace" or "A package
from the registry". The first lists only the libraries that are valid for the project, with
their kind and language, so nothing is typed.

## Common errors

| Message | Cause and remedy |
|---|---|
| `adding a package needs a target project: smith install -w <project> <package>. A bare smith install (no package) installs the whole workspace` | Add `-w <project>`. |
| `no project "nope": name it by its name or its directory (projects: api, tools, ui, web, web2)` | Check the name with `smith projects`. |
| `core is a project of this workspace (libs/core) and a package on the registry. Say which one you mean: ...` | Use `workspace:core` for the project or `--registry core` for the package. |
| `web is an app (apps/web), and an app cannot be depended upon.` | Link a lib or a package instead. |
| `the jvm plugin does not link an internal library yet: add the dependency in libs/x/pom.xml or build.gradle by hand` | That language has no `link` capability yet. |

## Related

[`smith uninstall`](uninstall.md), [`smith up`](up.md), [`smith sync`](sync.md),
[`smith ci setup`](ci.md#setup) (what CI runs instead).
