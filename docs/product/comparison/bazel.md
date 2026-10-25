# Bazel compared with MonoSmith

Bazel is an open-source build and test tool, comparable to Make, Maven and
Gradle, that describes software in BUILD files written in Starlark. It is built
for scale and correctness.

## What Bazel is best at

- **Scale.** Its introduction claims builds with "100k+ source files" and
  codebases worked on by tens of thousands of users.
- **Hermetic builds.** Builds and tests can be configured to run hermetically
  through sandboxing, which reduces skew and improves reproducibility.
- **Remote execution and caching.** An open gRPC protocol distributes build and
  test actions across machines, and shares outputs. The documentation links to
  commercially supported remote execution services and self-service tools.
- **Many languages.** Rules exist for Java, C++, Android, iOS, Go and others, and
  Bazel can be extended to other languages with its own extension language.
- **Open source.** Apache-2.0.

## Where MonoSmith differs

| | Bazel | MonoSmith |
|---|---|---|
| Unit of work | targets declared in BUILD files | projects found from manifests |
| What you write | BUILD files in Starlark | `smith.toml` with a name; the rest is inferred |
| Hermetic sandboxing | configurable | none: tasks run the project's own tools as child processes |
| Remote execution | yes | no |
| Affected | no changed-files command found; `bazel query` can compute reverse dependencies (`rdeps`, `allrdeps`) of targets you name | `smith affected` from a base branch, with everything affected when unsure |
| CI | the FAQ says Bazel "returns a non-zero exit code if the build or test invocation fails", enough for basic CI integration; no generator found | generated pipelines for three providers |
| Release | no release or versioning command found (searched "publish", "versioning", "release") | releases from commits |
| Editors | an IntelliJ plugin, per the FAQ | a VS Code extension, a terminal interface and an MCP server |
| Dependency versions | bzlmod resolves versions with Minimal Version Selection; no alignment or lint command found | `smith sync`; `smith doctor` checks one version per production dependency |
| Repository status | Apache-2.0, not archived, last push 2026-10-25 (GitHub API, 2026-10-25) | MIT, not archived, first commit 2026-10-15; release 1.62.0 |
| Tools | rules, written in Starlark or provided with Bazel | runs the language's own tools (`npm`, `go`, `cargo`, `dotnet`) as child processes |

## When to choose Bazel instead

- You need hermetic, reproducible builds, remote execution, or builds at the scale
  Bazel's introduction describes.
- Your codebase is large enough that incremental correctness is worth the cost of
  maintaining BUILD files.

## When MonoSmith is the better fit

- You want native language tooling to keep working as it does, and want a tool that
  starts from the manifests you already have.
- The team cannot afford to maintain a second description of every dependency.

## Moving between Bazel and MonoSmith

There is no automated path in either direction. Moving to smith means giving each
project its normal manifest again (if it has none), then `smith adopt --dry-run`;
moving to Bazel means writing BUILD files. Neither was tried for this page.

## Sources

Retrieved 2026-10-25.

The documentation was read as MDX source files from the repository's `docs` folder
(the current docs, not the `versions/` archives), not through a summarising tool.

- About Bazel: <https://bazel.build/about/intro> (builds with "100k+" source files)
- Remote execution overview: <https://bazel.build/remote/rbe>
- FAQ: <https://bazel.build/about/faq> (CI by exit code; the IntelliJ plugin)
- Query language: <https://bazel.build/query/language> (`rdeps`, `allrdeps`)
- Local cache: <https://bazel.build/reference/flag-cheatsheet> (`--disk_cache`)
- Module versions: <https://bazel.build/external/faq> ("Bazel instead uses Minimal Version Selection")
- Repository (licence, README): <https://github.com/bazelbuild/bazel>. GitHub API on 2026-10-25: licence Apache-2.0, archived false, last push 2026-10-25.

### Added for the master table

Retrieved 2026-10-25, from the same MDX source files as above, for the cells of [Everything compared](everything-compared.md) that the other tables do not cover. Paths are relative to the documentation folder named above.

- `docs/user-manual.mdx` (`--explain`: "explain, for each build step, either why it is being executed, or that it is up-to-date").
- `basics/artifact-based-builds.mdx` (`visibility`); `about/faq.mdx` (sandboxed execution); `advanced/performance/build-performance-metrics.mdx` (Build Event Protocol); `install/ide.mdx` (a VS Code plugin on the Visual Studio Marketplace).
- `docs/bazel-and-javascript.mdx`, `external/overview.mdx` and `concepts/runfiles.mdx` (separately maintained `rules_nodejs`, `rules_rust` and `rules_python`).
