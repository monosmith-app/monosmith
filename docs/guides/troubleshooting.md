# Troubleshooting

Find your symptom, run the command, apply the remedy. Two commands solve most
problems, so start with them:

```sh
smith doctor      # every failed check names its remedy
smith why-miss <project>:<task>   # why a task ran instead of coming from the cache
```

If this page does not help, `smith explain <project>:<task>` shows how a task
resolved and where each value came from, and `smith info` shows the smith
version and workspace in use.

## smith says it cannot find the workspace

```text
✖ this folder is not a smith workspace yet (there is no smith.toml); it already holds a repository
  → smith adopt --dry-run shows what bringing it under smith involves
```

The advice depends on the folder: an Nx or Turborepo workspace is sent to
`smith convert --dry-run`, any other existing repository to
`smith adopt --dry-run`, and an empty folder to `smith new <name>`.

smith looks for a `smith.toml` in the current folder, then in each parent.
Change into the workspace, or set `SMITH_WORKSPACE_ROOT` to its path. Commands
that make or inspect a workspace (`new`, `convert`, `adopt`, `help`) work
outside one.

## A task ran again and I expected the cache

Ask:

```sh
smith why-miss web:test
```

```text
web:test would run: 1 change(s) since its last cached run (2026-10-19 07:21)
  upstream   ui changed
```

The answer names the cause. What each label means:

| Label | Cause | What to do |
|---|---|---|
| an added, removed or changed file | an input changed | expected. If the file should not matter, narrow the task's `inputs` in `smith.toml` ([tasks](../reference/tasks-and-cache.md#tasks)) |
| `upstream` | a project this one imports changed | expected |
| `dependency` | a task this one waits for has a new key | run `smith why-miss` on that task |
| `command` | the command changed, for example after a plugin or stack change | expected |
| `env` | a declared environment variable changed (names only) | expected |
| a new smith or plugin version | the key includes both | expected after an upgrade |
| `toolchain  node 20.11.0 → 22.3.0` | the compiler or runtime changed | expected; pin the version in CI and locally |
| `serve` | the app a task serves changed | expected |

Surprising misses usually come from three places:

- **A task rewrites its own inputs.** A build that writes a file inside the
  project, or a formatter that rewrites sources, changes the key every run.
  `smith doctor` fails with "tracked files are not generated files" and "every build
  writes to dist/" for the two common forms.
- **A task reads files its `inputs` do not list.** Then the cache serves a stale
  result, not a miss. Add the files to `inputs`.
- **The cache is shared with another machine that has other tool versions.**
  Compiler versions are part of the key, so a cache from CI on Node 24 does not
  answer a laptop on Node 22. Align the versions (`smith doctor` warns when
  Node or Python differ from CI).

To skip the cache once, use `smith run <task> --no-cache`. To remove it, delete
`.smith/cache`; or `[cache] max_size_mb` (default 2048) trims the least recently
used entries first.

## A task is wrong, but it passes from the cache

Run with `--no-cache` to see whether the task really passes. If it fails
then, the key is missing an input: list it under `inputs` for that task (see
the previous section).

## Remote cache does nothing

A remote cache is off unless `SMITH_CACHE_KEY` is set, because smith signs
every object and never shares anything unsigned. `smith doctor` says so, with
the remedy: export `SMITH_CACHE_KEY` with the value your team shares, plus the
backend's credentials (see [the remote cache](../reference/tasks-and-cache.md#remote-cache)),
or leave it unset to work with the local cache. A remote that cannot be reached
is a plain miss, never an error. `SMITH_CACHE_READ_ONLY=1` makes a machine
read the shared cache but never write it.

## What `smith doctor` reports, and what to do

Every failed check prints a `→` line with the remedy, one per way out. Checks
that pass are summarised: a check that repeats for every project is one line
("npm workspaces include each project (23 projects)"); `smith doctor --verbose`
lists them all. The checks, grouped by who owns them:

### Workspace checks (owned by smith)

