# Plugins and flexibility: extend your monorepo tool in any language

**Teach MonoSmith a new language, a team rule or a deploy target with a small program in the language you already use.** No new build language to learn, no fork, no rebuild. The 24 plugins that ship inside the `smith` binary go through the same door your plugin will, so what you write is not second class.

This page shows how that works, proves it with commands you can run, and sets it against the fourteen other tools in the [comparison](README.md). It is the extension and configuration half of the case for smith; the [language environments](languages-and-environments.md) and the [dependency graph](dependency-graph.md) are the other two.

## Why it matters to a monorepo owner

Every monorepo ends up with something the tool did not ship: an in-house language, a layout rule, a deploy target, a linter nobody else uses. What the tool asks of you then decides how long it takes and who can do it.

- **Write the extension in a language your team already knows.** Not the tool's own.
- **Keep it out of the tool's process.** A bug in your extension should not take the tool down.
- **Share a standard without code.** Most teams want "all our services look like this", which is data, not a program.
- **Know what you are running.** An extension named by a repository you just cloned is code that runs as you.

## Proof you can run

A language plugin written in Python, with no SDK and nothing but the standard library, is 102 lines (`docs/examples/plugins/python/plugin.py`). `smith plugin check` runs the conformance suite against it:

```text
$ smith plugin check docs/examples/plugins/python/plugin.cmd
widget 0.1.0
  ✔ names itself
  ✔ has a known kind
  ✔ plays a known role: language
  ✔ a language plugin discovers projects
  ✔ declares a concurrency the protocol defines
  ✔ answers 3 requests sent together, one at a time
  ✔ discover answers with no manifests
  ✔ shuts down when asked
```

A workspace that names an external plugin does not start it until you have trusted it. Run on a plugin named in `smith.toml` and not yet trusted:

```text
$ smith projects
✖ the external plugin "widget" is not trusted yet, so smith did not start it
  → run `smith plugin trust widget` to review it and trust it
  → run `smith plugin list` to see every external plugin and what it may see

$ smith plugin trust widget --yes --pin
Trusted widget in <workspace>.
Pinned sha256 a0b4ee27…6743402 in smith.toml [plugins.external.widget], which is what CI checks.

$ smith plugin list        # the part about external plugins
External plugins this workspace uses (smith starts them only when trusted):
  PLUGIN         FROM               TRUSTED  PINNED  MAY SEE BEYOND THE BASELINE
  widget         smith.toml         yes      yes     -
```

Both runs were made with smith 1.65.0 on 2026-10-25 in a scratch workspace.

## How smith is extended

A plugin is any executable that speaks JSON-RPC 2.0 on its standard input and output. smith starts it when a command begins, asks it questions and stops it when the command ends. A plugin does not run your tasks: it describes them (an argv, inputs, outputs), and smith runs, hashes and caches them. The built-in plugins are not special: they run as child processes (`smith plugin serve <name>`) and speak the same protocol.

### Five roles

| Role | It lets you... | A project has |
|---|---|---|
| `language` | find projects from manifests, turn imports into graph edges, write versions, publish | exactly one |
| `toolchain` | decide how a project builds, tests and lints (fills a task slot such as `lint`) | any number, one per slot |
| `scaffold` | write the starter files of a new project | at most one |
| `structure` | check layout and code rules, with fixes | any number |
| `deploy` | add pipeline steps, credentials and a release target | any number |

What a plugin contributes is the set of methods it advertises: projects and their tasks, tasks added or replaced in other plugins' projects, generated files, versions, publishing, CI steps and toolchain setup, health checks with fixes, adding and removing dependencies, newest releases. smith never calls a method a plugin did not advertise. The wire format and the roles are in [What a plugin is](../../plugins/using/what-is-a-plugin.md); the field-level detail is in the [protocol cheat sheet](../../plugins/creating/protocol-cheat-sheet.md).

