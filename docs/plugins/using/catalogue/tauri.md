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
| Capabilities | `generate`, `contribute`, `version`, `ci`, `doctor` |
| Stack it publishes | `node-vitest-tauri`: scaffold `typescript`; toolchain `typescript`, `eslint`, `vitest`, `tauri`. The `vite` plugin publishes `react-tauri`: scaffold `vite`; toolchain `vite`, `eslint`, `vitest`, `tauri` |
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

`smith run tauri-dev -p studio` runs the app in development mode: it starts the web dev server
itself, because the scaffold's `tauri.conf.json` has `beforeDevCommand` `npm run start` (Vite on
`http://localhost:5173`, the `devUrl`), opens the app window, and reloads it as you edit; stop it
with Ctrl+C. It is never cached. An app whose `tauri.conf.json` has no `beforeDevCommand` (one made
with `npx tauri init`) must have its dev server running already: start it with
`smith run start -p studio` in another terminal, or add the command to `beforeDevCommand`.

The cargo target directory, and so the installers, go under `dist/tauri/<project dir>`, beside the
web build rather than in it. The path is absolute: cargo runs inside `src-tauri`, and the Tauri CLI
reads the same flag to find the bundles, so a relative path would depend on which of them resolves it.

## What `smith doctor` checks

| Check | Fails when |
|---|---|
| `the Rust toolchain is installed` | `rustc` or `cargo` is not on `PATH`. Remedy: install Rust from rustup.rs |
| `the tauri CLI is installed` | `node_modules` has no `@tauri-apps/cli`. Remedy: `smith install -D @tauri-apps/cli` |
| `src-tauri is outside the root Cargo workspace` | A warning: the root `Cargo.toml` `[workspace] members` glob matches `<project>/src-tauri` and `exclude` does not name it. Remedy: narrow `members`, or list it under `exclude` |
| `frontendDist of <project> is the project's build output` | A warning: `build.frontendDist` in `src-tauri/tauri.conf.json` is a folder other than the one smith builds the web app into (`dist/<project dir>`), so `tauri build` fails with "Unable to find your web assets". A URL is not checked. Remedy: set it to the path the check prints |
| `the versions of <project> agree` | A warning: `package.json`, `src-tauri/Cargo.toml` and `src-tauri/tauri.conf.json` carry different versions (a `tauri.conf.json` version that is a path to `package.json` counts as the same). Remedy: use `"version": "../package.json"` and the same version in `Cargo.toml`, or let `smith release` write them |
| `the system libraries Tauri needs on Linux` | `pkg-config` does not find `webkit2gtk-4.1`, `gtk+-3.0` or `librsvg-2.0` (Linux only). Remedy: `sudo apt-get install -y pkg-config libwebkit2gtk-4.1-dev libgtk-3-dev librsvg2-dev` |

## Which toolchains combine with `tauri`