| Check | Meaning and remedy |
|---|---|
| the cache directory is ignored by git | `.smith/` would be committed. Add `.smith/` to `.gitignore` (`smith doctor --fix` does it) |
| the repository has commits | releases and `--affected` need history. Make the first commit |
| the clone has full history | a shallow clone cannot compute versions or affected projects. `git fetch --unshallow`; in CI, check out with `fetch-depth: 0` |
| tracked files are not generated files | git tracks a file that .gitignore ignores or a task writes (a build output, a cache). Stop tracking generated files (`git rm --cached <file>`), or commit the change on purpose |
| every build writes to dist/ | a task's output is outside `dist/`. Point the tool there (tsconfig `outDir`, `--outdir`, `-o`) and set the task's `outputs` in `smith.toml` |
| no two tasks write the same output directory | give each task its own directory (`dist/drop/<project dir>/` for a package task) |
| every task that serves an app can start it | `serve` must name an existing task, `serve_url` must be an `http(s)` URL, `serve_timeout` a duration like `60s`, and both only with `serve` |
| projects do not import each other in a cycle | smith cannot order a cycle. Move what both need into a project neither imports back, or merge the two. The finding prints the path (`ui → core → ui`) |
| CI pipelines match the workspace | the generated pipeline is stale. Run `smith ci generate` and commit it (`smith doctor --fix` does the same) |
| release tags resolve under each project's tag prefix | a project's release tags use another name, so its next release would start from scratch. `smith adopt --tags` creates them |
| release tags are not behind the registry | a version is already published beyond the latest tag. See [releasing](../reference/releasing.md) |
| `<project> can be verified` (a warning) | the project has no `lint` or no `test` task, so CI never checks it. Add the task, or declare it under `[projects.<name>.tasks]` |
| `<project>:<task> keeps the dependencies its plugins add` (a warning) | a `depends_on` override in `smith.toml` replaces the plugin's list, dropping dependencies it did not repeat. Add them back, or leave them out on purpose (`smith explain <project>:<task>` shows where each field came from) |

### Language checks (owned by each language plugin)

| Language | Checks |
|---|---|
| Go | `go.work` exists and lists exactly the workspace's Go modules (`go work use ./libs/x`, or `smith doctor --fix`); Go is installed and at least the version `go.work` names; `golangci-lint` is installed (`smith ci setup` installs it) |
| Node | Node.js matches CI's major version (a warning); dependencies are installed (`npm install` at the root); every project is in the root `package.json`'s `workspaces` (add `"libs/*"`, or leave a standalone project out with `exclude`); a package imports only what its `package.json` declares (the `deps-check` task); the scripts of the root and of each project call files and tools that exist (a `node ./scripts/x.mjs` whose file is gone, a tool such as `oxfmt` that is not installed), and the packages the ESLint config imports are installed; `tsconfig` `outDir` points into `dist/` |
| Python | the `.venv` exists (`smith ci setup`); `requirements-dev.txt` exists; the venv's Python matches CI (delete `.venv` and run `smith ci setup`) |
| Rust | `cargo` is installed; the root `Cargo.toml` is a workspace and lists every crate in `members` |
| .NET | the .NET SDK is installed; `nuget.config` matches `[registries]` (`smith doctor --fix`) |
| Ruby, PHP | the interpreter and Bundler or Composer are installed |
| C and C++ | CMake, Ninja, a compiler and cppcheck are installed; `VCPKG_ROOT` is set when a `vcpkg.json` exists |
| Dart, Flutter | the SDK is installed, and lives outside the workspace |
| JVM | a JDK, Maven, Gradle or the Android SDK, as the projects need |
| Bicep | the Azure CLI and the Bicep CLI |
| Docker convention | Docker is running; image descriptions fit the registry's length limit |

`smith doctor --fix` shows each fix first (diffs, the commands it runs, the
questions it asks) and applies the safe ones once you confirm. These have a fix
beyond the safe ones, which doctor asks about:

| Check | The fix |
|---|---|
| every build declares the directories it writes as outputs | declares them in `smith.toml` (safe), for all the projects at once |
| CI pipelines match the workspace | regenerates them, keeping your slot contents (safe when no slot would be lost) |
| tracked files are not generated files | `git rm --cached` on them; asks whether to add the output directory to `.gitignore` |
| `<project> can be verified` (no lint task) | adds a `lint` task running ESLint to `smith.toml`, for every Node project without one |
| `<project>` declares the packages it imports | adds each package to `package.json` at the range the workspace already uses (asks which when projects disagree), then `npm install` |
| the packages the ESLint config imports are installed | `npm install`, or `npm install -D` for a package nothing declares |
| the scripts of the root or a project call files and tools that exist | asks: install the missing tool, or remove the script that starts it |
| Node dependencies are installed | `npm install` |

