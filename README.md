<p align="center"><strong>MonoSmith</strong> — a simple monorepo tool</p>

# MonoSmith

MonoSmith runs a monorepo in Node, Go, Python, Rust, .NET, Ruby, PHP, C++,
Dart, the JVM and Bicep from one command, `smith`. It scaffolds projects, runs
their tasks in dependency order from a cache, generates the CI pipeline from
what is in the repository, and releases from conventional commits.

`smith` is a single static binary written in Go. Each language is a plugin
that runs out-of-process over a small JSON-RPC protocol; the built-in ones
ship inside the binary, and a plugin for another language can be written in
any language without touching `smith`.

```sh
smith new acme && cd acme
smith add app web --lang node        # deployed, never published
smith add lib core --lang go         # imported internally, never published
smith add package sdk --lang python  # published on release
smith add app api --stack react      # a stack: the plugins a project is built with

smith build                          # every project, in dependency order, cached
smith test --affected                # only what changed since main
smith graph                          # how projects depend on each other (--format tui to browse)
smith release --dry-run              # what the next release would publish
smith doctor                         # is the workspace healthy?
smith                                # the interactive shell, in a terminal
```

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/monosmith-app/monosmith/main/install.sh | sh     # Linux, macOS, Git Bash
irm https://raw.githubusercontent.com/monosmith-app/monosmith/main/install.ps1 | iex           # Windows PowerShell
```

Or with the package manager of your language; each installs the same `smith`
binary:

```sh
npm install --global monosmith
pipx install monosmith                  # or: uv tool install monosmith, pip install monosmith
dotnet tool install --global monosmith  # .NET 10 SDK
gem install monosmith
dart pub global activate monosmith
```

Or download a zip from the [releases](https://github.com/monosmith-app/monosmith/releases).
Each release also carries a `.deb` and an `.rpm`. Linux, macOS and Windows on
x64 and arm64 are supported.
[Installing](docs/guides/installing.md) covers each channel, upgrading, checking a
download, and the unsigned-binary prompts a browser download gets.

MonoSmith continues mnci, which was built on Nx; `smith convert` converts Nx
workspaces and those made by mnci.

## What it does

| | |
|---|---|
| **Three kinds** | `app` (deployed), `lib` (internal), `package` (published). The kind decides what release does with a project. |
| **Eleven languages** | Node (JavaScript and TypeScript), Python, Go, Rust, .NET (C#, F#), Ruby, PHP, C and C++ (CMake), Dart and Flutter, the JVM (Java with Maven, Kotlin with Gradle, Android) and Bicep, each through a built-in plugin. |
| **Stacks** | A project records the plugins it is built with: React, Vue, Svelte, Angular, an Astro docs site or a VS Code extension next to a plain Node library, in the same workspace. |
| **No configuration to write** | Projects, tasks and dependencies come from the manifests (`package.json`, `go.mod`, `Cargo.toml`, `*.csproj`, …) and from what the sources import. `smith.toml` names the workspace; everything else has a default. |
| **Cache** | A task whose inputs did not change replays its output and log from a content-addressed cache, locally or from a shared remote (HTTP, S3, Azure Blob). |
| **Affected** | `--affected` runs only the projects a change can break, and runs everything when it cannot be sure. |
| **CI generated** | GitHub Actions, Azure Pipelines and GitLab CI, derived from the plugins in use, calling `smith ci <phase>` — the same commands you run locally. |
| **Release** | feat → minor, fix → patch, breaking → major. Packages release in dependency order to their registries (npm, PyPI, crates.io, NuGet, RubyGems, pub.dev, a Maven repository, Go module tags); an app releases per-platform binaries to GitHub Releases and, if you ask, to npm, PyPI, NuGet and RubyGems. A failed release resumes when re-run. |
| **Conventions** | `vertical-slices` checks that code is laid out by what it does, in each language's own naming style. `docker` builds and pushes an image for every app with a `Dockerfile`. |
| **Convert** | `smith convert` turns an Nx workspace, or one made by the old Nx-based mnci, into a smith workspace, keeping its projects and release tags. It reports anything it cannot map. |
| **Editors and assistants** | Stable `--json` contracts for editors, a JSON Schema for `smith.toml`, and `smith mcp` for AI assistants. |

## Documentation

Start with [getting started](docs/guides/getting-started.md), a hands-on tour.
Everything is in [docs/](docs/README.md), organised by what you want to do, and
in four sections:

- **Guides:** [installing](docs/guides/installing.md),
  [getting started](docs/guides/getting-started.md),
  [converting from Nx](docs/guides/converting.md),
  [editors](docs/guides/editors.md),
  [writing a plugin](docs/guides/writing-a-plugin.md),
  [troubleshooting](docs/guides/troubleshooting.md), [FAQ](docs/guides/faq.md)
- **Reference:** [CLI](docs/reference/cli.md), [glossary](docs/reference/glossary.md),
  [configuration (`smith.toml`)](docs/reference/configuration.md),
  [tasks, cache and affected](docs/reference/tasks-and-cache.md),
  [CI pipelines](docs/reference/ci.md), [releasing](docs/reference/releasing.md),
  [stacks](docs/reference/stacks.md), [recipes](docs/reference/recipes.md),
  [conventions](docs/reference/conventions.md),
  [plugins and the protocol](docs/reference/plugins.md)
- **Design:** [architecture](docs/design/architecture.md), [plugin system](docs/design/plugin-system.md),
  [decisions](docs/design/decisions.md), [research](docs/design/research/)
- **Contributing:** [working on MonoSmith](docs/contributing/development.md)

## Status

The golden path — a new workspace, projects in every built-in language,
build, test, run, CI and release — is tested end to end on Windows, macOS and
Linux, and MonoSmith builds and releases itself. Plans and bugs are
tracked in [issues](https://github.com/monosmith-app/monosmith/issues), and
their progress on the [MonoSmith board](https://github.com/orgs/monosmith-app/projects/1).

## License

MIT