### The range: 24 plugins built with it

`smith plugin list` shows 24: 12 language plugins (`bicep`, `cpp`, `dart`, `dotnet`, `go`, `jvm`, `node`, `php`, `python`, `recipe`, `ruby`, `rust`) and 12 add-ons (`angular`, `astro`, `docker`, `eslint`, `go-smith-plugin`, `jest`, `playwright`, `typescript`, `vertical-slices`, `vite`, `vitest`, `vscode`). One uses the language role to find Cargo crates, another uses the toolchain role to run Playwright, another uses the deploy role to build and push container images. Each has a page in the [catalogue](../../plugins/using/catalogue/README.md), written from the code and from runs.

### Several plugins, one project, checked for conflicts

A project is built from one language plugin and any toolchain, structure, scaffold and deploy plugins on top, called a **stack**. Plugins declare what they `require`, `provide` and `conflict` with, and smith checks the whole combination in one place:

```text
jest and vitest both provide the test task; keep one
jest applies to [node], not to widget projects
```

`smith explain web:test` then says which plugin set each field, so a surprising task is traceable to its source.

### Flexibility without code: stacks and recipes

Most customisation is a choice, not a program.

- **16 stacks** ship with smith (`smith stacks`): `react`, `vue`, `svelte`, `angular`, `docs-site` (Astro), `vscode-extension`, `kotlin`, `android`, `flutter`, `node-vitest-playwright`, `go-smith-plugin` and more. Each project records its own, so one repository can hold a React app next to a Go service and a VS Code extension.
- **Your own stacks** are a few lines under `[stacks.<name>]` in `smith.toml`.
- **A recipe** is a TOML file, with no code, that records your team's choice of plugins and settings and asks a few questions when someone runs `smith add`. `smith recipe check` shows what it resolves to and where each value came from. `smith doctor` reports a project that drifted from its recipe, and `smith upgrade` brings it back. See [Recipes](../../reference/recipes.md).

### Safe to run what a repository names

An external plugin is code that runs as you, so smith treats it that way ([decision 0036](../../architecture/decisions/0036-external-plugins-are-trusted-and-get-a-minimal-environment.md)):

| Risk | What smith does |
|---|---|
| A cloned repository names a plugin in `smith.toml` | it is not started until you run `smith plugin trust`; even a read-only command is refused |
| The plugin changes after you trusted it | trust is for one workspace, one name, one command and the SHA-256 of the files that command runs; a change asks again |
| The plugin reads secrets it does not need | it starts with a minimal environment plus the variables it declares and you grant |
| A pull request adds a plugin and CI runs it | in CI a plugin starts only when `smith.toml` pins its SHA-256 and the file still matches |

Trusting a plugin is not sandboxing it. A trusted plugin runs with your privileges; smith says so in [Plugin security](../../plugins/using/plugin-security.md).

## Compared with fourteen tools

