# Releasing

> Terms such as project, kind, task and stack are defined in the [glossary](glossary.md). How it works inside: [release and CI internals](../architecture/release-and-ci-internals.md). Things that go wrong: [troubleshooting](../guides/troubleshooting.md#a-release-did-not-do-what-i-expected).

```sh
smith release --dry-run    # the plan; changes nothing
smith release --dry-run --comment   # in a pull request build: also post the plan on it (docs/reference/ci.md)
smith release              # asks before releasing (CI: smith ci release)
smith release --status     # the release journal: what was published where
smith release --rehearse   # ask every registry whether the release would be accepted; publishes nothing
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

Packages publish to their ecosystem's public registry (npmjs.com, PyPI,
nuget.org, …), or to the private registry smith.toml names with `[release]
registry` or `[projects.<name>] registry` (npm, Python, .NET and Maven, where
`maven = "github"` publishes to GitHub Packages with the job's token and no
secrets, and consumers need their own token to read it; see
[configuration](configuration.md#private-registries)). Either way a version
the registry already has is skipped, so a re-run completes a release.

An app is never published to a registry, in any language: it is released by its tag, its
GitHub Release with the files it attaches (`build`, `assets`, `platforms`; a plugin's own
archives) and its deploy plugins. No `registry = "none"` is needed for it, and a Python
app, a Rust binary crate and a .NET app behave alike. To publish a program to a registry,
make it a package (`kind = "package"`), or give it an npm, PyPI, NuGet or RubyGems
distribution under `[projects.<name>.release]`.

A package that a registry builds from its tag (JitPack for Maven, Packagist,
Go modules) sets `registry = "none"` under `[projects.<name>]` or `[release]`:
the release versions it, tags it and creates the GitHub Release, publishes to
no registry, and no credentials are listed for it. The summary says
"published by its tag". JitPack uses the tag name as the version, so
`bare_tags = true` on that one project tags `0.1.0` instead of `java@0.1.0`
(or `tag_prefix = "v"` for `v0.1.0`). Only one project of a workspace can use
bare tags, since bare tags of two projects could not be told apart.

For JitPack, smith only tags: JitPack builds the artifact itself from a
`jitpack.yml` at the repository root, for example
`install: mvn -pl packages/java -am install`. Its own Maven is older than
3.6.3, so a build that needs a newer one installs it first in that file (for
example through SDKMAN: `before_install: sdk install maven 3.9.11`).
Consumers use the tag as the version.

A project changes between internal `lib`, published `package` and `app` with
[`smith kind`](../manual/kind.md): it records the kind, drops or restores the publish
guard, moves the folder and updates what names it. `smith kind core package
--dry-run` lists every edit first. A package that was released is refused (the
newest tag is named) unless `--force`, because its published versions would never
be updated; `smith release --dry-run` shows the project in the plan once it is a
package and out of it once it is not.

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
assets, one zip per platform and a `checksums.txt`, all under
`dist/release/<app>/`. `platforms` are `os/arch` pairs in Go's spelling
(`linux/amd64`, `windows/arm64`, `darwin/arm64`, the six of the default) whatever
the language, and `asset_name` names each file the same way. The default is these
six: `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`
and `windows/arm64`. A **native** app (cgo; Go, Rust) is the exception: `platforms`
does not apply to it. Each runner of `smith ci native` builds its own platform
only, so with the default `[ci] native_runners` it gets three archives
(`linux/amd64`, `windows/amd64`, `darwin/arm64`). Add `ubuntu-24.04-arm`,
`windows-11-arm` or `macos-15-intel` to `native_runners` for the other three.

| Language | What the plugin builds |
|---|---|
| Go | Cross-compiles every platform with `CGO_ENABLED=0`, stamps the version into `main.version` (`-ldflags -X`), zips each binary. |
| Rust | Runs `cargo build --release -p <crate> --target <triple>` per platform (the version is the one in the crate's `Cargo.toml`, which the release has set, so `env!("CARGO_PKG_VERSION")` and `--version` agree with the tag), and zips every program the crate builds with the README and licence files (the crate's, else the workspace's). |
| .NET | Runs `dotnet publish -r <rid> --self-contained -p:PublishSingleFile=true -p:Version=<version>` per platform (win-x64, linux-arm64, osx-arm64, ...) and zips the output with the README and licence files. Every platform is built on one machine; each needs its runtime downloaded from NuGet once. |

An app in another language with a release section is tagged and gets a GitHub
Release, without assets, unless it builds them with a command (next section).
The settings are in
[Configuration](configuration.md#a-projects-release-settings).

### Release files from a command

Any language, and any tool, can make an app's release files: name the command
and the files it leaves.

```toml
[release]
github_releases = true

[projects.music-fixer.release]
build     = "vcw app build examples/apps/music-fixer/app.json --out out --zip"
assets    = ["out/*.zip", "out/*.zip.sha256"]
platforms = ["windows-x64", "macos-arm64", "macos-x64", "linux-x64", "linux-arm64"]
```

`build` is a shell command, run at the workspace root once on a runner of each
platform, and `assets` are workspace-relative globs of what it leaves. The
command sees `SMITH_PLATFORM` (one of the five names above), `SMITH_VERSION`
and `SMITH_TAG`, so it can name its files after the platform. This replaces what
the project's plugin would build: there are no aliases, package manager
manifests, Linux packages or npm, PyPI, NuGet, RubyGems and pub packages for
it, and smith refuses a `release` table that sets them next to `build`, that has
`build` without `assets` or `platforms`, `assets` without `build`, a platform it
does not know, an `asset_name` placeholder the files do not have, or no
`github_releases`. Each message says how to fix it. (When the project's plugin
supplies `build`, `assets` and `platforms`, the table may leave them out; see
below.)

The release job tags the app and creates its GitHub Release, without files.
The generated GitHub Actions workflow then has an `assets` job that runs
`smith ci assets` on one runner per platform (`windows-latest`, `macos-latest`
for arm64, `macos-15-intel`, `ubuntu-latest`, `ubuntu-24.04-arm`) after the
release. Each runner runs the build, then uploads its own files with
`gh release upload --clobber`, so a re-run replaces a file and no job uses
`actions/upload-artifact` (no artifact storage is spent). Give the files names
that differ by platform: two runners uploading the same name replace each other.

A glob in `assets` may match nothing on a platform, as `dist/drop/app_*.dmg` does
on Windows and Linux, as long as another glob of the list matches a file there:
the runner prints `no file matches <glob> on <platform>: skipped` and uploads the
rest. When no glob matches, the build left nothing and the job fails with
`the build left no file matching assets "..."`, so a build that broke cannot pass
as a platform without that file. See
[ADR 0069](../architecture/decisions/0069-the-assets-phase-can-run-again-for-a-tag.md).

Two settings tidy the files on upload. `asset_name` renames each file, with the
placeholders `{product}` (default: the project name), `{version}`, `{platform}`
(`windows-x64`) and `{ext}`; two files that get the same name keep the name of
the folder they were in (`app-1.0.0-windows-x64-nsis.exe`), so a runner never
uploads a name twice. `checksums = true` adds a `<product>-<version>-<platform>.sha256`
file (`sha256sum` format) of what that runner uploaded. Both rename and
checksum copies kept under `dist/release/<project>-assets/<platform>`. Before the
build, each runner also writes the release's version into the project's version
files (`package.json`, and the files a plugin keeps a version in, such as a Tauri
app's `tauri.conf.json` and `Cargo.toml`), so what it builds carries the version
it uploads; the files are put back afterwards.

`install = false` says the build command needs none of the project's dependencies
(it only zips files, say). The assets job then runs `smith ci setup --assets`,
which asks the plugins to install every project except those, so a dependency
conflict that shows on one runner only (a Windows resolver failure for a Python
package that is only zipped) cannot stop the job before the build. If no project
of a plugin is left, that plugin adds no setup step, its toolchain included. The
`ci` job and the native job install every project. `install` needs `build`.
See [ADR 0068](../architecture/decisions/0068-the-assets-job-can-skip-installing-a-project.md).

A plugin can supply the settings a table leaves out. The `tauri` plugin does for
a Tauri app, so `[projects.studio.release]` with nothing in it is enough (with
`github_releases = true`): see [the tauri plugin](../plugins/using/catalogue/tauri.md#releasing-the-installers).
What the table sets wins, field by field. A table that sets only things the
plugin's archives use (`os_alias`, `package_managers`, `npm`, ...) is left to the
plugin's archives.

`smith release --dry-run` says which command and files each release has.
`smith release assets` runs the build and the upload for the platform you are on
(`--platform` to name it), for a release made by hand. Not covered: GitLab CI,
Azure Pipelines, CircleCI and Bitbucket Pipelines do not run this, and
`smith ci generate` says so when one of them is configured; run
`smith release assets` on each platform there. The checksums file, SBOM,
attestation and platform packages that smith builds from a plugin's archives are
not made for these files; have `build` write what you need.

### Cross-building, and apps built on each system

A program is linked against the system it runs on, so one machine builds only
some platforms. For Rust, the release machine builds its own system and its own
processor, the other Windows or Mac processor with just the Rust target
(`rustup target add`), and the other Linux processor once a linker is installed
(`gcc-aarch64-linux-gnu` for `linux/arm64`; or set the linker in
`.cargo/config.toml` or `CARGO_TARGET_<TRIPLE>_LINKER`). A platform of another
operating system is refused, in `smith release --dry-run` too, with the way out:

```
cli: this machine (linux/amd64) cannot build windows/amd64: a Rust program is built
on the system it is for. Mark the app native (native = true under
[package.metadata.smith] in its Cargo.toml) and each runner of `smith ci native`
builds its own archive
```

A **native** app is the way to get every platform: `smith ci native` runs a job
per runner (`ubuntu-latest`, `windows-latest` and `macos-latest` by default,
`[ci] native_runners` to add `ubuntu-24.04-arm`, `windows-11-arm` or a macOS
Intel runner for the other processors), each builds only its own platform and
attaches its zip and a `<zip>.sha256` file to the GitHub Release the release job
created. Its dry run lists what this machine's native job would attach. A native
app has no combined `checksums.txt` (builds made on different runners cannot be
summed in one place), no package manager manifests and no npm, PyPI, NuGet or
RubyGems packages.

`smith doctor` checks the same things beforehand: every platform has a Rust
target, the target is installed (the fix runs `rustup target add`), the Linux
linker is there, and an app that cannot be built here is marked native.

## Container images

With the docker convention, an app with a `Dockerfile` releases too: the
release builds its image tagged with the version and `latest` (and the commit and branch, see below) and pushes them
to the registry, or each of the registries, `smith.toml` names.

```toml
[conventions.docker]
registry = "ghcr.io/acme"   # unset: the image task builds it, release never pushes

[projects.api.docker]
image = "acme-api"          # default: the project's name, lowercase
push = false                # build api's image, never push it
context = "apps/api"        # default: ".", the workspace
```

### Several registries

`registries` replaces `registry` (setting both is an error) to publish the
same image in more than one place, for visibility on Docker Hub, say:

```toml
[conventions.docker]
registries = [
  "ghcr.io/open-craw",                  # GITHUB_TOKEN, packages: write
  "docker.io/opencraw",                 # DOCKERHUB_USERNAME, DOCKERHUB_TOKEN
  { url = "registry.acme.dev/team", username_env = "ACME_USER", password_env = "ACME_TOKEN" },
]
```

- Each entry is `host/namespace`, with the host: a name with no dot is a
  Docker Hub namespace to Docker, so Docker Hub is written `docker.io/acme`.
- A string entry logs in the way its host does (the table below). An entry
  as a table names its own variables, both of them, whatever the host.
- Several registries on hosts with no default would all read
  `CONTAINER_USERNAME` and `CONTAINER_PASSWORD`, so all but one must be a table.
  Two namespaces on `docker.io` share `DOCKERHUB_*`, which is fine.
- The image is built once and pushed under the same name and tags (the
  version, `latest`) to each registry. `image` names it under every one.
- `[projects.<app>.docker] registries` (or `registry`) replaces the
  convention's list for that app.
- Every registry's secrets are in the generated pipeline and in
  `smith ci check`, and a release missing one stops before it starts.

A push that fails names its registry, and the others are still tried. A
re-run asks each registry whether it has the version's tag
(`docker manifest inspect`) and pushes only to those that do not, so a
release never pushes twice to the same place. Per registry `latest` is pushed
first and the version last, so the version tag shows the registry is complete.

### Tags for deployment platforms, and digests

Deployment platforms and GitOps agents (Keel, Argo CD, Flux) follow the commit
or the branch, not a version. Two settings add those tags to every image, in
`[conventions.docker]` or per app in `[projects.<app>.docker]`:

```toml
[conventions.docker]
tag_commit = true     # also sha-<first 7 characters of the commit>
tag_branch = true     # also the release branch's name (main); / and other odd characters become -
```

The image is built once with all its tags and each registry gets them in this
order: `sha-…`, the branch, `latest`, the version last. A tag smith cannot name
(no commit, no branch) is left out with a warning.

After the push smith reads the manifest digest each registry holds and names it
in the release output, so a deploy step or a GitOps manifest update can pin
`ghcr.io/acme/api@sha256:…` instead of a tag that moves:

- the log says `docker published 1.2.0 to ghcr.io/acme`, then one line per
  registry: `ghcr.io/acme/api@sha256:… [tags: sha-0123456, latest, 1.2.0]`;
- the release summary lists them under "Images pushed for api";
- `smith release --json` and `.smith/release-report.json` carry them as
  `results[].images`: `{"reference": "ghcr.io/acme/api", "tags": [...], "digest": "sha256:…"}`.

A digest smith could not read is left out with a warning (the image is pushed
already).

The directories a platform should watch per app (Railway and Render's watch
paths, a GitOps path filter) come from `smith affected --paths` and the
`watch` object of `smith affected --json` (see
[Tasks, cache and affected](tasks-and-cache.md#watch-paths-for-deployment-platforms)).

Only Node apps have a `prune` task. It exists because a Node image must install
production dependencies from a pruned manifest. A Go app's image copies the
binary its `package` task builds, and a Python app's the wheel, so neither needs
a slimmed copy of the workspace; a Dockerfile written for them copies from
`dist/drop` (see [Packing](ci.md#packing)).

### The Docker Hub description

Docker Hub shows a repository's description on its page, and a pushed image
does not set it. `description_from` names a file whose text becomes the
description of the image's repository on every `docker.io` registry, after the
push (a release that pushed nothing new leaves it alone):

```toml
[conventions.docker]
registries = ["ghcr.io/open-craw", "docker.io/opencraw"]
description_from = "README.md"        # or per app: [projects.api.docker] description_from = "apps/api/README.md"
```

It signs in to Docker Hub's API with the registry's own credentials
(`DOCKERHUB_USERNAME`, `DOCKERHUB_TOKEN`). The description is limited to 25,000
characters, and longer text is cut. Editing a description may need an access
token with Read, Write & Delete permission, which a plain push does not: the
images are already published when this runs, so a refusal, a missing file or
missing secrets is a warning in the release log and the release goes on. Other
registries ignore the setting.

A docs-only commit releases nothing, so the release alone would never send a
changed file. `smith docker description` sets the description on its own:

```sh
smith docker description               # every app with a description_from file or a short description
smith docker description -p api        # one app
smith docker description --dry-run     # name the repositories and both descriptions, call nothing
smith docker description --strict      # exit non-zero if a description could not be set
```

The generated pipeline runs it at the end of `smith ci release`, on every push
to the release branch, whether or not something was released, when an app has a
`description_from` file or a short description and a `docker.io` registry. That is safe because
setting a description is an idempotent PATCH: sending the same text again
changes nothing, so smith keeps no record of what was last sent and needs no
new state in git. A failure is a warning (the images are published already),
as in the release.

The short description is the one line under the image's name on Docker Hub,
in search results and on the repository card. It is sent in the same call as
the full one. `short_description` sets it, per convention or per app; without
it smith uses the project's own description, from the manifest
(`package.json`, `pyproject.toml`, `pom.xml`, `Cargo.toml`, `pubspec.yaml`,
`composer.json`), so a project that has one needs no setting:

```toml
[conventions.docker]
short_description = "Crawl the web with recipes: the OpenCraw engine over HTTP."
```

Docker Hub takes at most 100 characters. A longer text is put on one line,
cut at a word with `…`, and the release log warns; `smith doctor` reports a
`short_description` over 100 (a long manifest description is cut without a
finding, as it is not a setting). An app with neither a setting nor a
description leaves the short description on Docker Hub alone, and so does an
app with no `description_from` leave the full one.

### The image and its login

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
request rather than the release. The image is built, never pushed, under the
name release pushes (the first registry's, `ghcr.io/acme/acme-api`) with the
local tags `local` and `latest`; it never carries the manifest's version. Where
Docker is not installed the image is skipped with a warning; a runner with
the CLI but no daemon (GitLab's default image) fails, so set
`verify = false` under `[conventions.docker]` there, or for one app under
`[projects.<app>.docker]`.

| Registry | Logs in with | The pipeline gets |
|---|---|---|
| `ghcr.io/<owner>` | `GITHUB_TOKEN` as `GITHUB_ACTOR` | `packages: write` on GitHub Actions, nothing to set; on Azure Pipelines set `GITHUB_TOKEN` to a token with `write:packages` |
| `docker.io/<namespace>` (Docker Hub) | `DOCKERHUB_USERNAME` and `DOCKERHUB_TOKEN` | those two secrets (a Docker Hub access token) |
| any other | `CONTAINER_USERNAME` and `CONTAINER_PASSWORD`, or the variables the entry names | those secrets |
| on this machine (`localhost`, `127.0.0.1`) | nothing | nothing |

A workspace that pushed to `docker.io` with a single `registry` and
`CONTAINER_*` secrets now reads `DOCKERHUB_*`: set those, or keep the old
names with `{ url = "docker.io/acme", username_env = "CONTAINER_USERNAME",
password_env = "CONTAINER_PASSWORD" }`.

A release that would push without its credentials stops before it starts and
names the setting. `smith release --dry-run` names each image it would push
(`would push ghcr.io/acme/acme-api:1.2.0, ghcr.io/acme/acme-api:latest`, then
the same for every other registry); it does not ask the registries what they
have.

`CONTAINER_REGISTRY` still names the registry when `registry` is unset, so a
pipeline that sets it keeps pushing; it is deprecated, and release says so.

## GitHub Releases

With `[release] github_releases = true`, every released tag gets a GitHub
Release named `<project> <version>`, and an app's assets are attached. It
needs `GITHUB_TOKEN` (or `GH_TOKEN`) with `contents: write`, which the
generated GitHub pipeline passes. A re-run completes a release that is missing
assets and never duplicates one. Another repository's CI downloads a private
repository's files with a fine-grained token:
[Private release files](../guides/private-release-assets.md).

The notes come from the releasing commits. Breaking changes are listed first,
with the text of their `BREAKING CHANGE:` footer. Then come Features, Fixes,
Performance and Reverts, each entry with its scope and short hash. A release
caused only by a dependency says so. The notes end by thanking the commits'
authors, bots aside. `smith release --dry-run --notes` prints them before
anything is released.

### SBOM and build provenance

Every app released with a GitHub Release also gets a CycloneDX 1.5 SBOM,
`<product>_<version>_sbom.cdx.json`, attached next to its archives. smith
writes it from what the project's plugin lists: each language plugin answers
the `sbom/components` method of the plugin protocol with the project's
components as package URLs, read from its lockfile where the ecosystem has one
and from its manifest otherwise. Go lists the modules in `go.mod` (direct and
indirect), npm the packages in the nearest `package-lock.json` (the project's
or the workspace's), falling back to the ranges in `package.json`; Python
(`uv.lock`, `poetry.lock`), Rust (`Cargo.lock`), .NET (`packages.lock.json`),
Ruby (`Gemfile.lock`), PHP (`composer.lock`), Dart (`pubspec.lock`), the JVM
(`gradle.lockfile`, `pom.xml`) and C++ (`vcpkg.json`, `conanfile.txt`) follow
the same rule, each falling back to its manifest. A project that declares no
dependency gets an SBOM that says so. smith does not read lockfiles itself: an
app whose plugin does not implement `sbom/components` (a plugin written before
the method, or one that has no packages to list) gets no SBOM, and the release
carries on without one. Each plugin's page says which files it reads.

On GitHub Actions the generated workflow then attests everything in
`dist/release` (archives, Linux packages, checksums and the SBOM) with
[`actions/attest-build-provenance`](https://github.com/actions/attest-build-provenance),
SLSA build provenance signed through GitHub's Sigstore. The workflow gets
`id-token: write` and `attestations: write` for it. Anyone can check a
download:

```sh
gh attestation verify acme_1.2.0_linux_amd64.tar.gz --repo acme/acme
```

Azure Pipelines and GitLab CI produce the SBOM too, since smith writes it,
but have no attestation service: the provenance is not available there, and
the SHA-256 `checksums.txt` is what verifies a download. Native apps, whose
archives are attached by the per-OS jobs, get neither yet. Registry
publishes keep their own provenance (npm provenance above, PyPI attestations
through trusted publishing).

## Package managers

An app released with assets can also get the manifests package managers
install it from, generated from the archives and their checksums, and attached
to the GitHub Release beside them:

```toml
[projects.cli.release]
package_managers = ["homebrew", "scoop", "winget", "aur", "nix"]
description = "What the app does"
license = "MIT"
homepage = "https://cli.acme.dev"   # default: the GitHub repository
```

| Manager | Files | Where they go |
|---|---|---|
| Homebrew | `<product>.rb` (macOS and Linux, Apple silicon and Intel) | `Formula/` in your tap |
| Scoop | `<product>.json` (64-bit and arm64) | your bucket |
| winget | `<Owner>.<product>.yaml`, `.installer.yaml`, `.locale.en-US.yaml` (a portable zip) | a pull request to `microsoft/winget-pkgs` |
| AUR | `<product>-bin.PKGBUILD` and `<product>-bin.SRCINFO` (x86_64 and aarch64, from the Linux archives) | the AUR git repository `<product>-bin`, as `PKGBUILD` and `.SRCINFO` |
| Nix | `<product>.flake.nix` (x86_64 and aarch64, Linux and macOS) | `flake.nix` in a repository, then `nix run github:<owner>/<repo>` |

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

#### RubyGems without a key: trusted publishing

A RubyGems account that asks for a one-time password on pushes (its MFA level
is "UI and API"), or a gem with `rubygems_mfa_required`, refuses a push made
with `GEM_HOST_API_KEY` from CI. Trusted publishing replaces the key with the
workflow's OIDC token:

```toml
[release]
rubygems_trusted_publishing = true     # GitHub Actions only
```

On each gem's page on rubygems.org, under **Trusted publishers**, add the
repository, the workflow `ci.yml` and no environment (for a gem that does not
exist yet, add a *pending* trusted publisher from your account's settings).
With the setting on, `smith ci generate` gives the workflow `id-token: write`,
adds a `Sign in to RubyGems` step (`rubygems/configure-rubygems-credentials`)
right before the release step, only on a push to the release branch, and no
longer passes `GEM_HOST_API_KEY` as a secret; the step exports a short-lived
key that `gem push` and smith's own gem uploads use. The release does not ask
for the secret either. The setting needs `providers = ["github"]`.

### VS Code extensions

An extension is an app: nothing imports it, and its version is what it
contains. It is versioned by its tag like any app, so it needs a release
section, and it takes the `vscode-extension` stack (or `[conventions.vscode]`),
whose plugin publishes it:

```toml
[projects.vscode.release]
```

The release stamps its version into the manifest of the build
(`dist/<dir>/package.json`; the manifest in the repository is not changed),
packs `dist/release/<name>/<name>.vsix`, publishes it to the Visual Studio
Marketplace and, when `OVSX_PAT` is set, to Open VSX. Both skip a version they
already have, so a re-run completes a release. Install `@vscode/vsce` (and
`ovsx` for Open VSX) as a devDependency: smith runs them with `npx
--no-install`. The extension must build into `dist/`.

An extension with native modules needs one .vsix per platform: mark it
`"smith": {"native": true}` in its package.json. The release job then only tags it,
and `smith ci generate` gives the workflow a job on each OS that builds and
publishes that platform's .vsix (`vsce package --target <platform>`), signing in
to Azure first when `vsce_entra` is set. See [CI pipelines](ci.md).

An extension that is a package (`"smith": {"kind": "package"}`) is still
published to the Marketplace by the node plugin, as before.

### The Marketplace without a token: Microsoft Entra ID

A VS Code extension (an app, or a package) is published to the Visual Studio Marketplace by
`vsce publish`, with the token in `VSCE_PAT`. Azure DevOps retires its global
personal access tokens on 1 December 2026, so a release can instead publish as
a Microsoft Entra ID app registration, signed in with the workflow's OIDC token
(workload identity federation): no token is stored anywhere.

```toml
[release]
vsce_entra = true      # GitHub Actions only
```

Set up once:

1. In the Azure portal (Microsoft Entra ID, App registrations), register an
   application. Note its **Application (client) ID** and the **Directory
   (tenant) ID**.
2. Under **Certificates & secrets, Federated credentials**, add one for
   *GitHub Actions deploying Azure resources*: the organisation and
   repository, and the entity **Branch** with the release branch (`main`). If a
   [release channel](#release-channels) publishes the extension, add one more for
   the channel's branch (`smith-v2`, say): the sign-in step runs there too, and
   Azure trusts only the branches it has a credential for.
3. Find the identity's Azure DevOps ID, which is what the Marketplace
   knows it by (not the client ID): a federated app has no password to sign in
   with from a terminal, so run this once in a workflow, after `azure/login`
   (an `az login` as the app works too), and read `id` from the answer:
   `az rest -u https://app.vssps.visualstudio.com/_apis/profile/profiles/me
   --resource 499b84ac-1321-427f-aa17-267ca6975798`.
4. On https://marketplace.visualstudio.com/manage, open the publisher, then
   **Members**, and add that ID with the **Contributor** role.
5. Add the repository secrets `AZURE_CLIENT_ID` and `AZURE_TENANT_ID`. They
   identify the app and are not credentials.

With the setting on, `smith ci generate` gives the workflow `id-token: write`,
adds a `Sign in to Azure` step (`azure/login`, with no subscription) right
before the release step, only on a push to the release branch or a release
channel's branch, tells the
release step to publish with that sign-in (`VSCE_AZURE_CREDENTIAL`, which makes
smith run `vsce publish --azure-credential`), and no longer passes `VSCE_PAT`
as a secret. The release does not ask for `VSCE_PAT` either. The same holds
for the native job when a native extension publishes from it. The setting
needs `providers = ["github"]`.

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
   and `{version}`), and pushes it. `{name}` is the project's `tag_prefix` without a trailing `@`, or
   without a trailing `-v`: with `tag_prefix = "acme-v"` the release tag
   `acme-v1.4.0` already is the pub tag, so no second tag is made and the
   workflow starts on it. If the pub tag already exists on another commit than
   the release's, the release stops (the run would start from that old commit);
2. starts `.github/workflows/pub.yml`, which `smith ci generate` writes, on
   that tag. A tag pushed with the workflow's own token starts no workflow,
   but a `workflow_dispatch` does, so the release job gets `actions: write`;
3. the pub workflow runs `smith release pub --project <app> --version
   <version>`: for an app it downloads the archives from the GitHub Release
   and builds the package; for a Dart package it writes the version into the
   pubspec (the tag's commit has the placeholder, because a release restores
   the files it versions). It then runs `dart pub publish`, authenticated by
   `setup-dart` with GitHub's OIDC token.

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
| every language but Go, and every app | `<name>@<version>` | override with `tag_prefix`, or for all projects at once with `[release] tag_format` |
| Go packages | `<dir>/v<version>` (e.g. `packages/gokit/v0.1.0`) | how `go get` finds a nested module's version |

`[release] tag_format` is the prefix of every project that sets no `tag_prefix` and no `bare_tags`.
`{name}` is the project's name (required) and `{scope}` is `[workspace] scope` as written, so with
`scope = "@acme"`, `tag_format = "{scope}/{name}@"` tags the project `kit` as `@acme/kit@1.0.0`. A
project's own `tag_prefix` wins, and so does the prefix a plugin reports (a Go module's `<dir>/v`).
`smith convert` writes the format when most of the workspace's tags follow `{scope}/{name}@` (two or
more projects, and more than the ones that do not), and a `tag_prefix` only on the projects that differ.

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
C++, Bicep, Lua, AppleScript, VBA, Swift and Objective-C projects have none). An unreachable
registry is a warning, not a stop. `smith doctor` runs the same checks without
fetching.

## The bump against the public API

A commit that says `fix` can delete an export. Before showing the plan, `smith release` asks the
plugin of each package that has a last release for the package's public API at that release's
tag and now, and compares the two lists of names. A removed name needs a major (a minor before
1.0). A new name needs a minor. When the commits give a smaller bump, the plan item gets a warning
that names the package, the names and the version to use instead:

```text
  ui    1.2.3 → 1.2.4  (1 releasing commit(s))
  warning: ui: the commits release a patch (1.2.4), but the public API lost 1 export (parse);
  that needs a major (2.0.0). Mark a commit as breaking (`feat!:` or a `BREAKING CHANGE:` footer),
  or restore the export.
```

The warning shows in the text plan, in `--format markdown`, in the `--comment` preview on the
pull request, in `--json` (`apiWarnings` on the item) and in the `smith ci release` log. It is
advice. A release never fails because of it, unless you ask:

```toml
[release]
api_check = "warn"   # "warn" (default), "error" stops the release too, "off" does not look
```

The check skips a package that was never released, a version a `version(...)[x]` directive forced,
and a plugin that cannot list an API. A failing plugin, a tag that was not fetched or a package
with no entry file give no warning. Fix the bump with a commit marker, not with `api_check`:
`feat:` for an added export, `feat!:` or a `BREAKING CHANGE:` footer for a removed one.

**Node.** The Node plugin reads the entry points `package.json` declares (`exports`, `types`,
`typings`, `module`, `main`), takes the built declaration file when the repository holds it and
the source it comes from otherwise (`dist/index.d.ts` is read as `src/index.ts`), and follows
`export * from`. It never builds. A private package has no public API. A subpath export is listed
as `./sub:name`.

**Go.** The Go plugin parses the `.go` files of the module with `go/parser` (at the tag, through
git, and now from disk). The API is every exported function, type, constant and variable, the
exported methods of exported types, the exported fields of structs and the methods of interfaces.
A name in a package below the module root is listed as `dir:Name`. Test files, `internal`,
`testdata`, `vendor`, hidden folders, `package main` and the packages of a nested module are not
part of the API. Build tags are not evaluated.

**Languages checked.** Node and Go. Python, C# and the other languages give no warning yet.

**Not detected.** A changed signature or type with the same name (only names are compared).
Wildcard subpath exports (`"./*"`). A second declarator in one `export const a = 1, b = 2` and
destructured exports. The implicit exports of an ambient `.d.ts` file. A regular expression
literal with an unbalanced brace. Other languages: a plugin can add the same method
(`api/surface`, see [the protocol](../architecture/plugin-protocol.md)) with no change to smith;
until then they are not checked. The reason is in
[ADR 0044](../architecture/decisions/0044-the-release-checks-the-bump-against-the-public-api.md).

## Release channels

A long-lived branch can release a line of prerelease versions of its own, while the release branch
keeps releasing stable ones:

```toml
[[release.channels]]
branch = "next"
prerelease = "beta"
major = 2
```

On `next` the plan numbers releases `2.0.0-beta.1`, `2.0.0-beta.2` and so on, from the tags that
the branch contains. A breaking change does not start a new major there, because the major is the
channel's. The release branch ignores these tags and keeps its `1.x` line. `smith release
--dry-run`, `smith report` and the pull request preview plan a channel branch like any other.

A real release from `next` publishes each package under the channel's name, `beta` here (set
`dist_tag` to change it; the default is the prerelease id, not `next`, and it is never `latest`),
so `latest` stays on the stable version. The GitHub Release of any prerelease version is marked as
a prerelease on its own, which GitHub never offers as the latest release: there is no setting for
it (`github_prerelease` is an unknown key). Install it with `npm install acme-kit@beta`.

To promote the line, merge the channel branch into the release branch: the next release there is
stable (`1.4.2 → 2.0.0`). Removing `prerelease` from the channel is an error, because a channel
without a prerelease id would release stable versions from the wrong branch.

Smith checks this before it publishes anything, and refuses the release when a project in the plan:

- is a package whose plugin does not declare the `channels` capability (the Node, Python, .NET,
  Rust, JVM, Ruby and Dart plugins, and those that release by tag alone, do: npm gets the
  channel's tag, and the other registries order a prerelease below the release). The Go plugin does
  not: a Go module needs `/vN` in its module path from major 2 on, so a `v2.0.0-beta.1` tag of a
  module whose path lacks `/v2` is unusable for `go get`. A Go release is only a tag, so smith
  cannot check the path before the tag exists; keep such a package out with `channels = false`,
- is deployed by a deploy plugin (a container image's `latest` and branch tags).

`smith release --dry-run`, `smith release --dry-run --comment` and the pull request preview show
the same refusal as a real release (`Blocked:`), once.

An app's binaries sent to npm (under the channel's tag), PyPI, NuGet, RubyGems and pub.dev, and its
Linux packages, are allowed: all but npm order a prerelease version below the release. The package
manager manifests (Homebrew, Scoop, winget) are attached to the GitHub Release as usual; that
release is a prerelease, which GitHub never offers as the latest, so the package managers keep
pointing at the stable version.

A VS Code extension on a channel is published as a pre-release under a plain version, because the
Marketplace takes no semver prerelease tag: `2.0.0-beta.3` becomes `2.1.3` (see the
[vscode plugin](../plugins/using/catalogue/vscode.md)).

Keep a project out of the channel with `channels = false` in its `[projects.x.release]`: on a
channel branch it is neither planned nor published, and it keeps releasing from the release branch.
Use it for a project that cannot take a prerelease yet. A Go module whose path lacks `/v2` is one:
a `v2.0.0-beta.1` tag could not be used by `go get`.

`smith ci generate` makes the pipelines of every provider trigger on a channel branch and run the
release step there, with the sign-in steps before it (not pack). `smith release --dry-run` and the pull
request preview name the channel, as the first line says: ``Release channel `beta`: prerelease
versions, published under the `beta` tag; the stable version stays the latest.`` A pull request
build plans for the branch it targets (GitHub, Azure, GitLab and Bitbucket say which one), so a
pull request into `next` previews `2.0.0-beta.n`. See
[ADR 0051](../architecture/decisions/0051-a-release-channel-gives-a-branch-its-own-line-of-prerelease-versions.md).

## Checked before anything is published

Before the first publish of a real release, smith asks npm, PyPI, NuGet and
RubyGems whether they accept it, with authenticated calls that publish
nothing: the credential is accepted (npm `/-/whoami`, an upload to PyPI or a
push to NuGet with no file, RubyGems' list of your gems), the registry is
reachable, and the package name is yours (a gem or an unscoped npm package
that exists must be owned by the key; a scoped npm package owned by someone
else is only a warning, since a team can own it). Any failure stops the
release before a single package is published, naming the registry, the
secret and the fix. A key that is merely unset is already refused by the
credential check above; npm's own login (no `NODE_AUTH_TOKEN`) cannot be
asked and is left to `npm publish`.

A version a registry already has is a warning: it is expected when the
[journal](#the-release-journal) shows this release published it, and said
otherwise, since the content cannot be compared before it is built. Smith
does not run `npm publish --dry-run`, `twine check` or `cargo publish
--dry-run` ahead of time (they need the built packages), and packages
published by a language plugin (not an app's binaries) are not asked yet.
A dry run does not check the registries.

## Rehearse a release

`smith release --rehearse` finds a wrong feed setting before the first real push, so the
first CI run on a new feed needs no fix-and-retry commit. It looks at every package project
(not at the commits), and for each registry its plugin publishes to it asks, with read-only
requests:

| Check | Question | How |
|---|---|---|
| `registry` | Does the URL resolve, and which credential does it expect? | An unauthenticated request; the `WWW-Authenticate` header gives Basic or Bearer. |
| `credential` | Is the credential set? | The variable the publish reads, or the one `smith.toml` names (see below). |
| `access` | Does the registry accept it? | npm `/-/whoami`, or a package read where the feed has no `whoami` (Azure Artifacts, GitHub Packages); an upload with no file to PyPI, which the index refuses after it checks the credential; a read of the feed for a private NuGet feed; `/api/v1/me` on crates.io. |
| `identity` | In CI, which identity publishes? | The pipeline (GitHub Actions or Azure Pipelines) is detected from its environment. |

The URL is what the publish uses: the registry `smith.toml` names (`registry = "<name>"`
and `[registries.<name>]`), else `NPM_CONFIG_REGISTRY` and the `registry=` and
`@scope:registry=` lines of `.npmrc` (user, workspace and project), `TWINE_REPOSITORY_URL`,
`NUGET_SOURCE`, else the public registry. Credentials by sign-in kind:

| Sign-in | Credential |
|---|---|
| `azure-artifacts` | `SYSTEM_ACCESSTOKEN` (map `$(System.AccessToken)` into the step), or an `.npmrc` token. Outside Azure Pipelines this is a warning, since the pipeline supplies it. |
| `github-packages` | `GITHUB_TOKEN`. |
| `token` | the variable `token_env` names. |
| none | npm `NODE_AUTH_TOKEN` or `NPM_TOKEN` or the `.npmrc` `_authToken` for the registry (a `${VAR}` in it is expanded); PyPI `PYPI_TOKEN` or `TWINE_PASSWORD`; NuGet `NUGET_API_KEY`; crates.io `CARGO_REGISTRY_TOKEN`. |

A GitHub Actions job with `id-token: write` and no token is a warning, not a failure: npm,
PyPI and crates.io accept OIDC (trusted publishing), which cannot be checked without
publishing. A NuGet push key on nuget.org cannot be checked without pushing either, so it is
reported as set and nothing more. Whether the Build Service identity has Contributor on an
Azure feed cannot be read with a read request: a refused request names the setting to change.
Maven, RubyGems, pub.dev and Packagist projects are not rehearsed yet (the existing
[pre-publish check](#checked-before-anything-is-published) covers RubyGems for an app's binaries).

The output is in the style of `smith doctor`: what failed, why, and `→` the exact fix.
`--verbose` lists passed checks too and `--json` prints the report (`results` with `project`,
`registry`, `check`, `status`, `message`, `fix`). The exit code is 1 when a check fails, 0
otherwise. No secret is printed or encoded.

## The release journal

Each package's release appends what it did to `dist/release/<project>/journal.json`:
the plugin's registry, each npm, PyPI, NuGet or RubyGems package (published, or
already there), images, the tag, the push, the mirror and the GitHub Release,
and the step that failed. `smith release --status` prints it (`--json` too).
The journal only grows and there is no rollback: registries do not let a
published version be removed (npm unpublish is restricted, PyPI and crates.io
are immutable), so recovery is to re-run, which publishes the rest.

## When something fails

The release stops at the failing package; everything before it is tagged and
pushed. Re-run it: tagged packages have no new commits and are kept, and a
package that was published but not tagged is skipped by the registry check and
then tagged. Tags left local by an interrupted push are pushed by the next run.

A release that stopped after its tag was pushed, when GitHub refused to create
the Release (a 500, say), is finished by a re-run even on a fresh CI checkout,
which has no journal. When no package has anything new to release, smith asks
GitHub (with `GITHUB_TOKEN`) whether each project's latest release tag has a
GitHub Release, and finishes the ones that do not: it creates the Release and
attaches the assets, rebuilding them if the Release lacks any. It publishes to
no registry again. Only a tag at the commit being run is finished, because the
assets are built from the checkout: the Release of an older tag, with commits
after it, is skipped (smith prints `not creating the GitHub Release of
smith@1.121.0: HEAD is newer than its tag`) and stays missing. This needs `[release] github_releases = true`. The notes of
a finished Release are empty, as the commits are not replanned.

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
