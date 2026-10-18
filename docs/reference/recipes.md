# Recipes

A recipe is a team's description of which plugin fills each level of a project
and with what settings. It is data: a TOML file with no code, so it can only
choose and configure plugins, never run anything. Recipes can extend one
another and ask a few choices, and the answers decide the rest.

Recipes are being built in steps (epic #312). What exists today is the file
format, loading, resolving, validating and `smith recipe check`. Creating
projects from a recipe (`smith new` and `smith add`), recording the recipe in a
project, upgrades, the creation wizard and versioned recipes are still to come;
the whole design is in [recipes and CI](../design/recipes-and-ci.md).

## The file

```toml
# team.recipe.toml
name = "team"
version = "1.2.0"
extends = "company/base.recipe.toml"   # optional, one parent: a path or an https URL

[choices.language]
ask = "Language"
options = ["python", "csharp", "node"]

[choices.kind]
ask = "What are you creating?"
options = ["library", "service"]
default = "library"

[when.language.python.use]
language = "python"
structure = "vertical-slices"

[when.language.csharp.use]
language = "dotnet"
structure = "vertical-slices"

[when.language.csharp.settings.dotnet]
target_framework = "net10.0"

[when.language.csharp.when.kind.service.use]
deploy = "docker"

[when.language.node.use]
language = "node"
lint = "eslint"
test = "vitest"
```

| Key | Does |
|---|---|
| `name` | required; names the recipe in messages |
| `version` | the recipe's version |
| `extends` | the parent recipe: a path relative to this file, or an `https://` URL. At most one |
| `[choices.<name>]` | a question: `ask` (the text), `options` (the answers allowed) and `default` |
| `[use]` | fills a level or a slot with a plugin (below) |
| `[settings.<plugin>]` | configures a plugin; any TOML, tables merge |
| `[when.<choice>.<answer>]` | a block (`choices`, `use`, `settings`, more `when`) that applies only for that answer |

An unknown key is an error with its file, line and column, and the keys that
would have been accepted there. So is a `default` that is not one of the
`options`, a missing `name`, and a `use` value that is not a plugin name.

### Use: levels and slots

Each key of `use` is a **level** or a **slot**, and its value is a plugin name
(`language = "python"`) or, for the levels that take several, a list
(`structure = ["vertical-slices"]`).

| Key | Is | Takes |
|---|---|---|
| `language` | exclusive level | one plugin |
| `scaffold` | exclusive level | one plugin |
| `structure` | additive level | any number |
| `deploy` | additive level | any number |
| `toolchain` | toolchain plugins, without naming their slots | any number |
| anything else (`lint`, `test`, `typecheck`, `e2e`, ...) | a **slot**: the task a toolchain plugin provides | one plugin, which must provide that slot |

A plugin may carry a version constraint after `@`: `lint = "eslint@^1.2"`
(`1.2.3`, `^1.2`, `~1.2` and `>=1.2` work, as in `[plugins.external]`).

Which plugins can be named, and which level each plays, is what
`smith plugin list` shows. See [plugins](plugins.md).

## Choices and `when` blocks

A choice is answered by `--choice name=answer`, else by its `default`. One
with neither is an error that names it and says how to answer. A given answer
must be one of the `options`.

A `when.<choice>.<answer>` block applies only when the choice has that answer.
There is no logic beyond that: no comparisons and no computed values. A block
can hold its own `choices`, asked only when the block applies, and its own
`when` blocks, so combinations nest:

```toml
[when.language.csharp.when.kind.service.use]
deploy = "docker"          # only for a C# service
```

A `when` block must name a choice a recipe in the chain declares, and an
answer among its options; a typo is an error, not a block that never applies.

## Resolving

`Resolve` takes the chain of recipes and the answers and produces one result:

1. The choices of every recipe in the chain are merged key by key, the
   child's over the parent's: a child can change only the `default` of a
   choice and keep its `options`.
2. The recipes are applied **oldest first**. Within one recipe the
   unconditional `use` and `settings` come first, then the `when` blocks that
   match, **in file order**. A block's own `use` and `settings` come before
   its nested `when` blocks.
3. A later value replaces an earlier one, key by key. Settings tables merge
   recursively; a value that is not a table, including a list, replaces.
4. Every value records the recipe, file and line it came from.

## Validating

Before anything is written, the result is checked with the same rules a stack
is checked with:

- every plugin named exists, as a built-in plugin, one declared in
  `smith.toml` or one on the plugin search path, and at a version its
  constraint accepts
- a slot is filled by a plugin that provides it
- the language and the scaffold are one plugin each, and a language is chosen
- one plugin per exclusive level and per slot, every requirement met, no
  conflicts, and every plugin applies to the chosen language

An error names both sides and the recipe lines that chose them:

```
error: eslint applies to [node], not to python projects
error:   chosen at:
error:     language = "python"  broken.recipe.toml:4 (recipe broken)
error:     lint = "eslint"  broken.recipe.toml:5 (recipe broken)
```

Settings for a plugin the recipe does not use are carried, not checked.

## Recipes from URLs

`extends` (and the recipe you check) may be an `https://` URL. Plain `http://`
is refused. A recipe loaded from a URL can extend only other URLs, relative to
its own.

In a workspace a URL recipe is **pinned**: its content is cached under
`.smith/recipes/` (build output, ignored by git) and its SHA-256 recorded in
`smith.recipes.lock`, which is committed:

```toml
# Written by smith. It pins the content of every recipe loaded from a URL, so
# a build does not change because a URL's content did. Review a change, then
# accept it with smith upgrade.

[[recipe]]
url = "https://example.com/recipes/team.recipe.toml"
sha256 = "9f2c..."
```

- The first load fetches the recipe and pins it.
- Later loads use the cached copy without the network, so a URL changing
  changes nothing.
- On a machine without the cache (a fresh clone) the recipe is fetched and
  must match the pinned hash. A recipe that changed is refused, with a message
  pointing at `smith upgrade`.
- Accepting the new content is `Loader.Accept(url)` (all pinned URLs:
  `AcceptAll`). It fetches again and replaces the cached copy and the pin, and
  reports the old and new hash. `smith upgrade` will call it; no command does
  yet.

Outside a workspace nothing is cached or pinned; the URL is fetched each time.

## `smith recipe check`

```sh
smith recipe check team.recipe.toml --choice language=csharp --choice kind=service
smith recipe check https://example.com/recipes/team.recipe.toml --json
```

Loads the recipe and its parents, answers the choices, resolves and validates
it. It prints the recipes in the chain, the choices asked and how each was
answered, the plugin for each level and the merged settings, each with its
recipe, file and line, or why the recipe is not valid. The exit code is
non-zero when it is not. A choice with no answer is listed as needing one.

`--json` prints the same for tools: `name`, `version`, `chain`, `choices`,
`use`, `settings`, `origins`, `pinned` and `errors` (empty when valid).
