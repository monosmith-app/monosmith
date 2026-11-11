# ADR 0092: A project ships from one folder, and a tool keeps its work files in its own folder

- Status: accepted; plan items 1, 3 and 4 and the `--release` part of item 2 implemented; the CMake and Swift work-folder move open; owner decisions recorded below
- Date: 2026-11-10
- Amends: [0081](0081-a-web-app-builds-once-per-environment-and-overlap-fixes-are-presented-choices.md),
  [0088](0088-doctor-checks-and-moves-the-build-folder-of-every-language-it-can-read.md)

## Context

Every build goes to the workspace's `dist/` (ADR 0087, 0088). But where in `dist/` depends on the
tool. A Go app, a Python wheel or a Vite app writes what it ships to `dist/<project dir>`
(`protocol.ProjectBuildDir`). A Rust app writes its binary to `dist/cargo/debug/`, a .NET app to
`dist/dotnet/bin/<project>/debug/`, a Gradle module to `dist/gradle/<id>/`, a Tauri app to
`dist/tauri/<project dir>/release/bundle/`. A CMake or Swift project uses `dist/<project dir>` as
its build tree, so the shipped file sits among object files. An Electron app packs into
`dist/<project dir>/out/`, inside the folder of its own build.

So a person, a Dockerfile or a release step cannot find "the app" with one rule. They must know
the tool. The inventory below lists every build tool smith knows, where it writes today and what
must change.

smith also has no release profile. `smith build` runs each tool's default: cargo and dotnet build
debug, `dotnet pack` and Tauri build release, and Angular builds production. Release assets build
release on their own (`cargo build --release`, `dotnet publish -c Release`).

## Options considered

- **One folder per project for what it ships, one folder per tool for work files**: chosen.
- **`dist/<tool>/<type>/<name>/<env-or-profile>`**: rejected. The path changes when the tool
  changes, so every reader (Dockerfile, pipeline, release, the person) must follow the tool. It
  repeats what smith knows already (the type and the name are in `smith.toml`). Deep paths reach the
  Windows `MAX_PATH` limit sooner, and most levels are empty (a Go app has no tool folder, no
  profile and no environment).
- **Point each tool at `dist/<project dir>` directly** (cargo `--artifact-dir`, `dotnet build -o`):
  not chosen. Cargo shares one target folder per Cargo workspace; a folder per crate rebuilds every
  dependency per crate. `dotnet build -o` is not supported for a solution and mixes the outputs of
  all referenced projects. The work files then also land in the shipped folder.
- **Symbolic links from `dist/<project dir>` into the tool's folder**: not chosen. On Windows a
  symbolic link needs administrator rights or Developer Mode.

## Decision

1. **The shipped folder.** `dist/<project dir>` (`protocol.ProjectBuildDir`, for example
   `dist/apps/web`) holds what a project ships or runs: a binary, a package, a bundle, a packaged
   app. There is one such path per project, whatever the tool. A level below it,
   `dist/<project dir>/<env>`, exists only for a project with build environments (ADR 0081: lcl,
   dev, uat, prd).
2. **The work folder.** `dist/<tool>/...` holds the work files of a tool: intermediates and
   incremental state. It is shared when the tool shares its state: `dist/cargo` (one Cargo target
   folder per Cargo workspace), `dist/dotnet` (the MSBuild artifacts folder), `dist/gradle/<id>`,
   `dist/maven/<artifactId>`, `dist/tauri/<project dir>`, and for an adopted Electron setup
   `dist/electron/<name>` (the Vite plugin of Electron Forge writes `.vite/`; electron-builder
   writes `*-unpacked/` and `builder-debug.yml`).
3. **The link step.** When a tool writes what a project ships into its work folder (Rust, .NET,
   JVM, CMake, Swift, Tauri, Electron packaging), the build task links the shipped files into the
   shipped folder after the tool finishes.
   - It makes a hard link (`os.Link`). A hard link works on NTFS without administrator rights. Both
     paths must be on the same volume, and `dist/` is.
   - It makes the links again after every build. Cargo and other tools replace a file with a new
     one, and a hard link to the old file then keeps the old content.
   - It copies the file when the link fails (another volume, a file system without hard links).
   - It never makes a symbolic link.
