# Creating a recipe and a stack

A **stack** is a named set of plugins a project is built with. A **recipe** is a team's choice of
which plugin fills each level of a project, and with what settings, written as a TOML file with no
code. Neither is a plugin, and neither runs anything: they choose and configure the plugins you
(or smith) wrote. This guide shows both, using the widget plugins from the other guides.

## When to write one

| You want | Write |
|---|---|
| every project of a language to be built with the same plugins, chosen by name | a stack in `smith.toml` (`[stacks.<name>]`) |
| a plugin to bring its own ready-made stack | a published stack: the plugin returns `stacks` from `initialize` |
| new projects to be made from a team standard, with a few questions whose answers choose plugins and settings | a recipe |
| a standard shared between repositories and versioned | a recipe, released like a package (the `recipe` plugin) |

## A stack

### Declared in `smith.toml`

```toml
[stacks.widget-strict]
description = "widget projects with the house lint and layout rules"
language = "widget"
toolchain = ["widget-lint"]
structure = ["widget-layout"]
# scaffold = "widget-service"     # default: the language plugin generates
# deploy   = ["widget-bundle"]
```

| Key | Level | Takes |
|---|---|---|
| `language` | the language plugin that discovers the stack's projects | one name |
| `toolchain` | toolchain plugins, in the order they apply | a list; each slot has one plugin, and a later plugin may replace a task an earlier one added |
| `structure` | structure plugins | a list |
| `scaffold` | the plugin that writes new projects | one name; empty means the language plugin does |
| `deploy` | deploy plugins | a list |

