# Decisions

The decisions that shaped MonoSmith, newest last. Each entry says what was
decided, why, and what follows from it. The research the first decisions rest
on is in [research/](research/); how the result is built is in
[architecture.md](architecture.md).

Product north star: a super simple monorepo tool where you ask for a new app,
an internal package or a publishable package and it just works; CI and release
are first-class and everything is generated.

## Open: not decided yet

- **Plugin registry and installation.** How plugins are published, found and
  installed by name (Yeoman/Homebrew style). Plugins can be declared by command
  or found on a search path today; nothing more is designed. See #164.
- **Plugin repository.** Where plugins live (this monorepo, their own
  repositories, a registry) and who publishes them.
- **Plugin flow.** How a user discovers, adds and upgrades plugins and stacks
  end to end.

The plugin protocol and roles described in [plugins.md](../reference/plugins.md) are what
exists today, not a settled design.

---

## 1. Build from scratch in Go

2026-10-15 · [research](research/) · #1

**Context.** The old mnci (TypeScript) was a thin CLI over Nx. Nx brings a
Node runtime, a daemon and its own plugin model; Please has no JavaScript,
.NET or Dart support and no native Windows.

**Decision.** Write a new tool in Go that *is* the engine: discovery, task
graph, hashing, cache, scheduler, affected, release, TUI, doctor. Nx and Please
are sources of ideas, not forks.

**Consequences.** One static binary with no runtime to install. Everything Nx
did for mnci has to be rebuilt, and converting from Nx becomes a feature
(`smith convert`). Native Windows is first-class from the start.

## 2. A language-agnostic core with out-of-process plugins

2026-10-15 · [research 03](research/03-prior-art-and-stack.md) · [plugins.md](../reference/plugins.md)

**Context.** The tool has to cover many languages without the core knowing any
of them, and plugin authors should not have to write Go.

**Decision.** Everything language-specific lives in plugins: separate
processes speaking a versioned JSON-RPC 2.0 protocol over stdio, framed like
the Language Server Protocol, in any language. Built-in plugins ship in the
`smith` binary but run out of process too (`smith plugin serve <name>`), so they
speak the same protocol an external plugin does.

**Consequences.** A new language is a plugin, with no change to the core. The
protocol is versioned by an integer, and additions are new capabilities that old
peers ignore. The Go SDK (`packages/plugin-sdk`) derives a plugin's
capabilities from what it implements. Roles and per-project stacks (#163,
2026-10-16) are how plugins combine today; the rest of the plugin design is
open (see above).

## 3. A content-addressed cache behind a Storage interface

2026-10-15 · [research 03](research/03-prior-art-and-stack.md) · [tasks and cache](../reference/tasks-and-cache.md)

**Context.** The cache must be correct across machines and operating systems,
and shareable later.

**Decision.** A task's key is a SHA-256 over a versioned, canonical encoding of
everything that can change its result, with no absolute paths or timestamps.
Results are stored the way Bazel's remote cache stores them: `ac/` entries
pointing at `cas/` blobs, behind a `Storage` interface. A local directory comes
first; remote backends come later.

