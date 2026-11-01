# Decision records

The decisions that shaped MonoSmith, one page each, newest last. Each record says what was
decided, the context, the options considered, the consequences and where the decision can be
checked (issues, pull requests, commits). The research the first decisions rest on is in
[research](../research/); how the result is built is in the
[architecture overview](../README.md).

Product north star: a super simple monorepo tool where you ask for a new app, an internal
package or a publishable package and it just works; CI and release are first-class and
everything is generated.

Where the history does not say why something was decided, the record says "Rationale not
recorded" or "Not recorded" instead of supplying a reason.

## Index

Records 0001 to 0020 are the original decisions log, kept in its order. Records 0021 and
later were reconstructed from the issues and pull requests that made them; their dates are the
merge dates.

| No. | Decision | Status | Date | Summary |
|---|---|---|---|---|
| [0001](0001-build-from-scratch-in-go.md) | Build from scratch in Go | accepted | 2026-10-15 | A new Go engine, one static binary; Nx and Please are sources of ideas |
| [0002](0002-out-of-process-plugins.md) | Out-of-process plugins | accepted | 2026-10-15 | Everything language-specific is a plugin speaking JSON-RPC over stdio |
| [0003](0003-content-addressed-cache.md) | Content-addressed cache | accepted | 2026-10-15 | SHA-256 task keys, `ac/` and `cas/` behind a Storage interface |
| [0004](0004-minimal-toml-configuration.md) | Minimal TOML configuration | accepted | 2026-10-15 | One `smith.toml`, every setting optional |
| [0005](0005-generated-ci-pipelines.md) | Generated CI pipelines | accepted | 2026-10-15 | One provider-neutral model rendered per provider, calling `smith ci <phase>` |
| [0006](0006-conventional-commits-tag-only-releases.md) | Conventional commits and tag-only releases | accepted | 2026-10-15 | Versions from commits; the tag is the record |
| [0007](0007-npm-only-for-node.md) | npm only for Node | accepted | 2026-10-15 | pnpm and yarn workspaces are converted to npm, not supported |
| [0008](0008-containers-through-docker-plugin.md) | Containers through a docker plugin | accepted | 2026-10-16 | An app with a Dockerfile gets an `image` task |
| [0009](0009-rename-to-monosmith.md) | The rename to MonoSmith | accepted | 2026-10-16 | Product MonoSmith, command `smith`, first release 1.0.0 |
| [0010](0010-one-binary-for-every-package-manager.md) | One binary for every package manager | accepted | 2026-10-16 | One static binary re-packed for npm, PyPI, NuGet, RubyGems, pub.dev and Linux packages |
| [0011](0011-no-paid-code-signing.md) | No paid code signing | accepted | 2026-10-16 | Checksums, npm provenance, SBOM and SLSA attestations instead |
| [0012](0012-no-java-registry.md) | No Java registry | accepted | 2026-10-16 | Not published to Java registries; OS package managers deferred |
| [0013](0013-vertical-slices-in-each-languages-naming.md) | Vertical slices in each language's naming | accepted | 2026-10-16 | The slice check uses each language's common file naming |
| [0014](0014-convert-end-to-end-test-on-demand.md) | Convert end-to-end test on demand | accepted | 2026-10-17 | The slow convert job runs only when started by hand |
| [0015](0015-npm-audit-blocks-only-fixable.md) | npm audit blocks only what can be fixed | accepted | 2026-10-17 | Unfixable advisories are notes, not failures |
| [0016](0016-license-holder.md) | The license holder | accepted | 2026-10-17 | MIT, copyright held by Eduardo Russo |
| [0017](0017-end-to-end-suites-run-daily.md) | End-to-end suites run daily | accepted | 2026-10-17 | Pull requests run unit tests; golden-path suites run at 03:00 UTC |
| [0018](0018-judge-improvements-by-ease.md) | Judge improvements by ease | accepted | 2026-10-18 | Priority goes to what makes monorepos easier than competitors do |
| [0019](0019-no-go-install.md) | No `go install` | accepted | 2026-10-18 | The build needs `go.work`; supported channels carry stamped binaries |
| [0020](0020-toolchain-pins-checked-not-downloaded.md) | Toolchain pins checked, not downloaded | accepted | 2026-10-20 | `[toolchains]` pins, doctor checks, mise or asdf provisions |
| [0021](0021-plugin-roles-and-per-project-stacks.md) | Plugin roles and per-project stacks | accepted | 2026-10-16 | Each project picks its plugins through a stack |
| [0022](0022-mcp-server-runs-smith-commands.md) | MCP server runs smith commands | accepted | 2026-10-16 | Every MCP tool is a visible smith command line |
| [0023](0023-recipes-with-levels-and-slots.md) | Recipes with levels and slots | accepted | 2026-10-18 | Recipes are data; plugins declare levels, slots, requirements |
| [0024](0024-release-without-publishing-and-bare-tags.md) | Release without publishing, bare tags | accepted | 2026-10-18 | `registry = "none"` and `bare_tags` |
| [0025](0025-ide-plugins-run-the-installed-smith.md) | IDE plugins run the installed smith | accepted | 2026-10-18 | The editor, the terminal and CI run the same smith |
| [0026](0026-event-stream-and-smith-serve.md) | Event stream and `smith serve` | accepted | 2026-10-20 | One run-event contract, a JSON-RPC server over stdio |
| [0027](0027-one-version-per-production-dependency.md) | One version per production dependency | accepted | 2026-10-20 | `[dependencies] one_version`, `deps up --align` |
| [0028](0028-graph-edges-from-shared-schemas.md) | Graph edges from shared schemas | accepted | 2026-10-20 | Schema owners and consumers; `graph --blast-radius` |
| [0029](0029-one-list-of-actions.md) | One list of actions | accepted | 2026-10-22 | Shell, MCP and serve read one catalog, with drift tests |
| [0030](0030-vscode-extension-is-a-serve-client.md) | VS Code extension is a serve client | accepted | 2026-10-23 | The extension carries no engine |
| [0031](0031-marketplace-publishing-with-entra-id.md) | Marketplace publishing with Entra ID | accepted | 2026-10-23 | `vsce_entra = true` replaces `VSCE_PAT` |
| [0032](0032-single-key-tui-keymap.md) | Single-key TUI keymap | accepted | 2026-10-24 | One keymap, no menu bar, one adaptive pane |
| [0033](0033-doctor-fixes-with-steps-and-choices.md) | Doctor fixes with steps and choices | accepted | 2026-10-24 | Fixes can run commands and ask questions |
| [0034](0034-adopt-and-convert-depth-levels.md) | Adopt and convert depth levels | accepted (Nx and Turborepo); proposed (plain repositories) | 2026-10-25 | `--level conservative\|full` |
| [0035](0035-ci-branches-setting.md) | `[ci] branches` | accepted | 2026-10-25 | Where the generated pipelines run |
| [0036](0036-external-plugins-are-trusted-and-get-a-minimal-environment.md) | External plugins are trusted and get a minimal environment | accepted | 2026-10-25 | An external plugin starts only once trusted, with the baseline environment plus what it declares |
| [0037](0037-cancel-and-concurrency-in-the-plugin-protocol.md) | Cancel and concurrency in the plugin protocol | accepted | 2026-10-25 | A `cancel` notification and a `concurrency` declaration, additive, version stays 1 |
| [0038](0038-a-plugin-is-packed-as-a-checksummed-archive-and-installed-without-running-it.md) | A plugin is packed as a checksummed archive and installed without running it | accepted | 2026-10-26 | `smith plugin pack` and `install`: a deterministic zip with checksums.sha256, verified before anything is written, the pin recorded, trust left to the person |
| [0039](0039-a-plugin-declares-the-project-kinds-it-scaffolds.md) | A plugin declares the project kinds it scaffolds | accepted | 2026-10-28 | An optional `kinds` list on the initialize result; `smith kinds`, `smith add` and `plugin check --generate` use it |
| [0040](0040-plugins-are-found-by-name-in-a-marketplace-index.md) | Plugins are found by name in a marketplace index | accepted | 2026-10-31 | `smith plugin search` and `smith plugin install <name>[@version]` read a JSON index of archive URLs and SHA-256 digests; install and trust stay as in 0038 and 0036 |
| [0041](0041-smith-add-follows-the-dialect-of-sibling-projects.md) | `smith add` follows the dialect of sibling projects | accepted | 2026-10-31 | A project shape (build, test runner, exports, tsconfig) read by the language plugin; the host follows the dominant shape, asks on a split, `--fresh` opts out |
| [0042](0042-plugins-publish-shared-facts-the-core-merges.md) | Plugins publish shared facts that the core merges | accepted | 2026-10-31 | Typed, namespaced facts in `Contribution.facts`; narrowest scope wins; plugins read `Project.facts` and `GenerateParams.facts`; `smith explain --facts` |
| [0043](0043-a-release-is-rehearsed-against-its-registries-with-read-only-requests.md) | A release is rehearsed against its registries with read-only requests | accepted | 2026-10-31 | `smith release --rehearse` checks that each registry resolves, which credential it expects, that yours is accepted and that the CI identity may publish; nothing is published |
| [0044](0044-the-release-checks-the-bump-against-the-public-api.md) | The release checks the bump against the public API | accepted | 2026-10-31 | An optional `api/surface` method lists exported names at a revision; the host compares the last tag with now and warns; `[release] api_check` |
| [0045](0045-frameworks-tests-and-styles-are-choices-shipped-as-data.md) | Frameworks, tests and styles are choices shipped as data | accepted | 2026-10-31 | A choice has an axis, `requires`/`conflicts`, dependencies by name from a catalog, tasks, slots and files, in embedded TOML; the host solves `--with` (one per axis) and a stack is a saved selection |
| [0046](0046-derived-files-have-an-owner-and-a-record.md) | Derived files have an owner and an ownership record | accepted | 2026-10-31 | A `derive` plugin method renders a file, a region or keys; `smith.lock` records owner, renderer and hash; `smith derive` never overwrites a hand edit; opt-in `[derive] enabled` |
| [0047](0047-mutating-commands-show-a-plan-and-can-be-undone.md) | Mutating commands show a plan, make one commit and can be undone | accepted | 2026-10-31 | `add`, `remove`, `upgrade` and `derive` print a plan and ask in a terminal (`--yes`, `--dry-run`); opt-in `[workspace] commit_operations` makes one commit each, logged in `.smith/operations.json`; `smith undo` reverts the last with `git revert` |
| [0048](0048-why-report-and-agent-file-explain-the-workspace-to-people-and-agents.md) | `smith why`, `smith report` and `smith agent file` explain the workspace to people and agents | accepted | 2026-10-31 | `why <file>` is its own command (owner, source, rule, how to change); `report` writes one self-contained accessible HTML page; `agent file` writes AGENTS.md between region markers with `--check` |
| [0049](0049-a-scaffold-is-also-the-renderer-of-its-own-files.md) | A scaffold is also the renderer of its own files | accepted | 2026-10-31 | A plugin renders the files it scaffolded from its own template in `file` mode, so a smith-made workspace has derived files and a deleted one comes back whole; `smith add` records them when derive is on |
| [0050](0050-config-that-sibling-projects-share-is-derived-from-their-shape.md) | Config that sibling projects share is derived from their shape | accepted | 2026-10-31 | When three or more projects, and two thirds of those that have a config file, hold it in the same shape, smith derives it for each; a file or key that differs is the project's own and is reported as diverging |
| [0067](0067-groups-replace-task-groups-and-also-set-tags-and-release.md) | `[[groups]]` replaces `[[task_groups]]` and also sets tags and release | accepted | 2026-11-01 | One selector table for tasks, tags and the release; `smith doctor --fix` renames the old key. |
| [0068](0068-the-assets-job-can-skip-installing-a-project.md) | The assets job can skip installing a project | accepted | 2026-11-01 | `[release] install = false` plus `smith ci setup --assets` leave a project out of the assets job's dependency install. |
| [0069](0069-the-assets-phase-can-run-again-for-a-tag.md) | The assets phase can run again for a tag, and a glob can be optional | accepted | 2026-11-01 | An `assets` glob may match nothing if another matches; `smith ci assets --tag` and the `assets_tag` input rerun a tag; no release means a skip. |
| [0066](0066-smith-add-like-copies-the-shape-of-one-named-project.md) | `smith add --like` copies the shape of one named project | accepted | 2026-11-01 | The named project's shape is the shape to follow, for the keys the stack follows; printed key by key; excludes `--fresh`; no three-sibling minimum |
| [0065](0065-smith-swap-changes-one-choice-of-a-project.md) | `smith swap` changes one choice of a project, and stops where the migration is not automatic | accepted | 2026-11-01 | Replaces the choice on an axis, checks it with the choice solver, prints the plan; refuses to write until a plugin offers the migration |
| [0064](0064-reasons-and-exceptions-are-data-before-they-are-screens.md) | Reasons and exceptions are data before they are screens | accepted | 2026-11-01 | `x-reason` in the schema, `smith why <setting>`, reason in `config list`, `smith divergences` (override, ejected, drift, differs); JSON for the editors, no screen yet |
| [0063](0063-the-architecture-rules-are-one-source-for-four-outputs.md) | The architecture rules are one source for four outputs | accepted | 2026-11-01 | The agent file reads roles, forbidden folders and layers from the rules `smith lint` enforces; tests tie the Go scaffold and the conventions page to them |
| [0062](0062-one-word-for-one-meaning-a-vocabulary-audit.md) | One word for one meaning, a vocabulary audit | accepted | 2026-11-01 | Audit of 13 terms; "level" (as role) becomes "role" in help, docs and `plugin list`; the five-noun model is rejected; the glossary maps the words that keep two meanings |
| [0061](0061-teach-by-example-drafts-the-facts-not-the-plugin.md) | Teach by example drafts the facts, not the plugin | accepted | 2026-11-01 | `smith plugin draft <sample> --lang x` writes the manifests, version line, tasks and publish candidates of a sample as a TOML file and lists what the author still writes; no declarative plugin format |
| [0060](0060-a-shared-preset-is-a-package-smith-up-groups.md) | A shared preset is a package that smith up groups | accepted | 2026-11-01 | `[dependencies] presets` names the shared-config packages; `smith up` shows their bumps first as one `preset` group; no new packaging format |
| [0059](0059-measure-the-copies-before-removing-them.md) | Measure the copies before removing them | accepted | 2026-11-01 | `smith copies` counts identical per-project files and their lines (22 Node packages: 286 lines, 13 once); removal is left to the plugin that knows the tool, not built yet |
| [0058](0058-folder-rules-are-task-groups-with-a-path-selector.md) | Folder rules are task groups with a path selector | rejected | 2026-11-01 | No `[folders]` table: `[[task_groups]] select = ["packages/*"]` is already the folder rule, and `explain` names it |
| [0057](0057-try-a-change-in-a-sandbox-before-it-touches-the-tree.md) | Try a change in a sandbox before it touches the tree | accepted | 2026-11-01 | `smith try <command>` runs a file-changing command in a detached git worktree with the uncommitted changes overlaid, runs doctor (or build, test), and offers to run it for real |
| [0056](0056-repeated-overrides-say-what-the-default-should-be.md) | Repeated overrides say what the default should be | accepted | 2026-11-01 | `smith doctor` warns when three projects repeat a task override (lift it to `[tasks.<task>]`) or nearly every project overrides a workspace default; report only |
| [0055](0055-a-committed-launcher-runs-the-pinned-smith.md) | A committed launcher runs the pinned smith | accepted | 2026-11-01 | `smith bootstrap` writes static `tools/smith/smith` and `.cmd`; they read `[ci] smith_version` at run time and use PATH, a cached copy, `install.sh` or `npx` |
| [0054](0054-defaults-have-vintages.md) | Defaults have vintages | accepted | 2026-11-01 | `[workspace] defaults = "2026.11"` records the defaults a workspace follows; a newer smith keeps them; `smith doctor` warns when behind and `smith upgrade --defaults` moves with a plan |
| [0053](0053-every-language-plugin-follows-one-contract.md) | Every language plugin follows one contract | accepted | 2026-11-01 | One written contract (manifest or declared project, apps never to a registry, channels only with the capability, same task names and failure words), tested by `smith plugin check`; `declared` joins `channels` as a modifier capability shown by `smith plugin list` |
| [0052](0052-a-project-without-a-manifest-is-declared-in-smith-toml.md) | A project without a manifest is declared in smith.toml | accepted | 2026-11-01 | `[projects.<name>]` with `root` and `language` declares a project with no manifest; the host passes it to the language's plugin in `DiscoverParams.declared`, and the Go plugin describes it as a package of the module above it |
| [0051](0051-a-release-channel-gives-a-branch-its-own-line-of-prerelease-versions.md) | A release channel gives a branch its own line of prerelease versions | accepted | 2026-11-01 | `[[release.channels]]` names a branch, a prerelease id and a major; on it the plan counts `<major>.0.0-<id>.n` from the channel's own tags and a breaking change does not start a new major |

