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
| Capabilities | `generate`, `contribute`, `ci`, `doctor` |
| Stack it publishes | `node-vitest-tauri`: scaffold `typescript`; toolchain `typescript`, `eslint`, `vitest`, `tauri` |
| Project kinds | any Node project with a `src-tauri` folder |

## What it detects

A Node project has a Tauri app when `src-tauri/` holds a `tauri.conf.json` (or `.json5`,
`Tauri.toml`), or a `Cargo.toml` that depends on `tauri`.

The Rust side is part of that Node project. The `rust` plugin therefore does not discover
`<project>/src-tauri` as a crate of its own when the folder next to it holds a `package.json`.
Keep `src-tauri` out of the root Cargo workspace too (it has its own `Cargo.lock`): do not list it
in the root `[workspace] members`. `smith doctor` warns when a member glob matches it.

## Tasks

| Task | Runs | Cached |
|---|---|---|
| `tauri-build` | `tauri build -- --target-dir {workspaceRoot}/dist/tauri/<project dir>`, after `^build` and the project's own `build` | yes; output `dist/tauri/<project dir>` |
| `tauri-dev` | `tauri dev` | no |

The cargo target directory, and so the installers, go under `dist/tauri/<project dir>`, beside the
web build rather than in it. The path is absolute: cargo runs inside `src-tauri`, and the Tauri CLI
reads the same flag to find the bundles, so a relative path would depend on which of them resolves it.

## What `smith doctor` checks

| Check | Fails when |
|---|---|
| `the Rust toolchain is installed` | `rustc` or `cargo` is not on `PATH`. Remedy: install Rust from rustup.rs |
| `the tauri CLI is installed` | `node_modules` has no `@tauri-apps/cli`. Remedy: `smith install -D @tauri-apps/cli` |
| `src-tauri is outside the root Cargo workspace` | A warning: the root `Cargo.toml` `[workspace] members` glob matches `<project>/src-tauri` and `exclude` does not name it. Remedy: narrow `members`, or list it under `exclude` |
| `the system libraries Tauri needs on Linux` | `pkg-config` does not find `webkit2gtk-4.1`, `gtk+-3.0` or `librsvg-2.0` (Linux only). Remedy: `sudo apt-get install -y pkg-config libwebkit2gtk-4.1-dev libgtk-3-dev librsvg2-dev` |

## Scaffolding

A project made with `--stack node-vitest-tauri` gets `@tauri-apps/cli` and a `tauri` script. It
does not write `src-tauri`: run `npx tauri init` in the project.

## Releasing and CI

The plugin gives every pipeline with a Tauri app a Rust toolchain (`tauri-rust`, the `stable`
channel, named apart from the `rust` plugin's so that plugin keeps its own version) and a setup
step that, on a Linux runner, installs `pkg-config`, `libwebkit2gtk-4.1-dev`, `libgtk-3-dev`,
`libayatana-appindicator3-dev` and `librsvg2-dev`. `smith ci setup` runs it, so the release's
`assets` job gets it too.

The installers are released through [release files from a command](../../../reference/releasing.md#release-files-from-a-command).
Tauri names its bundles after the platform and processor (`Studio_1.0.0_x64-setup.exe`,
`Studio_1.0.0_aarch64.dmg`, `studio_1.0.0_amd64.deb`), so one glob serves every runner:

```toml
[release]
github_releases = true

[projects.studio.release]
build     = "smith run tauri-build studio"   # or any command that also builds what the app bundles
assets    = ["dist/tauri/apps/studio/release/bundle/**/*.{msi,exe,dmg,deb,rpm,AppImage}"]
platforms = ["windows-x64", "macos-arm64", "linux-x64"]
```

A `build` that needs another program of the same release (VideoCraw's studio bundles the `videocraw`
binary) builds it first in the same command. No setting does this for you yet: the plugin has no
`publish` method.

## Verified

Unit tests only. **Not verified:** a real `smith run tauri-build` or a release's `assets` job (they
need Rust, the platform's libraries and the network), the installer names in the glob above, and
that the Tauri CLI finds the bundles under an absolute `--target-dir`.

## See also

- [vitest](vitest.md), [typescript](typescript.md), [playwright](playwright.md)
- [Plugins (reference)](../../../reference/plugins.md)
