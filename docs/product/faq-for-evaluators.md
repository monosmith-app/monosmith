# FAQ for evaluators

Questions a team lead, a security reviewer or a procurement contact asks before
adopting a tool. Answers are checked against the repository and the reference
pages as of smith 1.62.0 (2026-10-25); the answers on languages, plugins and the
graph were checked again against smith 1.65.0 the same day. Where something could
not be checked, the answer says so.

## Languages, plugins and the graph

### Which languages work out of the box, and what does "out of the box" mean?

Node and TypeScript, Python, Go, Rust, .NET, Ruby, PHP, C and C++, Dart and Flutter, the
JVM (Java, Kotlin, Android) and Azure Bicep. `smith add <kind> <name> --lang <language>` writes
the project with a worked example and a test, adds the root files the language wants once per
repository, and regenerates the CI pipeline and the dev container with that language's toolchain.
In a scratch workspace with all eleven on 2026-10-25, every test and build passed. Two cells are
not fully proven: Gradle, Kotlin and Android were not run, and the Dart lint task did not pass on
the test machine because `dart analyze` itself crashed. See
[Languages and environments](comparison/languages-and-environments.md).

### Can we add a language or a rule smith does not have?

Yes. A plugin is any program that speaks JSON-RPC 2.0 on stdin and stdout, so it can be written in
any language; SDKs are published for Go, TypeScript, Python and C#, and a 102-line hand-framed plugin with no SDK
passes `smith plugin check`. A plugin can add a language, tasks, a scaffold, CI steps or a release
target. A team standard that needs no code is a recipe. There is no plugin registry yet, and plugins
are placed on a search path or named in `smith.toml`. See
[Plugins and flexibility](comparison/plugins-and-flexibility.md).

### Can a plugin in a repository I cloned run code on my machine?

