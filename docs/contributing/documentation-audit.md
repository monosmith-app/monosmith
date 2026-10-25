# Documentation audit, October 2026

In October 2026 every page of these docs was checked against smith 1.62.0: the code
(`apps/smith`, `libs/*`, `packages/plugin-sdk`), the installed binary (`smith <command>
--help`, `smith commands --json`, `smith schema`, `smith serve`) and runs in throwaway
workspaces. About 290 commits had landed since the previous pass (the single-key
terminal interface, the adopt and convert wizard with `--level conservative|full`, the
doctor fix wizard, `[ci] branches`, plugin changes, SBOM and provenance, prompts for
missing arguments). This page records what was found, so the next audit starts from it.

Verdicts:

- **accurate**: every claim checked, nothing changed.
- **fixed**: claims corrected, added or deleted; the page's shape stayed.
- **rewritten**: a large part of the page was replaced or the page was split.
- **removed**: the page or the thing it described no longer exists.
- **new**: written in this pass.

## Pages that existed before

| Page | Verdict | What changed |
|---|---|---|
| `getting-started/installing.md` | fixed | Moved from `guides/`. Channels, install script variables, checksums and attestations match `install.sh`, `install.ps1` and `smith.toml`; the decision reference now links the record (ADR 0019). |
| `getting-started/first-workspace.md` | rewritten | Was "Getting started" under `guides/`. The adopt section moved to its own page; the 125-line description of the terminal interface (out of date: tabs, keys, wizards) became a short summary that links the manual; added the `native` phase; added `graph --blast-radius`; stated correctly what `smith new --dir .` commits (a folder that is not a repository is committed whole, a clone only smith's files); replaced the claim that `smith graph` warns about a cycle (every command stops with the cycle, see "Defects found" below); captured the healthy `smith doctor` output that was missing. Screenshots re-captured. |
| `getting-started/converting.md` | fixed | Moved from `guides/`. The flag table lacked `--level`, `--lint`, `--test`, `--build`, `--ci`, `--skip-verify` and the `--allow-dirty` limit; added them and a link to the manual page. Corrected the adopt links and the issue references (#548 levels, #553 plain repositories). Nx 23 and Turborepo 2 as tested versions confirmed in `convertplan`. |
| `guides/editors.md` | fixed | Added the VS Code extension as the shipped integration. The `--json` field lists, the MCP read-only and writing tools and their inputs match `mcpserver` and `clicontract`. |
| `guides/faq.md` | fixed | Adopt link repointed; the Nx answer now covers Turborepo and `--level full`; added `smith dev --all` and a "do I have to remember the commands" entry pointing at the terminal interface. |
| `guides/troubleshooting.md` | fixed | Added the real "no smith workspace here" message and the commands that work outside a workspace; documented `doctor --fix --only/--choose` and the fix wizard; corrected the cycle check and the cache size remedy (see "Defects found"). |
| `guides/writing-a-plugin.md` | fixed | Replaced a real user path in sample output; timeouts, error codes, doctor fix steps and choices checked against `packages/plugin-sdk`. Restructuring is left to the plugin-manual work. |
| `reference/ci.md` | fixed | Added the `before-publish` slot of the `pub.yml` workflow. Branches, native, SBOM, provenance, audit policies and the caches checked. |
| `reference/cli.md` | rewritten | Now a generated quick reference of every command and flag, filled by `reference/cli-generate.mjs` from `smith commands --json`; the hand-written examples moved to the manual. The generator was stale: it exited 1 silently on drift, misreported variadic arguments and had wrong defaults. |
| `reference/configuration.md` | fixed | Every key of `smith schema` is documented. Added `SMITH_VERBOSE` and the task environment (`SMITH_TASK`, `SMITH_TASK_KEY`, `SMITH_WORKSPACE_ROOT`). Marked `[cache] max_size_mb` as not applied. |
| `reference/conventions.md` | accurate | Slice-check languages, roles and `smith slices check --lang` match the binary. |
| `reference/glossary.md` | fixed | Added depth level, fix, native project and SBOM. |
| `reference/plugins.md` | fixed | The `vscode` plugin's roles now include deploy; "a recipe later" became "a stack or a recipe"; pointer to `smith plugin enable/disable/set`. Protocol methods, capabilities, `plugin.json`, search path and `SMITH_HOME` verified against the SDK. Restructuring is left to the plugin-manual work. |
| `reference/recipes.md` | fixed | Epic #312 is closed, so "being built in steps" is now past tense. |
| `reference/releasing.md` | accurate | Bump rules, `--status`, `version(x)[y]` and the Entra ID section match the code. |
| `reference/serve.md` | fixed | Method list matches `initialize` of 1.62.0 exactly; `smith lsp` marked planned (#303), it does not exist. |
| `reference/stacks.md` | accurate | Built-in stacks match `smith stacks`. |
| `reference/tasks-and-cache.md` | fixed | Cache layout, `why-miss` format and `affected --paths` verified by running them; the pruning claim corrected. |
| `architecture/architecture.md` (was `design/`) | fixed | All 45 slice names checked; about 25 core slices and several libraries missing from the tables added; links to the new pages. |
| `architecture/execution-and-cache.md` | fixed | Corrected the store layout (`<aa>/` shards, `tmp/`), the key example, the affected section and the remote backends; pruning documented as unwired. |
| `architecture/plugin-system.md` | fixed | The handshake table listed six plugin kinds; `protocol.PluginKind` has two (`language`, `convention`). Limits, timeouts, `SMITH_PLUGIN=1` and the search path match the code. |
| `architecture/recipes-and-ci.md` | fixed | Added a status table: built (URL pinning in `smith.recipes.lock`, `[registries]`, one-version), not built (recipe workspace scope, a CI-provider role, deploy environments). |
| `architecture/release-and-ci-internals.md` | fixed | The build step applies only to projects without a release section; added the journal, the SBOM, the `before-publish` slot and the on-demand `convert` job. |
| `architecture/technology.md` | fixed | Decision links now point to records; corrected the S3 signing size, the libraries table and what `slices check` verifies for Go. |
| `architecture/tui-and-ide.md` | fixed | A design proposal. Added an "As built, October 2026" box: `smith lsp` and the JetBrains plugin are planned; the context menu, pane-border drag, Huh, Glamour and ntcharts were never built; the Run view parses text, not the event stream; the keymap table is a proposal. |
| `architecture/decisions.md` | rewritten | Split into the records under `architecture/decisions/` (the 20 entries kept as 0001 to 0020, 15 new records from the history). |
| `architecture/research/01` to `03` | accurate | Dated snapshots about mnci, Nx and prior art, not claims about the current tool. Left as they are. |
| `contributing/development.md` | rewritten | Layout now lists `apps/vscode`, `packages/eslint-config`, `libs/actions`, `libs/plugin-recipe`; about 30 slices missing from the table added; the doc-generation paragraph covers the manual and `SMITH`. |
| `assets/README.md` and the images | fixed | `doctor`, `doctor-failing`, `doctor-fix`, `explain`, `why-miss`, `release-preview` and `ci-yml` re-captured with `render-terminal.mjs` because their output changed; `run-cache-hit`, `graph-text` and `graph-html` still match. `doctor.png` was unlisted and unused; it now shows a healthy `smith doctor`. |
| `assets/render-terminal.mjs` | accurate | Used for the re-captures (needs `playwright` and `RENDER_CHANNEL=chrome` without a downloaded Chromium). |
| `examples/plugins/{go,python,node}` | accurate | All three pass `smith plugin check` on 1.62.0 (6 of 6 checks). |
| `apps/smith/README.md` | accurate | Paths and claims hold. |
| `apps/smith/tui/README.md` | fixed | Added the `adoptwizard` subfeature; corrected the file count; the design link follows the move. |
| `packages/plugin-sdk/README.md` | accurate | All nine packages exist; the `Describe` signature is correct. |
| `apps/vscode/README.md` | fixed | Publishing text named only `VSCE_PAT`; it now says `vsce_entra = true` signs in as an Entra ID app and `OVSX_PAT` adds Open VSX. Added a manual link. Settings and keybindings match `package.json`. |
| `apps/vscode/CHANGELOG.md` | accurate | Only an "Unreleased" section. |
| `docs/README.md` | rewritten | The docs home: who reads what, the sections, a task table. |
| Root `README.md` | fixed | The documentation section lists the new tree. |
| Code comments and messages naming a docs path | fixed | The "nothing to release" hint pointed at `docs/guides/getting-started.md`; `runview/doc.go` said there was no event stream; `nxworkspace/doc.go` said Nx was not run (it runs `nx graph --file`). |

## Pages written in this pass

| Pages | Verdict | Note |
|---|---|---|
| `getting-started/README.md`, `getting-started/adopting.md` | new | `adopting.md` was split out of the first-workspace guide and updated for the depth levels and the wizard. |
| `guides/README.md`, `reference/README.md`, `architecture/README.md`, `contributing/README.md` | new | Index pages; the folder tree is the page tree. |
| `manual/` (59 pages) | new | The index, one page per command, the subcommand pages (`<command>-<sub>.md`), the terminal interface and the VS Code extension. Flag tables are generated from `smith commands --json`; prose and examples were written by hand from the code and from real runs. |
| `architecture/` (13 explainer pages) | new | System context, discovery, project graph, task graph, cache, affected, slice architecture, CI generation, release, plugin protocol, adopt and convert, the terminal interface, the VS Code extension. |
| `architecture/decisions/` (36 pages) | new | The index and records 0001 to 0035. |

## Counts

The first table has 39 rows (a row can cover several files): 9 accurate, 25 fixed, 5 rewritten, 0 removed. One page, `decisions.md`, was replaced by the records; one image, `doctor.png`, was unlisted and unused and now shows a healthy `smith doctor`. New pages: 59 in the manual, 14 in the architecture manual (13 explainers and the index), 36 decision records, and 6 others (the getting-started and the guides, reference and contributing indexes, the adopting guide, and this page).

## Defects found in smith while checking (not fixed here)

| What | Evidence |
|---|---|
| `[cache] max_size_mb` is read but never applied: `taskcache.Prune` has no caller outside tests | `grep` for callers; `smith config list` and the schema still describe pruning |
| A project import cycle stops every command that opens the workspace, so `smith doctor`'s cycle check and the warning in `smith graph` cannot be reached | `session.Open` fails with `projects import each other in a cycle: ui → web → ui`; reproduced in a scratch workspace |
| `smith test --json` (and any `run --json`) prints task log lines before the JSON on standard output | observed; `--json-events` is clean |
| `smith plugin check <path> --json` prints `"detail": "<nil>"` for one check | observed |
| `smith doctor --fix --yes` regenerated pipelines with `go-version: '1.24'` where they had `'1.27'` when `go.work` was missing | observed |
| `smith ci generate` says "wrote .devcontainer/devcontainer.json" when it removes an unconfigured dev container | observed |
| `smith release --format markdown` prints the text plan line before the markdown | observed |
| `smith recipe check --all` printed the same as without `--all` | observed |
| `smith adopt --init` on an adopted workspace says "something blocks adoption; see smith adopt", and `smith adopt` then says "already a smith workspace" | observed |
| `SMITH_NONINTERACTIVE` is set by the terminal interface's runner but nothing reads it | `grep` |
| `libs/plugin-node-toolchains` generates a VS Code extension README that names only `VSCE_PAT` | `extension_templates_config.go:92` |
| `smith commands --json` lists the arguments of `why-miss` and `explain` as separate, but both take `project:task`; variadic arguments of `build`, `test`, `lint`, `typecheck`, `plugin enable/disable/set` and `slices check` are not marked variadic | compared with `--help`; the generator carries overrides |
| Seven slices exceed the 12-file review threshold with no written exception (`convertplan` 29, `release` 21, `cipipeline` 17, `projectgraph` 15, `workspace` 14, `doctor` 13, `plugin-sdk/protocol` 17); only `tui` has one | `smith slices check`; see [the slice architecture](../architecture/slice-architecture.md) |
| `libs/plugin-node-toolchains` imports `libs/plugin-node/scaffolding`, an undocumented dependency exception | `go list` |
| `~/.local/bin/smith.exe` (an old install script copy) can shadow the npm install on Windows | found while running the generator; set `SMITH` to the binary to use |

## What was not verified

- The terminal interface and the VS Code extension were not run: their manual pages and
  architecture pages are written from source (`apps/smith/tui`, `apps/vscode`) and the
  commands they stand for.
- Publishing paths (`smith release` to real registries, `smith docker description`,
  Marketplace and Open VSX) were read in code and not exercised.
- The Ruby, PHP, Dart, JVM and Bicep rows of the built-in plugin table in
  `reference/plugins.md` were checked against `smith plugin list` only.
- The 30 Mermaid diagrams were checked against the code by reading, and all parse with
  Mermaid 10.9.1 (two used the reserved word `graph` as a node name and were fixed);
  none was rendered to an image.
- `SMITH_NONINTERACTIVE`, an Nx conversion and `smith new` inside an existing repository
  were not exercised.

## Re-running the checks

```sh
node docs/reference/cli-generate.mjs --check    # the generated flag tables
```

Set `SMITH` to the binary of the version being documented. Relative links and heading
anchors were checked with a throwaway script; there is no link check in CI yet.