**Consequences.** The remote cache (2026-10-16, #195) slots in behind the local
store: HTTP, S3 or Azure Blob, every object signed with HMAC-SHA256, and a
failed check counts as a miss. `smith why-miss` can explain a miss because each
entry keeps what its key was computed from.

## 4. Configuration in one minimal TOML file

2026-10-15

**Decision.** `smith.toml`: every setting optional, so most workspaces only
name themselves. Overrides record their source for `smith explain`, and the TUI
edits settings without losing comments.

## 5. Generated CI pipelines

2026-10-15 · [ci.md](../reference/ci.md)

**Decision.** Pipelines are generated from the workspace. Plugins contribute
toolchains, setup and audit steps and credentials. smith builds one
provider-neutral model and renders it for GitHub Actions and Azure Pipelines,
later also GitLab CI. The YAML calls `smith ci <phase>`, so a laptop runs what
CI runs. Steps a team adds between `smith:slot` markers survive regeneration.

## 6. Conventional commits and tag-only releases

2026-10-15 · [releasing.md](../reference/releasing.md)

**Context.** Release commits back to the branch race with other merges and
clutter history.

**Decision.** Versions come from conventional commits. The model is tag-only:
manifests are stamped for the build and publish, then restored, and the tag is
the record. Releases run in dependency order and are resumable (a re-run skips
what is tagged and what a registry already has). `--dry-run` follows the same
code path. A commit `version(<project>)[x.y.z]` forces a version, the way
`version(smith)[1.0.0]` started MonoSmith at 1.0.0. A released dependency gives
its dependents a patch, and no setting changes that (2026-10-16).

**Consequences.** Every merge to `main` that touches a released project
releases it. Commit subjects matter, so commitlint guards them.

## 7. npm only for Node

2026-10-15

**Decision.** The Node plugin runs npm workspaces. pnpm and yarn are not
supported, and `smith convert` reports them as such.

**Clarified 2026-10-18.** This is not a gap to close. People do not call the
package manager themselves: `smith install` and `smith uninstall` find the
project's ecosystem and run the right tool through its plugin, so not
supporting pnpm or yarn does not make smith harder to use. A workspace coming
from another tool (Turborepo, a pnpm or yarn Nx workspace) is converted to npm
workspaces instead of being refused (#353).

## 8. Containers through a docker plugin

2026-10-16 · #83 · #213

**Decision.** An app with a `Dockerfile` gets an `image` task from the
`docker` plugin, and the release pushes the image tagged with the version.
This replaced the old mnci's separate container project.

## 9. The rename to MonoSmith

2026-10-16 · a17bcfd

**Context.** The project outgrew the MoNecromanCI name, and its first release
was a chance to start clean.

**Decision.** The product is **MonoSmith**, its command is `smith`, the
repository is `monosmith-app/monosmith`, the packages are named `monosmith`
(`@monosmith/cli` too on npm), and the first release is `smith@1.0.0`. The old
Nx-based tool keeps the name **mnci**, and `smith convert` still reads its
workspaces.

**Consequences.** References to the old tool (`@mnci/*` packages, `mnci:slot`
markers, the `mnci` block of `nx.json`) stay where convert needs to recognise
them.

## 10. One binary, re-packed for every package manager

2026-10-16 → 2026-10-17 · #121 · #122 · #217 · #238 · #240 · #216

**Context.** People install tools with the package manager of their language.
Each wrapper that downloads a binary at install time breaks offline and behind
proxies.

**Decision.** smith is one static binary per OS/arch (`CGO_ENABLED=0`, six
targets). The release builds the archives once, and re-packs the same binaries
for each channel, with nothing downloaded after install:

- npm: per-platform packages plus wrappers, with provenance
- PyPI: platform wheels with the binary in their scripts
- NuGet: a .NET 10 RID-specific tool
- RubyGems: platform gems with a Ruby launcher
- pub.dev: one package holding every platform's binary
- Linux: .deb and .rpm attached to the GitHub Release

**Consequences.** pub.dev 1.0.0 was published by hand; its automation is #216.
The .deb/.rpm packages are #240. Each registry has rules that only showed up
on the first real upload, and are now in the code:

- PyPI rejects wheels whose zip entries use data descriptors, so entries are
  written with their sizes up front (#233).
- nuget.org needs a client version of 4.1.0 or later in `X-NuGet-Client-Version`
  (#245), and a license expression needs `licenseUrl` too (#247). The nuspec is
  written as `dotnet pack` writes it.

## 11. No paid code signing for now

2026-10-16 · #124

**Decision.** No Apple Developer ID and no Windows code-signing certificate
for now. Free provenance instead: SHA-256 `checksums.txt` and npm
provenance today, with GitHub artifact attestations to follow.

**Consequences.** Only a binary downloaded from GitHub Releases in a browser
is affected: macOS quarantines it, and Windows SmartScreen warns. Installs
through package managers and the install scripts are not affected.

## 12. No Java registry; OS package managers later

2026-10-16 · #225

**Decision.**
- **Java:** smith is not published to a Java registry. Java has no CLI
  registry worth an approval request: Maven Central and SDKMAN! both require
  one. The jvm plugin publishes a workspace's own libraries to a Maven
  repository it is given.
- **OS package managers** (Homebrew tap, Scoop bucket, winget, Chocolatey,
  AUR, Snap, apt/dnf repositories) and crates.io and Packagist are deferred to
  #225. The Homebrew, Scoop and winget manifests are already attached to each
  GitHub Release.

## 13. Vertical slices in every language's own naming

2026-10-16 → 2026-10-17 · #87 · #236 · [conventions.md](../reference/conventions.md)

**Decision.** The vertical-slices check uses the most common naming of each
language:

| Language | Naming |
|---|---|
| Go, Python | `<name>_<role>` (snake_case) |
| Dart, Rust, Ruby | `<name>_<role>` (snake_case) |
| C#, Kotlin, Java, PHP | a PascalCase role suffix (`PlaceOrderUseCase.cs`, tests `PlaceOrderUseCaseTests.cs`) |
| C++ | either style, but one style per folder |

Node gets the same rules through the ESLint config.

## 14. The convert end-to-end test runs on demand

2026-10-17 · #235

**Context.** Building the old-mnci and Nx workspaces that the test converts
took 30–44 minutes on Windows and held every pull request. Converting itself
takes seconds, and nobody else converts yet.

**Decision.** The `convert` job in `golden-path` runs only when started by
hand. Unit tests and the other end-to-end suites still run on every pull
request.

**Since 2026-10-17** the other end-to-end suites run once a day too (decision 17).

## 15. npm audit blocks only what can be fixed

2026-10-17 · #246

**Decision.** `smith ci audit` fails on a production advisory of moderate
severity or above only when a plain `npm audit fix` would change something.
An advisory with no upstream fix, one whose only fix is a semver-major
`--force`, and dev-only advisories are printed as notes. This ports the old
mnci's rule.

**Consequences.** CI no longer stays red while an upstream package is
unpatched. The answer to a real finding is a targeted `overrides` entry.

## 16. The license holder

2026-10-17 · #239

**Decision.** MonoSmith is MIT licensed, with the copyright held by Eduardo
Russo.

## 17. The end-to-end suites run once a day

2026-10-17

**Context.** With a plugin for every language, the golden-path suites took up
to 45 minutes on Windows, and every pull request waited for them. Several
pull requests at once queued behind GitHub's concurrency limits.

**Decision.** Pull requests and pushes run the `unit` job (vet and test of
every module on three operating systems, a few minutes). The golden-path
suites run once a day at 03:00 UTC and when started by hand.

**Consequences.** A change can reach `main`, and a release, before the
end-to-end suites have run on it; the daily run catches it within a day, and
a change the unit tests can't judge gets a manual run before merging.

## 18. Improvements are judged by how much easier smith is than its competitors

2026-10-18 · #336

**Context.** The product owner keeps research on Nx, Turborepo, Bazel, Moon
and Pants in Google Drive. Its purpose is to find where MonoSmith must improve
to be an easier and better tool than they are, not to copy their feature lists.

**Decision.** When ideas from that research (#341 to #353) or elsewhere
compete for time, the first question is how much easier, safer or more
predictable they make setting up, running and releasing a polyglot monorepo
for a team. A feature a competitor has that does not answer that is not a
priority (distributed task execution, hermetic sandboxing, a Starlark-style
configuration language).
