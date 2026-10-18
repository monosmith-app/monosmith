# Getting started

## Prerequisites

- **smith**: see [installing](installing.md).
- **git**, always.
- The toolchain of each language you use: **Node 24** (with npm), **Go 1.26**,
  **Python 3.12+**, **Rust** (with clippy and rustfmt), the **.NET 10 SDK**,
  **Ruby 3.2+** with Bundler,
  **PHP 8.2+** with Composer,
  **CMake 3.25+** with Ninja, a C++ compiler and cppcheck,
  **Dart** (or Flutter, which carries it; installed outside the workspace),
  a **JDK 21** with **Maven** (or **Gradle 9** for Kotlin, and the
  **Android SDK** for Android),
  the **Azure CLI** with Bicep.
  smith tells you what is missing.

## Create a workspace

```sh
smith new acme
cd acme
```

This writes `smith.toml`, `.gitignore`, `.gitattributes`, a README and a CI
pipeline, initialises git on `main` and makes the first commit. In a clone that
already exists (with a LICENSE, say), run `smith new acme --dir .`: files smith
does not own are kept, and the commit holds only the files smith wrote, so
your other changes stay uncommitted. A workspace created in a folder below the
top of an existing repository is left uncommitted; commit it there yourself.

Options: `--scope @acme` (npm scope), `--module-prefix github.com/acme/acme`
(Go module prefix; by default it is derived from the `origin` remote) and
`--ci github,azure` (or `gitlab`).

## Or adopt an existing repository

