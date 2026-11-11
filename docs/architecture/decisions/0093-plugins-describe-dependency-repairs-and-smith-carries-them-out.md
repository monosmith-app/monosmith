# ADR 0093: Plugins describe dependency repairs, and smith carries them out

- Status: accepted; implemented for Node, Python, Go, Rust, .NET, Dart, Ruby, PHP, Swift and Maven
- Date: 2026-11-11

## Context

Installed dependencies break. A `node_modules` folder keeps a package no manifest declares, a
half-finished install leaves a tree npm cannot read, a corrupt download sits in a cache. Before,
smith had no answer: people deleted `node_modules` and `package-lock.json` by hand. Every
language has the same problem with other folders and other commands (`.venv`, `vendor`,
`.dart_tool`, the Go module cache, the NuGet caches).

The owner asked for all four usual repairs, for every language, and on the Deps tab.

## Decision

1. **Four repair kinds**, lightest first. The person picks one each run. smith never picks.

   | Kind | Does | Versions |
   |---|---|---|
   | `prune` | removes the installed packages that no manifest declares | stay |
   | `reinstall` | deletes the installed dependencies, then installs from the lockfile | stay |
   | `lockfile` | deletes the installed dependencies and the lockfile, then installs | can change |
   | `cache` | clears the tool's download cache, then installs | stay |

2. **A plugin capability, `dependencies/repair`** (`Capabilities.Repair`). The host sends the
   projects. The plugin answers one `RepairPlan` for each kind its tool has an honest equivalent
   for: the workspace-relative paths to delete, the commands to run, whether to install after,
   whether the change is machine-wide, and whether versions can change. A kind with no
   equivalent is left out. A plan for a lockfile that does not exist is left out. The plugin
   only describes. It never deletes. Third-party plugins get the capability through the Go,
   TypeScript, Python and C# SDKs (`DependencyRepairer`).

3. **smith carries the plan out**, in one order per tool: delete, run, install
   (`dependencies/install`). It stops at the first step that fails and names it.

4. **smith checks every path twice**, when the plan is read and just before the deletion. It
   refuses an absolute path, a `..` step, `.git`, `.smith`, `smith.toml`, the root of a project
   or a folder that holds one, and a path below a link. The plugin conformance check applies the
   same shape rule.

5. **A link is deleted as a link.** The deletion reads each entry without following it
   (`os.Lstat`, the folder's own entry types). A symlink or a Windows junction is removed as one
   entry. smith never walks into it. npm links each workspace package into `node_modules` (a
   junction on Windows), so following a link would delete source. A test makes a real junction
   (Windows) or symlink and checks that its target keeps its files.

6. **The command is `smith deps repair`** (`--kind`, `-p`, `--dry-run`, `--yes`, `--json`). It
   always prints the plan first and asks. Without a terminal it refuses unless `--kind` and
   `--yes` are given. A machine-wide kind says that other workspaces on the machine use the
   cache. A kind that can change versions says so. It never commits.

7. **The TUI** opens the same walk from the Deps tab (`H`, a button, the palette): the kinds,
   each explained, none marked as suggested, a kind no tool has disabled with the reason; then
   the plan; then the progress per step. It reuses the fix wizard, so `Esc` goes back one screen.

## Node

npm installs a workspace as one tree, so the root `node_modules` is part of each plan that
deletes, also with `-p`.

| Kind | Deletes | Runs |
|---|---|---|
| `prune` | nothing | `npm prune --no-audit --no-fund` |
| `reinstall` | every `node_modules` that exists (the root, then each project's) | `npm install` (smith's install) |
| `lockfile` | the same, and `package-lock.json` (only when it exists) | `npm install` |
| `cache` | nothing in the workspace | `npm cache clean --force`, then `npm install`. Machine-wide |

smith's install is `npm install`, not `npm ci`. With an intact lockfile both install the
locked versions.

## The other languages

Each language offers only the kinds its tool has an honest equivalent for.

| Language | prune | reinstall | lockfile | cache |
|---|---|---|---|---|
| Python (pip) | no command | delete `.venv`, install (no lockfile, versions can change) | no lockfile | `pip cache purge` |
| Go | `go mod tidy` | nothing installed in the workspace | delete `go.sum`, `go mod tidy` | `go clean -modcache` |
| Rust | nothing installed | nothing installed | delete `Cargo.lock`, `cargo fetch` | no stable command |
| .NET | nothing to prune | `dotnet restore --force` | delete `packages.lock.json`, `--force-evaluate` | `dotnet nuget locals all --clear` |
| Dart | no command | delete `.dart_tool`, `pub get` | and `pubspec.lock` | `pub cache clean --force` |
| Ruby | `bundle clean` is unsafe | `bundle install --redownload` | delete `Gemfile.lock` | no command |
| PHP | install already prunes | delete `vendor` | and `composer.lock` | `composer clear-cache` |
| Swift | no command | `swift package reset` + `resolve` | delete `Package.resolved` | `swift package purge-cache` |
| Maven | nothing installed | nothing installed | no lockfile | `dependency:purge-local-repository` |

Gradle, C and C++ and Lua offer no repair. Not verified by a real run: the Dart, Swift, Ruby,
PHP and Maven commands (taken from each tool's documentation), and `dotnet restore
--force-evaluate` on a workspace with lockfiles.

## Consequences

- A new language gets the four repairs by implementing one method. smith's checks and the
  link-safe deletion apply to every plugin, also a third-party one.
- A plugin cannot run a repair smith did not show. The plan the person confirms is the plan that
  runs (the host reads it again before it runs, and refuses a path that changed into a link).
- The repair of a single project still deletes shared folders when the tool installs the
  workspace as one (npm). The plan says so.
- `smith deps repair` is not an MCP tool or a `smith serve` method yet.
