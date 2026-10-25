# Research 02 — Nx subsystems to replace

Source: Nx 23.3.0 as installed in the TypeScript mnci repo. Hashing, cache I/O, globbing and touched-project
location are native Rust in Nx, so **Nx's hash values cannot and need not be reproduced**; we define our own.

## 1. Project graph

- Discover projects from workspace manifests (package.json workspaces globs, project.json) honouring ignore files.
- Merge precedence (later wins): plugins < `targetDefaults` < package.json < project.json. Per-target, per-key merge;
  arrays replaced unless spread; an *incompatible* target (different executor/command) REPLACES the base. This is
  the main source of Nx confusion — we replace it with one simple rule (see config epic).
- Edges: workspace package deps (static), import analysis resolved through tsconfig paths / package names,
  `implicitDependencies` (with `!name` removal). File→project is the longest-root match, POSIX-normalised.
- Drop: JS plugin execution, lockfile external nodes, graph cache/incremental recompute, isolated plugin worker.

## 2. Task graph

- Task id `project:target[:configuration]`. `dependsOn` forms: `^t` (direct deps' `t`, bridging through deps that
  lack `t` via dummy nodes), `t`, `proj:t`, and object form (`projects`, `dependencies`, params forwarding).
- Defaults: `cache=false`, `parallelism=true`, inputs `["default","^default"]`. Cycle detection required.
- Drop initially: continuous/batch tasks, `ultracache`. Add continuous tasks later for `dev`/`serve`.

## 3. Hashing

Inputs: filesets, named inputs (`^named` = dependency projects), `env`, `runtime` (hash command stdout),
`externalDependencies`, `dependentTasksOutputFiles` (deferred until upstream tasks ran), `json` field subsets.
Hash = sorted file hashes + env values + runtime outputs + external dep versions + dependency task hashes +
command + overrides + tool version. Only *declared* env affects the hash. CRLF/autocrlf and case-insensitive
paths matter on Windows.

## 4. Local cache

`.cache/<hash>/{outputs/<relpath>, terminalOutput, code}` plus a commit marker written last (rename into place).
Only `code===0` results are stored (optional failure caching). Hit: delete existing outputs, copy cached, replay
terminal output. Miss: run, capture output, store. Trim by size/LRU. Windows: long paths (`\\?\`), delete retries
(AV/watchers), no partial writes on Ctrl+C.

## 5. Affected

Base = `--base` | env | `defaultBase` | `main`; base becomes `git merge-base base head` (fallback fork-point, then
raw). Touched files: `git diff --name-only --no-renames --relative -z base head` (+ uncommitted + untracked when
head is default). File→project; non-project files via workspace implicit deps; deleted manifests affect all;
lockfile change affects dependents of changed externals (or all); then reverse closure over all edges.
Shallow clones need history; unresolvable base ⇒ verify everything.

## 6. Run-many / scheduling

Selection by name, glob, `tag:x`, negation. Parallelism default 3, `N`, `N%`. Priority: more reverse
dependents, more deps, longer duration. `parallelism:false` runs exclusively. On failure dependents are
`skipped`, independents continue; `--bail` stops starting. Env injected: `NX_TASK_*`-style task id/hash, PATH
gets `node_modules/.bin`. Windows: `taskkill /T /F` for trees, `.cmd` shims via PATHEXT, case-insensitive env,
cmd.exe vs sh quoting.

## 7. Release

- Tag pattern default independent `{projectName}@{version}` (fixed: `v{version}`). Current version from latest
  matching tag (fallback disk). Commits `git log <tag>..HEAD` (use NUL delimiters, not literal separators).
- Parse `type(scope)!: description`; breaking via `!` or `BREAKING CHANGE:`; reverts. Drop automated release
  commits. Map commits to projects by changed files through the graph (incl. dependents) and by scope.
- Bumps: feat→minor, fix→patch, breaking→major, others none; highest wins; optional 0.x adjustment.
- Dependents get range updates and a patch bump; topological order. Manifest edits must preserve formatting,
  key order, EOL and trailing newline (text edits, not re-serialisation).
- Order: version → changelog → `git add` (chunk long path lists) → commit (skip if nothing staged) → annotated tags
  → push `--follow-tags --atomic`. Publish: skip private, check already-published, tag from preid. Dry-run runs the
  same code path and writes nothing.

## 8. Sync

Only one generator matters: TypeScript project references mirroring graph deps, plus a `--check` mode. In CI,
check only.

## 9. Daemon

Drop. Build the graph per invocation; persist keyed by config hash + file stat if slow. Named-pipe daemons and
file watchers are a top source of Nx Windows bugs.

## Implementation priority

1. Config + discovery + merge. 2. Task graph. 3. Runner + scheduler + failure skipping. 4. Hashing + cache.
5. Affected. 6. Release with dry-run. 7. Sync check, changelog, TUI later.
