# CI pipelines

> Terms such as project, kind, task and stack are defined in the [glossary](glossary.md). How it works inside: [release and CI internals](../architecture/release-and-ci-internals.md). Things that go wrong: [troubleshooting](../guides/troubleshooting.md#ci-differs-from-my-machine).

smith generates the pipeline from what is in the workspace and keeps it in step:
`smith new` writes it, every `smith add` regenerates it, and `smith ci generate`
does it on demand. `smith doctor` (and `smith ci generate --check`) fails when the
committed pipeline is out of date.

## How it is built

Each plugin contributes, for its own projects:

- **toolchains** — the runtimes to install, at versions the plugin pins. The
  plugin says how: a GitHub action and an Azure task with their inputs
  (`actions/setup-node@v7`, `UsePythonVersion@0`, …), and a shell command for
  everything else (GitLab, CircleCI and Bitbucket install every toolchain this way). smith itself knows no runtime, so a plugin for a new
  language installs it the same way; a runtime no plugin says how to install
  is left as a comment in the pipeline, to fill through the `before-setup`
  slot.
- **setup** — e.g. `npm ci`, the Python venv and editable installs, `go mod download`
- **audit** — `npm audit` (blocking only on a production advisory at moderate
  or above that `npm audit fix` can apply; advisories with no fix, a
  semver-major-only fix, or only in dev dependencies are printed as notes),
  `pip-audit` and `govulncheck` (reported). A step that passed is not run
  again while its lockfile is unchanged and the day is the same (see
  [Cached audit verdicts](#cached-audit-verdicts))
- **credentials** — the secrets publishing needs (`NODE_AUTH_TOKEN`, `PYPI_TOKEN`,
  `CARGO_REGISTRY_TOKEN`, …; see [Secrets](#secrets))

smith merges them into one provider-neutral model, and a renderer per provider
writes the YAML. Every provider carries the same phases, in the same order.

## Branches

The GitHub and Azure pipelines run on pushes to, and pull requests into, the
release branch (`[release] branch`, default `main`). A workspace with another
long-lived branch lists it:

```toml
[ci]
branches = ["main", "dev"]   # push and pull_request triggers; default: the release branch only
```

The release branch is always one of them, even if the list leaves it out. It
stays the only branch a release runs from: the pack, upload and release steps
keep their condition on it, so a push to `dev` runs verify and nothing more.
Pull requests into any of the listed branches get the release preview comment.

A [release channel](releasing.md#release-channels) adds its branch to the triggers and to the
release step only: a push to the channel branch runs verify and `smith ci release`, which
publishes the channel's prerelease versions. The sign-in steps (Azure, RubyGems) run there too, since the release publishes from the branch; pack and upload stay on the release
branch. The preview comment on a pull request into a channel branch names the channel and shows
the prerelease versions, because a pull request build plans for the branch it targets.
`smith doctor` reports a pipeline that does not match the list, and
`smith ci generate` rewrites it. GitLab, CircleCI and Bitbucket are not affected.

## Sibling repositories

A build that needs another repository beside the workspace (a plugin pack built
from source, say) declares it:

```toml
[[ci.sibling_repos]]
repo = "acme/VideoPlugins"             # owner/name
ref = "main"                           # branch, tag or commit; default: the repository's default branch
env = "VIDEOPLUGINS"                   # holds $GITHUB_WORKSPACE/VideoPlugins in every later step
dir = "VideoPlugins"                   # default: the repository's name
pre_build = "cargo build --release -p videoplugins-music-sync"   # optional, runs in the checkout
token_secret = "VIDEOPLUGINS_TOKEN"    # optional: a secret that can read a private repository
```

The generated workflow checks the repository out right after the workspace's own
checkout, exports `env`, and runs `pre_build` after the toolchains and smith are
installed and before `smith ci setup`. The `ci` job, the native job and the
pipeline templates (`smith pipeline`) all do this. `smith.toml` is refused,
with how to fix it, when `repo` is not `owner/name`, `env` or `dir` repeats or
is not valid, or `dir` leaves the workspace.

Only GitHub Actions is covered: with another provider in `[ci] providers` the
configuration is refused. The Azure, GitLab, CircleCI and Bitbucket pipelines,
the pub workflow and the devcontainer do not check the repository out, so check
it out in your own pipeline there. Without `token_secret` the job's own token is
used, which can read only public repositories (and this one).

## CircleCI

With `"circleci"` in `[ci] providers`, smith writes `.circleci/config.yml`
(configuration version 2.1). It is the same pipeline as GitLab's, spelled for
CircleCI:

- an `ci` job on a `debian:bookworm` image, with one shared command
  (`install-tools`) that installs system packages, checks out the code,
  installs each toolchain through its plugin's shell command, restores the
  package stores and installs smith (npm, pip or `dotnet tool` for Node, Python or .NET workspaces, else the vendored `tools/smith/install.sh`);
- then `smith ci setup`, `audit` and `verify`, the `before-setup` and
  `after-verify` slots, and the release preview;
- on the release branch only (a `when` condition on the branch), `smith ci
  pack` (its `dist/drop` kept as the `drop` artifact) and `smith ci release`;
- a `native` job after it, when a project is native;
- one workflow that runs both.

Each CircleCI step is its own shell, while the plugins' install commands set
`PATH` (and `JAVA_HOME`, `DOTNET_ROOT`) for "the rest of the job". Each
install step therefore appends what its command changed in the environment
to `$BASH_ENV`, which every later step reads.

**What CircleCI cannot express, and what smith does about it:**

- **Pull requests.** CircleCI builds branches; it has no pull request trigger.
  The job runs on every branch, the release preview runs when
  `CIRCLE_PULL_REQUEST` is set, and a release runs only from the release
  branch. `[ci] branches` does not apply; `smith ci generate` says so.
- **The release preview comment.** CircleCI has no comment service of its
  own, so smith comments through the host of the pull request: on a GitHub
  repository (GitHub.com or Enterprise Server) when `GITHUB_TOKEN` is a
  project environment variable that can comment on pull requests; on any
  other host the preview stays in the job's log. There is no job summary.
  Pull requests, including from forks, read the remote cache and never write
  it.
- **Native projects on other systems.** The toolchains' install commands are
  written for Debian, so the `native` job builds on Linux only. A configured
  `windows-latest` or `macos-latest` runner is named in a comment in the file
  and in a note when `smith ci generate` runs; add a job with a macOS or
  Windows executor that runs `smith ci native` yourself.
- **Pushing tags.** CircleCI's checkout key is read-only. Add a user key with
  write access (Project Settings, SSH Keys, Add user key) so `smith ci
  release` can push its tags.
- **Secrets.** CircleCI passes project environment variables (and the
  variables of a context the job uses) to every step, so the file maps none
  of them; the credentials the release reads are listed in a comment at its
  end and by `smith ci check`.
- **GitHub-only features.** SBOM provenance attestation, `pub.yml`, npm
  provenance, RubyGems trusted publishing and `vsce_entra` are GitHub Actions
  features. The last two are refused in `smith.toml` when another provider is
  listed.

`smith ci generate --check` and `smith doctor` cover the file like the
others: an edited file is reported as out of date, and steps you put between
the `smith:slot` markers are kept.

## Bitbucket Pipelines

With `"bitbucket"` in `[ci] providers`, smith writes `bitbucket-pipelines.yml`
at the repository root. It is the same pipeline as GitLab's, spelled for
Bitbucket Pipelines:

- a `Verify` step on a `debian:bookworm` image, defined once under
  `definitions` and reused: it installs system packages, each toolchain through
  its plugin's shell command and smith (npm, pip or `dotnet tool` for Node,
  Python or .NET workspaces, else the vendored `tools/smith/install.sh`), then
  runs `smith ci setup`, `audit` and `verify`, with the `before-setup` and
  `after-verify` slots and the release preview;
- on the release branch only (a test of `BITBUCKET_BRANCH` in the script),
  `smith ci pack` (its `dist/drop` kept as the step's artifacts) and
  `smith ci release`;
- a `Native (Linux)` step after it, when a project is native;
- the same steps for `pull-requests` (`'**'`) and for `branches` of the release
  branch, so a push to any other branch builds nothing.

A step's script runs in one shell, so the plugins' install commands that set
`PATH` (and `JAVA_HOME`, `DOTNET_ROOT`) for "the rest of the job" work as they
are. The package stores are Bitbucket caches under `definitions`, keyed on the
store's lockfiles.

**What Bitbucket Pipelines cannot express, and what smith does about it:**

- **Branches.** Only the release branch and pull requests build, so `[ci]
  branches` does not apply; `smith ci generate` says so.
- **The release preview comment.** Bitbucket Cloud has a pull request comments
  API, so smith comments there when `BITBUCKET_TOKEN` is a secured repository
  variable (a repository access token with pull request write). Without it the
  preview stays in the step's log. There is no job summary. Pull requests read
  the remote cache and never write it.
- **Native projects on other systems.** Bitbucket's cloud runs Linux, so the
  `Native (Linux)` step is the one native step. A configured `windows-latest` or
  `macos-latest` runner is named in a comment in the file and in a note when
  `smith ci generate` runs; add a step with `runs-on: [self.hosted, ...]` on a
  self-hosted macOS or Windows runner that runs `smith ci native` yourself.
- **Pushing tags.** `smith ci release` pushes with `BITBUCKET_TOKEN`
  (`https://x-token-auth:...@bitbucket.org/...`), so the token also needs
  repository write.
- **Secrets.** Bitbucket passes repository and workspace variables to every
  step, so the file maps none of them; add the credentials the release reads,
  listed in a comment at the file's end and by `smith ci check`, as secured
  variables (Repository settings, Pipelines, Repository variables).
- **Cache keys.** A Bitbucket cache key hashes at most ten files; a store with
  more lockfiles takes the first ten.
- **GitHub-only features.** The same as for CircleCI: SBOM provenance, `pub.yml`,
  npm provenance, RubyGems trusted publishing and `vsce_entra`.

`smith ci generate --check` and `smith doctor` cover the file like the others:
an edited file is reported as out of date, and steps you put between the
`smith:slot` markers are kept.

## GitLab CI

With `"gitlab"` in `[ci] providers`, smith writes `.gitlab-ci.yml`. It has
one job, run for merge requests and on pushes to the release branch, on a
`debian:bookworm` image. Each toolchain installs through its plugin's shell
command: Node.js from NodeSource, Python through uv at the pinned version, Go
from its release tarball. Then come the same phases as GitHub and Azure, with
the same slots.

Two masked CI/CD variables complete it. `GITLAB_TOKEN` is a project access
token with the `api` scope, for the release preview on merge requests, and
`write_repository`, to push release tags. Then add the publishing secrets the
file lists at its end. Native projects build per OS in the GitHub and Azure
pipelines only: on GitLab, add a job per runner OS that runs
`smith ci native`.

## The phases

The YAML is thin: each step calls `smith ci <phase>`, so what CI runs is what you
can run locally.

| Phase | Does |
|---|---|
| `smith ci setup` | every setup step the plugins contributed |
| `smith ci audit` | dependency vulnerability checks; non-blocking ones warn |
| `smith ci verify` | `lint`, `typecheck`, `test`, `build` — affected projects only on a pull request, everything otherwise; then the tasks a plugin marks for verify, such as the docker convention's `image` (skipped, with a warning, where their program is not installed). When `[derive] enabled = true` it first runs `smith derive --check` and fails on a derived file or `smith.lock` that is out of date. It fails when the tasks left a tracked file changed, naming the file and the project whose tasks ran, since the release refuses a dirty tree (in CI, a file setup changed fails it too; locally, your own edits do not). `smith doctor` warns about the usual causes: a tracked file `.gitignore` ignores, and a tracked file inside a task's outputs |
| `smith release --dry-run --comment` | on a pull request: the release preview, in the job summary and as a comment on the pull request |
| `smith ci pack` | on a push to the release branch: every project's `package` task into `dist/drop`, kept as the run's `drop` artifact |
| `smith ci native` | in a job per OS, after the ci job: lint, test, build and package the projects tagged native |
| `smith ci assets` | in a job per platform, after the ci job (GitHub Actions only): run an app's `[release]` build command and upload its files to the GitHub Release ([Releasing](releasing.md#release-files-from-a-command)) |
| `smith ci release` | on a push to the release branch: version, publish and tag; then the post-release steps plugins declare, whether or not anything was released (the docker plugin's `smith docker description`, when an app has a `description_from` file or a short description for a `docker.io` registry), each only a warning when it fails |

On GitHub Actions and Azure Pipelines output is grouped, failures are annotated,
and each phase adds a line (passed or failed, and how long it took) and the
release its report to the job summary. When a blocking setup or audit step fails, its whole captured output (the last 64 KB) is printed again in a log group named after the step, so a cause that scrolled past (such as pip's conflict report) is one click away; the error keeps its last 30 lines.

## Packing

`smith ci pack` runs the `package` task of every project that has one and
leaves what they make in `dist/drop`, which the pipeline keeps as the `drop`
artifact (GitHub `actions/upload-artifact`, Azure `publish`). The steps are in
the pipeline only when some project has a `package` task. The built-in
plugins give every app one:

| App | `package` makes |
|---|---|
| Node | `npm pack` of the app, after its build: `dist/drop/<scope>-<name>-<version>.tgz` |
| Go | the binary: `dist/drop/<name>/<name>` (a plain `smith build` writes it to `dist/<project dir>/`) |
| Python | the app's wheel: `dist/drop/<name>-<version>-py3-none-any.whl` |
| VS Code extension (`vscode-extension` stack) | `dist/drop/<name>-<version>.vsix` |

A stack's plugin, or `[projects.<name>.tasks.package]` in `smith.toml`, can
replace it, for example with a zip a deploy step expects.

On a pull request, `smith ci verify` also runs the `package` task of every
affected **app**, building into `dist/drop` and publishing nothing, so a
packaging step that breaks (a staging script that misses a new workspace
dependency) shows on the pull request and not on the release branch. Native
projects are left to the native job. `verify = false` under
`[projects.<name>.tasks.package]` turns it off for one app, or under
`[tasks.package]` for all.

## The release preview comment

On a pull request the preview is posted as one comment, which later runs of
the same pull request update instead of adding more: an issue comment on
GitHub (the workflow grants `pull-requests: write` and passes
`GITHUB_TOKEN`), a closed thread on Azure Repos (the step passes
`SYSTEM_ACCESSTOKEN: $(System.AccessToken)`; the build service needs
"Contribute to pull requests"). A pull request from a fork gets a read-only
token, so its preview stays in the log and the job summary; failing to post
is a warning, never a failure.

## SBOM and build provenance

When an app attaches archives to a GitHub Release, `smith ci release`
attaches a CycloneDX SBOM as well (see [releasing](releasing.md)). On GitHub
the workflow adds `id-token: write` and `attestations: write` and, after the
Release step, an `Attest build provenance` step
(`actions/attest-build-provenance`) over `dist/release/*/*`, which signs the
archives, packages and SBOMs with SLSA provenance. Verify one with
`gh attestation verify <file> --repo <owner>/<repo>`. A private repository of a free
GitHub organisation has no attestations: the step fails and takes the release job with it,
after the releases exist. Set `[ci] attest = "optional"` to let the job go on when GitHub
refuses (the step gets `continue-on-error: true`), or `"off"` to leave the step and its two
permissions out. The default is `"on"`. Azure Pipelines and
GitLab CI, CircleCI and Bitbucket get the SBOM but no provenance step: they have no attestation
service.

## Pinned versions and the dev container

Nothing installs "latest". Each plugin pins what it installs: for example
Node.js 24, Python 3.12 and its tools (`requirements-dev.txt`), Go at the
version `go.work` names and golangci-lint, the .NET 10 SDK, a JDK 21; actions
are pinned to a major.
Generated pipelines install smith at the version that generated them.

With `[ci] devcontainer = true` (which `smith new` sets), smith also writes
`.devcontainer/devcontainer.json`: a base image plus one dev container
feature per toolchain, at the versions CI installs, and a post-create command
that installs the same smith and runs `smith ci setup`. Like the pipelines, it
is rewritten by `smith add` and `smith ci generate`, and `ci generate --check`
and doctor report it when it is out of date.

`smith doctor` compares the tools on your machine with the pins: an older Go
than `go.work` requires fails; a different Node.js major, venv Python or
golangci-lint warns, since results may then differ from CI's.

## Native projects

A project tagged `native` must be built on each operating system rather than
cross-compiled: a Go app using cgo, marked with a `// smith:native` line at the
top of its `go.mod` (its tasks then build with `CGO_ENABLED=1`), or a Node
project whose build differs per platform (prebuilt addons, a bundled runtime,
a VS Code extension with a build per target), marked with `"smith": {"native":
true}` in its `package.json`. `smith ci
verify` and `smith ci pack` leave native projects out; the generated pipelines
add a `native` job that runs after the ci job on each OS and calls `smith ci
native`, which lints, tests, builds and packages them there and keeps
`dist/drop` as the `native-<os>` artifact.

A native app that releases (`[projects.<name>.release]`) gets no
cross-compiled assets from the release: on a push that released it, each
native job builds the app for its own platform and attaches the archive to
the app's GitHub Release (GitHub Actions; the job needs `GITHUB_TOKEN`). A
Node app's archive is its build directory, `dist/<project dir>`, zipped and
named by `asset_name`.

A native package is tagged by the release but not published there: on the
push that released it, each native job stamps the version, publishes its own
platform's build and restores the manifest. A VS Code extension is published
with `vsce publish --target` (`win32-x64`, `darwin-arm64`, `linux-x64`, …), so
the Marketplace gets one build per platform. The native job is given the
release credentials (`VSCE_PAT`, `NODE_AUTH_TOKEN`, …) for this, on GitHub
Actions and on Azure Pipelines alike (a push to the release branch, the
condition the release step has). Attaching a native app's assets needs a
GitHub Release, so it happens on GitHub Actions only.

```toml
[ci]
native_runners = ["ubuntu-latest", "windows-latest", "macos-latest"]  # the default
native_linux_packages = ["libsqlite3-dev"]   # apt packages installed before building on Linux
```

Every AppleScript project is native, because `osacompile` exists on macOS only: with only AppleScript projects native, set `native_runners = ["macos-latest"]`.

Every Objective-C project is native too, because `pod lib lint` needs Xcode: with only such projects native, set `native_runners = ["macos-latest"]`.

On GitHub Actions the native and assets jobs wait for the ci job, but they do not need all of it to succeed. `smith ci release` writes `released=true` to the step output as soon as it has released a project, and right after the Release step a "Mark the release done" step sets the same output on success. The jobs run when either is set, so a release that stops half way (six projects released, the seventh failed) still gets its files, and a step that fails after the release (the attestation on a repository with none, say) does not leave the releases without their files. On a pull request they still need the whole ci job to succeed. A re-run of the failed ci job finishes the release and then runs them.

On Azure Pipelines a workspace with a native project gets a `jobs:` pipeline:
the ci job, and the native job with a matrix of runner images. Steps kept in
the `smith:slot` markers move with the job.

## Installing smith in CI

Generated pipelines install the smith version that generated them with the
install script (`SMITH_VERSION`), so a pipeline does not change behaviour until
you regenerate it with a newer smith. Each release of smith therefore makes
`smith doctor` report the pipelines as out of date until you regenerate them.

A workspace that does not need that sets the version itself:

```toml
[ci]
smith_version = "latest"    # or an exact version such as "1.70.0"
```

smith's own repository does: its pipelines run the smith built from the commit,
and the installed one only bootstraps.

## Secrets

### Your own secrets

A build command that reads a token (a fetch step before a native build, say) needs it in
the job that runs it. Name the repository secrets in `[ci] secrets`:

```toml
[ci]
secrets = ["FETCH_TOKEN"]
```

The generated pipelines hand each one, as an environment variable of the same name, to
every step that runs smith: setup, verify, release, the native jobs and the assets jobs
(GitHub Actions `${{ secrets.FETCH_TOKEN }}`, Azure `$(FETCH_TOKEN)`; on GitLab, Bitbucket
and CircleCI a CI/CD variable of that name is already in the environment). A name may not
start with `GITHUB_`. None is required, so a local run needs none.

### Secrets smith uses

| Secret | Used for |
|---|---|
| `NODE_AUTH_TOKEN` | `npm publish` (the generated `.npmrc` reads it), and an app's npm packages |
| `PYPI_TOKEN` | twine upload as `__token__` (or set `TWINE_USERNAME`/`TWINE_PASSWORD` and `TWINE_REPOSITORY_URL` for another index), and an app's wheels |
| `NUGET_API_KEY` | `dotnet nuget push` of .NET packages, and an app's .NET tool packages (a nuget.org API key that can push the tool's package and `<name>.*`) |
| `GEM_HOST_API_KEY` | `gem push` of Ruby gems, and an app's platform gems (rubygems.org, or `RUBYGEMS_HOST`) |
| `CARGO_REGISTRY_TOKEN` | `cargo publish` of Rust crates |
| `PUB_TOKEN` | a Dart package whose `publish_to` names another repository; pub.dev itself publishes with GitHub's OIDC token |
| `MAVEN_REPOSITORY_URL`, `MAVEN_USERNAME`, `MAVEN_PASSWORD` | `mvn deploy` and Gradle's `maven-publish` of JVM packages, unless `[registries]` names a `maven` repository (GitHub Packages needs only the job's `GITHUB_TOKEN`) |
| `OVSX_PAT` | `ovsx publish` of a VS Code extension app to Open VSX, besides the Marketplace; optional: without it the extension is published to the Marketplace only |
| `VSCE_PAT` | `vsce publish` of a VS Code extension package to the Marketplace (on the release step, and on the native job's step for a native extension). With `[release] vsce_entra = true` it is not needed: the workflow signs in to Azure with `AZURE_CLIENT_ID` and `AZURE_TENANT_ID` instead (see releasing.md) |
| `DOCKERHUB_USERNAME`, `DOCKERHUB_TOKEN` | the docker convention's image push to Docker Hub (a `docker.io/...` registry); the token is a Docker Hub access token |
| `CONTAINER_USERNAME`, `CONTAINER_PASSWORD` | the docker convention's image push to a registry other than ghcr.io and docker.io, which logs in with `GITHUB_TOKEN` (giving the job `packages: write`) and `DOCKERHUB_*`. Only one such registry can use these names; further ones name theirs in a `registries` table, and those variables are listed here too (see releasing.md). `CONTAINER_REGISTRY`, deprecated, names the registry when `[conventions.docker] registry` is unset |
| `SPLIT_REPO_TOKEN` | pushing a project with `split_repo` to its mirror repository; a token with write access to that repository (see releasing.md) |
| `SMITH_CACHE_KEY` | the remote cache (`[cache] remote`): the secret every object is signed with. Given to every step that runs tasks, with the backend's credentials below (see [The remote cache](#the-remote-cache)) |
| `AZURE_STORAGE_SAS_TOKEN` | an `azblob://` remote cache: a SAS token for the container |
| `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION` | an `s3://` remote cache |
| `SMITH_CACHE_TOKEN` | an `https://` remote cache whose server wants a bearer token |
| `GITHUB_TOKEN` | GitHub Releases and their assets, with `[release] github_releases`; GitHub Actions provides it. Also installs from and publishes to a `github-packages` registry (given to setup too) |
| the `token_env` of a `token` registry | installing from (given to setup too) and publishing to that private registry |

The pipeline lists only the secrets its projects need. Go packages are
published by their tag and need no secret; PHP packages are read by Packagist
from the tags (of a split repository, with `split_repo`, when the package lives in a subdirectory), and C++, Bicep, Lua, AppleScript, VBA, Swift and Objective-C projects release by their tag only. On GitHub, the job
has `contents: write` so it can push tags; on Azure the checkout keeps its
credentials (`persistCredentials`).

An app published to pub.dev needs no secret either. With one, GitHub gets a
second workflow, `.github/workflows/pub.yml`, which the release starts on the
app's pub tag (the ci workflow gets `actions: write` for that). It sets up
Dart with pub.dev's OIDC token and runs `smith release pub`; see
[Releasing](releasing.md#pubdev) for the one-time setup on pub.dev.

## Package download stores

smith caches task outputs, but `npm ci`, `pip install`, `go mod download`,
`cargo fetch` and `dotnet restore` would still download everything on every
run. Each language plugin declares the directory its package manager keeps
downloads in and the lockfiles that decide what is in it, and the generated
pipelines restore it before `smith ci setup` and save it after a successful
run, keyed on the lockfiles:

| Plugin | Store | Keyed on |
|---|---|---|
| node | `~/.npm` | `package-lock.json` |
| python | `~/.cache/pip` | `requirements-dev.txt` and each project's `pyproject.toml` |
| go | `~/go/pkg/mod` | each module's `go.sum` |
| rust | `~/.cargo/registry` | every `Cargo.lock` |
| dotnet | `~/.nuget/packages` | every `.csproj` and `.fsproj` |
| php | `~/.cache/composer` | every `composer.lock` |
| dart | `~/.pub-cache` | every `pubspec.lock` |

The key is `smith-<store>-<OS>-<hash of the lockfiles>`. When a lockfile
changed, the run restores the latest cache of the same store and OS and only
downloads the difference. Only the ci job caches: the native jobs run on other
operating systems whose stores live elsewhere.

How each provider does it:

- **GitHub Actions**: an `actions/cache` step per store. A pull request from a
  fork restores the cache and never writes the base repository's.
- **Azure Pipelines**: a `Cache@2` task per store. The stores live under
  `$(Pipeline.Workspace)/.smith-stores/`, which the pipeline's `variables:`
  point each package manager at (`NPM_CONFIG_CACHE`, `PIP_CACHE_DIR`,
  `GOMODCACHE`, `CARGO_HOME`, `NUGET_PACKAGES`, `COMPOSER_CACHE_DIR`,
  `PUB_CACHE`), because Azure caches only paths it can name.
- **GitLab CI**: a `cache:` entry per store under `.smith/stores/` in the
  project (the same variables point there; `.smith/` is smith's own ignored
  directory). GitLab keys on at most two files, so a store with more takes
  the first two, and allows four caches per job: stores beyond the fourth are
  named in a comment. A merge request from a fork runs in the fork's project
  and uses its caches.
- **CircleCI**: a `restore_cache` and a `save_cache` step per store, keyed on
  the OS and a checksum of all the store's lockfiles (concatenated into one
  file, since a key takes one file), falling back to the latest cache of the
  same store. The stores stay where each tool keeps them (`~/.npm`,
  `~/go/pkg/mod`).

`smith ci generate --check` reports a pipeline that is missing them.

## Cached audit verdicts

A pull request that touches only Go code should not re-audit the Python and npm
trees. Each audit step names the files that decide what it finds (a plugin
declares them: `package-lock.json` for npm audit, a module's `go.mod` and
`go.sum` for govulncheck, `requirements-dev.txt` and the projects'
`pyproject.toml` for pip-audit). When the step passes, smith records it in the
task cache, keyed on the content of those files, the step's command and
directory, smith's version, the platform and **the day (UTC)**. The next
`smith ci audit` with the same files on the same day prints
`✓ npm audit: no change in package-lock.json since it passed on 2026-10-18`
and skips the step. With a [remote cache](#the-remote-cache) configured, the
verdict is shared by every runner and by pull requests, which read the remote
and never write it.

The day stands in for the advisory database: a new advisory shows up within a
day, so no verdict is replayed on the next one. A failing step is never
remembered, blocking or not, and always runs again. govulncheck also reads the
module's code, which the key leaves out; a change that newly reaches a
vulnerable function is found at the next day's run.

## Audit policy checks

Beyond the vulnerability scan, a language plugin can contribute policy checks
to `smith ci audit`. They are off by default, so an existing workspace does not
change; opt in per ecosystem in `[ci]`:

```toml
[ci]
audit_policies = ["node", "go", "python", "rust"] # which ecosystems' policy checks run
audit_warn = ["go"]                      # these only report; the rest block
audit_allow_install_scripts = ["esbuild"] # node: packages allowed an install script
audit_max_unsafe = 0                     # rust: unsafe blocks allowed in all (default 0)
```

| Policy | Check | Fails on |
|---|---|---|
| `node` | `smith npm-policy` reads `package-lock.json` | a registry package with no integrity hash, a package fetched from git or over plain `http://`, a package with an install script that is not in `audit_allow_install_scripts` (the install scripts of the default stack's own tools, `@parcel/watcher` and `unrs-resolver` from Jest and `@monosmith/eslint-config`, are allowed without it) |
| `go` | `go mod verify` in every Go module | a downloaded module that no longer matches `go.sum` |
| `python` | `smith py-policy` reads `requirements-dev.txt` and each Python project's `requirements.txt` (a missing file is skipped) | a requirement with no `--hash` (not in the generated `requirements-dev.txt`, which holds top-level tool pins and is exempt from the hash rule), one fetched from a direct URL or `git+`, an `--extra-index-url`, or an `--index-url` other than the expected index (the workspace's private registry, else PyPI) |
| `rust` | `cargo deny check licenses bans` (a setup step installs cargo-deny; your `deny.toml` decides what is allowed) and `smith rust-policy`, which counts `unsafe` blocks, functions, impls and traits in each Rust project's `.rs` files (not `target/`, comments or strings) | a licence or ban cargo-deny rejects, or more `unsafe` than `audit_max_unsafe` |

Each check is an ordinary audit step, so it is cached like the others (see
[Cached audit verdicts](#cached-audit-verdicts)) and shows in the generated
pipeline's single `smith ci audit` call. A plugin reads the choice from the CI
request's `audit` field (`protocol.AuditSettings`), so a plugin for another
language adds its checks the same way. `govulncheck` already runs as the Go
vulnerability scan and `pip-audit` as the Python one. Hashed requirements come
from `pip-compile --generate-hashes` or `uv pip compile --generate-hashes`.
The unsafe count reads sources, so it always runs; the other checks are cached.

## The remote cache

With `[cache] remote` set (see [Tasks, cache and affected](tasks-and-cache.md#remote-cache)),
the pipelines pass the cache's key and its backend's credentials to every step
that runs tasks: `smith ci verify`, `smith ci pack`, `smith ci native` and the
release, and to `smith ci audit`, which keeps its verdicts there. The setup runs
no cached task and gets none. Which
credentials follows from the remote's scheme:

| `remote` | Secrets |
|---|---|
| `azblob://account/container/prefix` | `SMITH_CACHE_KEY`, `AZURE_STORAGE_SAS_TOKEN` |
| `s3://bucket/prefix` | `SMITH_CACHE_KEY`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION` |
| `https://host/path` | `SMITH_CACHE_KEY`, `SMITH_CACHE_TOKEN` (when the server wants one) |

`smith ci check` lists them, in each provider's terms, and what each one is.
None of them is required: a run without the key uses the local cache and says
so (`the remote cache is off: SMITH_CACHE_KEY is not set`).

- **Pull requests read the cache and never write it.** A pull request (GitHub
  `pull_request`, Azure `Build.Reason` of `PullRequest`, GitLab merge requests,
  CircleCI's `CIRCLE_PULL_REQUEST`, Bitbucket's `BITBUCKET_PR_ID`)
  is read-only, so a contributor's build cannot store a result that every later
  run replays; only a push to the release branch writes. This is smith's own
  rule, not YAML: `smith run` in a pull request behaves the same on a laptop
  that sets `GITHUB_EVENT_NAME=pull_request`. `[cache] remote_read_only = true`
  and `SMITH_CACHE_READ_ONLY` still force read-only everywhere.
- **A fork's pull request has no secrets.** GitHub passes an empty
  `SMITH_CACHE_KEY`; Azure leaves the variable's own name, `$(SMITH_CACHE_KEY)`,
  in its place. smith reads both as unset, so the remote is off and the run
  works with the local cache.
- **Azure needs the mapping.** Azure Pipelines exposes a secret variable to a
  step only through `env:`, so each of those steps carries
  `SMITH_CACHE_KEY: $(SMITH_CACHE_KEY)`. GitHub reads `${{ secrets.X }}`, and
  GitLab's CI/CD variables, CircleCI's and Bitbucket's environment variables are already environment.
- `smith doctor` warns on a machine where the remote is configured and the key
  is not set.

### Example: Azure Pipelines with Azure Blob

1. Create a storage account and a container for the cache, for example
   `smith-cache` in `acmecache`.
2. Create a SAS token for the container with read, write and create
   permission, and a long random string for the key.
3. Configure the remote (a recipe can set this once for every repository):

   ```toml
   [cache]
   remote = "azblob://acmecache/smith-cache/main"
   ```

4. In the pipeline's variables (Edit → Variables) add `AZURE_STORAGE_SAS_TOKEN`
   and `SMITH_CACHE_KEY`, each marked **keep this value secret**.
5. `smith ci generate`. The Verify step becomes:

   ```yaml
     - bash: smith ci verify
       displayName: Verify
       env:
         AZURE_STORAGE_SAS_TOKEN: $(AZURE_STORAGE_SAS_TOKEN)
         SMITH_CACHE_KEY: $(SMITH_CACHE_KEY)
   ```

   and the Pack, Native and Release steps carry the same two lines.
6. On a developer's machine, export the same two variables (with
   `remote_read_only = true`, or a SAS token with read permission only, when
   only CI should write).

## Private registries

With `[registries]` in smith.toml (see
[configuration](configuration.md#private-registries)), the pipeline signs in
before `smith ci setup`:

- **Azure Pipelines:** `npmAuthenticate@0` on the `.npmrc`, `PipAuthenticate@1`
  and `TwineAuthenticate@1` for the feed, `NuGetAuthenticate@1`. Each signs in
  with the pipeline's own token, so an Azure Artifacts feed needs no secret.
- **GitHub Actions:** `GITHUB_TOKEN`, with `packages: read` (or `packages:
  write` when a package publishes to GitHub Packages).
- **GitLab:** a `token` registry's variable, as a masked CI/CD variable. GitLab
  has no sign-in task, so use `auth = "token"` there.
- **CircleCI:** the same as GitLab: a `token` registry's variable, as a
  project environment variable.

A registry's credential reaches setup only when installs read from it.
Publishing to the feed drops `NODE_AUTH_TOKEN`, `PYPI_TOKEN` and
`NUGET_API_KEY` from the pipeline.

`smith ci check` lists what to set up by hand before the pipelines run: the
secrets to create, in each provider's terms, and for each registry its
permissions, such as the Feed Publisher role an Azure Artifacts feed must give
the project's build service.

## Ready-made pipelines

Beside the generated CI, `smith pipeline` adds a pipeline that does one thing:

| Template | Providers | What it does |
|---|---|---|
| `e2e` | GitHub, Azure | runs every project's `e2e` task, which `smith ci verify` leaves out, one at a time (`--parallel 1`: suites share ports and a display), on pull requests too, with the Playwright browsers cached (on Azure Repos, the pull-request run needs a build validation policy) |
| `package-zip` | GitHub, Azure | packs the projects (`smith ci pack`) and keeps `dist/drop` as the `drop` artifact |
| `deploy-pages --project <app>` | GitHub | builds the app and deploys its build output to GitHub Pages |
| `deploy-azure-function --project <app>` | GitHub, Azure | builds the app and deploys it to an Azure Function App (OIDC on GitHub, a service connection on Azure) |

```sh
smith pipeline                          # list them
smith pipeline e2e                      # for the workspace's providers; --ci github|azure|both
smith pipeline deploy-pages --project web
```

Each starts the way the generated jobs do (the toolchains, smith at the same
version, `smith ci setup`) and goes to `.github/workflows/<name>.yml` or
`azure-pipelines/<name>.yml`. The file is written once and yours from then
on: `smith upgrade` never rewrites it, and adding it again keeps it unless
`--force`.

## Your own steps

Put team steps between the slot markers; regeneration keeps them:

```yaml
      # smith:slot after-verify
      - name: End-to-end tests
        run: npm run e2e
      # smith:slot-end after-verify
```

Slots: `before-setup` and `after-verify` in the ci workflow, and `before-publish`
in the `pub.yml` workflow (see [Secrets](#secrets)).