The marks follow the [comparison's legend](README.md#legend): ✅ the tool's documentation says it has it, ◐ partly or with a condition, ❔ not verified in the sources searched (not "absent"), ❌ checked and not provided. These rows are also in the [master table](everything-compared.md).

**MonoSmith lets you write an extension in any language against a documented protocol, with published SDKs for Go, TypeScript, Python and C#, and the extension can add a language, a task, a scaffold, a CI step or a release target.** moon (any language that compiles to WASM) and Cargo (any program named `cargo-<name>`) are the nearest ideas, and each is credited below.

| Row | **MonoSmith** | Nx | Turbo | Bazel | Pants | Lerna | Rush | moon | Buck2 | Gradle | Earthly | Lage | JS pkg mgrs | Cargo | Please |
|---|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| Has a plugin model | **✅** | ✅ | ❔ | ✅ | ✅ | ❔ | ◐ | ◐ | ✅ | ✅ | ❔ | ❔ | ◐ | ✅ | ✅ |
| Extensions can add a new language | **✅** | ✅ | ❔ | ✅ | ✅ | ❔ | ❔ | ◐ | ✅ | ✅ | ❔ | ❔ | ❔ | ❔ | ✅ |
| Extension written in any language | **✅** | ❌ | ❔ | ❌ | ❌ | ❔ | ❌ | ◐ | ❌ | ❌ | ❔ | ❔ | ❌ | ✅ | ❌ |
| Each extension runs in its own process | **✅** | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ✅ | ❔ |
| Separate programs speaking a documented protocol | **✅** | ◐ | ❔ | ❔ | ❔ | ❔ | ❔ | ◐ | ❔ | ❔ | ❔ | ❔ | ❔ | ◐ | ❔ |
| An SDK for writing extensions | **✅** | ✅ | ❔ | ✅ | ✅ | ❔ | ◐ | ❔ | ✅ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ |
| A command that checks an extension | **✅** | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ |
| Team presets as data (recipes) | **✅** | ◐ | ◐ | ❔ | ❔ | ❔ | ❔ | ◐ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ |
| A registry or catalogue | **❌** | ✅ | ❔ | ◐ | ❔ | ❔ | ❔ | ◐ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ |
| Trust step and pins for repository-named extensions | **✅** | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ◐ | ❔ | ◐ | ❔ | ❔ | ❔ | ❔ | ❔ |
| Integrity-checked or sandboxed | **◐** | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ |

### How each tool is extended

| Tool | Mechanism | You write it in | Credit and condition |
|---|---|---|---|
| **MonoSmith** | plugins in five roles, over JSON-RPC on stdio; stacks and recipes as data | any language; SDKs for Go, TypeScript, Python and C# | no registry yet; not sandboxed |
| Nx | plugins with generators, executors, `createNodes` and `createDependencies` | TypeScript or JavaScript: "The plugin is a Node package" | the largest ecosystem in this set: a plugin registry and community plugins; each plugin runs "in a separate forked process" |
| Turborepo | no plugin API found; custom generators use Plop | JavaScript | focused on tasks and caching; `turbo gen` covers scaffolding |
| Bazel | macros, rules, repository rules and module extensions in `.bzl` files | Starlark | "Support for a new language is typically done with a rule"; the Bazel Central Registry hosts modules |
| Pants | Plugin API (Target API and Rules API) | "typed Python 3 code" | "all of Pants's built-in functionality uses the same API"; plugins come from PyPI or a `pants-plugins` folder |
| Lerna | none found; it runs on the Nx project graph | – | – |
| Rush | plugins; the documentation page is titled "(experimental)" | an npm package | share configuration across monorepos and add features |
| moon | WASM plugins: toolchains and extensions | "any language that compiles to WASM" | sandboxed file system; the authors say the implementation is "still experimental" |
| Buck2 | rules and BXL scripts | Starlark | BXL "enables integrators to inspect and interact with the Buck2 graph" |
| Gradle | binary, convention and precompiled-script plugins | Java, Kotlin or Groovy | the Gradle Plugin Portal, a registry of plugins |
| Earthly | none found | – | – |
| Lage | custom reporters only | JavaScript | – |
| npm, pnpm, Yarn | Yarn plugins; pnpm hooks in `.pnpmfile.mjs`; none found for npm | JavaScript | – |
| Cargo | custom subcommands | any language: "extensible with new subcommands" as `cargo-<name>` programs on `PATH` | commands, not languages or tasks |
| Please | plugins that add build rules | its BUILD language | "build rules for additional languages or technologies"; `plz init plugin` scaffolds one |

## What this costs you with other tools

Every tool above is good at its own job. What differs is what you still have to write, learn or glue together to extend it.

- **A new language in Bazel, Buck2 or Please** means rules in Starlark or the BUILD language. Bazel's own documentation says rules are "for advanced users" and that most users never write one, which is the point: someone on your team becomes the person who does.
- **A new language in Nx, Rush or Gradle** means a plugin in TypeScript or in a JVM language, packaged the way that ecosystem packages code. Nx and Gradle repay that with a registry and a large body of community plugins.
- **A Pants plugin** is Python against the Rules API, which is powerful for a Python shop and one more thing to learn for everyone else.
- **moon's WASM plugins** let you choose the language, as long as it compiles to WASM and you accept a plugin system its authors call experimental.
- **A team standard** is, in most of these tools, a custom generator or a template you write and maintain. In smith it is a TOML recipe, and `doctor` tells you which projects drifted from it.

## Where another tool is the better choice

- **You want a large ecosystem of ready-made third-party plugins, today.** Nx has a plugin registry and community plugins; Gradle has the Plugin Portal; Bazel has its Central Registry. smith has no registry yet ([#164](https://github.com/monosmith-app/monosmith/issues/164), [#565](https://github.com/monosmith-app/monosmith/issues/565)): plugins are executables you place on a search path or name in `smith.toml`.
- **You need extensions to be sandboxed.** moon's WASM plugins get a sandboxed file system. smith's trust gate, minimal environment and digest pins reduce the risk and do not remove it: a trusted plugin runs with your privileges ([#565](https://github.com/monosmith-app/monosmith/issues/565)).
- **You want an SDK in your language.** Go, TypeScript, Python and C# have published SDKs. For any other language you write the message framing yourself, about 40 lines, as the 102-line hand-framed Python example does ([#564](https://github.com/monosmith-app/monosmith/issues/564)).
- **You need a plugin to run tasks in parallel with its own requests, or to be cancelled mid-call.** External plugins can declare both; the built-in plugins are called one request at a time and cannot be cancelled ([decision 0037](../../architecture/decisions/0037-cancel-and-concurrency-in-the-plugin-protocol.md)).
- **The plugin API must be stable.** The protocol is versioned (today, 1) and the Go SDK is below 1.0, so a minor version may change its API.
- **You are JavaScript only and want Nx's depth.** Nx's generators, executors and graph plugins are a mature, documented extension surface for that stack.

## Try it

```sh
smith plugin list                              # the 24 built-ins, and any external ones
smith stacks                                   # the 16 ready-made stacks
smith plugin check ./path/to/your-plugin       # the conformance suite
```

Write your first plugin in ten minutes with the [quickstart](../../plugins/creating/quickstart.md), pick your language in [Writing a plugin in Go, Node.js or Python](../../plugins/creating/plugin-sdks.md), or read the [TypeScript SDK](../../plugins/creating/typescript-sdk.md). Not ready to write code? A [recipe](../../reference/recipes.md) needs none. To see what moving an existing repository involves, run `smith adopt --dry-run`; coming from Nx, see [Moving from Nx](nx.md#moving-from-nx-to-monosmith).

## Method and evidence

- **MonoSmith.** Everything above about smith was checked on 2026-10-25 against smith 1.65.0 and this repository: `smith plugin list`, `smith plugin check`, `smith plugin trust`, `smith stacks`, the protocol documents and decisions 0036 and 0037. The SDKs were checked on their registries the same day: `@monosmith/plugin-sdk` 0.2.0 on npm; `github.com/monosmith-app/monosmith/packages/plugin-sdk` v0.34.0 on the Go module proxy; no `monosmith-plugin-sdk`, `monosmith-sdk` or `smith-plugin-sdk` package on PyPI. The Python example is `docs/examples/plugins/python`, run with Python 3 on Windows.
- **Other tools.** Read raw from each tool's own documentation source on 2026-10-25 (the repositories named below), searched for "plugin", "extension", "extending", "isolat", "sandbox", "trust" and the tool's own vocabulary. A quoted phrase is the text relied on. ❔ means the sources were searched and neither state nor exclude the feature. Third-party plugins, blogs and commercial services were not searched.
- **Not verified.** Linux and macOS behaviour of the trust flow; the Go, TypeScript, Python and C# SDK examples were not re-run for this page; Nx, Bazel, Pants, Gradle and moon plugins were read, not built.

### Sources

Retrieved 2026-10-25 from the default branch of each repository.

- Nx (`nrwl/nx`, `astro-docs/src/content/docs`): `features/multi-language-support.mdoc` ("Plugins build that graph by reading existing configuration files"; "Browse the plugin registry for the full list."); `kb/add-language-support.mdoc` ("The plugin is a Node package"; `createNodes`, `createDependencies`); `features/generate-code.mdoc` ("Generators come as part of Nx plugins"); `kb/nx-sandbox-unix-sockets.mdoc` ("runs each plugin in a separate forked process").
- Turborepo (`vercel/turborepo`, `apps/docs/content/docs`): `guides/generating-code.mdx` ("create your own custom generator using Plop configurations"); a search for "plugin" found only CI-vendor and migration pages.
- Bazel (`bazelbuild/bazel`, `docs`): `extending/concepts.mdx` ("Bazel extensions are files ending in `.bzl`."; "Support for a new language is typically done with a rule."; "Rules are for advanced users"); `external/extension.mdx` and `external/repo.mdx` (module extensions, repository rules); `external/faq.mdx` (the Bazel Central Registry).
- Pants (`pantsbuild/pants`, `docs/docs`): `writing-plugins/overview.mdx` ("Plugins are written in typed Python 3 code."; "From a published package in a repository such as PyPI."; "Add support for new languages.").
- Rush (`microsoft/rushstack-websites`, `websites/rushjs.io/docs/pages`): `extensibility/creating_plugins.md` (title "Creating Rush plugins (experimental)"; "A plugin package is an NPM package").
- moon (`moonrepo/moon`, `website/docs`): `guides/wasm-plugins.mdx` ("any language that compiles to WASM"; "a sandboxed file system (via WASI)"; "still experimental"); `guides/extensions.mdx`; `commands/toolchain/add.mdx`; `proto/commands/plugin/add.mdx` ("must be trusted with `proto trust` before the plugin is loaded").
- Buck2 (`facebook/buck2`, `docs`): `rule_authors/writing_rules.md`; `bxl/index.md` ("Starlark-based script that enables integrators to inspect and interact").
- Gradle (`gradle/gradle`, `platforms/documentation/docs/src/docs/userguide`): `fundamentals/developing-plugins/plugin_introduction_advanced.adoc` ("A Gradle plugin is an add-on that enhances your build"); `fundamentals/authoring-builds/plugins_intermediate.adoc` (the Gradle Plugin Portal); `best-practices/best_practices_testing.adoc` ("Groovy or Kotlin"); `securing-builds/securing_dependencies.adoc` ("library or plugin from a trustworthy author").
- Lage (`microsoft/lage`, `docs/docs/reference/cli.md`): "create and use your own custom reporters".
- Yarn (`yarnpkg/berry`, `packages/docusaurus/docs/features/extensibility.mdx`): "small scripts listed in your configuration that Yarn will dynamically require at startup"; pnpm (`pnpm/pnpm.io`, `docs/pnpmfile.md`): "Hooks can be declared in a file called `.pnpmfile.mjs`".
- Cargo (`rust-lang/cargo`, `doc/book/src/reference/external-tools.md`): "Cargo is designed to be extensible with new subcommands".
- Please (`thought-machine/please`, `docs/plugins.html`): "Plugins are a way to extend Please with build rules for additional languages or technologies."
- Not found, after searching "plugin" and "extension": Lerna, Earthly, npm.
- MonoSmith: `docs/plugins/using/what-is-a-plugin.md`, `plugin-security.md`, `docs/plugins/creating/plugin-sdks.md`, `docs/architecture/decisions/0036-*.md`, `0037-*.md`, and the registries named above.
