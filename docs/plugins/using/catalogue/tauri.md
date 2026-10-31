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
| `tauri-prepare` | only with sidecars or resources in `[projects.<p>.settings.tauri]`: builds the sidecars and copies them and the resources ([below](#sidecars-and-resources)); `tauri-build` and `tauri-dev` run it first | when everything it reads is in the cache key |
| `tauri-verify` | only with sidecars or resources: checks the built app holds the sidecars; after `tauri-build` | no |

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
| `the tauri settings of <project> are valid` | `[projects.<p>.settings.tauri]` has an unknown key or an entry that is not a table of the shape above. Remedy: fix it in `smith.toml` |
| `the sidecar <name> of <project> is a project that exists` | The sidecar's `project` is not a project of the workspace, or not a Rust project. Remedy: name one of `smith projects`, or give a `from` |
| `the sidecar <name> of <project> exists` | A sidecar given as `from` is not there (a warning when it is under `dist/`, which a build makes, or when its environment variable is unset). Remedy: build it first, or fix `from` |
| `the resource <from> of <project> exists` | The resource's file or folder is missing or the pattern matches nothing (a warning under `dist/` or with an unset environment variable). Remedy: fix `from` |
| `the sidecar <name> of <project> is named for this machine` | A warning: `tauri.conf.json` lists `<name>` in `bundle.externalBin`, smith does not manage it, and `<name>-<target triple>[.exe]` for this machine is missing, which Tauri requires. Remedy: declare it in `settings.tauri` sidecars |
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

### Adding Tauri to an existing app

One command makes an app that already exists, such as a Vite React app, a Tauri app:

```sh
smith tauri add studio
```

It writes the same `src-tauri` as above (never over a file that is already there), adds the same
dependencies and script, keeps `src-tauri` out of a root Cargo workspace, installs, regenerates
the pipelines and prints what it wrote and the commands to run next. The project must be a Node app
with an `index.html`. Run it again and it changes nothing that is there.

The project's stack gains `tauri` last in its toolchain, so its own ESLint, test and build
tooling stay: `react` becomes `react-tauri` (the same make-up), and any other stack becomes
`<stack>-tauri`, declared in `smith.toml`, with `[projects.studio] stack` naming it. See
[`smith tauri add`](../../../manual/tauri-add.md). An app made with `npx tauri init` already has
`src-tauri`: `smith tauri add` then only gives it the stack, the dependencies and the script (set
`frontendDist` to the path `smith doctor` prints).

### Icons

The scaffold's icons are plain placeholders so the first build runs. Make the real ones from a
square PNG (1024x1024 is best) in the project:

```sh
smith tauri icon logo.png -p studio
```

It runs the Tauri CLI's `icon` command in the app's folder (`npx tauri icon logo.png` does the
same) and overwrites `src-tauri/icons/`.

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

### Sidecars and resources

An app that bundles another program (a Tauri *sidecar*, `bundle.externalBin`) or extra files
(`bundle.resources`) declares them in `smith.toml`, and `tauri-build` prepares them. The table is
the generic plugin settings table of the project, `[projects.<p>.settings.tauri]`:

```toml
[projects.studio.settings.tauri]
sidecars = [
  { name = "videocraw", project = "cli", bin = "videocraw" },   # a Rust project of the workspace
  { name = "ffmpeg-lite", from = "tools/ffmpeg-lite" },          # a file something else made
]
resources = [
  "assets/models",                                               # a folder: its files, under models/
  { from = "$VIDEOPLUGINS/dist", to = "plugins/" },              # from outside the repository
]
```

| Key | Meaning |
|---|---|
| `sidecars[].name` | The program's name as the app starts it (`shell.sidecar("videocraw")`): no folder, no `.exe`. |
| `sidecars[].project` | A Rust project of the workspace (its `smith projects` name). smith builds it with `cargo build --release -p <crate>` for this machine's platform and copies the program it makes. |
| `sidecars[].bin` | Which program to take when the crate builds several. Default: the only one, else the one named like the sidecar, the crate or the project. |
| `sidecars[].from` | Instead of `project`: a file that something else built. Relative to the workspace; `$NAME` is an environment variable; on Windows `.exe` is found without being written. |
| `resources[]` | A path, or `{ from, to }`. `from` is a file, a folder (its contents are copied) or a pattern with `*` and `?`; relative to the workspace, `$NAME` or `${NAME}` is an environment variable. `to` is where it goes inside the app's resources (`plugins/` with a trailing slash puts a file in that folder); default the file or folder name. |

A key it does not know, a sidecar with both or neither of `project` and `from`, and a resource that
is missing, matches nothing, is an empty folder or names an unset environment variable all fail
with a message that names the entry, so nothing empty is ever bundled.

With the table, the plugin adds a `tauri-prepare` task that `tauri-build` (and `tauri-dev`) run
first:

1. builds each sidecar `project` in release mode (cargo is incremental, a built program costs a
   moment);
2. copies the program to `dist/tauri-prepare/<project dir>/binaries/<name>-<target triple>[.exe]`.
   The triple is the host's from `rustc -vV` (`x86_64-pc-windows-msvc`, `aarch64-apple-darwin`, ...) and
   `.exe` is added for Windows, so the same setting works on every runner of a release;
3. copies the resources under `dist/tauri-prepare/<project dir>/resources/`;
4. writes `dist/tauri-prepare/<project dir>/tauri.config.json` with the `bundle.externalBin` and
   `bundle.resources` entries, and `tauri-build` passes it as `tauri build --config <file>`.

**tauri.conf.json is not edited.** Tauri merges the `--config` file into it, replacing arrays, so
smith writes the developer's own `externalBin` entries and `resources` first and adds only the ones
it lacks: a sidecar you already list in `tauri.conf.json` stays where you put it (smith copies the
program to the path you gave, with the triple), and nothing of yours is overwritten. `resources`
written as an array in `tauri.conf.json` cannot take smith's entries (they need the map form, which
puts a folder at a name): use the map form `{ "folder/": "target/" }` or declare all of them in
`smith.toml`; smith says so and stops.

`tauri-prepare` and `tauri-build` are cached when everything they read is in the cache key: sidecars
that name a project (any change to a Rust file or `Cargo.toml` in the workspace counts), and
resources that are tracked files of the workspace. A sidecar given as `from`, a resource under
`dist/`, outside the workspace or named by an environment variable is not in the key, so then both
tasks run every time. They are never wrong, only slower.

`tauri-verify` runs after `tauri-build` and the release's `build` runs both (`smith run tauri-build
tauri-verify -p studio`). It fails when a declared sidecar is not beside the built app's program
(`dist/tauri/<project dir>/release/<name>[.exe]`) or inside a bundle that is unpacked on disk: the
`.app` of a macOS build, the data folder of a `.deb`, the AppDir of an AppImage. It does not open
the installers that need an installer program to read them (`.exe`, `.msi`, `.dmg`, `.rpm`, the
`.AppImage` file); Tauri itself stops the build when an `externalBin` file is missing.

In CI the environment variable of a resource from another repository comes from
`[[ci.sibling_repos]]` (see the [CI reference](../../../reference/ci.md)): the checkout is exported as a variable that the
setting reads.

Not supported yet: a sidecar that is not a Rust project (give its file as `from`), a build for a
platform other than the runner's (`tauri build --target`), and glob patterns with `**` (name the
folder instead).

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

Sidecars and resources: on Windows, with a Rust project as the sidecar and a folder inside and a
folder outside the repository (an environment variable) as resources, `smith run tauri-build
tauri-verify` built the app, the MSI holds the sidecar and both resource folders (read with
`msiexec /a`), the second run came entirely from the cache, and a change to the sidecar's source
ran the prepare step again. **Not verified:** the same on macOS and Linux (`.app`, `.deb` and
AppImage layouts the verify step reads), the sidecar inside the NSIS installer, and a sidecar's
`bin` when the crate builds several programs outside the unit tests.

## See also

- [vitest](vitest.md), [typescript](typescript.md), [playwright](playwright.md)
- [Plugins (reference)](../../../reference/plugins.md)