4. **The release profile.** `smith build --release` builds the optimised profile: `cargo build
   --release`, `dotnet build -c Release`, `swift build -c release`, CMake `Release`. A tool without
   profiles (Go, Python, Lua, Bicep, AppleScript) ignores the flag.
   - The tool keeps both profiles in its work folder (cargo `debug/` and `release/`, .NET
     `debug/` and `release/`).
   - The shipped folder holds the last build. The profile is part of the cache key of the task, so a
     switch back to the other profile restores it from the cache.
   - `smith release` and the generated pipelines always build with `--release`.
   - **Web environments.** `--release` does not replace an environment, and an environment does not
     replace `--release`. Today the environment alone picks the optimisation of a Vite build:
     `vite.build.ts` keeps sourcemaps and does not minify for `lcl`, does not minify for `dev`, and
     minifies for every other environment (`libs/plugin-node-toolchains/vite/webenv/
     environment_files_mapper.go`, the `build(mode)` template). The task `build` builds the default
     environment (`prd`) and `build:<env>` builds each other one, each with `vite build --mode <env>
     --outDir dist/<project dir>/<env>`. So for a web app with environments, `--release` changes
     nothing: the person picks the environment, and the environment holds its optimisation. A web
     app without environments, Angular and Astro build their production build in both profiles,
     as today.
5. **The doctor.** Each build-folder check (`libs/plugin-*/pipeline/check_build_dir_use_case.go`)
   checks both folders of its tool: the work folder is in `dist/<tool>/...`, and the build task
   declares the shipped folder `dist/<project dir>` as its output.

## The cache and the two folders

The cache code (`apps/smith/taskcache/store_result_use_case.go`) sets three limits.

- `Save` walks each output folder and stores every regular file by content. A hard link is a
  regular file, so it is stored like any file. Two links to one file are stored once (the store is
  addressed by content).
- `Restore` first deletes each output folder (`os.RemoveAll`), then writes each file as a new
  copy. It does not make links again. After a cache hit the shipped folder holds copies, which is
  correct.
- So a **shared** work folder must never be a task output. If `dist/cargo` were an output of one
  crate's build, a restore of that crate would delete the build state of every other crate, and a
  save would store the files of every crate.

The rule: a build task lists **only the shipped folder** as its output. The work folder stays the
tool's own state, and the tool's incremental build keeps it current. Today the builds of Cargo,
.NET, Gradle, Maven and CMake are not cached (`Cache` is not set), so nothing breaks now. A later
choice to cache them must keep this rule. A task that reads the work folder after the build
(`dotnet test --no-build`, `dotnet run --no-build`, `ctest --test-dir`) needs the work folder. A
cache hit on its build skips the tool, so such a build stays uncached.

The profile is in the cache key without a new key field when the release profile has its own
command (`cargo build --release`): the key holds the command (`KeyInputs.Command`). A tool that
ignores the profile keeps one command, so both profiles share one cache entry.

## Inventory

How each build tool writes today, read from the code on 2026-11-10. "Unverified" marks a fact
about the tool itself that this record did not test on a machine.

