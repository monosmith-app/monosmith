# Technology: what smith is made of, and why

A tour of the technical choices, with the reason for each. For the structure of
the code see [architecture.md](architecture.md); for the decisions log (the
ones the product owner made) see [decisions.md](decisions.md).

## One static binary

smith is written in **Go** and ships as a single static executable per platform
(Linux, macOS and Windows, amd64 and arm64). The reasons are about the person
installing it:

- nothing to install first: no Node, no JVM, no Python. A monorepo tool that
  needs a runtime to start would fail the first time on a CI image that lacks
  it;
- the binary is stamped with its version and reproducible (`CGO_ENABLED=0`,
  `-trimpath`), and starts in milliseconds, which matters for a tool that runs
  on every commit and every shell prompt;
- cross-compilation from one machine builds all six targets, so a release is
  one job;
- goroutines make the scheduler ("one goroutine per task, gated by a semaphore")
  nearly the whole of its implementation.

`go install` is deliberately not a supported way to install it (decision 19):
the install channels (script, npm, PyPI, NuGet, RubyGems, pub.dev, `.deb`,
`.rpm`, GitHub Release) all carry the stamped binary.

### Libraries

Few, and each does one thing:

| Library | Used for |
|---|---|
| `spf13/cobra` | the command tree, flags, shell completion |
| `pelletier/go-toml/v2` | reading `smith.toml` strictly (unknown keys are errors) and editing it in place, keeping comments |
| `bmatcuk/doublestar` | `**` globs for task inputs and selection |
| `charmbracelet/bubbletea`, `lipgloss`, `bubbles` | the terminal UI (a bare `smith`) and the interactive graph browser |
| `google/rpmpack` | building `.rpm` packages without `rpmbuild` |
| `golang.org/x/sys`, `x/term` | Windows Job Objects, terminal detection |

Everything else is the standard library: HTTP clients, `crypto/hmac` for the
remote cache, ZIP/TAR/GZIP for release archives, `encoding/json` for the
protocol. S3 request signing (AWS Signature V4) is implemented in ~150 lines
rather than pulling in an SDK.

## Process boundaries

| Boundary | Mechanism | Why |
|---|---|---|
| smith ⇄ plugin | JSON-RPC 2.0 over stdio, LSP framing | language-neutral, no ports, crash-isolated ([plugin-system.md](plugin-system.md)) |
| smith ⇄ tool (`npm`, `go`, `cargo`, `docker`) | child process, **argv arrays, never a shell** | the same command means the same on Windows and Linux; nothing is word-split or globbed by accident; no injection through a project name |
| smith ⇄ git | the `git` binary, through one slice (`gitclient`) | git is the source of truth for the file list, history and tags; shelling out is more faithful than a reimplementation and keeps the door narrow |
| smith ⇄ GitHub | REST, through one slice (`githubrelease`) | releases, comments |
| CI ⇄ smith | `smith ci <phase>` | the YAML stays thin; a laptop runs what CI runs |
| machine ⇄ machine | the remote cache, HMAC-signed objects | share results, never trust the store |

Nothing in smith runs a long-lived daemon. State lives in files: `smith.toml`
(configuration), `.smith/cache` (results), git tags (releases), and the
manifests of the projects themselves. There is no database and no hidden state
to repair.

## Cross-platform, taken seriously

A monorepo tool that works only on the maintainer's OS is not a monorepo tool.
The recurring problems and their answers:

- **Paths.** Inside smith every path is POSIX and workspace-relative (project
  roots, `cwd`, inputs, outputs, cache entries, JSON output). Conversion to the
  OS form happens at the file system door only.
- **Process trees.** Unix process groups, Windows Job Objects
  (`taskrun/process_tree_client*.go`): cancelling a task leaves no orphans
  anywhere.
- **Line endings and quoting.** The repository pins `eol=lf`; commands are
  argv arrays; scripts that need a shell are written as `package.json` scripts
  (which npm runs through the right shell on each OS) by `smith convert`.
- **Tool lookup.** The `PATH` a task sees is the one smith builds (venv first),
  honouring `PATHEXT` on Windows so `npm` finds `npm.cmd`.
- **CI.** Unit tests run on all three operating systems on every pull request,
  and the end-to-end suites daily on all three. Windows is the slowest and the
  one that finds the most.

## Data formats

- `smith.toml`: TOML, every setting optional, unknown keys rejected with a
  message naming the key. A JSON Schema ships for editors
  (`apps/smith/workspace/smith.schema.json`).
- Recipes: TOML, extendable, resolved with provenance — each value remembers the
  file and line it came from, and `smith explain` shows it.
- Machine output: `--json` on every reporting command, with a stable
  contract (`clicontract`) that editors and `smith mcp` rely on.
- The cache: JSON entries plus content-addressed blobs.
- Release records: `.smith/release-report.json`.

## Security model

- **No shell, no eval.** Tasks are argv; plugins are separate processes.
- **Secrets only through the environment**, named in the generated pipeline and
  listed by `smith ci check`. smith never writes a secret to disk or to a log.
  Forked pull requests, which get no secrets, still build and test.
- **The remote cache cannot make a build wrong:** signed objects, verified
  content hashes, failures are misses.
- **Publishing uses short-lived identity where the registry allows it:** npm
  provenance and pub.dev through GitHub's OIDC token; PyPI, NuGet and RubyGems
  through scoped tokens.
- **Plugins are code you run.** An external plugin is a program on your
  `PATH`-like search path; smith.toml pins its version. The built-ins are part
  of the binary. `smith plugin check` verifies behaviour, not trustworthiness.
- **MCP tools** refuse inputs that look like options and never take a path where
  a name suffices.
- No code signing or notarisation yet (decision log); checksums are published
  with every release.

## Testing

| Layer | What | Where |
|---|---|---|
| Unit | pure policies, algorithms, mappers, use cases with fakes | next to the code (`*_test.go`), run on three OSes on every pull request |
| Plugin conformance | a plugin obeys the protocol | `packages/plugin-sdk/conformance`, `smith plugin check` |
| Golden path | the real thing: a new workspace, every project kind in every language, build, test, release | `apps/smith/goldenpath`, daily and on demand |
| Convert | Nx and old-mnci workspaces convert and still build | `apps/smith/convert` fixtures on three OSes |
| Self-hosting | smith builds, lints, tests and releases itself | this repository's CI |

The code layout is enforced by a tool, too: `smith slices check` verifies
every file's name carries its role and every package has a public surface
(`slicecheck`), and CI runs it on every Go module.

## Why a monorepo tool needs all of this in one program

Teams that mix languages usually end up with three build systems and a Makefile
holding them together. smith's bet is that the things that are the same across
languages — *what depends on what, what has to be rebuilt, what to run in CI,
what to release and in which order* — deserve one owner, while everything that
differs lives behind a narrow interface. The technology above is the cost of
keeping that interface narrow: a binary with no runtime, plugins as processes,
git as the database, and a cache that is allowed to be wrong only by being
absent.