Not before you allow it. An external plugin named in `smith.toml` does not start until you run
`smith plugin trust`; it then starts with a minimal environment, and in CI only when `smith.toml`
pins its SHA-256. Trusting is not sandboxing: a trusted plugin runs with your privileges, and
smith has no sandbox yet (issue #565).

### How does smith know what depends on what, across languages?

Each language plugin reports a project's imports and exports, and smith matches them; shared
OpenAPI, AsyncAPI, protobuf and GraphQL files link projects of different languages when a
consumer names the schema in `smith.toml`. The graph is per project, not per file.
`smith graph --blast-radius <file>` lists what a change reaches and `smith graph --diff <rev>` shows
what a branch changed in the graph. Bazel, Pants and Buck2 model individual targets and files;
smith does not. See [The dependency graph](comparison/dependency-graph.md).

## Cost

### What does MonoSmith cost?

Nothing to license. It is MIT licensed and free. There is no paid tier, hosted
service or account in the documentation or the repository.

### What costs money, then?

What you already pay for: CI minutes, and storage if you run a remote cache. The
remote cache is an HTTP server, an S3 bucket or an Azure Blob container that you
operate; the documentation names no smith-run cache service. `--affected` and
the cache are how smith reduces the CI minutes.

### Does the tool phone home?

Not that was found. The documentation describes no telemetry setting, and a search
of the Go sources found no telemetry code. This is a search, not an audit. One
network call is documented: `smith info` asks GitHub for the newest release, and
`SMITH_NO_UPDATE_CHECK=1` stops it. Commands that talk to a registry (`smith up`,
releasing) need the network by their nature.

## Lock-in

### What would I be tied to?

- **`smith.toml`**, which is small, with every setting optional.
- **The generated CI pipelines**, which call `smith ci <phase>`. CI needs the `smith`
  binary, which the pipeline installs at a pinned version: from npm, PyPI or NuGet in a Node, Python or
  .NET workspace, else with `tools/smith/install.sh`, a copy of the installer committed in your
  project (it checks the download against the release checksums).
- **The directory convention:** builds go to `dist/` at the workspace root, and
  `smith doctor` fails a task that writes elsewhere.
- **Release tags** named `<project>@<version>` (a Go package: `<dir>/v<version>`).
- **npm** as the package manager for Node projects.

### What is not locked in?

Projects keep their own manifests (`package.json`, `go.mod`, `pyproject.toml`,
`Cargo.toml`, `*.csproj`) and their sources. Tasks run the language's own tools
(`npm`, `go`, `cargo`, `dotnet`) as child processes with argument arrays, never
through a shell. Published packages are ordinary packages on their registries.

### How do I leave?

There is no `smith eject` command. Leaving means removing `smith.toml` and the
generated pipelines and replacing them with your own. The projects, manifests and
tags stay as they are. This has not been tried as a documented procedure, so
rehearse it on a branch before relying on it.

## Licence

### Which licence?

MIT, in the repository's `LICENSE` file (copyright 2026 Eduardo Russo). The
third-party libraries smith uses are listed in
[Technology](../architecture/technology.md); their licences were not audited here.

### Who owns it?

The repository belongs to the `monosmith-app` GitHub organisation. No company,
foundation or governance document was found in the repository.

## Maturity

### How mature is smith?

Young, and it is better to hear that here. The repository's first commit is dated
2026-10-15 and this page was written on 2026-10-25. In that time it reached release
1.62.0: the repository holds 102 tags named `smith@<version>`, and almost all of the
commits are by one author. smith continues mnci, an earlier tool built on Nx, whose
workflows it replaces, but the engine was rewritten from scratch in Go
([Decisions](../architecture/decisions/README.md), entry 1).

What the project can show:

- A golden-path test (a new workspace, projects in every built-in language, build,
  test, run, CI and release) that the README says is tested end to end on Windows, macOS
  and Linux, and a `golden-path.yml` workflow in the repository (the decisions log
  says these suites run once a day and on demand, not on every change).
- MonoSmith builds and releases itself.
- Reference documentation for every command, setting and plugin role.

What it cannot show yet: years of production use, a large user community, a public
list of adopters, or a record of how it behaves in a very large repository. Nothing
in this folder reports a measurement of smith on one.

### How fast does it change?

Quickly. Expect settings, commands and generated files to change between versions.
`smith upgrade` migrates `smith.toml` to the current schema, keeps its comments and
regenerates the files smith owns; `smith upgrade --check` fails in CI when something
would change.

### Which platforms?

Linux, macOS and Windows on x64 and arm64. Windows is a first-class target from the
start. You also need git and the toolchain of each language you use; smith does not
install compilers.

## Security and supply chain

### Does smith produce an SBOM and provenance?

Yes, with limits. From [Releasing](../reference/releasing.md#sbom-and-build-provenance):

- Every **app released with a GitHub Release** gets a CycloneDX 1.5 SBOM
  (`<product>_<version>_sbom.cdx.json`) beside its archives, written from the
  lockfile or manifest. **Only Go and npm apps get one today;** Python, Rust, .NET
  and the rest are listed as to follow.
- **On GitHub Actions** the generated workflow attests the release files (archives,
  Linux packages, checksums and the SBOM) with SLSA build provenance, signed through
  GitHub's Sigstore. Anyone can verify a download with `gh attestation verify`.
- **On Azure Pipelines and GitLab CI** there is an SBOM but no provenance; the SHA-256
  `checksums.txt` verifies a download.
- **Native apps** (per-OS jobs) get neither yet.
- **Registry publishes** keep their own provenance: npm provenance, and PyPI
  attestations through trusted publishing.

### How is smith itself distributed and checked?

Every smith release has a `checksums.txt` with SHA-256 for each archive, a CycloneDX
SBOM, and SLSA provenance on GitHub; the install scripts check the archive against
`checksums.txt` before installing. The npm packages carry provenance
(`npm audit signatures`). The binaries are **not signed** with an Apple or Microsoft
certificate: a zip downloaded in a browser meets Gatekeeper on macOS and SmartScreen
on Windows. The install scripts and package managers avoid that prompt. See
[Installing](../getting-started/installing.md).

### How does CI treat vulnerable dependencies?

The generated pipelines have an `audit` phase: `npm audit` (blocking only on a
production advisory at moderate or above that `npm audit fix` can apply),
`pip-audit` and `govulncheck` (reported). The repository also runs Dependabot.

### Is the remote cache safe?

Every object is signed with HMAC-SHA256 under a shared secret (`SMITH_CACHE_KEY`),
the signature covers the object's name, a hit is verified (blobs are digest-checked),
and a failed check is a miss. Without the key, nothing is shared. A remote cache
"can make a run slower, never wrong". That holds for tampering with stored objects;
a person who holds the key can write valid objects, so protect it like any CI secret.

### Is there a security policy or contact?

None was found: the repository has no `SECURITY.md`. Ask in the issues before you rely
on a private disclosure channel existing.

## Support

### Who supports it?

The project's GitHub issues
(<https://github.com/monosmith-app/monosmith/issues>). No commercial support,
service-level agreement or paid plan was found. Plan to be self-reliant: the
documentation is extensive, and `smith doctor`, `smith explain` and `smith why-miss`
are built to answer the common questions without a human.

### Where do I get help with a problem?

[Troubleshooting](../guides/troubleshooting.md) lists symptoms and remedies, and the
[FAQ](../guides/faq.md) answers usage questions.

## Roadmap

### What is planned?

Plans and bugs are tracked in the repository's issues and on the MonoSmith project
board; no dated roadmap is published. What the documentation itself lists as not
done yet:

- **Plugin discovery and sandboxing.** How plugins are published, found and installed by name is open
  ([Decisions](../architecture/decisions/README.md)); plugins are found on a search path or
  declared by command. A registry and a sandbox are tracked in #565, an SDK for C# in #564.
- **More install channels.** Homebrew, Scoop, winget, Chocolatey, the AUR, Snap and an
  apt or dnf repository are tracked in issue #225; manifests are generated but no
  tap, bucket or submission uses them yet.
- **SBOMs for more ecosystems,** and provenance outside GitHub Actions
  ([Releasing](../reference/releasing.md)).
- **`convert` beyond Node, Go and Python,** and the `--level full` clean-up beyond Node
  workspaces ([Converting](../getting-started/converting.md)).
- **Not planned: distributed execution and hermetic sandboxing.** The design notes
  record both as deliberately not done: tasks run on one machine, only results are
  shared, and smith trusts a task's declared inputs instead of sandboxing it
  ([Execution and cache](../architecture/execution-and-cache.md), section 9; [Decisions](../architecture/decisions/README.md), entry 18).
  The price, in the design notes' words: wrong `inputs` give wrong cache hits.

Treat all of this as intent, not commitment.

### Can I influence it?

Open an issue. [Working on MonoSmith](../contributing/development.md) explains how to
build and test the repository.

## Next

- [Choosing a monorepo tool](comparison/choosing.md)
- [The case for MonoSmith](why-monosmith.md)
