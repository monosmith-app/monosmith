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

Records 0004, 0005, 0008 and 0016 state their rationale as not recorded; so do parts of
0002, 0009, 0011 to 0015 and 0032.

## Open: not decided yet

- Plugin registry and installation. How plugins are published, found and installed by name
  (Yeoman/Homebrew style). Plugins can be declared by command or found on a search path
  today; nothing more is designed. See [#164](https://github.com/monosmith-app/monosmith/issues/164).
- Plugin repository. Where plugins live (this monorepo, their own repositories, a registry)
  and who publishes them.
- Plugin flow. How a user discovers, adds and upgrades plugins and stacks end to end.
- `smith lsp`, the language server for `smith.toml` and recipes, planned in
  [#303](https://github.com/monosmith-app/monosmith/issues/303) and not built.
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
