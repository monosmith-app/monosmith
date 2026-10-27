# smith kind

`smith kind` makes a project another kind after it was created: an **app** (deployed), an
internal **lib** (shipped inside what uses it) or a published **package** (installed by
others from a registry). A project often starts internal and one day becomes the thing others
install, or a published package is pulled back in. That is a line in the manifest, a folder
move and a dozen references, and smith does all of it.

## When to use it

- A core library gets a public SDK on top, and the core becomes a package.
- A package is no longer published and goes back inside the projects that use it.
- An app turns into a library, or the other way round.

To see what a change involves without making it, add `--dry-run`: it lists every file that
would be edited and the move, and changes nothing.

## Synopsis

```sh
smith kind <project> <app|lib|package> [--dry-run] [--force] [--json]
```

## Arguments and flags

<!-- smith:begin kind -->
Make a project an app, an internal lib or a published package

Usage: `smith kind <project> <kind> [flags]`

| Argument | Required | Meaning |
|---|---|---|
| `project` | yes |  |
| `kind` | yes |  |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--dry-run` | boolean | off | list every file edit and the move; change nothing |
| `--force` | boolean | off | go ahead although other projects import it or it was already released |
| `--json` | boolean | off | print the plan and what was done as JSON |

<!-- smith:end kind -->

- `<project>` is a project name, as [`smith projects`](projects.md) lists it.
- `<kind>` is `app`, `lib` or `package`.
- `--dry-run` prints the plan and changes nothing.
- `--force` goes ahead against the reasons listed under *When it refuses*. It never overrides a
  folder that is in the way or a kind that is already set.
- `--json` prints the plan and what was done: the edits, the move, the warnings and the files
  still naming the old folder.

## What it changes

The project's language plugin says which files change; smith applies them, moves the folder and
regenerates the CI pipelines, which name projects by kind.

| Step | What happens |
|---|---|
| The kind | Recorded where the language keeps it: `[package.metadata.smith] kind` in `Cargo.toml`, `"smith": {"kind"}` in `package.json`, `// smith:kind` in `go.mod`, `[tool.smith] kind` in `pyproject.toml`. |
| The publish guard | A package is published, so the guard goes: `publish = false` (Rust) and `"private": true` (Node) are dropped. A lib or an app gets it back. |
| The folder | Moved to `libs/<name>`, `packages/<name>` or `apps/<name>` with `git mv`, so the history follows. A project that is not in one of those folders stays where it is, and only its kind changes. |
| References | Everything that names the folder follows it: workspace members and path dependencies (Rust), `workspaces`, `file:` dependencies, tsconfig paths and references, the lockfile (Node), `go.work`, `replace` lines and, because a Go module's path ends in its folder, the module path and every import of it (Go), uv workspace members, editable requirements and `uv.lock` (Python), and `smith.toml`. |
| CI | The pipelines are regenerated, so a package's registry credentials appear and a lib's go away. |

The project keeps its name: it is the folder's last part. The project's own relative
references (to a sibling project, to the shared tsconfig) are rewritten to reach the same
place from the new folder, and the build folder (`dist/<folder>`) follows the folder.

## When it refuses

smith names the reason and changes nothing. These can be overruled with `--force`:

- **An app that others import.** An app is run, not imported.
- **A package that published packages import.** Once it is not published, installing them would
  fail.
- **A package that was released.** The newest release tag is named: the versions on the
  registry stay there but would never be updated.

These cannot be overruled:

- The project is already that kind.
- The folder it would move to exists.
- The language plugin cannot change kinds yet (it says so, and you change the manifest and move
  the folder by hand), or says that kind makes no sense for the project.

## What it warns about

After the plan it lists what to check: the registry fields a package needs (a description and a
license), what an app has that a lib does not (a binary, a start script, a Dockerfile), release
settings in `smith.toml` that only apply to an app, internal libs a new package imports (they
ship inside it), the Go module path that changes, and the install to run so the toolchain
learns the new folder (`smith install` for Node and Python).

Files that still name the old folder and that smith does not know how to rewrite (a Dockerfile,
a script, a document) are listed too. Search and change them yourself.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The project was changed, or the dry run printed. |
| 1 | No such project, a refusal, or a step failed. |

## Files and environment

Edits manifests and workspace files, moves the project folder, and regenerates the CI files.
Run `git status` afterwards: nothing is committed. `smith release --dry-run` shows the project
appearing in the plan once it is a package (after a `feat:` or `fix:` commit since the change),
and leaving it when it is not.

## Examples

```sh
smith kind core package --dry-run
```

```text
Would make core a package (it was a lib)
  moves   libs/core -> packages/core (git mv)
  edits   Cargo.toml
  edits   apps/cli/Cargo.toml
  edits   libs/core/Cargo.toml
  the CI pipelines would be regenerated
Check:
  - crates.io asks for a description in libs/core/Cargo.toml before the first release; add it
  - the next smith release will publish core: check its registry credentials and the version it starts from
```

```sh
smith kind core package
smith kind core lib       # and back: the workspace is as it was
```

## Languages

Rust, Node, Go and Python change kinds. The other languages (.NET, JVM, Ruby, PHP, C++, Dart,
Bicep) answer that they cannot yet: edit the kind in the manifest and move the folder by hand,
then run [`smith doctor`](doctor.md).

## See also

- [`smith kinds`](kinds.md) lists the kinds of project `smith add` can create.
- [`smith add`](add.md) and [`smith remove`](remove.md).
- [Releasing](../reference/releasing.md): what a package is and what a lib is.
- Assistants and editors: the MCP tools `smith_change_kind_preview` and `smith_change_kind`
  ([`smith mcp`](mcp.md)) and the `smith serve` methods `kind.preview` and `kind`
  ([`smith serve`](serve.md)). Neither offers `--force`: a refusal is shown, and a person
  overrules it on the command line.