| Tool | Output today (file:line) | Work files today | What ships | Debug or release | Change to meet the rules |
|---|---|---|---|---|---|
| Vite (plain) | `dist/<dir>`: `plugin-node-toolchains/vite/choices/base.toml:64`, task output `plugin-node/discovery/script_tasks_config.go:41` | `node_modules/.vite` (unverified) | the static bundle | none; `vite build` is always the production build (unverified) | none |
| Vite (environments) | `dist/<dir>/<env>`: `vite/webenv/environment_files_mapper.go:14-19,41`; output from `--outDir`: `plugin-node/discovery/script_tasks_config.go:43-50` | as plain Vite | one bundle per environment | the environment picks it (`build(mode)`, same file :97-125) | none; `--release` is ignored |
| Angular | `dist/<dir>`: `angular/angular_templates_config.go:64` | `.angular/cache` in the project (unverified) | the browser bundle | `defaultConfiguration: production` (:70) | none |
| Astro | `dist/<dir>`: `astro/astro_templates_config.go:46` | `.astro/` in the project (unverified) | the static site | none | none |
| tsc / esbuild (Node app, library, package) | `dist/<dir>`: `typescript/typescript_templates_config.go:175,186` | none outside the output | JS and `.d.ts` | none | none |
| Node functions | `dist/<dir>/dist`: `functions/choices/base.toml:54` | none | the compiled function | none | none |
| VS Code extension | build to `dist/<dir>`; `.vsix` to `dist/release/<app>`: `vscode/publish_extension_use_case.go:121` | none | the `.vsix` | none | none (`dist/release` is the release drop, not a build) |
| Tauri | frontend to `dist/<dir>` (`tauri/frontend_dist_check_use_case.go:42`); cargo to `dist/tauri/<dir>`: `tauri/contribute_tauri_use_case.go:77,97-103`, `tauri/extras/extras_contract.go:44-52`; sidecars to `dist/tauri-prepare/<dir>` | `dist/tauri/<dir>` (the whole cargo target is the task output today) | installers in `dist/tauri/<dir>/release/bundle/` (`contribute_tauri_use_case.go:52`) | `tauri build` builds release (unverified) | link the bundles into the shipped folder; the task output becomes the shipped folder, not the target; see owner decision 1 |
| Electron Forge (smith-made) | `dist/<dir>` holds compiled code, `package.json`, `forge.config.cjs`: `electron/choices/base.toml:53,68-76`; `package` writes `dist/<dir>/out/`: `base.toml:93-102` | Forge's `out/` inside the shipped folder | the packed app folder | none | move Forge's `out/` to `dist/electron/<dir>/out` and link the app into the shipped folder; see owner decision 2 |
| Electron Forge with Vite plugin, electron-builder (adopted) | no code in smith reads them (no match for `forge.config`, `electron-builder`, `.vite/`) | `.vite/`, `out/`, `*-unpacked/`, `builder-debug.yml` in the project | the packed app or installer | electron-builder `--dir` and the full build (unverified) | new: point them at `dist/electron/<name>` and link; PR 4 |
| Cargo | no output; `cargo build -p <crate>`, not cached: `plugin-rust/discovery/crate_tasks_config.go:44`; target `dist/cargo` from `.cargo/config.toml`: `plugin-rust/scaffolding/generate_crate_use_case.go:107,119` | `dist/cargo` (shared) | the binary of an app crate (`dist/cargo/debug/<crate>`); a library ships nothing | `debug/` and `release/`; release assets use `--release --target`: `plugin-rust/releasing/build_assets_use_case.go:180` | link step for app crates; output = shipped folder; `--release` command |
| .NET build | no output; `dotnet build`, not cached: `plugin-dotnet/discovery/project_tasks_config.go:22`; `ArtifactsPath` `dist/dotnet`: `plugin-dotnet/scaffolding/project_templates_config.go:20` | `dist/dotnet` (shared: `bin/`, `obj/`) | the app's `bin/<project>/debug/` folder | `Debug` by default; `-c Release` | link step for apps; output = shipped folder; `-c Release` command |
| .NET pack | `dist/<dir>`, always `-c Release`: `project_tasks_config.go:32-33` | `dist/dotnet` | the `.nupkg` | release only | none |
| .NET functions | `dist/dotnet/publish/<id>/release`: `plugin-dotnet/functions/function_tasks_config.go:22-27`, `function_template_fields_mapper.go:47` | `dist/dotnet` | the published folder | release only | link (or publish) into the shipped folder; output = shipped folder |
| Gradle (JVM, Kotlin, Android) | output `dist/gradle/<id>` when the script names it: `plugin-jvm/discovery/gradle_tasks_config.go:32-35`; set by `scaffolding/gradle_templates_config.go:69`, `android_templates_config.go:35` | `dist/gradle/<id>` | the jar in `libs/`, or the APK in `outputs/apk/<variant>/` | Android `assemble` builds every variant (unverified); plain JVM none | link the jar or APK; output = shipped folder; Android: pick the variant by profile |
| Maven | output `dist/maven/<artifactId>`: `plugin-jvm/discovery/module_tasks_config.go:15,29`; set by `scaffolding/module_templates_config.go:52` | `dist/maven/<artifactId>`, and the local repository (`mvn install`) | the jar | none | link the jar; output = shipped folder |
| Go | `dist/<dir>/<binary>`: `plugin-go/discovery/module_tasks_config.go:41-42`; `package` to `dist/drop/<name>` (:44-46); a library builds no file (:49) | Go's own cache outside the workspace | the binary | none | none |
| Python build | `dist/<dir>`: `plugin-python/discovery/project_tasks_config.go:49`; `package` to `dist/drop` (:52-53) | `build/` and `*.egg-info` in the project with setuptools (unverified) | wheel and sdist | none | none for the shipped folder; the work files in the project stay (owner decision 4) |
| Python functions | `dist/<dir>/app` and a zip in `dist/<dir>`: `plugin-python/functions/function_tasks_config.go:15-20` | none | the staged folder and the zip | none | none |
| Dart CLI | `dist/<dir>`: `plugin-dart/discovery/package_tasks_config.go:53`, `declared_projects_use_case.go:119-120` | `.dart_tool/` in the project | the executable bundle | none (AOT) | none |
| Flutter web | `dist/<dir>`: `package_tasks_config.go:47-49` | `.dart_tool/`, `build/` in the project (unverified) | the web bundle | release by default (unverified) | none |
| CMake (C++) | the build tree is `dist/<dir>`, `-DCMAKE_BUILD_TYPE=Debug` fixed, build not cached: `plugin-cpp/discovery/project_tasks_config.go:35-38` | the same `dist/<dir>` | the executable or the library | `Debug` fixed | move the tree to `dist/cmake/<dir>`; link the executable; `Release` with `--release` |
| Swift | `--scratch-path dist/<dir>`, cached: `plugin-swift/discovery/project_tasks_config.go:15-17`, `declared_projects_use_case.go:87,97` | the same `dist/<dir>` | the product in `debug/` (unverified) | `swift build -c release` | move the scratch path to `dist/swift/<dir>`; link the product; output = shipped folder |
| Lua (LuaRocks) | `luarocks make --tree dist/<dir>`: `plugin-lua/discovery/project_tasks_config.go:13,16` | none outside the tree | the installed rock tree | none | none |
| Bicep | `dist/<dir>/template` and `dist/<dir>/parameters`: `plugin-bicep/discovery/project_tasks_config.go:18-24` | none | the ARM template and parameters | none | none |
| AppleScript | `dist/<dir>/<name>.scpt`: `plugin-applescript/discovery/project_tasks_config.go:19-24` | none | the `.scpt` | none | none |
| Objective-C (CocoaPods) | no file; `pod lib lint`: `plugin-objc/discovery/project_tasks_config.go:21` | Xcode's DerivedData (unverified) | nothing | none | none |
| PHP, Ruby, VBA, recipe | no build task (`plugin-php/discovery/package_tasks_config.go`, `plugin-ruby/discovery/gem_tasks_config.go`) | none | nothing | none | none |
| Docker | reads `dist/<dir>`: `plugin-docker/scaffolding/dockerfile_templates_config.go:35-47` | none | the image | none | none; after the link step a Rust or .NET image can copy `dist/<dir>` too |