A stack declared here wins over a published one of the same name. smith checks it against the loaded
plugins as soon as it loads `smith.toml`; a stack that cannot be used is an error that names both
sides (see [Troubleshooting plugins](../using/troubleshooting-plugins.md#stacks-and-conventions)).

### Published by a plugin

A plugin returns `Stacks` in its `initialize` answer; the stack needs a unique `Name` and a
`Language`. The `widget-lint` plugin publishes `widget-linted`, the `widget-service` scaffold publishes
`widget-service`; `smith stacks` lists every stack and where it came from:

```text
$ smith stacks
widget-linted  (plugin widget-lint)
  widget projects that must describe and license themselves
  language   widget
  toolchain  widget-lint

widget-strict  (smith.toml)
  widget projects with the house lint and layout rules
  language   widget
  toolchain  widget-lint
  structure  widget-layout
```

### Using a stack

```sh
smith add lib strict --stack widget-strict     # a new project, built with it
smith add lib plain --stack none               # the language plugin alone
```

A project records its stack in its manifest (for widgets, `"stack"` in `widget.json`, which the
language plugin writes and reads back). `[projects.<name>] stack = "..."` in `smith.toml` replaces
the recorded one, and `[default_stacks]` gives a language's projects a stack when they record none.
See [Stacks](../../reference/stacks.md).

## A recipe

A recipe asks questions; the answers decide the plugins and settings. It lives in a file, usually
in a repository of its own so that teams share it. The example is
[`recipe-widget-team`](../../examples/plugins/recipe-widget-team):

```toml
name = "widget-team"
version = "0.1.0"

[choices.kind]
ask = "What are you creating?"
options = ["library", "service"]
default = "library"

# Whatever the answer: the widget language, the house linter in the lint
# slot, and the layout rules.
[use]
language = "widget"
lint = "widget-lint"
structure = "widget-layout"

# A service is made by the widget-service scaffold, listens on 9000 and
# is deployed by the bundle plugin.
[when.kind.service.use]
scaffold = "widget-service"
deploy = "widget-bundle"

[when.kind.service.settings.widget-service]
port = 9000
```

| Key | Does |
|---|---|
| `name`, `version` | the recipe's identity; `version` is what a release of the recipe bumps |
| `extends` | one parent recipe: a path relative to this file, or an `https://` URL |
| `[choices.<name>]` | a question: `ask`, `options` (the allowed answers) and `default` |
| `[use]` | fills a **level** (`language`, `scaffold`, `structure`, `deploy`, `toolchain`) or a **slot** (`lint`, `test`, `typecheck`, `e2e`, ...) with a plugin; `name@^1.2` adds a version constraint |
| `[settings.<plugin>]` | configures a plugin: any TOML, tables merge |
| `[when.<choice>.<answer>]` | a block (`use`, `settings`, more `choices` and `when`) that applies only for that answer |

A key of `use` that is not a level is a **slot**, and the plugin must provide it. The recipe above
fills `lint` with `widget-lint` because that plugin lists `lint` in `provides`.

### Check it

```text
$ smith recipe check recipes/widget-team.recipe.toml --choice kind=service
Recipe widget-team 0.1.0
  chain: widget-team 0.1.0 (recipes/widget-team.recipe.toml)

Choices
  kind         service        (answered; options library, service) What are you creating?

Plugins
  deploy       widget-bundle            recipes/widget-team.recipe.toml:24 (recipe widget-team)
  language     widget                   recipes/widget-team.recipe.toml:16 (recipe widget-team)
  lint         widget-lint              recipes/widget-team.recipe.toml:17 (recipe widget-team)
  scaffold     widget-service           recipes/widget-team.recipe.toml:23 (recipe widget-team)
  structure    widget-layout            recipes/widget-team.recipe.toml:18 (recipe widget-team)

Settings
  widget-service.port                      9000             recipes/widget-team.recipe.toml:27 (recipe widget-team)

The recipe is valid.
```

`--all` checks every combination of answers (up to 64) and is what a recipe project's `test` task
runs. Every value shows the file and line it came from.

### Create projects from it

```toml
# smith.toml
[workspace]
recipe = "recipes/widget-team.recipe.toml"      # or an https URL
```

```text
$ smith add app billing --choice kind=service
Recipe widget-team 0.1.0 (kind=service) will use:
  deploy     widget-bundle   recipes/widget-team.recipe.toml:24 (recipe widget-team)
  language   widget          recipes/widget-team.recipe.toml:16 (recipe widget-team)
  lint       widget-lint     recipes/widget-team.recipe.toml:17 (recipe widget-team)
  scaffold   widget-service  recipes/widget-team.recipe.toml:23 (recipe widget-team)
  structure  widget-layout   recipes/widget-team.recipe.toml:18 (recipe widget-team)
  setting widget-service.port = 9000
Creating app billing in apps/billing: widget-team-service
Adding widget-lint
Wrote 5 file(s):
  apps/billing/README.md
  apps/billing/WIDGET-LINT.md
  apps/billing/service.json
  apps/billing/src/handler.txt
  apps/billing/widget.json
billing listens on port 9000: edit apps/billing/service.json
```

smith turns the resolved plugins into a stack (`widget-team-service`, shared by every project made
with the same answers), records the project's recipe, version and answers, and gives each plugin the
`settings.<plugin>` table it names, when the project is generated (`GenerateParams.Settings`) and
afterwards in `Project.Settings`. The scaffold wrote `"port": 9000` because the recipe's setting
reached it. What smith.toml got:

```toml
[stacks.widget-team-service]
description = "From the recipe widget-team (kind=service); written by smith"
language = "widget"
toolchain = ["widget-lint"]
structure = ["widget-layout"]
scaffold = "widget-service"
deploy = ["widget-bundle"]

[projects.billing.recipe]
name = "widget-team"
version = "0.1.0"
source = "recipes/widget-team.recipe.toml"

[projects.billing.recipe.answers]
kind = "service"

[projects.billing.settings.widget-service]
port = 9000
```

`smith explain billing --recipe` lists what the recipe decided for the project, with the line each value
came from; `smith doctor` and `smith upgrade` keep projects current when the recipe changes.

### Write it with `smith recipe new`

`smith recipe new` exists. It asks for the name, version and parent, then walks through the levels
(language, scaffold, each slot, structure, deploy), offering only the plugins that fit what is chosen
so far, and checks the result for every answer before writing it. Without a terminal, give it the same
answers as a file:

```toml
# answers.toml
name = "widget-team-generated"
version = "0.1.0"

[choices.kind]
ask = "What are you creating?"
options = ["library", "service"]
default = "library"

[levels.language]
fixed = ["widget"]

[levels.lint]
fixed = ["widget-lint"]

[levels.structure]
fixed = ["widget-layout"]

[levels.scaffold]
choice = "kind"
plugins = { service = "widget-service" }

[levels.deploy]
choice = "kind"
plugins = { service = "widget-bundle" }
```

```text
$ smith recipe new /tmp/x.recipe.toml --from-answers answers.toml
Wrote /tmp/x.recipe.toml (recipe widget-team-generated 0.1.0).
Check it any time with: smith recipe check /tmp/x.recipe.toml --all
```

The file it wrote has the same `[use]` and `[when.kind.service.use]` as the one above. The wizard does
not write `settings` or nested `when` blocks: add them by hand and run `smith recipe check --all`.

### Release it like a package

Put the recipe in a folder of its own and the built-in [`recipe`](../using/catalogue/recipe.md)
plugin makes it a package project: its `test` task is `smith recipe check --all`, and `smith release`
tags `widget-team@1.0.0` and attaches the recipe to the GitHub Release, with `version` stamped.
Teams then pin that file by URL: see [Recipes](../../reference/recipes.md#versioned-recipes). A
`*.recipe.toml` beside other files, or at the workspace root, is **not** a project: in the example
workspace the file in `recipes/` became the project `widget-team` (kind `package`) because it was the
only file in its folder.

## Common mistakes

| You see | Cause and remedy |
|---|---|
| `error: eslint applies to [node], not to widget projects` followed by `chosen at:` and the two lines | a plugin that does not apply to the chosen language. The lines name the recipe, file and line of each side |
| `use.lnt names widget-lint, which does not provide a "lnt" slot (it provides: lint). A key of use is a level (...) or the task ...` | a typo in a slot name |
| `use.lint names the plugin "nope": no such plugin: "nope" is not built in, declared in smith.toml or installed (built in: ...)` | the plugin is not declared: add `[plugins.external.<name>]` or install it |
| `choice "k" has the default "c", which is not one of its options (a, b)` | the `default` must be one of `options` |
| `use.lint wants widget-lint ^2, but the plugin is 0.1.0` | the `@constraint` of the plugin is not met |
| `the recipe chooses no language plugin; set language = "..." under [use], or in the when block of each language` | no `language` in any branch |
| `smith.toml: ... unknown field` / a recipe key error with file, line and column | recipes reject unknown keys and name the keys that would be accepted there |
| `--choice` with `--stack` is an error | `--stack` bypasses the recipe; use one or the other |
| `smith recipe new` says there is no terminal | pass `--from-answers <file>` |

## See also

- [Recipes](../../reference/recipes.md), [Stacks](../../reference/stacks.md), [`smith recipe`](../../manual/recipe.md)
- [Creating a toolchain plugin](toolchain-plugin.md): the `lint` slot filled here
- [Creating a scaffold plugin](scaffold-plugin.md): the settings a recipe hands over
- [The recipe plugin](../using/catalogue/recipe.md)