A check with no fix (a file a script runs that is gone, a build that writes
outside `dist/`) needs a decision only you can make; its remedy says what.
`smith doctor --json` prints every finding for tools.

## Windows

- **Run smith from PowerShell, Git Bash or Windows Terminal.** Paths smith
  prints use forward slashes (`{workspaceRoot}/.venv/Scripts/python.exe`) on
  every system. That is deliberate and works on Windows.
- **Line endings.** `smith new` writes a `.gitattributes`, so git normalises line
  endings. The `CRLF will be replaced by LF` warnings git prints on Windows are
  harmless.
- **Long paths.** smith stores cache files with the `\\?\` prefix, so files past
  260 characters are cached correctly. Git itself needs `git config core.longpaths
  true` to check such files out.
- **Ctrl+C** stops every process smith started, through a Job Object.
- **SmartScreen** warns about an unknown publisher the first time a downloaded
  `smith.exe` runs. Choose More info, then Run anyway, or install with the script
  or a package manager. See [installing](installing.md).
- **Plugins written in Python or Node.** Windows cannot execute a `.py` file
  ("%1 is not a valid Win32 application"). Launch it with a `.cmd` file, and write
  that path with backslashes. See [writing a plugin](writing-a-plugin.md).
- **npm prints `EBADENGINE` and `ERESOLVE` warnings** while `smith add` installs
  a Node project. They come from npm (a dependency wants a newer Node patch
  version, or a peer range), not from smith. They do not fail the command.

## CI differs from my machine

- **Everything is affected in CI but not locally**, or the reverse. On a pull
  request smith compares with the pull request's target branch; locally it
  compares with `main`. A shallow clone cannot find the merge base, so smith
  treats everything as affected. Check out with `fetch-depth: 0` (the generated
  pipeline does).
- **The pipeline is out of date.** `smith ci generate` rewrites it; put your own
  steps between the `smith:slot` markers and they survive. `smith ci generate
  --check` fails on drift.
- **Secrets.** `smith ci check` lists the secrets to create and each private
  registry's permissions.
- **Toolchain versions.** The pipeline installs the versions the plugins pin
  (Node 24, Python 3.12, the Go version in `go.work`). `smith doctor` warns when
  yours differ, because different versions mean different cache keys.
- **Slow installs.** Generated pipelines cache each package manager's download
  store (npm, pip, Go modules, cargo, NuGet, Composer, ...), keyed on the
  lockfiles. See [CI pipelines](../reference/ci.md#package-download-stores).
- **Releases do not run on pull requests.** A pull request only comments a
  preview (`smith release --dry-run --comment`). Releases run on pushes to the
  release branch (`[release] branch`, default `main`).

## Private registries

Installs and publishes go to the public registries unless you name a private one.
Declare it once and point installs and releases at it:

```toml
[registries.acme-libs]
npm = "https://pkgs.dev.azure.com/acme/_packaging/libs/npm/registry/"

[workspace]
registry = "acme-libs"      # installs resolve from it
```

smith writes the matching `.npmrc` (and `nuget.config`) between `# smith:registry`
markers and keeps the rest of the file yours. `smith doctor` fails when the file
and `smith.toml` disagree, and `smith doctor --fix` repairs it. Tokens come from
environment variables you name, and `smith ci check` lists what to create as
secrets. The full format is in
[configuration](../reference/configuration.md#private-registries).

## A release did not do what I expected

```sh
smith release --dry-run
```

shows every package, its new version, why, and its status. Packages with no
`feat`, `fix` or breaking commits since their last tag do not release. Only
`package` projects release to registries; libs are released inside the packages
that use them. A release refuses to run off the release branch. A release that
failed half way resumes when you run it again.

## An editor or tool cannot read smith's output

Use the `--json` flag. Output meant for people changes between versions; the
JSON contracts are stable. See [editors](editors.md).

## Still stuck

Open an issue at
[github.com/monosmith-app/monosmith/issues](https://github.com/monosmith-app/monosmith/issues)
with the output of `smith info` and `smith doctor`.
