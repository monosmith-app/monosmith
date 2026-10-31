# Writing choices

A **choice** is one option on an axis of a project: a framework, a test runner, a style library.
A plugin ships its choices as data files, not as code. A new choice is a new file. This page
shows the format, the rules the host applies, and how to test a choice. The decision behind it is
[ADR 0045](../../architecture/decisions/0045-frameworks-tests-and-styles-are-choices-shipped-as-data.md).

Read first: [Stacks](../../reference/stacks.md#choices-a-stack-as-a-saved-selection) (what a person
does with choices) and [the toolchain plugin guide](toolchain-plugin.md).

## What a plugin ships

A plugin embeds a folder of TOML files. The Node plugins use `libs/plugin-node/choicedata` to read
it; a plugin in another language reads the same files, or builds the same `protocol.Choice` values
in code.

```text
choices/
  catalog.toml   the version of every package a choice names
  base.toml      what every choice of the plugin builds on (optional)
  stacks.toml    the stacks the plugin publishes, as saved selections (optional)
  react.toml     one file per choice; the file name is the choice name
  bootstrap.toml
```

```go
//go:embed choices
var choiceFiles embed.FS

set, err := choicedata.Load(choiceFiles, "choices")           // checks every file
// In Describe:   Choices: set.Choices(), Stacks: set.Stacks()
// In Generate:   plan, err := set.Plan(params.Choices)         // the merge of the base and your choices
```

`Load` refuses an unknown member, a package with no version in the catalog, a path that leaves the
project, a name used twice, and a choice that `protocol.ValidateChoice` refuses.

## A choice file

```toml
axis = "style"
name = "bootstrap"
description = "Bootstrap 5 styles and components"
language = "node"
dependencies = ["bootstrap"]

[slots]
"entry.imports" = ["import 'bootstrap/dist/css/bootstrap.min.css'"]
```

| Member | Meaning |
|---|---|
| `axis`, `name`, `description` | What it answers, how a person names it, one line for `smith choices`. Lowercase letters, digits and dashes. The name is unique among all loaded plugins. |
| `language` | The language plugin it applies to. |
| `roles` | `scaffold`, `toolchain`, both, or none. See below. |
| `needs` | Axes the selection must have a choice on: `needs = ["lint", "test"]`. |
| `default` | `true` when this is the choice added for a needed axis the person left out. One per axis. |
| `many` | `true` when it may share its axis with other choices that are `many` too. |
| `requires`, `conflicts` | Names of other choices. |
| `dependencies` | Packages for the project's `package.json`. Names only. |
| `dev_dependencies` | Packages for the workspace root's `package.json`. |
| `overrides` | Packages the workspace root pins with npm `overrides`. |
| `[tasks]` | The project's scripts: `typecheck = "tsc -p tsconfig.json --noEmit"`. No quotes in the command. |
| `[slots]` | Lists of text for the templates (below). |
| `[files]` | Go templates by project-relative path (below). |

There is no `facts` member yet: the loader refuses it.

### Roles

- `scaffold`: the plugin writes the project. Exactly one choice of a selection has it.
- `toolchain`: the plugin joins the stack's toolchain, so it contributes tasks (`vite` builds,
  `vitest` tests).
- none: the choice only adds to a project that another choice of the same plugin writes. A style
  has no role: `bootstrap` needs a framework of the plugin that owns it, and the solver says which
  frameworks fit when none is chosen.

### The catalog

```toml
[versions]
react = "^19.3.1"
bootstrap = "^5.3.8"
# npm cannot place Babel 7 and 8 together (issue 584), so the workspace stays on 7.
"@babel/plugin-transform-runtime" = "^7.29.0"
```

A choice never writes a version. To bump one, edit one line.

### Slots

A slot is a named list. The base and each choice append to it, in selection order. A template reads
it with `[[ range index .Slots "name" ]]`. A new choice adds to a slot and needs no change in any
other file. The slots of the Vite plugin:

| Slot | Read by | Holds |
|---|---|---|
| `entry` | `index.html` | the file under `src/` that the page loads (one item) |
| `entry.imports` | the entry file | lines to import first, for example a stylesheet |
| `vite.imports`, `vite.plugins` | `vite.config.ts` | imports, and the calls in `plugins: [...]` |
| `tsconfig.options`, `tsconfig.include` | `tsconfig.json` | extra compiler options, and globs |

### Files

A file is a Go template with `[[ ]]` delimiters. It sees `.Dir`, `.Name`, `.Package`, `.Stack`,
`.Slots` and the alias values from the shared facts. A path is relative to the project.

A choice may replace a file of the base, and a file of a choice it `requires`. `chakra` requires
`react`, so it may write `src/main.tsx`. Two choices that write one file with no `requires` between
them are an error that names both.

## The rules the host applies

The host knows only axes, names and rules. See the table in
[ADR 0045](../../architecture/decisions/0045-frameworks-tests-and-styles-are-choices-shipped-as-data.md#selecting-the-solver).
Run `smith choices` to see what the host sees, and what is wrong with a choice that cannot be used.

## Testing a choice

1. `smith plugin check <plugin>` validates the choices that the plugin publishes.
2. Write a golden path: a workspace, `smith add app web --with <your choices>`, and then
   `smith run lint typecheck test build`. The repository's example is
   `apps/smith/goldenpath/web_choices_use_case_test.go`; it reads its selections from
   `testdata/web-selections.txt`, so a new choice adds one line there.
3. A choice with no golden path is marked unsupported in its `description`.

## If your choice needs a change in the loader

Stop. Two choices in a row that need a change in the loader or the solver mean the format is wrong.
Open an issue that names the choices, and fix the format first.