A repository with npm workspaces, Go modules or Python projects, but no
`smith.toml`, can be adopted (an Nx workspace is `smith convert`'s):

```sh
smith adopt                 # read-only: what adopting involves
smith adopt --tags --init   # take the steps
```

The report lists the projects smith would see, then each blocker and warning
with the step that clears it. For example, a project whose release tags were
made under another name (`core@1.2.0` where smith expects `libs/core/v1.2.0`)
would otherwise start its next release from scratch. A pipeline of yours sits
at a path smith's uses. Root `package.json` runtime dependencies belong to the
projects that import them. `smith adopt` exits non-zero while something blocks.

Both steps need a clean git tree. `--tags` creates the missing tags locally, on
the same commits; pushing them is yours. `--init` writes `smith.toml` (the
name comes from the directory, the npm scope from the root `package.json`,
and the CI providers are the ones the repository already uses) and the root
files smith needs, keeping yours and adding `.smith/` to your `.gitignore`. It
keeps your pipeline as `<file>.pre-smith`, generates smith's for the projects
it found, and commits it all as `chore: adopt smith`.

## Add projects

```sh
smith add app web --lang node
smith add lib ui --lang node
smith add package sdk --lang node
```

| Kind | Directory | Release |
|---|---|---|
| `app` | `apps/<name>` | never published; deployed by your pipeline |
| `lib` | `libs/<name>` | never published; imported by other projects |
| `package` | `packages/<name>` | published when it has releasable commits |

`--lang` picks the language plugin: `node`, `go`, `python`, `rust`, `dotnet`,
`ruby`, `php`, `cpp`, `dart`, `jvm` or `bicep`. Once a workspace has only one
language, it is the default. `--stack` picks the plugins the project is built
with instead (`smith add app web --stack react`); `smith stacks` lists them,
and [stacks](../reference/stacks.md) explains them.

If your team keeps a [recipe](../reference/recipes.md) (`smith new acme
--recipe <path or URL>`, or `recipe` under `[workspace]` in `smith.toml`),
`smith add` creates the project from it instead: it asks the recipe's few
questions one at a time (a language, say), shows the plugins the answers pick,
and makes the project with them. `--choice language=python` answers without
asking, which is how scripts and CI add projects; `--stack` bypasses the recipe.
To write a recipe, run `smith recipe new` (it asks, and checks the result for
every answer before writing it); a team releases its recipes like packages
([versioned recipes](../reference/recipes.md#versioned-recipes)).

Each project starts with one worked example slice (`greeting`) laid out as
vertical slices, and passes lint, type-check, test and build immediately.

To take a project out again:

```sh
smith remove web --dry-run    # what would change
smith remove web              # asks first; --yes in scripts
```

It deletes the project's directory, lets its plugin clean what it owns
elsewhere (the `go.work` entry, the npm lockfile, the Python venv), drops the
project's tables from `smith.toml` and regenerates the pipelines. A project
another one imports is refused unless `--force`; its release tags are kept.

## Use one project from another

Just import it — no manifest to edit:

```ts
// apps/web/src/main.ts
import { greet } from '@acme/ui'
```

```go
// apps/api/main.go
import "example.com/acme/libs/core/greeting"
```

```python
# apps/worker/src/acme_worker/__main__.py
from acme_pyutil.greeting import greet
```

npm workspaces, `go.work` and the workspace's Python venv (editable installs)
make every project importable by every other; smith reads the imports to learn
the dependency graph:

```sh
smith graph                         # a tree in the terminal
smith graph --format tui            # browse it: what a project uses, what uses it
smith graph --focus web             # web, everything it uses and everything that uses it
smith graph --affected              # what changes since main affect, the changed ones marked
smith graph --format html -o graph.html   # a page to share; also json and dot
smith graph --diff v1.4.0           # how the graph changed since a revision
```

Every format shows the same projects; `--focus` and `--affected` narrow all
of them the same way. `--diff` discovers the graph at another revision in a
temporary git worktree, leaving yours untouched. It lists projects and imports
added (`+`) and removed (`-`), and kinds, stacks, plugins and directories that
changed (`~`), as text or `--format json`.

Projects that import each other in a cycle cannot be ordered. `smith graph`
warns with each cycle's path (`ui → core → ui`), and `smith doctor` fails with
the same paths and how to break them.

## Add dependencies

Each project owns its dependencies, so adding one names the project:

```sh
smith install -w web zod                    # npm install zod --workspace apps/web
smith install -w worker -D pytest-cov       # pyproject.toml's dev extra, into .venv
smith install -w api github.com/pkg/errors  # go get, in apps/api
smith uninstall -w web zod
```

`-w` takes a project's name or its directory, and can be repeated. `-D` adds a
development dependency: an npm devDependency, the Python `dev` extra, or a Go
tool (`go get -tool`). A bare `smith install` installs everything a fresh clone
needs: `npm install`, the Python venv with every project in it,
`go mod download`, and the same for the other languages that support it.
Without `-w`, a package is refused instead of being added at the root.

`smith install` and `smith uninstall` work for Node, Python, Go, Dart, PHP and
Ruby projects. `smith sync` compares Node and Python declarations, and
`smith up` checks Node, Python and Go. For the other languages, use their own
tool (`cargo add`, `dotnet add package`, …) in the project's directory.

When projects declare the same dependency at different versions, `smith sync`
makes them agree, language by language. The target is the version installed,
with the operator most declarations use, or else the highest version declared.
Sync then installs, so lockfiles and the venv follow. `smith sync --check`
only reports, and fails, which makes it a CI check. Peer dependencies, the
workspace's own packages and specs smith cannot safely rewrite (URLs, git refs,
compound ranges) are left alone, and the ones that disagree are reported. Go
is not compared: a `go.work` builds every module with the highest version any
of them requires.

`smith up` lists every declared dependency with a newer release, grouped as
patch, minor and major, with the projects that declare each one:

```sh
smith up            # pick in the terminal (patch and minor start selected)
smith up zod jest   # exactly these, no prompt
smith up --yes      # everything
smith up --check    # report only (the default without a terminal)
```

Each chosen declaration keeps its operator (`^`, `~`, `>=`…) and moves to the
latest version, then the workspace is installed again. Npm is asked through
`npm view`, so `.npmrc` applies. Python uses the index pip uses
(`PIP_INDEX_URL`, else PyPI). A final section compares the tools smith pins
(Node.js LTS, Python, the Go version in `go.work`, golangci-lint) with their
newest releases.

## Run tasks

```sh
smith build                     # every project
smith test -p web               # web and what it depends on
smith run lint typecheck test   # several tasks
smith run start -p web          # run an app
smith dev web api               # start several apps together; Ctrl+C stops them all
smith test --affected           # only what changed since main
```

`smith dev` builds what the apps use, then runs every `start` task at once,
each line prefixed with its project. With no names it lets you pick from the
apps that have a `start` task; `--all` starts every one of them.

A task whose inputs did not change replays from the cache. See
[tasks and cache](../reference/tasks-and-cache.md).

## Commit and release

Commit with [Conventional Commits](https://www.conventionalcommits.org):

```sh
git commit -m "feat(sdk): add retries"
smith release --dry-run
```

On `main`, CI runs `smith ci release`, which publishes every package with
releasable commits. See [releasing](../reference/releasing.md).

## Check the workspace

```sh
smith doctor          # every failed check names its remedy
smith doctor --fix    # show the fixes doctor can make as diffs, apply the safe ones
smith explain web     # how web's tasks resolved and where each value came from
```

`--fix` shows every fix as a unified diff before anything changes, then asks
for confirmation (`--yes` skips the question; without a terminal it only
shows). Only safe fixes are applied. A safe fix restores what smith or a plugin
owns without losing anything of yours: a `go.work` that lists exactly the
workspace's Go modules, `.smith/` in `.gitignore`, or pipelines regenerated with
your slot contents kept. A fix that could lose something is shown, never
applied.

## The interactive shell

`smith` with no arguments, in a terminal, opens a shell with these tabs (tab and
shift+tab switch between them):

- **Menu**: the common tasks, asking one question at a time.
- **Commands**: every command; type to find one, press enter, then type its
  arguments.
- **Doctor** (inside a workspace): every check, problems first. Press enter on
  one to see its detail and remedy, and r to check again.
- **Release** (inside a workspace): what the next release would do, or why it
  is blocked. Press p to preview it and R to release.
- **Settings** (inside a workspace): the settings of `smith.toml`, with their
  values and what they do. Enter edits one (or flips a yes/no setting), u
  returns it to its default. A value the schema rejects is refused on the
  spot. The change is made by `smith config set` or `unset`, which keep the
  file's comments.

The shell never acts by itself. Each action ends as a smith command line,
printed before it runs: you can type the same command in a script or in CI. R,
for example, hands over to `smith release`, which asks for confirmation and then
shows its progress. When stdin or stdout is not a terminal, or `CI` is set,
`smith` prints its help instead.

## Upgrade smith

After installing a newer smith, run it once in the workspace:

```sh
smith upgrade             # migrate smith.toml, regenerate the pipelines, update recipe projects
smith upgrade --dry-run   # show the plan and the change to smith.toml; change nothing
smith upgrade --check     # the same, and fail when something would change (CI)
```

`smith.toml` is migrated to the current schema with its comments and layout
kept. The files smith owns (the pipelines, which start with "Generated by
smith") are regenerated, with your steps between `smith:slot` markers kept, and
one the workspace no longer asks for (a provider dropped from `[ci] providers`)
is deleted. Files smith wrote once, like `README.md` and `.gitignore`, and your
own files are never touched. It also says what an older workspace may want,
for example a default stack for Node.

Projects made from a [recipe](../reference/recipes.md#keeping-projects-current)
are brought up to date with the recipe as it is now: the recipe version, the
stack and the settings recorded in `smith.toml`, and the files their plugins
own. Starter files you edited are left as they are. `smith doctor` warns about
a project that has drifted before you upgrade.
