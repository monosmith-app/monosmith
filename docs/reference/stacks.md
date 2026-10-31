# Stacks

> Terms such as project, kind, task and stack are defined in the [glossary](glossary.md). How plugins are loaded and combined: [the plugin system](../architecture/plugin-system.md).

A **stack** is the set of plugins a project is built with. Each project
records its own, so one monorepo can hold a React app built with nothing but
Vite next to an Azure Function built with a strongly opinionated stack:
scaffold, lint rules, test library, layout rules and deploy steps.

## Plugin roles

Every plugin plays one or more roles:

| Role | Does | Per project |
|---|---|---|
| **language** | finds projects from their manifests, turns imports into graph edges, writes versions, publishes | exactly one |
| **toolchain** | how the project builds, tests and lints | any number; two may not provide the same task |
| **structure** | layout and code rules, checked as tasks | any number |
| **scaffold** | the starter files of a new project | at most one |
| **deploy** | pipeline steps and deploy targets | any number |

Any plugin may add `doctor` checks and CI steps for the projects it serves.

smith itself knows no language: the graph, the cache, affected and releases work
the same for every project, because dependencies always come from the language
plugin. A project built with one stack can depend on a project built with
another.

A language plugin without explicit roles also brings a default toolchain and
scaffold, so every built-in language plugin works on its own: a project with
no stack is built by its language plugin alone. `node` on its own is plain
JavaScript; TypeScript is the `typescript` plugin's, through a stack.

## Declaring a stack

Plugins publish ready-made stacks, and `smith.toml` declares more (a stack
declared there wins over a published one of the same name):

```toml
[stacks.ts-func]
description = "Azure Function in TypeScript, the team's way"
language = "node"
toolchain = ["eslint", "vitest"]
structure = ["team-layout"]
scaffold = "azure-function"
deploy = ["azure-functions"]
```

List one provider per task: `typescript` and `vite` both provide `build` and `typecheck`, so a stack
with both is refused ("typescript and vite both provide the build task; keep one"). A Vite app names
`vite` alone; `tauri`, `vitest`, `eslint` and `playwright` combine with either.

Toolchain plugins apply in order, so a later one may replace a task an earlier
one added. smith checks every stack against the loaded plugins: each is loaded,
plays the role it is listed under and applies to the stack's language, and no
two provide the same task (`jest` and `vitest` both provide `test`).

`smith stacks` lists every stack, where it comes from, its plugins, what is
wrong with it if it cannot be used, and the projects using it.

## Choices: a stack as a saved selection

A stack that fixes the framework, the test runner and the style for every combination would need a
stack per combination. Instead, plugins publish **choices**: one option on an **axis**
(`framework`, `test`, `style`, `lint`, `library`). `smith choices` lists them:

```sh
smith choices
smith add app web --with react,vitest,bootstrap
```

smith checks the selection before it writes a file, and gives one message that names the choices
and the fix when it cannot be used:

- One choice per axis, unless the choices on the axis say they can share it. `react,vue` is refused
  (`react and vue are both on the framework axis: keep one`).
- A choice may require another (`chakra requires react, which is not in the selection: add react
  to --with, or drop chakra`) or conflict with one.
- A choice may need an axis (`react` needs `lint` and `test`). A selection that leaves the axis out
  takes its default (`eslint`, `vitest`), and smith says so.
- Exactly one choice writes the project, and all choices are for one language.

A stack is a saved selection. Give it `choices`; the language, the scaffold and the toolchain come
from them, and `toolchain`, `structure` and `deploy` add plugins after those:

```toml
[stacks.shop-web]
description = "React with Vitest and Bootstrap"
choices = ["react", "vitest", "bootstrap"]
```

`smith add --with` saves the selection in `smith.toml` this way, under a name made of the choices
(`react-vitest-bootstrap`), and the project records that stack. A stack with the same choices is
reused. The ready-made `react`, `vue`, `svelte` and `html` stacks are saved selections too.
`--with` and `--stack` exclude each other. How a plugin ships choices:
[Writing choices](../plugins/creating/choices.md). The decision:
[ADR 0045](../architecture/decisions/0045-frameworks-tests-and-styles-are-choices-shipped-as-data.md).

## Choosing a stack

```sh
smith add app api --stack ts-func
smith add app web --lang node          # the default stack for node, if any
smith add lib util --lang node --stack none   # the language plugin alone
```

The project records its stack in its manifest:

| Language | Where |
|---|---|
| Node | `"smith": { "kind": "app", "stack": "ts-func" }` in `package.json` |
| Python | `stack = "ts-func"` under `[tool.smith]` in `pyproject.toml` |
| Go | a `// smith:stack ts-func` line at the top of `go.mod` |
| Rust | `stack` under `[package.metadata.smith]` in `Cargo.toml` |
| .NET | a `<SmithStack>` property in the `.csproj` or `.fsproj` |
| Ruby | `spec.metadata['smith_stack']` in the `.gemspec` |
| PHP | `extra.smith.stack` in `composer.json` |
| Dart | `stack` under `smith:` in `pubspec.yaml` |
| JVM | a `smith.stack` property in `pom.xml`, or in `gradle.properties` |
| C, C++ | `set(SMITH_STACK ...)` in `CMakeLists.txt` |
| Bicep | `metadata smithStack = '...'` in `main.bicep` |
| Lua | `-- smithStack: ...` comment in the `.rockspec` |
| AppleScript | `-- smithStack: ...` comment in the main `.applescript` |
| VBA | `' smithStack: ...` comment in the main module (`.bas`, `.cls`, `.frm`) |
| Swift | `// smithStack: ...` comment in `Package.swift` |
| Objective-C | `# smithStack: ...` comment in the `.podspec` |

