# Releasing

```sh
smith release --dry-run    # the plan; changes nothing
smith release --dry-run --comment   # in a pull request build: also post the plan on it (docs/reference/ci.md)
smith release              # asks before releasing (CI: smith ci release)
```

## What releases

Every project of kind `package`, and every app with a release section (see
[Apps](#apps)). For each, smith reads the commits since its last release tag that
touched its directory, or the directory of an internal `lib` it imports (a lib
ships inside what uses it, so a fix there releases its users):

| Commit | Bump |
|---|---|
| `feat:` | minor |
| `fix:`, `perf:` | patch |
| `!` after the type, or a `BREAKING CHANGE:` footer | major (minor before 1.0) |
| `docs`, `chore`, `test`, `ci`, `refactor`, `build`, `style` | nothing |

The first release starts from the manifest's version (`0.0.0`), so a first
`feat` gives `0.1.0`. A package whose dependency releases is released too (at
least a patch), so its published dependency range stays correct.

To force a version, commit a directive anywhere in the repository:

```
version(sdk)[1.0.0]: first stable release
```

## Apps

An app releases when `smith.toml` gives it a release section:

```toml
[projects.cli.release]
asset_name = "{product}_{version}_{os}_{arch}.{ext}"
```

It is versioned by its tag (`cli@1.2.0`) only. Its plugin builds the release
assets. Today only the Go plugin builds them: it cross-compiles every platform
with `CGO_ENABLED=0`, stamps the version into `main.version` (`-ldflags -X`),
zips each binary and writes `checksums.txt`, all under `dist/release/<app>/`.
An app in another language with a release section is tagged and gets a GitHub
Release, without assets. The settings are in
[Configuration](configuration.md#a-projects-release-settings).

## Container images

With the docker convention, an app with a `Dockerfile` releases too: the
release builds its image tagged with the version and `latest` and pushes both
to the registry `smith.toml` names.

```toml
[conventions.docker]
registry = "ghcr.io/acme"   # unset: the image task builds it, release never pushes

[projects.api.docker]
image = "acme-api"          # default: the project's name, lowercase
push = false                # build api's image, never push it
context = "apps/api"        # default: ".", the workspace
```

The image is built with the workspace as its context, after the app's
`build` and, for a Node app, its `prune`. That task writes the app's deploy
output into its build directory: a production `package.json` and lockfile,
and the workspace packages it uses in `workspace_modules/`. The Dockerfile the
convention writes for a new Node app installs from it with
`npm ci --omit=dev`. A Dockerfile written for its own directory as the context
(Nx's `docker build apps/api`) takes `context = "apps/api"`.

To build the image after a task of your own (a hand-written prune script, a
bundler), list what it waits for:

```toml
[projects.api.tasks.image]
depends_on = ["build", "bundle"]
```

That replaces the convention's list, so name everything the image needs;
`smith doctor` names a dependency the override leaves out.

`smith ci verify` builds the image of every app a pull request affects, after
lint, typecheck, test and build, so a broken Dockerfile fails the pull
request rather than the release. The image is built, never pushed. Where
Docker is not installed the image is skipped with a warning; a runner with
the CLI but no daemon (GitLab's default image) fails, so set
`verify = false` under `[conventions.docker]` there, or for one app under
`[projects.<app>.docker]`.

| Registry | Logs in with | The pipeline gets |
|---|---|---|
| `ghcr.io/<owner>` | `GITHUB_TOKEN` as `GITHUB_ACTOR` | `packages: write` on GitHub Actions, nothing to set; on Azure Pipelines set `GITHUB_TOKEN` to a token with `write:packages` |
| any other | `CONTAINER_USERNAME` and `CONTAINER_PASSWORD` | those two secrets |
| on this machine (`localhost`, `127.0.0.1`) | nothing | nothing |

A release that would push without its credentials stops before it starts and
names the setting. `smith release --dry-run` names each image it would push
(`would push ghcr.io/acme/acme-api:1.2.0 and ghcr.io/acme/acme-api:latest`).

`CONTAINER_REGISTRY` still names the registry when `registry` is unset, so a
pipeline that sets it keeps pushing; it is deprecated, and release says so.

## GitHub Releases

With `[release] github_releases = true`, every released tag gets a GitHub
Release named `<project> <version>`, and an app's assets are attached. It
needs `GITHUB_TOKEN` (or `GH_TOKEN`) with `contents: write`, which the
generated GitHub pipeline passes. A re-run completes a release that is missing
assets and never duplicates one.

The notes come from the releasing commits. Breaking changes are listed first,
with the text of their `BREAKING CHANGE:` footer. Then come Features, Fixes,
Performance and Reverts, each entry with its scope and short hash. A release
caused only by a dependency says so. The notes end by thanking the commits'
authors, bots aside. `smith release --dry-run --notes` prints them before
anything is released.

## Package managers

An app released with assets can also get the manifests package managers
install it from, generated from the archives and their checksums, and attached
to the GitHub Release beside them:

```toml
[projects.cli.release]
package_managers = ["homebrew", "scoop", "winget"]
description = "What the app does"
license = "MIT"
homepage = "https://cli.acme.dev"   # default: the GitHub repository
```

| Manager | Files | Where they go |
|---|---|---|
| Homebrew | `<product>.rb` (macOS and Linux, Apple silicon and Intel) | `Formula/` in your tap |
| Scoop | `<product>.json` (64-bit and arm64) | your bucket |
| winget | `<Owner>.<product>.yaml`, `.installer.yaml`, `.locale.en-US.yaml` (a portable zip) | a pull request to `microsoft/winget-pkgs` |

They point at the release's own archives with their SHA-256, so each one can
be copied into a tap, a bucket or winget-pkgs as it is. A manager whose
platforms the release has no archive for is skipped, and the release says so.
Native apps attach one archive per OS from separate jobs and get no manifests.

### Linux packages

`linux_packages` builds a .deb and an .rpm from each linux archive (amd64 and
arm64), with the binary in `/usr/bin`, and attaches them to the GitHub
Release with their SHA-256 in `checksums.txt`:

```toml
[projects.cli.release]
linux_packages = ["deb", "rpm"]
```

```sh
sudo apt install ./acme_1.2.0_amd64.deb
sudo dnf install ./acme-1.2.0-1.x86_64.rpm
```

They are built before the app is tagged, so a failure stops the release, and
a resumed release attaches any the GitHub Release lacks. A prerelease
(`1.2.0-rc.1`) is versioned `1.2.0~rc.1`, which both formats order before
`1.2.0`.

## npm, PyPI, NuGet and RubyGems

An app released with assets can also be published to npm, PyPI, NuGet and
RubyGems from the same archives, so people install it with the package
manager of their language, and nothing is downloaded after install:

```toml
[projects.cli.release]
npm = ["@acme/cli", "acme"]   # npm install --global acme
pypi = "acme"                 # pipx install acme, uvx acme
nuget = "acme"                # dotnet tool install -g acme
rubygems = "acme"             # gem install acme
```

| Registry | Packages | Secret |
|---|---|---|
| npm | One package per platform, named after the first name (`@acme/cli-linux-x64`, `@acme/cli-win32-arm64`, ...), holding the binary and limited to its `os` and `cpu`. Then every listed name as a package whose `bin` runs the binary that npm installed from its optional dependencies. | `NODE_AUTH_TOKEN` |
| PyPI | One wheel per platform, the binary in its scripts, so pip, pipx and uv put it on PATH with no Python in between. One Linux wheel serves glibc and musl. | `PYPI_TOKEN` |
| NuGet | One package per platform (`acme.linux-x64`, `acme.win-arm64`, ...) holding the binary, then the .NET tool package `acme`, which points at them. `dotnet tool install` (.NET 10 SDK and later) picks the platform's package and the command runs the binary itself, with no .NET runtime. Set `NUGET_PUSH_URL` to push to another feed's v2 endpoint. | `NUGET_API_KEY` |
| RubyGems | One platform gem per platform (`x86_64-linux`, `x86_64-linux-musl`, `arm64-darwin`, `x64-mingw-ucrt`, ...), holding the binary and a Ruby launcher that runs it; RubyGems installs the gem matching the machine. | `GEM_HOST_API_KEY` |

They are published after the archives are built and before the tag, platform
packages before the packages that depend on them. A version a registry
already has is skipped, so a re-run completes a release. On GitHub Actions
npm packages carry [provenance](https://docs.npmjs.com/generating-provenance-statements):
the generated workflow gets `id-token: write` for it. Publishing to npm needs
the `npm` CLI on the runner (GitHub's and Azure's hosted images have it). A
prerelease goes to npm's `next` tag, to NuGet as is, to PyPI as `1.0.0rc1`
(only `-alpha.N`, `-beta.N` and `-rc.N` have a Python spelling), and to
RubyGems as `1.0.0.pre.rc.1`. Native apps are not published
this way.

### pub.dev

`pub = "acme"` publishes the app to pub.dev, so `dart pub global activate
acme` (or `flutter pub global activate acme`) installs it. The package
carries every platform's binary under `lib/src/bin/` and a launcher in
`bin/` that runs the one for the running platform, so nothing is downloaded
after activation.

pub.dev takes no token from CI: its automated publishing accepts a package
only from a GitHub Actions run on a tag that matches a pattern set on
pub.dev, and its patterns cannot contain the `@` of smith's tags. So after
the GitHub Release, the release:

1. tags the release commit with the pub tag, `<name>-v<version>`
   (`acme-v1.4.0` for `acme@1.4.0`; change it with `pub_tag`, using `{name}`
   and `{version}`), and pushes it;
2. starts `.github/workflows/pub.yml`, which `smith ci generate` writes, on
   that tag. A tag pushed with the workflow's own token starts no workflow,
   but a `workflow_dispatch` does, so the release job gets `actions: write`;
3. the pub workflow runs `smith release pub --project <app> --version
   <version>`: it downloads the archives from the GitHub Release, builds the
   package and runs `dart pub publish`, authenticated by `setup-dart` with
   GitHub's OIDC token.

A version pub.dev already has is skipped, so a re-run completes a release.
The release summary says when it started a publish; the pub workflow's run
says how it went. It works only on GitHub Actions with `[release]
github_releases = true`; on Azure Pipelines or GitLab the release says so
and skips it.

Once per package, on pub.dev:

- Publish the first version by hand (`dart pub publish` asks you to sign in).
- On the package's **Admin** tab, under **Automated publishing**, enable
  publishing from GitHub Actions for the repository, with the tag pattern
  `<name>-v{{version}}` (or your `pub_tag`), and tick **Enable publishing
  from workflow_dispatch events**.

pub.dev also needs a description of 60 to 180 characters (the release's
`description`, cut at a word when longer) and a `LICENSE` at the workspace
root.

## Tags

| Language | Tag | Why |
|---|---|---|
| every language but Go, and every app | `<name>@<version>` | override with `tag_prefix` |
| Go packages | `<dir>/v<version>` (e.g. `packages/gokit/v0.1.0`) | how `go get` finds a nested module's version |

The model is **tag-only**: nothing is committed back to the branch. For each
package, in dependency order, smith:

1. writes the version into the manifest (keeping its formatting), and the new
   versions of workspace dependencies into its dependency ranges
2. builds the package
3. publishes it — skipping a version the registry already has
4. restores the manifest
5. tags the commit and pushes the tag
6. pushes the package to its split repository, when it has one

## Split repositories

Some registries read a package at the root of a repository: Packagist reads
`composer.json` there and follows the repository's tags. A package in a
monorepo's subdirectory cannot be listed that way, so it is mirrored to a
read-only repository of its own:

```toml
[projects.client-php]
split_repo = "acme/client-php"   # a GitHub owner/repo
```

Each release splits the project's directory out of the history at its tag
with `git subtree split`, pushes it to the mirror's `main` and tags it
`v<version>` there (in one `--atomic` push), the spelling Composer expects.
The monorepo keeps its own tag. Register the mirror with the registry once;
later releases reach it by its tags.

- `SPLIT_REPO_TOKEN` is a token that can push to the mirror (`GITHUB_TOKEN`
  cannot: it is scoped to the repository the workflow runs in). It is sent
  as an HTTP header through git's environment, never on a command line, and
  the generated pipelines pass it to the release step. A release refuses to
  start without it.
- The split is taken from the pushed tag: a release whose tag was not pushed
  is not mirrored yet, and the re-run that pushes the tag mirrors it.
- A mirror that already has the tag is left alone, so a re-run resumes:
  `main` and the tag are pushed together, so the tag means the mirror is done.
- The mirror is read-only. The same history always splits into the same
  commits, so each release fast-forwards its `main`; a commit made there by
  hand stops the push, and the release says to reset the mirror.
- Splitting needs the full history, as a release does anyway (`fetch-depth: 0`),
  `git subtree` (part of git's contrib, which some git builds leave out) and
  git 2.31 or newer: the release checks all three, and the token, before it
  starts.
- The preview says what it would mirror (`would mirror packages/client-php to
  acme/client-php@v1.2.0`, and `Would mirror client-php to …` in the report).

## Tag checks

Versions come from tags, so before showing a plan, even for `--dry-run`,
`smith release` fetches the tags from the release remote and checks them for
every releasable project. A failed fetch is a warning. Each of these stops the
release, because the plan's versions would be wrong:

| Problem | Why it matters | Fix |
|---|---|---|
| The project's newest tag is under another name (its package name, scoped or not), not its tag prefix | the release would start from scratch, or from an older version | set `[projects.<name>] tag_prefix`, or tag the same commit under the current prefix |
| The registry has a version, but the project has no tag | the release would start from the manifest | `git fetch --tags`; if the tag really does not exist, tag the commit that published it |
| The newest tag is lower than the registry's version | the release would propose a version the registry refuses | tag the commit that published the registry's version |

The registry is asked through the project's plugin (`npm view` for Node, the
PEP 503 index for Python, and likewise crates.io, NuGet, RubyGems, pub.dev,
Packagist and the Maven repository; a Go module's tags are its registry, and
C++ and Bicep projects have none). An unreachable
registry is a warning, not a stop. `smith doctor` runs the same checks without
fetching.

## When something fails

The release stops at the failing package; everything before it is tagged and
pushed. Re-run it: tagged packages have no new commits and are kept, and a
package that was published but not tagged is skipped by the registry check and
then tagged. Tags left local by an interrupted push are pushed by the next run.

Before a real release smith refuses a shallow clone, a branch other than
`[release] branch`, uncommitted changes, and a missing publish credential —
before anything is tagged.

## The report

Every run writes `.smith/release-report.json` (also printed by `--json`) and, on
GitHub Actions, a summary table:

```json
{ "dryRun": false, "results": [
  { "project": "sdk", "current": "0.1.0", "next": "0.2.0", "tag": "sdk@0.2.0",
    "reason": "2 releasing commit(s)", "status": "released",
    "published": true, "registry": "https://registry.npmjs.org/", "pushed": true } ],
  "unchanged": ["gokit"] }
```

There are no changelog files: the release notes go to the GitHub Release.
