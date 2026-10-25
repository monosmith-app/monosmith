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
| Status | the README: "does not have a stable release tag at this time"; the docs: "early-stage". GitHub lists dated tags (the newest, 2026-10-15) and a `latest` tag, all flagged pre-release | release 1.62.0, repository history of ten days |
| Remote execution | yes | no |
| Remote cache | the remote-execution action cache and content-addressable storage | HTTP, S3 or Azure Blob, signed |
| What you write | BUCK files and rules in Starlark; `buck2 init` creates a project | `smith.toml`; the rest inferred |
| Languages | a prelude with rules for C/C++, Erlang, Go, Haskell, Java, Kotlin, OCaml, Python and Rust; the language-support table marks C#, Objective-C and Swift "Unavailable" and every language's "Documented" column reads "no" | eleven languages built in |
| Affected | `cquery` can select reverse dependencies (`rdeps`); the docs mention a "Target Determinator" only in passing | `smith affected` |
| Editors | the Starlark library gives LSP and DAP bindings; no editor extension is documented | a VS Code extension, a terminal interface and an MCP server |
| Release, CI generation | none found (searched the documentation source) | both |
| Repository status | MIT and Apache-2.0 (two licence files), not archived, last push 2026-10-25 (GitHub API, 2026-10-25) | MIT, not archived, first commit 2026-10-15; release 1.62.0 |

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

The documentation was read as Markdown source files from the repository's `docs`
folder, not through a summarising tool.

- Why Buck2: <https://buck2.build/docs/about/why/> ("it is early-stage software")
- Language support: <https://buck2.build/docs/about/language_support/>
- Architecture and remote execution: <https://buck2.build/docs/concepts/architecture/> ("RE action cache") and <https://buck2.build/docs/users/remote_execution/> ("Buck2 can use services that expose Bazel's remote execution API")
- Tutorial: <https://buck2.build/docs/getting_started/tutorial_first_build/> (`buck2 init`)
- Repository (licence, status): <https://github.com/facebook/buck2>. The README says Buck2 "does not have a stable release tag at this time" and "is licensed under both the MIT license and Apache-2.0 license". GitHub API on 2026-10-25: archived false, last push 2026-10-25; the licence field shows Apache-2.0 only, because it reports one; the repository holds both licence files. The releases API lists tags `latest` (2023-04-13), 2026-10-15, 2026-10-01 and 2026-09-15, each flagged pre-release.

### Added for the master table

Retrieved 2026-10-25, from the same Markdown source files as above, for the cells of [Everything compared](everything-compared.md) that the other tables do not cover. Paths are relative to the documentation folder named above.

- `about/benefits/compared_to_buck1.md` (sandboxed remote execution); `concepts/build_rule.md` (visibility).
