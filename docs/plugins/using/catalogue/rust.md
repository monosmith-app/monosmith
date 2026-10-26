# The rust plugin

The rust plugin makes every crate in a root Cargo workspace a smith project. It runs `cargo build`, `test`, `clippy`, `check`, `fmt --check` and `package` for one crate at a time (`-p <crate>`), keeps the root `Cargo.toml` `members` list in step with the crates smith creates and removes, sends every build to `dist/cargo` instead of `target/`, installs Rust in CI, and publishes package crates to crates.io. It is a built-in plugin: no installation needed.

## At a glance

| | |
|---|---|
| Plugin name | `rust` (built-in, version follows smith, 1.62.0 here) |
| Roles | `language`, `toolchain`, `scaffold` (all three by default for a language plugin) |
| `--lang` values | `rust` |
| Manifest that makes a directory a project | `Cargo.toml` with a `[package]` table. A root `Cargo.toml`, and a manifest without `[package]` (a virtual workspace), are not projects. |
| Requires / conflicts / provides | none / none / no task slots |
| Capabilities advertised | `discover`, `generate`, `version`, `publish`, `ci`, `doctor`, `remove`, `published`. Not advertised: `dependencies`, `declarations`, `latest`. |
| Stacks it publishes | none |
| Project kinds it can create | `app`, `lib`, `package` (`rust-app`, `rust-lib`, `rust-package`) |

Capabilities are from `smith plugin list --json`.

## What it detects

(`libs/plugin-rust/discovery/`)

| Field | How it is read |
|---|---|
| Project | each `Cargo.toml` below the workspace root that has `[package]` and a `name` |
| Name | the last element of the directory (`packages/kit` is `kit`); the crate name is different (see Exports) |
| Kind | `[package.metadata.smith] kind = "app" \| "lib" \| "package"`; else the top directory (`apps/`, `libs/`, `packages/`); else `app` if `src/main.rs` exists, otherwise `lib` |
| Stack | `[package.metadata.smith] stack` |
| Version | `[package] version`, when it is a literal string. `version.workspace = true` (a table) reads as empty. |
| Description | `[package] description`, when a string |
| Exports | the crate name, `[package] name` |
| Imports (graph edges) | the crates it depends on by `path` in `[dependencies]`, `[dev-dependencies]` and `[build-dependencies]`, by package name (a dependency renamed with `package = "..."` uses that name). Registry dependencies are not edges. |
| Toolchain | `rustc <version>` from `rustc --version`; left out when `rustc` is not on PATH |

Real output after `apps/cli` got `rs-demo-core = { path = "../../libs/core" }`:

```
$ smith graph
cli  app · rust · apps/cli
  └─ core
core  lib · rust · libs/core
kit  package · rust · packages/kit
```

## Tasks

All tasks run in the crate directory, scoped with `-p <crate name>`. Cargo builds a crate's path dependencies itself, so no task has `dependsOn`. Inputs of the cached tasks: `**/*`, `!target/**`, `{workspaceRoot}/Cargo.toml`, `{workspaceRoot}/Cargo.lock`.

| Task | Command | Appears for | Cached |
|---|---|---|---|
| `build` | `cargo build -p <crate>` | every crate | no (and no outputs declared) |
| `test` | `cargo test -p <crate>` | every crate | yes |
| `lint` | `cargo clippy -p <crate> --all-targets -- -D warnings` | every crate | yes |
| `typecheck` | `cargo check -p <crate> --all-targets` | every crate | yes |
| `format` | `cargo fmt -p <crate> --check` | every crate | yes |
| `start` | `cargo run -p <crate>` | `app` | no |
| `package` | `cargo package -p <crate> --allow-dirty` | `package` | no |

`build` is a debug build: it writes `dist/cargo/debug/`, and the app binary is named after the crate (real: `dist/cargo/debug/rs-demo-cli.exe`). There is no `--release` task. `package` writes the `.crate` under `dist/cargo/package/`, not under `dist/drop`, so `smith ci pack` yields no file for a Rust workspace (the task runs, but its output is not in the drop folder; not verified in a CI run). An `app` has no `package` task.

With `[conventions.vertical-slices]` each crate also gets `slice-check` (inputs `src/**/*.rs`, `tests/**/*.rs`) and `lint` runs after it; see [vertical-slices](vertical-slices.md).

Real output:

