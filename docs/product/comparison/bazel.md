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
| Release and CI | not covered by the pages read | generated pipelines and releases from commits |
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

- About Bazel: <https://bazel.build/about/intro>
- Remote execution overview: <https://bazel.build/remote/rbe>
- Repository (licence, README): <https://github.com/bazelbuild/bazel>
