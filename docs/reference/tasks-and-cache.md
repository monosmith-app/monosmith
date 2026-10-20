# Tasks, cache and affected

> Terms such as project, kind, task and stack are defined in the [glossary](glossary.md). How it works inside: [execution and cache](../design/execution-and-cache.md). Why a task missed the cache: `smith why-miss`, and the [troubleshooting guide](../guides/troubleshooting.md).

## Tasks

A task is a command a plugin infers for a project: `build`, `test`, `lint`,
`typecheck`, `start` and any npm script. Each has:

| Field | Meaning |
|---|---|
| `command` | argv, run without a shell, from the project root; `{workspaceRoot}`, `{projectRoot}` and `{smith}` (the running smith binary) are expanded |
| `dependsOn` | `build` — this project's build first; `^build` — the build of every project this one imports first (reaching through projects without a build); `web:build` — a specific task |
| `inputs` | globs relative to the project, `{workspaceRoot}/...` for root files, `!` to exclude; default: every file of the project |
| `outputs` | directories restored on a cache hit (`dist`) |
| `cache` | whether results are cached |

```sh
smith run <task>... [-p <projects>] [--affected] [--parallel N] [--bail] [--no-cache] [--json]
smith build | test | lint | typecheck [project...]
```

Project selection (`-p`, comma-separated or repeated): a name, a glob over names
or directories (`apps/*`), `kind:package`, `lang:go`, `tag:x`, and `!` to
exclude.

Every task runs with the workspace's Python environment (`.venv`) first on
`PATH`, the way npm puts `node_modules/.bin` first, so `python` and the tools
installed in the venv are the workspace's on every OS.

Tasks run in parallel (one per CPU by default) in dependency order. A failed task
skips the tasks that depend on it; independent tasks keep going unless `--bail`.
Ctrl+C stops every process tree smith started — on Windows through a Job Object,
elsewhere through a process group.

### deps-check

A Node package (`kind = package`) gets a `deps-check` task, and its `lint` (or,
without one, its `build` or `test`) runs after it, so CI's verify phase checks
it. It fails when a file imports a package that the package.json does not
declare where installing the package from npm installs it, which the
workspace hides because npm hoists every dependency to the root:

- a file that ships may import what `dependencies`, `peerDependencies`
  (optional peers in `peerDependenciesMeta` too) and `optionalDependencies`
  declare
- a test, e2e suite, test double, tool or config file (`*.test.*`,
  `*.spec.*`, `*.config.*`, and anything under `e2e/`, `tools/`, `test/`,
  `tests/`, `__tests__/` or `testing/`) may also import `devDependencies` and
  what the workspace root declares
- Node's built-ins, the package's own name, `#` subpath imports and, in a VS
  Code extension, `vscode` are always fine

Versions are not compared, and a declaration nothing imports is not reported.
`smith doctor` reports the same, per package. A project that defines its own
`deps-check` script keeps it.

## Builds go to dist/

Every build writes to one directory at the workspace root, `dist/`, in a path
that mirrors the project's: `apps/web` builds into `dist/apps/web/`. Nothing is
written into a source tree, and deleting `dist/` cleans everything.
`dist/drop/` holds what `smith ci pack` keeps, and `dist/release/` holds release
assets. A tool that builds the whole workspace into one directory gets its own
folder there: Cargo builds into `dist/cargo` (`target-dir` in
`.cargo/config.toml`, written with the first crate).

| Language | Builds to |
|---|---|
| Go | `dist/<project dir>/`: the app's binary (`go build -o`) |
| Python | `dist/<project dir>/`: a package's wheel and sdist (`python -m build --outdir`) |
| Node (TypeScript) | `dist/<project dir>/`: an app's or a package's compiled JavaScript (`tsc`, `outDir`); a Vite, Angular or Astro app's site |
| Dart, Flutter | `dist/<project dir>/`: a Flutter web app, or a Dart app's `dart build cli` output |
| C, C++ | `dist/<project dir>/`: the CMake build (Ninja) |
| Bicep | `dist/<project dir>/template` and `parameters` |
| Rust | `dist/cargo` for the whole workspace |
| .NET | `dist/dotnet` (`bin/` and `obj/`) for the whole workspace |
| JVM | `dist/maven` or `dist/gradle` for the whole workspace |

TypeScript libraries don't build at all. A library's `package.json` exports
its sources (`"exports": "./src/index.ts"`). TypeScript reads them, and so
does Node 24 when an app runs, stripping the types. That's why relative imports
name the file (`./greeting/index.ts`): `tsc` rewrites them to `.js` when it
emits, and `erasableSyntaxOnly` keeps the code to what Node can strip. A
package is read from its sources inside the workspace too. It is published from
`dist/packages/<name>/`, where smith stages a copy of its `package.json` whose
`exports`, `main` and `types` point at the built files, along with its README
and LICENSE.

