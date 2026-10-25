# Cargo workspaces compared with MonoSmith

A Cargo workspace is a set of one or more Rust packages (workspace members) that are
built and managed together. It is Rust's own answer to a monorepo, and smith's Rust
plugin uses it.

## What Cargo workspaces give you

- **One `Cargo.lock`** at the workspace root, shared by all members.
- **One output directory**, `target/`, shared by default.
- **Commands across members**, such as `cargo check --workspace`.
- **Inherited metadata and dependencies.** The root `[workspace]` table can define
  `package`, `dependencies` and `lints` that members inherit with
  `version.workspace = true` or `serde = { workspace = true }`.
- **A build cache in the target directory.** Cargo "stores the output of a build" there,
  and dep-info files record what a rebuild depends on. Sharing built dependencies
  across workspaces is documented through a third-party tool, `sccache`, set as
  `RUSTC_WRAPPER`.
- **Publishing all members.** `cargo publish --workspace` publishes every member;
  choosing versions is left to you.
- **Scaffolding and extension.** `cargo new` and `cargo init` create packages, and
  programs named `cargo-<name>` on `PATH` act as third-party subcommands.
- **Continuous integration samples.** The Cargo book has CI examples for GitHub
  Actions, GitLab CI, sr.ht and CircleCI.
- Licence of Cargo: MIT or Apache-2.0 (the repository holds both licence files).

## Where MonoSmith differs

| | Cargo workspaces | MonoSmith |
|---|---|---|
| Languages | Rust | eleven languages; Rust crates run through the `rust` plugin |
| Beyond Rust | not applicable | tasks across languages, cache, CI, release |
| Affected, remote execution | no changed-since selection and no remote execution found (searched the Cargo book) | `smith affected`; none |
| Health check, editors | none found in the Cargo book (searched "rust-analyzer", "IDE") | `smith doctor`; a VS Code extension, a terminal interface and an MCP server |
| Repository status | MIT or Apache-2.0, not archived, last push 2026-10-25 (GitHub API, 2026-10-25) | MIT, not archived, first commit 2026-10-15; release 1.62.0 |

smith's `rust` plugin treats the crates of a root Cargo workspace as projects and
runs `cargo build`, `test`, `clippy`, `check`, `fmt --check` and `package`. Builds go
to `dist/cargo`, and packages publish with `cargo publish`. The Cargo workspace stays
the source of truth for Rust; smith adds the cross-language parts.

## When to stay with Cargo alone

- The repository is Rust only. Cargo already gives you a workspace, a lockfile and
  inherited dependencies, and smith would add a layer you may not need.

## When MonoSmith is the better fit

- Rust crates live beside other languages in one repository, and you want one pipeline
  and one release process.

## Moving to MonoSmith

No migration step moves Rust code: the crates stay where they are and the root
`Cargo.toml` keeps its `[workspace]`. `smith adopt --dry-run` reports what adopting
involves. smith's `adopt --help` names npm workspaces, Go modules and Python
projects. Tried on 2026-10-25 with smith 1.62.0 on a scratch two-crate Cargo
workspace (one crate depending on the other by path): `smith adopt --dry-run` listed
both crates as `rust` projects and reported nothing that blocks. The help text does
not promise this, and nothing past the dry run was tried. `smith add` can add Rust
crates to a smith workspace.

## Sources

Retrieved 2026-10-25.

The Cargo book was read as Markdown source files from the repository's
`doc/book/src` folder, not through a summarising tool. Paths are relative to it.

- Cargo workspaces: <https://doc.rust-lang.org/cargo/reference/workspaces.html>
- Build cache: `reference/build-cache.md` ("Cargo stores the output of a build"; "A third party tool, sccache, can be used to share built dependencies").
- Publishing: `commands/cargo-publish.md` (`--workspace`: "Publish all members in the workspace.").
- Dependency graph: `commands/cargo-build.md` (`--keep-going`: "Build as many crates in the dependency graph as possible").
- Subcommands: `reference/environment-variables.md` ("3rd party subcommands (ie. programs named `cargo-foobar` placed in `$PATH`)").
- CI: `guide/continuous-integration.md`.
- Cargo repository (licence): <https://github.com/rust-lang/cargo>. It holds `LICENSE-MIT` and `LICENSE-APACHE`; the GitHub API on 2026-10-25 reports Apache-2.0 (one value), archived false, last push 2026-10-25.