Records 0004, 0005, 0008 and 0016 state their rationale as not recorded; so do parts of
0002, 0009, 0011 to 0015 and 0032.

## Open: not decided yet

- Plugin repository and flow. Packing, installing, trusting and finding plugins by name are decided
  ([ADR 0036](0036-external-plugins-are-trusted-and-get-a-minimal-environment.md),
  [0038](0038-a-plugin-is-packed-as-a-checksummed-archive-and-installed-without-running-it.md),
  [0040](0040-plugins-are-found-by-name-in-a-marketplace-index.md)). Where the plugins of the
  marketplace index live, who publishes them and how a user upgrades a plugin are not.
- `smith lsp`, the language server for `smith.toml` and recipes, tracked in
  [#303](https://github.com/monosmith-app/monosmith/issues/303). `smith.toml` diagnostics,
  completion and hover exist; recipes and doctor findings are not built.
- Plain-repository adoption at the full level,
  [#553](https://github.com/monosmith-app/monosmith/issues/553) (see [ADR 0034](0034-adopt-and-convert-depth-levels.md)).

The plugin protocol and roles described in [Plugins](../../reference/plugins.md) are what
exists today, not a settled design.

## How to write one

Copy this structure into `NNNN-short-name.md`, with the next free four-digit number, and add a
row to the table above. The page title is `ADR NNNN: <decision>` and must be unique in the
repository.

```markdown
# ADR NNNN: <the decision, as a noun phrase>

- Status: proposed | accepted | rejected | superseded by NNNN
- Date: YYYY-MM-DD (say "date approximate" when inferred)

## Context
What forces the decision. Facts only.

## Options considered
What else was possible and why it was not chosen.

## Decision
What was decided, in the present tense.

## Consequences
What follows, including what is not done.

## Sources
Issues, pull requests and commits as links; research and architecture pages.
```

Do not invent a rationale. If the history does not say why, write "Rationale not recorded."
and link the commit or pull request. To change a decision, write a new record and mark the old
one `superseded by NNNN`; do not edit its decision.
