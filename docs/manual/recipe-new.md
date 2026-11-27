<!-- lore-master
parent: recipe.md
-->
# smith recipe new

`smith recipe new [path]` writes a recipe by answering questions. It asks for the recipe's
name, version and the recipe it extends, then goes through the roles (language, scaffold,
each task slot such as lint and test, structure, deploy). For each role it lists the plugins
that fit what is chosen so far, and you say whether one is fixed, a choice the recipe's
users make (python or csharp, say, with a plugin per answer), or left unset.

The recipe is checked for every combination of answers before it is written, so a recipe
this command writes resolves. It is written as the TOML you would write by hand, to
`<name>.recipe.toml` or the path you give. An existing file is replaced only after you
confirm, or with `--force`.

## When to use it

To start a recipe without learning the format; then edit the file by hand. Without a
terminal, `--from-answers` reads the same answers from a file, so a script or a test gets
the same recipe the questions would.

A recipe whose plugins are all fixed needs neither questions nor a file: `--name` and one
`--fixed role=plugin` for each role write it. A structure, deploy or toolchain role takes
plugins joined by commas (`--fixed structure=slices,layers`).

A role the recipe's users choose is written with `--chosen role=question:answer=plugin,answer=plugin`,
for example `--chosen "language=Which language?:python=python,csharp=dotnet"`. The choice takes the role's
name, its options are the answers in the order given, and the first is the default. A structure, deploy or
toolchain role joins the plugins of one answer with `+` (`--chosen "structure=Which layout?:small=slices,big=slices+layers"`).
The question cannot contain a colon, and a role is either `--fixed` or `--chosen`. Choices that two roles
share, or that a parent recipe holds, still need the questions or an answers file. The interface's **Write a recipe** form
(the guided menu) and the VS Code command **MonoSmith: New Recipe** ask for these values as pick
lists and tick boxes, and show this command line and the file it writes before it runs.

`--list-levels` writes nothing. It lists the levels and the plugins each can take; `--json`
prints them for programs. The JSON has `levels` (each with `level`, `additive` and `plugins`,
each plugin with `name` and `description`) and `fits`, which names, for each language plugin,
the plugins of every other level that fit it. The forms of the interface and of the editors read
this list, so none of them holds a list of plugin names.

## A recipe from this workspace

`--from-workspace` writes the recipe the workspace already follows, so a team that set its
workspace up by hand can give the same choices to its next projects. It reads:

- the language that most projects use; `--language` picks another one of the workspace's languages
- the stack that most projects of that language use: its scaffold, toolchain, structure and deploy
  plugins
- a task slot such as `lint` or `test` for a toolchain plugin that is the only one to provide that
  task, when the slots keep the order the toolchain applies in; otherwise the plugin stays in `toolchain`
- the code layout that smith.toml turns on (`[conventions.<architecture>]`), when it checks that language

Every role is fixed, and the recipe has no choices. The command prints what it fixes before it
writes. On a terminal it asks the language (when the projects use several), the name and the
version (0.1.0 when left empty). Without a terminal, `--name` is needed. The file is
`recipes/<name>.recipe.toml` in the workspace, or the path you give. An existing file is replaced
only with `--force`; the command does not ask. `--from-workspace` does not go with
`--from-answers`, `--fixed` or `--chosen`. The Settings tab of the interface has the same action
(**Settings › Recipes › Create a recipe from this workspace**).

```sh
smith recipe new --from-workspace --name team-go
```

```text
From this workspace (26 go project(s), no stack, the language plugin alone), the recipe fixes:
  language   go
  structure  vertical-slices

Wrote recipes/team-go.recipe.toml (recipe team-go 0.1.0).
Check it any time with: smith recipe check recipes/team-go.recipe.toml --all
```

## Synopsis

```sh
smith recipe new [path] [--force] [--from-answers <file>]
smith recipe new [path] --name <name> [--recipe-version <x.y.z>] [--extends <recipe>] [--fixed <role=plugin>...] [--chosen <role=question:answer=plugin,...>...]
smith recipe new [path] --from-workspace [--name <name>] [--recipe-version <x.y.z>] [--language <language>] [--force]
smith recipe new --list-levels [--json]
```