What the inventory shows:

- 19 of 29 rows already meet the rules. Every tool that writes one product with an output flag
  (Go, Python, Dart, Lua, Bicep, AppleScript, every Node bundler) writes straight to the shipped
  folder.
- The link step is needed for Cargo, .NET, Gradle, Maven, CMake, Swift, Tauri and Electron.
- CMake and Swift today mix the work folder and the shipped folder. They must move their work
  folder to `dist/<tool>/<dir>`.
- Electron writes Forge's `out/` inside its shipped folder, so a restore of the build deletes it.

## Implementation plan

Each pull request is one capability. None changes the layout of a tool that already meets the
rules.

1. **The link step, and Rust, .NET and JVM.**
   - A declared list of files to ship on the task, so the host links them, one implementation for
     every plugin: `packages/plugin-sdk/protocol/project_contract.go` (task fields),
     `packages/plugin-sdk/protocol/build_dir_config.go` (`BuildOutputDir`), the protocol schema in
     `packages/plugin-sdk/protocolschema`.
   - The link, the relink after every build and the copy fallback: a new slice in `apps/smith`
     (for example `apps/smith/shiplink`), called from `apps/smith/taskrun`.
   - Rust: `libs/plugin-rust/discovery/crate_tasks_config.go` (app crates link
     `dist/cargo/<profile>/<crate>`).
   - .NET: `libs/plugin-dotnet/discovery/project_tasks_config.go`,
     `libs/plugin-dotnet/functions/function_tasks_config.go`.
   - JVM: `libs/plugin-jvm/discovery/gradle_tasks_config.go`,
     `libs/plugin-jvm/discovery/module_tasks_config.go`.
   - Golden paths: `apps/smith/goldenpath/{rust,dotnet,jvm,kotlin,android}_workspace_use_case_test.go`
     also check the shipped folder.
