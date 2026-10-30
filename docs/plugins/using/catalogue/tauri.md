# The tauri plugin

The `tauri` plugin wraps a desktop app built with [Tauri](https://tauri.app) in a Node project. It
owns the `tauri-build` task: a cached build of the app's installers, after the web build it
packages, plus a `tauri-dev` task and `smith doctor` checks for what the build needs from the
machine. It is both a toolchain a stack can name and a convention you can switch on
workspace-wide with `[conventions.tauri]`.

## At a glance

| | |
|---|---|
| Kind | `convention` |
| Roles | `toolchain` |
| Provides (slots) | `tauri-build` |
| Requires | language plugin `node` |
| Applies to | `node` projects |
| Capabilities | `generate`, `contribute`, `doctor` |
| Stack it publishes | `node-vitest-tauri`: scaffold `typescript`; toolchain `typescript`, `eslint`, `vitest`, `tauri` |
| Project kinds | any Node project with a `src-tauri` folder |

## What it detects

A Node project has a Tauri app when `src-tauri/` holds a `tauri.conf.json` (or `.json5`,
`Tauri.toml`), or a `Cargo.toml` that depends on `tauri`.

The Rust side is part of that Node project. The `rust` plugin therefore does not discover
`<project>/src-tauri` as a crate of its own when the folder next to it holds a `package.json`.
Keep `src-tauri` out of the root Cargo workspace too (it has its own `Cargo.lock`): do not list it
in the root `[workspace] members`.

## Tasks

| Task | Runs | Cached |
|---|---|---|
| `tauri-build` | `tauri build -- --target-dir <dist/tauri/<project dir>>`, after `^build` and the project's own `build` | yes; output `dist/tauri/<project dir>` |
| `tauri-dev` | `tauri dev` | no |

The cargo target directory, and so the installers, go under `dist/tauri/<project dir>`, beside the
web build rather than in it.

## What `smith doctor` checks

| Check | Fails when |
|---|---|
| `the Rust toolchain is installed` | `rustc` or `cargo` is not on `PATH`. Remedy: install Rust from rustup.rs |
| `the tauri CLI is installed` | `node_modules` has no `@tauri-apps/cli`. Remedy: `smith install -D @tauri-apps/cli` |
| `the system libraries Tauri needs on Linux` | `pkg-config` does not find `webkit2gtk-4.1`, `gtk+-3.0` or `librsvg-2.0` (Linux only). Remedy: `sudo apt-get install -y pkg-config libwebkit2gtk-4.1-dev libgtk-3-dev librsvg2-dev` |

## Scaffolding

A project made with `--stack node-vitest-tauri` gets `@tauri-apps/cli` and a `tauri` script. It
does not write `src-tauri`: run `npx tauri init` in the project.

## Releasing and CI

No `publish` and no `ci` method yet. Per-platform installers as release assets, and setting up the
Rust toolchain in the assets job, are not done.

## Verified

Unit tests only. **Not verified:** a real `smith run tauri-build` (it needs Rust, the platform's
libraries and the network), including that cargo resolves the relative `--target-dir` from
`src-tauri`.

## See also

- [vitest](vitest.md), [typescript](typescript.md), [playwright](playwright.md)
- [Plugins (reference)](../../../reference/plugins.md)