`smith doctor` enforces this. Any task output outside `dist/` fails, naming
the task and where its outputs were set (a plugin or `smith.toml`). For Node, a
`tsconfig` whose `outDir` is in the source tree fails too.

## Serving an app for a task

An end-to-end suite needs the app it tests running first and stopped after.
Any task can say so, whatever tool it runs (Playwright's runner, Jest or Vitest
driving `playwright`, Cypress, `pytest-playwright`):

```toml
[projects.web-e2e.tasks.e2e]
serve         = "web:start"               # a task started before the suite, stopped after
serve_url     = "http://127.0.0.1:5173"   # waited for (HTTP 2xx) before the suite starts
serve_timeout = "60s"                     # how long to wait; default 60s
```

`serve` is a task of the same project (`start`) or of another (`web:start`).
smith first runs what that task depends on (`web:start` depending on `build`
makes `build` a dependency of the suite), then starts it in the background with
its output prefixed `web:start │`, polls `serve_url` every 200 ms until it
answers with a 2xx, and runs the suite with `E2E_BASE_URL` set to that URL.
When the suite ends, passed or failed or cancelled, smith stops the app and its
whole process tree.

The suite does not run when the app exits before it answers or does not answer
within `serve_timeout`: the task fails and says why. Without `serve_url` the
suite starts as soon as the app has been started.

The served app is part of the task's cache key: the `serve` target and URL, and
the sources of the app's project and everything it depends on, so a cached
passing suite is a miss when the app changed. `smith why-miss` shows a changed
target as `serve`. `smith doctor` reports a `serve` that names no task,
`serve_url` or `serve_timeout` without `serve`, a `serve_url` that is not
http(s), and a `serve_timeout` that is not a duration.

## Cache

A task's **key** is a SHA-256 over a versioned, canonical encoding of:

- the task id and its command (before `{workspaceRoot}` and `{smith}` are expanded, so keys do
  not depend on where the repository is checked out)
- the declared env
- the content of every input file
- the sources of every project it depends on, transitively
- the keys of the tasks it depends on
- the OS/architecture, the smith version, and the version of every plugin
  that shapes the project (its language plugin, its stack's plugins, its
  conventions)
- the version of the compiler or runtime the task runs with (`node --version`,
  `go version`, `python --version`, `rustc --version`, `dotnet --version`,
  `java -version`, `php`, `ruby`, `dart`, `cmake`), which the language plugin
  reports as part of the project and probes once per run. A `build` cached on
  Node 20 is a miss on Node 22. A plugin that cannot tell leaves it out, and the
  key is then what it was before

On a hit, smith replaces the task's output directories with the cached files and
replays its log. On a miss it clears the outputs, runs the task, and stores the
result if it succeeded.

The store is a directory (`.smith/cache/v1`) split like Bazel's remote cache:
`ac/` holds one small JSON entry per key; `cas/` holds blobs named by their
SHA-256, shared between entries. Writes land by rename, so an interrupted run
never leaves a half-written entry; an entry with a missing blob is a miss. When
the cache grows past `max_size_mb`, the least recently used entries go first.

Each entry also keeps what its key was computed from (the files and their
digests, the command, env, upstream sources and dependency keys), and `last/`
remembers, per task, the key of its most recent cached run. That is what
`smith why-miss` compares with:

```sh
$ smith why-miss api:build
api:build would run: 2 change(s) since its last cached run (2026-10-16 09:12)
  upstream   core changed
  dependency core:build changed   (smith why-miss core:build)
```

It names added, removed and changed input files, a changed command, env
variables (names only), the projects upstream whose sources changed, the
dependency tasks whose keys changed, a new smith or plugin version, and a new
toolchain (`toolchain  node 20.11.0 → 22.3.0`). It runs nothing; `--json` prints the same report.

Everything goes through a `Storage` interface (`Has/Get/Put/Delete/Walk/Touch`)
with a conformance test. It stores three kinds (`ac`, `cas`, `last`); the
remote backends below implement the same interface.

### Remote cache

A remote store behind the local cache shares results between machines. CI
builds once and every developer's next run replays it:

```toml
[cache]
remote = "s3://acme-smith-cache/main"   # or https://cache.acme.dev/smith, azblob://acme/smith
remote_read_only = false               # true on developer machines, if only CI should write
```