```
$ smith explain cli
cli has 7 tasks:
  build        cargo build -p rs-demo-cli
  format       cargo fmt -p rs-demo-cli --check  (cached)
  lint         cargo clippy -p rs-demo-cli --all-targets -- -D warnings  (cached, after slice-check)
  slice-check  smith slices check --lang rust .  (cached)
  start        cargo run -p rs-demo-cli
  test         cargo test -p rs-demo-cli  (cached)
  typecheck    cargo check -p rs-demo-cli --all-targets  (cached)

$ smith explain core:build
core:build — cargo build
  command    cargo build -p rs-demo-core
  cwd        libs/core
  dependsOn  —
  inputs     **/*, !target/**, {workspaceRoot}/Cargo.toml, {workspaceRoot}/Cargo.lock
  outputs    —
  env        —
  cache      false

$ smith explain kit:package
kit:package — cargo package: the .crate crates.io would get
  command    cargo package -p rs-demo-kit --allow-dirty
  ...
  cache      false
```

`smith run build`, `test`, `lint`, `typecheck`, `format` and `package` all succeeded (3/3 or 6/6 tasks).

## Creating a project

```
smith add app|lib|package <name> --lang rust
```

The first crate creates the root `Cargo.toml` (`[workspace]`, `resolver = "2"`, `members`) and `.cargo/config.toml` (`[build] target-dir = "dist/cargo"`); later crates are appended to `members` (a single-line or multi-line array is kept in its style). A workspace that already has a `.cargo/config.toml` keeps its own. If the root `Cargo.toml` is a crate rather than a workspace, `smith add` refuses: `the root Cargo.toml is a crate (<name>), not a workspace: make it a [workspace] first`.

The crate name is the workspace name, a dash and the project name (`rs-demo` + `core` = `rs-demo-core`), unless the project name already starts with that, so a published crate does not clash on crates.io.

Real output:

```
$ smith add lib core --lang rust
Adding lib core (rust) in libs/core
Wrote 8 file(s):
  .cargo/config.toml
  .devcontainer/devcontainer.json
  .github/workflows/ci.yml
  Cargo.toml
  libs/core/Cargo.toml
  libs/core/src/greeting/greet_use_case.rs
  libs/core/src/greeting/mod.rs
  libs/core/src/lib.rs
Import it from another crate as rs-demo-core = { path = "<relative path>" }.
Try: smith test -p core

$ smith add app cli --lang rust
Wrote 5 file(s):
  Cargo.toml
  apps/cli/Cargo.toml
  apps/cli/src/greeting/greet_use_case.rs
  apps/cli/src/greeting/mod.rs
  apps/cli/src/main.rs
```

The root `Cargo.toml` after three crates:

```toml
[workspace]
resolver = "2"
members = ["libs/core", "apps/cli", "packages/kit"]
```

| Generated | What it is |
|---|---|
| `<dir>/Cargo.toml` | `edition = "2021"`, `version = "0.0.0"`, `[package.metadata.smith] kind` (and `stack`), empty `[dependencies]`. A `package` also gets `description` and `license = "MIT"`; apps and libs get `publish = false`. |
| `<dir>/src/greeting/` | a worked slice: `mod.rs` and `greet_use_case.rs` with `pub fn greet(name: &str) -> String` (`Hello, <name>!`) and a unit test |
| `<dir>/src/main.rs` (app) | prints `greeting::greet("world")` |
| `<dir>/src/lib.rs` (lib, package) | `pub mod greeting;` |

The plugin publishes no stack. It does not write `rust-toolchain.toml`.

## Settings

The plugin has no `[projects.<p>.settings.rust]` keys.

| Where | Key | Effect | Read in |
|---|---|---|---|
| `Cargo.toml` (crate) | `[package.metadata.smith] kind`, `stack` | kind and stack | `discovery/cargo_manifest_mapper.go` |
| `Cargo.toml` (crate) | `[package] version`, `description`, `name` | version, description, export | same |
| `Cargo.toml` (root) | `[workspace] members` | which crates cargo (and `-p`) can see; doctor checks it | `pipeline/check_workspace_use_case.go` |
| `.cargo/config.toml` | `[build] target-dir` | where builds go (`dist/cargo`) | written by `scaffolding/generate_crate_use_case.go` |
| `rust-toolchain.toml` | `channel = "..."` | the Rust version CI installs (default `1.90`) | `pipeline/ci_contribution_use_case.go` |
| `smith.toml` | `[ci] audit_policies = ["rust"]`, `audit_warn`, `audit_max_unsafe` | cargo-deny and an `unsafe` budget in the audit | same |
| `smith.toml` | `[toolchains] rustc = "1.90"` | version pin checked by doctor | host |
| `deny.toml` | whole file | what `cargo deny check licenses bans` allows (yours) | cargo-deny |
| env | `CARGO_REGISTRY_TOKEN` | crates.io token for `cargo publish` (read by cargo) | CI credential |

## Toolchain

The pin label is `rustc`, probed with `rustc --version`. Real output with `[toolchains] rustc = "1.90"` on a machine with 1.76.0:

```
✖ rustc matches the pinned version 1.90
    [toolchains] pins rustc 1.90, but this machine runs 1.76.0: tasks cached on another version are misses, and results can differ from CI
    → install rustc 1.90 (`mise use rust@1.90`, asdf, or your installer; `smith doctor --fix` writes .tool-versions for mise and asdf), or change [toolchains] rustc
```

smith does not install Rust. The machine needs `cargo`, `rustc`, `clippy` and `rustfmt` (the CI toolchain installs `clippy, rustfmt`). The plugin's doctor checks that `cargo` exists and compares `rustc` with the version CI installs (`1.90`, or the `channel` of `rust-toolchain.toml`): older fails, newer warns. A machine on 1.76.0 fails the check instead of failing later in `smith test` with `feature edition2024 is required`. To make local, CI and the dev container agree, pin the version once in `rust-toolchain.toml` (`channel = "1.99.1"`) and run `smith ci generate`.

Output goes to `dist/cargo` (set by the `.cargo/config.toml` the plugin writes): `dist/cargo/debug/` for builds, `dist/cargo/package/` for `cargo package`.

## Releasing

| Aspect | Behaviour |
|---|---|
| Version | `[package] version` in the crate's `Cargo.toml`. The release writes the literal version there (keeping formatting), and also updates the `version = "..."` requirement of inline path dependencies on other released crates in the workspace (`dep = { path = "...", version = "..." }`), then publishes and restores. A crate with `version.workspace = true` has no literal to write; nothing is changed for it. |
| Tag | `<name>@<version>`, name being the directory name: `kit@0.1.0` (real). |
| What is published | `package` crates only, with `cargo publish -p <crate> --allow-dirty` (the manifests carry the release version while smith releases, and cargo would refuse the dirty tree). Apps and libs are released by their tag. An app with a `[projects.<app>.release]` section is tagged and gets a GitHub Release, without assets: `Rust release assets are not built yet: the app is released by its tag`. |
| Registry | crates.io only. There is no setting for another registry. `published` reads the newest stable version from `https://crates.io/api/v1/crates/<crate>`. |
| Credentials | `CARGO_REGISTRY_TOKEN` (a crates.io API token that can publish the workspace's crates), listed as required when any crate is a package. |
| Dry run | `smith release --dry-run` runs `cargo publish -p <crate> --allow-dirty --dry-run` for each package crate, so cargo does package, verify and contact crates.io for the index ("Updating crates.io index"); it does not upload. The manifest still holds `0.0.0` at that point. |

Real dry run (excerpt):

```
$ smith release --dry-run
Dry run: nothing is published or tagged.
  kit                  unreleased → 0.1.0  (1 releasing commit(s))
    Updating crates.io index
warning: manifest has no documentation, homepage or repository.
   Packaging rs-demo-kit v0.0.0 (...\packages\kit)
   Verifying rs-demo-kit v0.0.0 (...\packages\kit)
    Packaged 5 files, 1.2KiB (912.0B compressed)
   Uploading rs-demo-kit v0.0.0 (...\packages\kit)
warning: aborting upload due to dry run
```

cargo's output goes to stderr in both the text and the `--json` run; the JSON reports `"tag": "kit@0.1.0"`, `"registry": "crates.io"`, `"published": false`.

## CI

Real output of `smith ci generate` (GitHub):

```yaml
      - uses: dtolnay/rust-toolchain@master
        with:
          toolchain: '1.90'
          components: 'clippy, rustfmt'
      - name: Cache the cargo download store
        uses: actions/cache@v4
        with:
          path: ~/.cargo/registry
          key: smith-cargo-${{ runner.os }}-${{ hashFiles('**/Cargo.lock') }}
      ...
        env:
          CARGO_REGISTRY_TOKEN: ${{ secrets.CARGO_REGISTRY_TOKEN }}
```

| Contribution | What |
|---|---|
| Toolchain | Rust at the `channel` of `rust-toolchain.toml`, else `1.90`, with `clippy` and `rustfmt`. GitHub: `dtolnay/rust-toolchain@master`. Azure and GitLab: `rustup` with `--profile minimal`, then `cargo` on PATH (`##vso[task.prependpath]` on Azure). Dev container: the `rust` feature, `profile: minimal`. |
| Package store | `~/.cargo/registry` (`CARGO_HOME`), keyed on every `Cargo.lock` |
| Setup (`smith ci setup`) | `cargo fetch` (blocking). Real: `▶ Fetch crates`, `smith ci setup passed in 100ms`. |
| Audit | by default nothing: `smith ci audit` printed `nothing to audit`. There is no `cargo audit` step. With `[ci] audit_policies = ["rust"]`: setup also runs `cargo install --locked cargo-deny --version 0.18.5`, and the audit runs `cargo deny check licenses bans` and `smith rust-policy --max-unsafe=<audit_max_unsafe> <crate dirs>` (counts `unsafe` blocks, functions, impls and traits in `.rs` files outside `target/` and hidden folders). Blocking unless `"rust"` is in `audit_warn`. |
| Credentials | `CARGO_REGISTRY_TOKEN` (required, only when a package crate exists). `smith ci check`: `GitHub: add the repository secrets CARGO_REGISTRY_TOKEN`. |

`smith rust-policy --max-unsafe=0` over the three generated crates printed nothing and exited 0. cargo-deny is not installed here, so the full opt-in audit was not run.

## Doctor, dependencies and removal

Doctor checks (`pipeline/check_workspace_use_case.go`), shown only when the workspace has Rust crates:

| Check | Fails when |
|---|---|
| `cargo is installed` | `cargo --version` fails; remedy: install Rust with rustup |
| `Rust is <version>, as in CI` | `rustc --version`, as rustup resolves it in the workspace (so a `rust-toolchain.toml` counts), is older than the version CI and the dev container install (`1.90`, or the `channel` of `rust-toolchain.toml`); a newer local Rust is a warning. Nothing is reported when `rustc` is absent (the `cargo` check says so) or the channel is `stable` or `nightly`. Fix: write `rust-toolchain.toml` (never over an existing one) and `rustup toolchain install` the pinned version; for a newer local Rust, pin what the machine runs |
| `the root Cargo.toml is a workspace` | there is no root `Cargo.toml` |
| `the root Cargo.toml is valid TOML` | it does not parse |
| `workspace member <path> exists` | a (non-glob) `members` entry has no `Cargo.toml`; remedy: remove it from `members` |
| `the Cargo workspace lists <dir>` | a crate directory is not in `members` (a glob member matches), so `cargo -p` cannot find it |

All passed in the test run (`21 passed, 0 failed`). Only the version check has a `--fix`, and it runs `rustup`, so it is asked about, never applied by `--yes`.

Dependencies: the plugin has no `dependencies` capability. `smith install -w kit serde` answers `smith: plugin rust cannot change kit's dependencies; edit its manifest instead` (real), and `smith up` has nothing to list for Rust (`Every declared dependency is at its latest release.`). Use `cargo add` / edit `Cargo.toml`.

Removal: `smith remove <project>` deletes the directory and takes its entry out of `members` in the root `Cargo.toml`. Real dry run: `Would remove core (libs/core)`, `changes Cargo.toml`, `the pipelines would be regenerated`.

## Requirements and limits

- `cargo` (with `rustc`, `clippy`, `rustfmt`) on PATH. Network for the first `cargo fetch` and for `release --dry-run`.
- A root workspace `Cargo.toml` is required. A repository whose single root `Cargo.toml` is a crate is not a smith project (and `smith add` refuses to turn it into a workspace for you).
- The root `Cargo.toml` is yours after the first write; smith edits only the `members` array and, at release time, the version lines.
- `build` is a debug, uncached build with no declared outputs; there is no release build task and no Rust release assets.
- No vulnerability audit by default (no `cargo audit`), no private cargo registries, no dependency management through smith.

## Verified

Windows 11 Pro 10.0.26300, smith 1.62.0, cargo 1.76.0, rustc 1.76.0, clippy 0.1.76, rustfmt 1.7.0-stable. Throw-away workspace `rs-demo` from `smith new rs-demo`. Run: `smith add lib|app|package ... --lang rust`, `smith projects`, `smith graph`, `smith explain`, `smith run build|test|lint|typecheck|format|package`, `smith doctor` (clean, and with a wrong `[toolchains] rustc` pin), `smith ci generate`, `smith ci setup`, `smith ci audit`, `smith ci check`, `smith release --dry-run [--json]` (including the real `cargo publish --dry-run` against crates.io), `smith install -w`, `smith up --check`, `smith remove core --dry-run`, `smith rust-policy`.

Not verified:

- A real publish (needs a crates.io token and would publish).
- The opt-in rust audit policy with cargo-deny, and the CI jobs themselves (Azure/GitLab `rustup` steps, `rust-toolchain.toml` overriding `1.90`).
- Rust 1.90 behaviour: the machine runs 1.76.0, so generated code was only compiled with that.
- Whether `smith ci pack` leaves anything in `dist/drop` for Rust (read from the task definitions: no).

## See also

- [The plugin catalogue](README.md)
- [vertical-slices](vertical-slices.md)
- [Releasing](../../../reference/releasing.md), [CI](../../../reference/ci.md)
- [Writing a language plugin](../../creating/language-plugin.md)
