# CI pipelines

smith generates the pipeline from what is in the workspace and keeps it in step:
`smith new` writes it, every `smith add` regenerates it, and `smith ci generate`
does it on demand. `smith doctor` (and `smith ci generate --check`) fails when the
committed pipeline is out of date.

## How it is built

Each plugin contributes, for its own projects:

- **toolchains** — the runtimes to install, at versions the plugin pins. The
  plugin says how: a GitHub action and an Azure task with their inputs
  (`actions/setup-node@v7`, `UsePythonVersion@0`, …), and a shell command for
  everything else (GitLab installs every toolchain this way). smith itself knows no runtime, so a plugin for a new
  language installs it the same way; a runtime no plugin says how to install
  is left as a comment in the pipeline, to fill through the `before-setup`
  slot.
- **setup** — e.g. `npm ci`, the Python venv and editable installs, `go mod download`
- **audit** — `npm audit` (blocking only on a production advisory at moderate
  or above that `npm audit fix` can apply; advisories with no fix, a
  semver-major-only fix, or only in dev dependencies are printed as notes),
  `pip-audit` and `govulncheck` (reported)
- **credentials** — the secrets publishing needs (`NODE_AUTH_TOKEN`, `PYPI_TOKEN`,
  `CARGO_REGISTRY_TOKEN`, …; see [Secrets](#secrets))

smith merges them into one provider-neutral model, and a renderer per provider
writes the YAML. Every provider carries the same phases, in the same order.

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
| `smith ci verify` | `lint`, `typecheck`, `test`, `build` — affected projects only on a pull request, everything otherwise; then the tasks a plugin marks for verify, such as the docker convention's `image` (skipped, with a warning, where their program is not installed). It fails when the tasks left a tracked file changed, naming the file and the project whose tasks ran, since the release refuses a dirty tree (in CI, a file setup changed fails it too; locally, your own edits do not). `smith doctor` warns about the usual causes: a tracked file `.gitignore` ignores, and a tracked file inside a task's outputs |
| `smith release --dry-run --comment` | on a pull request: the release preview, in the job summary and as a comment on the pull request |
| `smith ci pack` | on a push to the release branch: every project's `package` task into `dist/drop`, kept as the run's `drop` artifact |
| `smith ci native` | in a job per OS, after the ci job: lint, test, build and package the projects tagged native |
| `smith ci release` | on a push to the release branch: version, publish and tag; then the post-release steps plugins declare, whether or not anything was released (the docker plugin's `smith docker description`, when an app sets `description_from` for a `docker.io` registry), each only a warning when it fails |

On GitHub Actions and Azure Pipelines output is grouped, failures are annotated,
and each phase adds a line (passed or failed, and how long it took) and the
release its report to the job summary.

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

On Azure Pipelines a workspace with a native project gets a `jobs:` pipeline:
the ci job, and the native job with a matrix of runner images. Steps kept in
the `smith:slot` markers move with the job.

## Installing smith in CI

Generated pipelines install the smith version that generated them with the
install script (`SMITH_VERSION`), so a pipeline does not change behaviour until
you regenerate it with a newer smith.

## Secrets

| Secret | Used for |
|---|---|
| `NODE_AUTH_TOKEN` | `npm publish` (the generated `.npmrc` reads it), and an app's npm packages |
| `PYPI_TOKEN` | twine upload as `__token__` (or set `TWINE_USERNAME`/`TWINE_PASSWORD` and `TWINE_REPOSITORY_URL` for another index), and an app's wheels |
| `NUGET_API_KEY` | `dotnet nuget push` of .NET packages, and an app's .NET tool packages (a nuget.org API key that can push the tool's package and `<name>.*`) |
| `GEM_HOST_API_KEY` | `gem push` of Ruby gems, and an app's platform gems (rubygems.org, or `RUBYGEMS_HOST`) |
| `CARGO_REGISTRY_TOKEN` | `cargo publish` of Rust crates |
| `PUB_TOKEN` | a Dart package whose `publish_to` names another repository; pub.dev itself publishes with GitHub's OIDC token |
| `MAVEN_REPOSITORY_URL`, `MAVEN_USERNAME`, `MAVEN_PASSWORD` | `mvn deploy` and Gradle's `maven-publish` of JVM packages |
| `VSCE_PAT` | `vsce publish` of a VS Code extension package to the Marketplace (on the release step, and on the native job's step for a native extension) |
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
from the tags (of a split repository, with `split_repo`, when the package lives in a subdirectory), and C++ and Bicep projects release by their tag only. On GitHub, the job
has `contents: write` so it can push tags; on Azure the checkout keeps its
credentials (`persistCredentials`).

An app published to pub.dev needs no secret either. With one, GitHub gets a
second workflow, `.github/workflows/pub.yml`, which the release starts on the
app's pub tag (the ci workflow gets `actions: write` for that). It sets up
Dart with pub.dev's OIDC token and runs `smith release pub`; see
[Releasing](releasing.md#pubdev) for the one-time setup on pub.dev.

## The remote cache

With `[cache] remote` set (see [Tasks, cache and affected](tasks-and-cache.md#remote-cache)),
the pipelines pass the cache's key and its backend's credentials to every step
that runs tasks: `smith ci verify`, `smith ci pack`, `smith ci native` and the
release. The audit and the setup run no cached task and get none. Which
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
  `pull_request`, Azure `Build.Reason` of `PullRequest`, GitLab merge requests)
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
  GitLab's CI/CD variables are already environment.
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

Slots: `before-setup`, `after-verify`.
