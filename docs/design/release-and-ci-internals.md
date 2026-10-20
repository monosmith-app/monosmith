# Releases and CI: how they are built

Two features that look unrelated share one idea: **smith owns the process, and
the CI provider only starts it.** A pipeline is a few lines that call
`smith ci <phase>`; a release is a plan that the same code produces on a laptop
(`--dry-run`) and in CI. This page explains both. Code: `apps/smith/release`,
`releaselineage`, `releasepreview`, `cipipeline`, `nativebuild`.

## Part 1 — Releases

### The plan

`release.MakePlan` decides, for every publishable project (kind `package`, and
every app that has a release section), whether it releases and as what version.

1. **Find the last release.** The highest tag with the project's prefix
   (`<name>@` by default, `<dir>/v` for Go modules, `bare_tags`, or `tag_prefix`).
   No tag means the first release, which starts from the manifest's version.
2. **Read the commits** since that tag that touched the project's *source
   roots*: its own directory and the directory of every internal `lib` it
   imports, transitively. A lib is not published; it ships inside what imports
   it, so a fix to the lib releases its users.
3. **Map conventional commits to a bump:** `feat` → minor, `fix` / `perf` →
   patch, `!` or a `BREAKING CHANGE:` footer → major (minor while below 1.0),
   everything else → nothing. A `version(<project>)[x.y.z]` directive anywhere
   in the repository forces a version.
4. **Propagate.** A package whose published dependency releases is released too,
   at least a patch, so its declared dependency range stays correct.
5. **Order** the items by the dependency graph, so a library is released before
   what depends on it.

The plan is a value (`Plan{Items, Unchanged}`), printed by `--dry-run` and
posted on pull requests by `releasepreview` (a comment updated in place). A dry
run executes the same `Execute` function with `DryRun` set, so the preview
cannot drift from the real thing.

### Lineage: refusing a plan built on a wrong past

Versions are derived from tags, so a wrong tag gives a wrong version.
`releaselineage` compares three things per project before a release: the tags
under the current prefix, tags left under an *old* prefix (a project renamed),
and what the registry says it already has (`published`). It refuses to continue
when a version was published with no tag, or tags are behind the registry —
either would make smith propose a version that already exists. `smith doctor`
reports the same findings.

### Preflight

A real release is refused up front, with every problem listed, when it would
fail halfway or corrupt history: a shallow clone (versions cannot be computed),
the wrong branch (`[release] branch`, default `main`), uncommitted changes
(manifests are stamped and restored), or a missing credential — including the
credentials of deploy plugins, of distribution to other registries, and of a
split-repository mirror. It also asks each plugin what it needs *now* so that a
contradictory docker setting fails before anything is published, not after the
npm package is already out.

### Executing one project

`Execute` goes through the plan in order and stops at the first failure,
**keeping what already finished**. For each item (`releaseOne`):

1. *Already tagged?* If the tag exists, the project is done in git; the run only
   makes sure the tag is pushed, the GitHub Release exists, and any mirror and
   pub.dev dispatch happened. This is what makes a re-run after a failure
   resume.
2. **Stamp.** Ask the language plugin to `version/write`, passing the new
   versions of the project's dependencies so ranges can be updated. The plugin
   returns the files it changed. A deferred call restores them with git: **no
   commit is ever made**. The tag is the record.
3. **Build.** Run the project's `build` task through the task runner (without
   cache if files were stamped, since the manifest differs).
4. **Publish.** Ask the language plugin to `publish`. A registry that already has
   the version reports "skipped". For an *app* the plugin is asked for release
   assets instead of a registry upload (below).
5. **Deploy.** Each deploy plugin of the project (docker) is asked to `publish`
   too, and reports the images it pushed.
6. **Distribute** an app's binaries as npm packages, PyPI wheels, a .NET tool,
   RubyGems, a pub.dev package, `.deb`/`.rpm`, and Homebrew/Scoop/winget
   manifests, from the archives step 4 built.
7. **Tag** the commit (annotated), **push** the tag, **create the GitHub
   Release** with notes rendered from the commits and the archives,
   `checksums.txt` and manifests attached, mirror to a split repository if
   configured, and dispatch the pub.dev workflow.

The result is written to `.smith/release-report.json` (printed by `--json`) and
rendered as Markdown for the CI step summary: per project, the version, why, the
status, the registry, the assets, the packages, the images with their digests,
the GitHub Release URL, the mirror.

### Why apps release assets from the plugin

A Go app's release is not "upload to a registry"; it is six cross-compiled
archives. Only the language plugin knows how to cross-compile (for Go:
`CGO_ENABLED=0`, `-trimpath`, the version stamped into `main.version`). So
`publish` carries an `AssetSpec` (name pattern, platforms, output directory) and
returns `assets`, `assetPlatforms` (which archive is for which `os/arch`) and
`binary`. Everything after that — turning the archives into an npm package with
`optionalDependencies`, a PyPI platform wheel, a gem — is language-independent
and lives in the core (`npmbinaries`, `pypiwheels`, `nugettool`, `rubygems`,
`packagemanifests`, `linuxpackages`, `pubpackage`).