## Arguments and flags

<!-- smith:begin recipe new -->
Write a recipe by answering questions

Usage: `smith recipe new [path] [flags]`

| Argument | Required | Meaning |
|---|---|---|
| `path` | no | where to write the recipe; default `<name>.recipe.toml` |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--chosen <chosen>` | string, repeatable |  | a role the recipe's users choose, as role=question:answer=plugin,answer=plugin (language=Which language?:python=python,csharp=dotnet; + joins plugins of one answer); repeat for each role |
| `--extends <extends>` | string |  | the recipe this one extends: a path from the new file, or an https URL |
| `--fixed <fixed>` | string, repeatable |  | fix a role's plugin, as role=plugin (language=python, structure=slices,layers); repeat for each role |
| `--force` | boolean | off | replace the file if it exists |
| `--from-answers <from-answers>` | string |  | read the answers from a TOML file instead of asking (see docs/reference/recipes.md) |
| `--from-workspace` | boolean | off | fix the plugins this workspace's projects use now, and write recipes/<name>.recipe.toml |
| `--json` | boolean | off | with --list-levels, print the list as JSON |
| `--language <language>` | string |  the one most projects use | with --from-workspace, the language the recipe fixes (default: the one most projects use) |
| `--list-levels` | boolean | off | list the levels and the plugins each can take, and write nothing |
| `--name <name>` | string |  | the recipe's name; with --fixed, nothing is asked |
| `--recipe-version <recipe-version>` | string |  | the recipe's version (0.1.0 when not given) |

<!-- smith:end recipe new -->

The answers file is TOML: `name`, `version`, an optional `extends`, `[choices.<name>]` with
`ask`, `options` and `default`, and `[levels.<level>]` with either `fixed` plugins or a
`choice` and `plugins` by answer. See [Recipes](../reference/recipes.md#answers-file).

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The recipe was written. |
| 1 | No terminal and no `--from-answers`, the answers do not resolve, the file exists without `--force`, the answers file cannot be read, or `--from-workspace` finds nothing to read. |

## Files and environment

Writes `<name>.recipe.toml` (or `path`); with `--from-workspace`, `recipes/<name>.recipe.toml` in
the workspace. Reads the answers file and the parent recipe, and with `--from-workspace` the
workspace's `smith.toml` and projects.

## Examples

With an answers file (`answers.toml`, as in [Recipes](../reference/recipes.md#answers-file)):

```sh
smith recipe new --from-answers answers.toml
```

```text

Wrote team.recipe.toml (recipe team 0.1.0).
Check it any time with: smith recipe check team.recipe.toml --all
```

With fixed plugins only, as the interface's form runs it:

```sh
smith recipe new --name team --fixed language=python --fixed lint=ruff
```

The file it wrote:

```toml
# Recipe team, created with `smith recipe new`.
# Edit it by hand if you like; `smith recipe check <this file> --all` checks it
# for every answer.
name = "team"
version = "0.1.0"

[choices.language]
ask = "Language"
options = ["python", "node"]
default = "python"

[use]
structure = "vertical-slices"

[when.language.python.use]
language = "python"

[when.language.node.use]
language = "node"
lint = "eslint"
```

## Common errors

| Message | Cause and remedy |
|---|---|
| `the recipe file exists: team.recipe.toml` | Pass `--force`, or choose another path. |
| `smith recipe new asks questions, and there is no terminal to ask them on: write the answers in a file and use --from-answers (docs/reference/recipes.md)` | Use `--from-answers`. |
| `open nothing.toml: ...` | The answers file does not exist. |
| `--from-workspace needs the recipe's name: pass --name (lowercase words joined by dashes)` | There is no terminal to ask the name on. Pass `--name`. |
| `the workspace has no projects, so there is no language to read: ...` | Add a project first, or write the recipe with the questions or `--fixed`. |
| `no project of this workspace uses the language "rust"; the languages here are go, node` | Pass one of the languages listed to `--language`. |

## Related

[`smith recipe check`](recipe-check.md), [Recipes](../reference/recipes.md), [`smith add`](add.md).
