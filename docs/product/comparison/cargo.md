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
- Licence of Cargo: MIT or Apache-2.0.

## Where MonoSmith differs

| | Cargo workspaces | MonoSmith |
|---|---|---|
| Languages | Rust | eleven languages; Rust crates run through the `rust` plugin |
| Beyond Rust | not applicable | tasks across languages, cache, CI, release |

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
involves. Note that smith's `adopt` is documented for npm workspaces, Go modules and
Python projects; Rust adoption by `adopt` was not verified. `smith add` can add Rust
crates to a smith workspace.

## Sources

Retrieved 2026-10-25.

- Cargo workspaces: <https://doc.rust-lang.org/cargo/reference/workspaces.html>
- Cargo repository (licence): <https://github.com/rust-lang/cargo>