### Native apps

An app whose build differs per OS (prebuilt addons, a bundled driver, a
per-platform VS Code extension) is tagged `native`. It cannot be built on one
runner, so `smith ci native` builds it on each OS of a matrix job and each job
publishes *its own platform's* build; the release job tags and creates the
GitHub Release, and the native jobs attach to it.

## Part 2 — CI

### One provider-neutral model

`cipipeline.Collect` asks every plugin that advertises `ci` what its projects
need. Each answers a `CIResult`:

| Field | Becomes |
|---|---|
| `toolchains` | a setup step per provider: a GitHub action, an Azure task, or a plain install command (GitLab, or when the plugin gives no native step) |
| `setup` | steps of `smith ci setup` (`npm ci`, `go mod download`, the Python venv) |
| `audit` | steps of `smith ci audit` |
| `postRelease` | steps run at the end of `smith ci release` |
| `credentials` | secrets: for the release, for setup (private registries), for task steps (the remote cache) |
| `permissions` | GitHub token permissions (`packages: write`) |
| `authenticate` | provider-native sign-in steps for private registries |
| `stores` | package download stores to cache |

`Contributions.Add` merges them: toolchains and stores by name (first wins), a
credential once with the union of where it is needed, steps appended in plugin
order. `NewModel` sorts everything so the output is deterministic — regenerating
an unchanged workspace changes nothing, which is what lets
`smith ci generate --check` be a CI gate.

The core adds what is not a plugin's business: whether anything needs
`dist/drop` packing, whether there are native projects, whether a package
publishes to npm (OIDC provenance needs `id-token: write`), whether an app's archives go to a GitHub Release (the attestation step needs `id-token` and `attestations: write`) or to pub.dev (the
release job needs `actions: write` to dispatch the pub workflow), and the
distribution credentials of apps.

### Renderers

Three pure functions turn the model into text: `RenderGitHub`,
`RenderAzure`, `RenderGitLab`. They share one structure — checkout, toolchains,
sign-in, package-store caches, install smith, `before-setup` slot, `setup`,
`audit`, `verify`, `after-verify` slot, release preview, `pack`, `release` —
and differ only in syntax. The renderers never decide *what* to run; they only
spell it. Because the YAML is thin, a developer runs `smith ci verify` on a
laptop and gets what CI will get.

### The phases

| Phase | Does | Runs on |
|---|---|---|
| `setup` | the plugins' setup steps | every run |
| `audit` | dependency audits (npm, pip-audit, govulncheck), with cached passing verdicts | every run |
| `verify` | `lint`, `typecheck`, `test`, `build` for the affected projects on a pull request (everything on a push), then tasks plugins mark `verify` (an app's image build); fails if the tasks changed tracked files | every run |
| `pack` | every project's `package` task into `dist/drop`, kept as an artifact | push to the release branch |
| `native` | native projects on this OS | push, matrix |
| `release` | the release, then the post-release steps | push to the release branch |

### Slots: your steps survive regeneration

A generated file is smith's, but a team always has one more step.
`smith:slot <name>` / `smith:slot-end <name>` comment markers in the YAML
(`before-setup`, `after-verify`) mark regions that regeneration **preserves**:
`generate` reads the existing file, lifts the text between each pair of markers,
renders the new file and puts it back (re-indented when the slot moved, as when
an Azure pipeline gains a native job and nests its steps). A slot the new file
no longer has is reported, so a team's step is never dropped silently. Anything
outside a slot is smith's and is overwritten; `--check` reports drift without writing.

### What is cached in CI

- the *task* results, through the remote cache (signed objects; pull requests
  read it and never write it),
- *audit verdicts*, as small cache entries keyed on the lockfile content and the
  day,
- *package download stores* (npm, pip, Go modules, cargo, NuGet, Composer,
  pub), by the provider's own cache action, keyed on the lockfiles.

Fork pull requests get no secrets, so the remote cache is simply off for them,
and the other caches restore without writing the base repository's.

## Part 3 — How smith builds and releases itself

smith is its own first user. `.github/workflows/ci.yml` is generated by
`smith ci generate`. `golden-path.yml` is hand-written: it runs the unit tests of
every module on three operating systems on every pull request, and the
end-to-end suites (`goldenpath`: a new workspace, every kind in every language,
build, test, release) once a day and on demand — they take up to 45 minutes on
Windows, so they gate changes that touch release, discovery, doctor, the
plugin protocol or generated pipelines, by hand, before merging.

A merge to `main` releases: the conventional commits since the last tag decide
the version of `smith` (an app: six archives, npm, PyPI, NuGet, RubyGems, `.deb`
and `.rpm`, the GitHub Release, the package-manager manifests) and of the plugin
SDK (a Go module tag). A change that is `docs:` or `chore:` releases nothing.