2. **`--release`, in the CLI, the TUI and the pipelines.**
   - The flag, declared once (ADR 0084): `libs/actions/flagvocabulary/flag_vocabulary.json`, then
     the generated clients.
   - The run: `apps/smith/taskrun/run_handler.go`, `run_in_workspace_use_case.go`,
     `task_key_use_case.go`; the TUI: `apps/smith/tui/commands/parse_run_request_algorithm.go`.
   - The release command of each task: the plugins of Rust, .NET, C++, Swift, Gradle (Android).
   - `smith release` and the pipelines: `apps/smith/release`, `apps/smith/cipipeline`
     (`run_phase_use_case.go` and the rendered phases).
   - CMake and Swift move their work folder: `libs/plugin-cpp/discovery/project_tasks_config.go`,
     `libs/plugin-swift/discovery/project_tasks_config.go`, `declared_projects_use_case.go`.
3. **Doctor checks.** `libs/plugin-{rust,dotnet,jvm,cpp,go,python,dart,node}/pipeline/
   check_build_dir_use_case.go` check both folders; `packages/plugin-sdk/taskoutput/
   check_task_output_use_case.go` and `apps/smith/doctor/build_output_policy.go` accept a work
   folder under `dist/<tool>` and refuse it as a task output when the tool shares it.
   Implemented: Rust, .NET and Maven require the exact work folder the build ships from
   (`dist/cargo`, `dist/dotnet`, `dist/maven/<artifactId>`); Gradle already required
   `dist/gradle/...`. `taskoutput.CheckShippedOutput` refuses a shared work folder as an output and
   reports a shipping task that does not cache `dist/<project dir>`; `taskoutput.Check` (Go, Python,
   Dart) reports outputs that do not hold what the command writes; doctor's `SharedWorkOutputs`
   refuses an output that holds a folder another project ships from. Node and CMake are unchanged:
   Node writes straight to the shipped folder, and CMake's tree is still `dist/<project dir>` until
   its work folder moves. Not done: the JVM work folders keep the `<artifactId>` key, not the
   project folder of owner decision 5, because the scaffolding and the discovery name them so.