The project's kind sits beside it, under the same name with `kind`
(`SmithKind`, `smith_kind`, `smith.kind`, `SMITH_KIND`, `smithKind`).

A project's stack is, in order: `[projects.<name>] stack` in `smith.toml`, the
one its manifest records, or the default for its language. `none` means
the language plugin alone, even where there is a default:

```toml
[default_stacks]
node = "node-jest"
```

## How a project's tasks are built

1. The language plugin discovers the project and its tasks.
2. The stack's toolchain, structure and deploy plugins contribute, in order.
   Each adds tasks the project lacks, may **replace** one it has, and may make
   an existing task depend on one of its own.
3. Conventions apply to every project of their languages
   ([conventions](conventions.md)).
4. `smith.toml` task overrides apply last, so you always have the final word.

`smith explain <project>:<task>` shows where every field came from, for example
`from plugin vitest (stack node-vitest)`.

## Creating a project

`smith add` asks the stack's scaffold (or, without one, the language plugin) to
generate the project and record the stack, then asks each of the stack's other
plugins to add its part: files, scripts, settings, dependencies. These only
add: a file, script, setting or dependency that already exists is kept.

When siblings of the new project's kind and language share a build tool, a test runner
or where their entry points lead, the project follows them, not the stack's own default
(`smith add --fresh` gives the default). A plugin lists the shape keys it applies in
`initialize.follows`; see [`smith add`](../manual/add.md#following-the-siblings) and
[ADR 0041](../architecture/decisions/0041-smith-add-follows-the-dialect-of-sibling-projects.md).

## Built-in stacks

| Stack | Language | Scaffold | Toolchain | Published by |
|---|---|---|---|---|
| `node-ts` | node | `typescript` | `typescript` | `typescript` |
| `node-jest` | node | `typescript` | `typescript`, `eslint` (`@monosmith/eslint-config`), `jest` | `jest` |
| `node-vitest` | node | `typescript` | `typescript`, `eslint` (`@monosmith/eslint-config`), `vitest` | `vitest` |
| `node-vitest-playwright` | node | `typescript` | `typescript`, `eslint`, `vitest`, `playwright` | `playwright`: adds `playwright.config.ts`, a smoke test in `e2e/` (named `*.e2e.ts`, so Vitest never runs it), an `e2e` script and `@playwright/test`; reports and traces go to `dist/e2e/<project dir>`, the `e2e` task's outputs, beside the build rather than in it (a cached build replaces its own directory) |
| `node-vitest-tauri` | node | `typescript` | `typescript`, `eslint`, `vitest`, `tauri` | `tauri`: adds `@tauri-apps/cli` and a `tauri` script; writes `src-tauri` when the project has an `index.html`, else `npx tauri init` does |
| `react`, `react-p5`, `vue`, `svelte`, `html` | node | `vite` | `vite`, `eslint`, `vitest` | `vite`: a browser app built with Vite |
| `react-tauri` | node | `vite` | `vite`, `eslint`, `vitest`, `tauri` | `vite`: the `react` app; `tauri`: `src-tauri`, `@tauri-apps/api` and `@tauri-apps/cli` |
| `angular` | node | `angular` | `angular`, `eslint`, `vitest` | `angular` |
| `docs-site` | node | `astro` | `astro`, `eslint` | `astro`: an Astro Starlight documentation site |
| `vscode-extension` | node | `vscode` | `vscode`, `eslint`, `vitest` | `vscode`: bundled by esbuild, packed and published with vsce |
| `flutter` | dart | — | — | `dart`: a Flutter app for the web |
| `kotlin` | jvm | — | — | `jvm`: a Kotlin module built with Gradle, ktlint |
| `android` | jvm | — | — | `jvm`: an Android app in Kotlin (AGP), ktlint and Android Lint |
| `go-smith-plugin` | go | `go-smith-plugin` | — | `go-smith-plugin`: a smith plugin written in Go (see plugins) |

`smith stacks` prints the same, with the projects using each.

`smith new` makes `node-jest` the default for Node. The `node` plugin alone
scaffolds plain JavaScript (CommonJS), with nothing to compile. The
`typescript` plugin scaffolds a TypeScript project instead: `tsconfig`, and
`build` and `typecheck` scripts that run `tsc`. The other toolchain plugins add
the rest when a project is created:

- `eslint` writes `eslint.config.mjs` at the root (once), adds `eslint` and
  `@monosmith/eslint-config` to the root devDependencies, and a `lint` script.
- `jest` adds `jest`, `ts-jest` and `@jest/globals`, a `test` script, the
  Jest settings in `package.json`, and the starter's spec.
- `vitest` adds `vitest`, a `test` script (`vitest run`) and the starter's
  spec.

The starter's spec is written for the TypeScript starter; a JavaScript
project gets the test script and its runner, and you write the first spec.

A workspace created by smith 0.1 had ESLint and Jest built into the node
plugin. Add `[default_stacks]` with `node = "node-jest"` to its `smith.toml` to
keep new projects the same; existing projects are unchanged.

## Adding plugins

Any executable speaking the [plugin protocol](plugins.md) is a plugin,
written in any language. Declare it, then use it in a stack:

```toml
[plugins.external.azure-function]
command = ["smith-plugin-azure-function"]
```

A new language is a language plugin like any other.
`apps/smith/goldenpath/testdata/kiwiplugin` is a complete one, for a made-up
language, in about 200 lines: discovery and imports, generation, a published
stack whose toolchain replaces a task, CI setup, doctor, versions and
publishing. The end-to-end tests run it through all of smith. Installing
plugins by name from a registry is not decided yet (#164).