`tauri` provides only `tauri-build`, so it combines with any toolchain whose tasks do not clash.
The web build comes from exactly one of `typescript` or `vite`: both provide the `build` and
`typecheck` tasks, and smith refuses a stack that lists both ("typescript and vite both provide the
build task; keep one"). With a Vite app, name `vite` and leave `typescript` out:

```toml
[stacks.studio]
language  = "node"
toolchain = ["vite", "vitest", "eslint", "tauri"]   # not "typescript" as well
```

`vitest` and `eslint` combine with either, as does `playwright`. Do not name `jest` beside `vitest`
(both provide `test`).

## Scaffolding

One command makes a React desktop app:

```sh
smith add app studio --stack react-tauri    # vite, eslint, vitest, tauri
```

The `vite` plugin writes the React app; the `tauri` plugin then adds, for a project that has an
`index.html` (a web page to wrap), and never over a file that is already there:

- `src-tauri/`: `Cargo.toml` with its own empty `[workspace]`, `build.rs`, `src/main.rs`,
  `src/lib.rs`, `capabilities/default.json`, placeholder `icons/`, a `.gitignore` for `target`, `gen`
  and `binaries`, and `tauri.conf.json`.
- `@tauri-apps/api` in the project's dependencies, `@tauri-apps/cli` in the root's devDependencies
  and a `tauri` script.
- the project's `src-tauri` under `exclude` in the root `Cargo.toml`'s `[workspace]`, when its
  `members` would take it in.

`tauri.conf.json` is written to match how smith builds: `frontendDist` is the Vite output
(`../../../dist/apps/studio` from `apps/studio/src-tauri`), `devUrl` is Vite's `http://localhost:5173`
with `beforeDevCommand` `npm run start`, and there is no `beforeBuildCommand`: `tauri-build` runs
after the project's own `build`, so a second build would only repeat it. The bundle identifier is
`com.<workspace>.<project>`; change it before you ship.

A project made with `--stack node-vitest-tauri` has no web page: it gets `@tauri-apps/cli` and a
`tauri` script, and `npx tauri init` writes `src-tauri`.

An existing Node app does not join with a smith command yet: create `src-tauri` with
`npx tauri init` (answer the frontend questions with `../../../dist/apps/<project>` for the
dist folder and leave the build command empty), or copy the files of a `react-tauri` project.

### Icons

The scaffold's icons are plain placeholders so the first build runs. Make the real ones from a
square PNG (1024x1024 is best) in the project:

```sh
npx tauri icon logo.png
```

It overwrites `src-tauri/icons/`.

## Releasing and CI

The plugin gives every pipeline with a Tauri app a Rust toolchain and a setup step that, on a Linux
runner, installs `pkg-config`, `libwebkit2gtk-4.1-dev`, `libgtk-3-dev`,
`libayatana-appindicator3-dev` and `librsvg2-dev`. `smith ci setup` runs it, so the release's
`assets` job gets it too.

The Rust toolchain is named `rust`, like the `rust` plugin's, and the pipeline keeps the first of
two with the same name: the `rust` plugin's, with its version, components and targets. The plugin
never adds a second Rust step that would replace your pin. Its own version is the workspace's pin
(the `channel` of `rust-toolchain.toml`, else the `rust-version` of the root `Cargo.toml`), and
`stable` when there is none; that matters only when no Rust project of the `rust` plugin is in the
workspace.

The apt step is not a line of `ci.yml`: the workflow runs `smith ci setup`. To see it, read that
step's log: on a Linux runner it prints `installing the system libraries Tauri needs (apt-get)`,
and the steps table at the end lists `Install the system libraries Tauri needs`. Elsewhere it
prints nothing and passes.

### Releasing the installers

The installers are released through [release files from a command](../../../reference/releasing.md#release-files-from-a-command),
and the plugin supplies the command. Switch GitHub Releases on and name the project; that is all:

```toml
[release]
github_releases = true

[projects.studio.release]
```

For a project the plugin applies to, the table's missing settings default to:

| Setting | Default |
|---|---|
| `build` | `smith run tauri-build -p studio` |
| `assets` | `dist/tauri/apps/studio/release/bundle/**/*.{msi,exe,dmg,deb,rpm,AppImage}` |
| `platforms` | `windows-x64`, `macos-arm64`, `macos-x64`, `linux-x64` |
| `asset_name` | `{product}-{version}-{platform}.{ext}` |
| `checksums` | `true` |

`linux-arm64` is not a default (Linux bundles for it are rarely wanted and cost a slower runner);
add it with `platforms = [..., "linux-arm64"]`, which smith runs on `ubuntu-24.04-arm`. What the
table sets wins, field by field: `platforms = ["windows-x64"]` alone keeps the default `build` and
`assets`. A table that sets only what a plugin's own archives use (`os_alias`, `npm`, ...) is not
given the defaults. The table must exist: a Tauri app with none is not released, so building on
four runners is always your choice.

Each runner builds with `tauri-build`, then uploads its installers renamed after the project, the
version and the platform, plus a checksums file. Tauri's own names (`Studio_1.0.0_x64-setup.exe`,
with spaces, the same on every runner of an OS) would collide or confuse:

| Tauri's bundle | Uploaded as |
|---|---|
| `msi/Studio_1.0.0_x64_en-US.msi` | `studio-1.0.0-windows-x64.msi` |
| `nsis/Studio_1.0.0_x64-setup.exe` | `studio-1.0.0-windows-x64.exe` |
| `dmg/Studio_1.0.0_aarch64.dmg` | `studio-1.0.0-macos-arm64.dmg` |
| `deb/studio_1.0.0_amd64.deb`, `rpm/...rpm`, `appimage/...AppImage` | `studio-1.0.0-linux-x64.deb`, `.rpm`, `.AppImage` |
| the files of that runner | `studio-1.0.0-windows-x64.sha256` (`sha256sum` format) |

Two bundles of one runner that end in the same extension keep their folder's name
(`studio-1.0.0-windows-x64-nsis.exe`). The product in the name is the project's name; set
`product` in the table to change it.

A `build` that needs another program of the same release (VideoCraw's studio bundles the
`videocraw` binary) builds it first in the same command, and `assets` then lists what you want
(sidecars and resources are not handled by the plugin yet).

### Versions

smith owns the version, and a Tauri app keeps it in three files. `smith release` writes the new
version into `package.json` (the `node` plugin), `src-tauri/Cargo.toml` (`[package] version`) and
`src-tauri/tauri.conf.json` (the top-level `version`, not the ones nested in `bundle`), then puts
the files back after the build. The assets job of every runner does the same before it builds, so
the installers carry the version in their file names and in their metadata. The scaffold sets
`"version": "../package.json"` in `tauri.conf.json` (Tauri 2 reads the version from that file); a
`tauri.conf.json` with a path there is left alone. Only `tauri.conf.json` is edited, not
`tauri.conf.json5` or `Tauri.toml`.

`smith doctor` warns (`the versions of <project> agree`) when the files that carry a literal
version differ, with the versions it found.

## Verified

Unit tests, and on Windows `smith add app studio --stack react-tauri` followed by `smith doctor`
and `smith run tauri-build -p studio` in a scratch workspace: the MSI and the NSIS installer were
built under `dist/tauri/apps/studio/release/bundle`. **Not verified:** macOS and Linux builds
(including the generated `icon.icns`), a release's `assets` job, and the installer names in the
glob above on those platforms.

## See also

- [vitest](vitest.md), [typescript](typescript.md), [playwright](playwright.md)
- [Plugins (reference)](../../../reference/plugins.md)