4. **Electron and Tauri.**
   - Smith-made Electron: `libs/plugin-node-toolchains/electron/choices/base.toml` (Forge's
     `out/` to `dist/electron/<name>`, then the link).
   - Tauri: `libs/plugin-node-toolchains/tauri/contribute_tauri_use_case.go` (output and link of the
     bundles).
   - Adopted Electron (Forge with its Vite plugin, electron-builder): detection and the move in
     `apps/smith/adoptmove` and the Node doctor check. This is new code: smith reads neither setup
     today.

   Implemented:
   - Smith-made Electron: `build` writes `dist/electron/<dir>/app` (the folder `start` runs) and
     caches it; the generated `forge.config.cjs` sets `outDir: '../out'`, so Forge packs into
     `dist/electron/<dir>/out`; `package` ships `out/*-*-*` into `dist/<dir>` and caches that folder.
     The node plugin's discovery sets the outputs and the link (`libs/plugin-node/electronapp`).
   - Tauri: `tauri-build` caches `dist/<dir>` and ships `release/bundle/*/*.{msi,exe,dmg,deb,rpm,
     AppImage,app}` there; the release assets come from `dist/<dir>`. The web build goes to
     `dist/tauri/<dir>/frontend`: the scaffold writes that `frontendDist` and moves the `outDir` of
     `vite.config`, and `build` caches that folder when `frontendDist` names it. With sidecars or
     resources `tauri-build` is not cached, because `tauri-verify` reads the cargo target.
   - Adopted Electron: the detection, the link and the move are in `libs/plugin-node/electronapp`,
     not in `apps/smith/adoptmove`. `adoptmove` moves a root project into a folder of its own; the
     node plugin already reads `package.json` and owns the doctor fixes that move a build folder
     (ADR 0088), and its discovery is where task outputs are set. The packing scripts
     (`electron-forge package`/`make`, `electron-builder`) ship into `dist/<dir>` wherever the tool
     packs. `CheckElectronDir` warns when the tool packs outside `dist/electron/<dir>` and offers the
     move as a choice when it can edit the configuration safely (JSON, a YAML file without
     `directories`, a Forge file with one `outDir` string or one plain exported object).
   - Not done: Forge's Vite plugin writes `.vite/` in the project, and its documentation names no
     setting to move it, so it stays. A Tauri app made before this change keeps `frontendDist`
     `dist/<dir>` until the person moves it (doctor warns). Until then its `build` is not cached
     (`contribute_tauri_use_case.go`), so it and `tauri-build` never cache the same folder and a
     restore of one cannot delete the other's files.
   - Tool facts checked on 2026-11-11 against the tools' documentation and source: Forge `outDir`
     (default `./out`, resolved from the project folder: `packages/api/core/src/util/out-dir.ts`),
     electron-builder `directories.output` (default `dist`), Tauri `build.frontendDist` (relative to
     the configuration file) and `tauri build` builds release (CLI reference). Unverified: Forge
     `make` writes `<outDir>/make/<maker>/...`; the native bundle folder
     `target/release/bundle/<kind>/` (the documentation states it only for a cross build,
     `target/<triple>/release/bundle/nsis/`; smith already relied on it); the packed folder name
     `<name>-<platform>-<arch>`. No real Electron or Tauri build ran for this change.

## Owner decisions

The owner settled the open questions on 2026-11-10:

1. **Tauri:** the installers are the shipped files, in the shipped folder. The web frontend is a
   work file, in `dist/tauri/<dir>/frontend`.
2. **Electron:** the packed app is the shipped file. The folder that `start` runs is a work file,
   in `dist/electron/<dir>`, which also takes Forge's `out/` out of the shipped folder.
3. **Web builds:** `--release` changes nothing for Vite, Angular or Astro. Their build is already
   the production build, and a web app with environments takes the mode from the environment.
4. **Work files beside the sources:** they move to `dist/<tool>/<dir>` where the tool lets smith
   choose the folder (setuptools, `.angular/`, Flutter). This is a later PR.
5. **The key of a work folder:** every work folder uses the project folder, `dist/<tool>/<dir>`,
   the same key as the shipped folder.
6. **A file in use:** when Windows does not let smith replace a link to a running program, the
   build warns, names the program to close, and keeps the old file. The tool's own output is
   built correctly.
7. **Caching:** the Cargo and .NET builds stay uncached for now. `dotnet test --no-build` and
   `ctest` need the work folder, which a cache hit would skip.

## Consequences

- One rule finds any project's product: `dist/<project dir>`. Dockerfiles, release steps and people
  stop following each tool's layout.
- Tools keep their incremental builds and both profiles. The link step costs one hard link per
  shipped file.
- A cache hit restores copies, not links. That uses more disk for the shipped files, never for
  the work files.
- CMake and Swift projects move their build tree. The first build after the change is a full build.

## Where to check

- `packages/plugin-sdk/protocol/build_dir_config.go` (`ProjectBuildDir`, `BuildOutputDir`)
- `apps/smith/taskcache/store_result_use_case.go` (`Save`, `Restore`),
  `apps/smith/taskcache/cache_key_algorithm.go` (`KeyInputs`)
- The files in the inventory table.
