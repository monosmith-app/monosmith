# Buck2 compared with MonoSmith

Buck2 is Meta's build system, written in Rust, open source since 2023. Its own
documentation and README describe it as early-stage and pre-release, so read this
page as a comparison with a tool that says it is still settling.

## What Buck2 is best at

- **Remote execution first.** The documentation calls Buck2 "remote execution
  first", with local execution treated as a special case.
- **One graph.** It tracks changes precisely through a single dependency graph, and
  supports dep files and incremental actions to skip unneeded work.
- **Many languages.** Rules are written in Starlark and can combine languages;
  Meta's own monorepo spans C++, Python, Rust, Kotlin, Swift and more. C and C++
  support is written as a library.
- **Open source.** Dual-licensed under MIT and Apache-2.0.

## Where MonoSmith differs

| | Buck2 | MonoSmith |
|---|---|---|
| Status | the README: "does not have a stable release tag at this time"; the docs: "early-stage" | release 1.62.0, repository history of ten days |
| Remote execution | yes | no |
| What you write | BUCK files and rules in Starlark | `smith.toml`; the rest inferred |
| Release, CI generation | not covered by the pages read | both |

## When to choose Buck2 instead

- You want a Rust build system built around remote execution and are prepared for
  rough edges the README warns about for outside users.

## When MonoSmith is the better fit

- You want a tool that starts from your manifests and covers CI and release, and you
  do not need remote execution.

## Moving between Buck2 and MonoSmith

No automated path in either direction, and neither was tried.

## Sources

Retrieved 2026-10-25.

- Why Buck2: <https://buck2.build/docs/about/why/>
- Repository (licence, status): <https://github.com/facebook/buck2>
