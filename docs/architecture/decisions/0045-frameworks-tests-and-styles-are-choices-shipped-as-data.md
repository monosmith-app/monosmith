# ADR 0045: Frameworks, tests and styles are choices shipped as data

- Status: accepted
- Date: 2026-10-31

## Context

A web app can be React or Vue or Svelte, tested with Vitest or Jest, styled with Bootstrap or
Tailwind. Today smith has no place for most of these answers (epic
[#819](https://github.com/monosmith-app/monosmith/issues/819), phase 2, issue
[#821](https://github.com/monosmith-app/monosmith/issues/821)):

- The frameworks are Go struct literals in the `vite` plugin. Their templates and dependency
  versions are typed into the code. A new framework is a code change and a release of smith.
- There is no style axis. Bootstrap, Chakra and Tailwind have no home.
- A stack names one combination. `node-vitest-playwright` and `react-tauri` encode answers that
  are independent. Each new answer doubles the stacks.

The rules of the epic apply: the core stays free of any language, plugins never call each other,
the protocol stays at version 1, and there are no shims.

## Decision

### A choice

A **choice** is one option on an **axis**. An axis is a question: `framework`, `test`, `style`,
`lint`, `library`. The plugins agree on the axis names. The host only compares them.

A choice has these members:

| Member | Meaning |
|---|---|
| `axis`, `name`, `description` | The question it answers, its name in a selection, one line of text. The name is unique among all loaded plugins. |
| `language` | The language plugin it applies to, for example `node`. A selection holds choices of one language. |
| `roles` | `scaffold` (the plugin writes the project), `toolchain` (the plugin joins the toolchain), both, or none. A choice with no role only adds to a project that another choice of its plugin writes, as a style does. |
| `requires`, `conflicts` | Names of other choices. `chakra` requires `react`. |
| `needs`, `default` | Axes that a selection must have a choice on. An axis the person leaves out takes the choice marked `default`. `react` needs `lint` and `test`; `eslint` and `vitest` are the defaults. |
| `many` | The choice may share its axis with other choices that are `many`. Without it a selection has one choice per axis. |
| `dependencies`, `dev_dependencies`, `overrides` | Package names. The versions come from the catalog. |
| `tasks` | Scripts of the project, by task name (`typecheck = "vue-tsc ..."`). |
| `slots` | Lists of text for the templates: the Vite plugins, the imports of the entry file, the tsconfig include list. |
| `files` | Go templates by project-relative path. |

The first six rows reach the host through the protocol (`InitializeResult.choices`, the type
`protocol.Choice`). The rest stays inside the plugin: the host never reads a dependency, a task or a
file. So the host is language-free, and a plugin written in any language can publish choices.

### Facts

A template reads the shared facts through the values the scaffolding already gives it: the
`module.aliases` fact is `[[.TSPaths]]` and `[[.ViteAlias]]` ([ADR 0042](0042-plugins-publish-shared-facts-the-core-merges.md)).
A choice does not publish facts in this version. The choice file has no `facts` member, and the
loader refuses one, because no plugin reads it yet. When a choice must publish a default (for
example `source.roots`), the member is added with the reader in the same change.

### Where the data lives and why TOML

A plugin embeds a folder of data files with `go:embed`. A plugin of another language ships the same
files and reads them the same way, or returns the choices from code.

```text
libs/plugin-node-toolchains/vite/choices/
  catalog.toml   version of every package a choice names
  base.toml      what every choice of the plugin builds on: index.html, vite.config.ts, tsconfig
  stacks.toml    the stacks the plugin publishes, as saved selections
  react.toml     one file per choice
  vue.toml
  bootstrap.toml
```

The format is TOML, the format of `smith.toml`. A template is code in another language: it holds
quotes, backslashes and `{{ }}`. TOML multi-line literal strings (`'''`) hold it with no escape. JSON
would need `\n` and `\"` in every template, and has no comments. YAML makes indentation matter
inside the template. TOML also lets the catalog explain a version in a comment (the React override).

The loader reads each file strictly: an unknown member is an error. It checks every package against
the catalog, every path against the project folder, and every choice with `protocol.ValidateChoice`.
The check runs when the plugin loads, and `smith plugin check` runs the same validation.

### The version catalog

`catalog.toml` has one table, `[versions]`, from package name to range. A choice names a package and
never a version. A bump is one line in one data file. A package with no version in the catalog is an
error at load time. Dependency versions are not Go code any more.

### How choices combine into files

The plugin merges the base and the chosen choices, in selection order, into a plan:

1. **Slots grow.** A slot is a list. Each layer that sets it appends. The base templates read the
   slots, for example `plugins: [[ range index .Slots "vite.plugins" ]]`. So `bootstrap` adds one
   import to `entry.imports` and works with `react`, `vue`, `svelte` and `html` without a line about
   any of them.
2. **Dependencies and tasks join.** They go to the project's `package.json` and, for development
   dependencies, to the workspace root.
3. **A later file replaces an earlier file** only when the later choice is the base's or *requires*
   the earlier one. `chakra` requires `react`, so it may replace `src/main.tsx`. Two choices that
   write one file with no `requires` between them are an error that names both: the result would
   depend on the order of the names.

Pure Go stays where logic is needed: the tests a toolchain plugin writes (they depend on the
workspace's test suffix), the Jest compiler choice, the alias facts.

### Selecting: the solver

`smith add app web --with react,vitest,bootstrap` names choices. The host solves them against the
choices that the loaded plugins offer. The solver is pure and knows no language. It checks, in this
order, and reports the first problem as one sentence that names the choices and the fix:

| Rule | Message |
|---|---|
| a name no plugin offers | `no choice is named "reakt": did you mean react?` |
| one language | `react is for node projects and flask for python: keep choices of one language` |
| one choice per axis | `react and vue are both on the framework axis: keep one` |
| `conflicts` | `sass conflicts with bootstrap: drop one of them` |
| `requires` | `chakra requires react, which is not in the selection: add react to --with, or drop chakra` |
| `needs` with no default | `react needs a choice on the test axis, and none is the default: add one of jest, vitest to --with` |
| exactly one choice writes the project | `no choice in bootstrap writes the project: add one of react, vue` |
| a choice with no role has its plugin | `bootstrap is provided by the vite plugin, which no other choice here uses: add one of ...` |

The solver never adds a choice that `requires` asks for. A guess hides the person's mistake. It adds
only the default of an axis that a choice `needs`, and says so: `Choices: react, vitest, bootstrap
(eslint by default) as the stack react-vitest-bootstrap`.

`smith choices` lists the axes and the choices, the plugin that owns each, `requires` and
`conflicts`, and what is wrong with a choice that cannot be used (its language plugin is not
enabled, a choice it requires is offered by no plugin).

### A stack is a saved selection

`Stack` gains `choices`. A stack with choices needs no language, scaffold or toolchain: the host
derives them from the solved selection (the scaffold is the plugin of the choice that writes the
project; the toolchain is the plugins of the choices that play the toolchain role, in order). The
stack's own `toolchain`, `structure` and `deploy` lists add plugins after those.

- A plugin publishes a stack as `choices = ["react"]`. `smith stacks` shows the full list.
- `smith add --with` saves the selection in `smith.toml` as `[stacks.react-vitest-bootstrap]` with
  `choices`. The project records that stack, as any project does, so discovery, doctor, CI and
  `smith add --stack` need nothing new. A stack with the same choices is reused. A name that is
  taken by a stack with other choices gets a short hash after it.
- The host sends the full list as `GenerateParams.choices` to every plugin of the stack. A plugin
  applies the choices it owns and ignores the rest.

`--with` and `--stack` exclude each other, and so do `--with` and the recipe's `--choice`.

## Options considered

**Keep Go literals and add more of them.** Rejected: a new framework or style is a code change and
a release, and the versions stay in code.

**One stack, and one plugin, per combination.** Rejected: it is the growth this record removes.

**JSON.** Rejected: templates need escapes, and the catalog needs comments.

**The host reads the files, dependencies and tasks.** Rejected: the host would learn `package.json`
and Vite. The host would need the same for every language.

**The solver adds the choices that `requires` asks for.** Rejected: `--with vue,chakra` would turn
into React plus Vue, or fail in a confusing way. One message with the fix is clearer.

**Axes with a fixed list in the core.** Rejected: an axis is a name that plugins agree on, so a new
plugin can add `router` or `state` with no change in the core.

## Consequences

- A framework, a test runner or a style is a data file. The exit test of the epic runs
  `--with react,vitest,bootstrap` and `--with vue,jest,tailwind` through the same code. The second
  case was added as data only.
- If two choices in a row need a change in the loader or the solver, the format is wrong: fix the
  format first.
- The host and the protocol know axes, names and rules, never packages or files.
- A stack in `smith.toml` can be three lines: a name and `choices`.
- The loader needs a TOML reader, so `libs/plugin-node` depends on `go-toml`. The plugin SDK stays
  free of dependencies.
- A toolchain plugin (`vitest`, `jest`, `eslint`) publishes its choice as metadata and keeps its Go
  setup, because its files depend on the workspace.
- Limits: a choice has no settings of its own yet, a file is replaced as a whole (no patching of
  one file by two choices), and a choice cannot publish facts.

## Sources

- Epic [#819](https://github.com/monosmith-app/monosmith/issues/819), issue
  [#821](https://github.com/monosmith-app/monosmith/issues/821)
- [ADR 0021](0021-plugin-roles-and-per-project-stacks.md): roles and stacks
- [ADR 0041](0041-smith-add-follows-the-dialect-of-sibling-projects.md): the dialect of siblings
- [ADR 0042](0042-plugins-publish-shared-facts-the-core-merges.md): shared facts
- [Writing choices](../../plugins/creating/choices.md), [Stacks](../../reference/stacks.md)