Every object is signed with HMAC-SHA256 under `SMITH_CACHE_KEY`, a secret the
team shares (a CI secret, a developer's environment), and the signature covers
the object's name too. A local miss asks the remote, and a hit is verified, and
for blobs digest-checked, before it is kept locally and used. An object that
fails verification is a miss, said once. So is a remote that cannot be
reached, and a remote that cannot be written is only a warning. A remote cache
can make a run slower, never wrong. Without `SMITH_CACHE_KEY` the remote is
off: nothing is shared unsigned.

| Store | `remote` | Credentials |
|---|---|---|
| any server taking PUT, GET and HEAD | `https://host/path` | `SMITH_CACHE_TOKEN`, sent as a bearer token, when set |
| S3 | `s3://bucket/prefix` | `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN`, `AWS_REGION`; `AWS_ENDPOINT_URL_S3` for an S3-compatible store |
| Azure Blob | `azblob://account/container/prefix` | `AZURE_STORAGE_SAS_TOKEN` |

The generated pipelines pass the key and the backend's credentials to the steps
that run tasks, and keep the cache read-only in pull requests; see
[The remote cache](ci.md#the-remote-cache) in CI pipelines, with a worked Azure
Blob example.

`SMITH_CACHE_REMOTE` and `SMITH_CACHE_READ_ONLY=1` override the settings for
one machine. The remote holds results only: which run was last (for
`smith why-miss`) and pruning stay local.

## Affected

```sh
smith affected [--base main]
smith test --affected
```

The affected projects are those owning a changed file (since the merge base
with the base branch, plus uncommitted and untracked files) and every project
that depends on them. The base is `--base`, else `SMITH_BASE`, else the pull
request target (`GITHUB_BASE_REF`, `SYSTEM_PULLREQUEST_TARGETBRANCH`,
`CI_MERGE_REQUEST_TARGET_BRANCH_NAME`), else the release branch.

When smith cannot be sure it affects **everything**: no base, an unknown base, no
common ancestor (a shallow clone), or a changed workspace-level file such as
`smith.toml` or a lockfile. Documentation (`*.md`, `docs/**`) never affects a
project. Too wide costs minutes; too narrow reports green for code nothing
checked.

### Shared schemas

A JVM service that exports an OpenAPI spec, with a Go client and a TypeScript
frontend generated from it, shares no import with them, so the language
plugins report no edge. smith links them through the schema. A project owns
the OpenAPI (`openapi.yaml`, `swagger.json`), AsyncAPI (`asyncapi.yaml`),
Protocol Buffers (`*.proto`) and GraphQL (`*.graphql`, `*.gql`) files inside
its directory, and a consumer names what it consumes:

```toml
[projects.orders-go]
consumes_schemas = ["services/orders/openapi.yaml"]

[projects.storefront]
consumes_schemas = ["services/*/openapi.yaml", "contracts/*.proto"]

[projects.orders]
schemas = ["contracts/*.proto"]     # owned, though outside its directory
```

A consumer depends on the owner of every schema it consumes, so changing the
spec marks the owner and its clients affected (`smith affected`,
`smith run --affected`), orders the owner's tasks before theirs, and shows in
`smith graph`. A schema outside every project's directory still names its
owner instead of affecting everything. The globs are workspace-relative; see
[configuration](configuration.md#shared-schemas).

`smith graph --blast-radius <project|file>` lists everything a change reaches,
across languages, and whether each step is an import or a schema:

```text
$ smith graph --blast-radius services/orders/openapi.yaml
services/orders/openapi.yaml starts at orders
  orders-go            via schema of orders
  storefront           via schema of orders
```

### Watch paths for deployment platforms

```sh
smith affected --paths      # every app, a tab, the directories to watch
smith affected --json       # the same under "watch", with the affected projects
```

A deployment platform that rebuilds on a push needs to know which directories
can change an app: the app's own and those of every project it depends on,
transitively. For each app:

```
web	apps/web libs/core libs/ui
api	apps/api libs/core
```

Both list every app, affected or not, and ignore `--base`.

## Paths

Inside smith every path is POSIX and relative: project roots, task `cwd`,
inputs and outputs, cache entries and keys, graph and JSON output. Paths keep
the case git reports for them. Whatever a plugin or `smith.toml` gives is put
in that form when the workspace is read (backslashes become slashes, `.` and
`..` segments are resolved), and an absolute path or one that climbs out of
the workspace is an error.

On disk smith joins them to the absolute workspace root, so files nested past
Windows' 260-character limit still save to and restore from the cache: Go
adds the `\?\` long-path prefix to absolute paths. Git itself needs
`git config core.longpaths true` to check such files out.
