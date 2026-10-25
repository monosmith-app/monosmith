<!-- lore-master
parent: recipe.md
-->
# smith recipe check

`smith recipe check <file-or-url>` loads a recipe and the recipes it extends (a path, or an
https URL pinned by content hash in `smith.recipes.lock`), answers its choices, and checks
the result with the rules a stack is checked with: one plugin per exclusive level and slot,
requirements met, no conflicts, every plugin installed at a version in range.

Choices are answered with `--choice name=answer`, else by their default. It prints the
choices asked, the plugin for each level and the merged settings, each with the recipe, file
and line it came from, or why the recipe is not valid.

## Synopsis

```sh
smith recipe check <file-or-url> [--choice name=answer]... [--all] [--json]
```

## Arguments and flags

<!-- smith:begin recipe check -->
Resolve and validate a recipe, showing what it would use and where each value came from

Usage: `smith recipe check <file-or-url> [flags]`

| Argument | Required | Meaning |
|---|---|---|
| `file-or-url` | yes | the recipe file or https URL |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--all` | boolean | off | check every combination of answers (up to 64), not only the one given or defaulted |
| `--choice <choice>` | name=answer, repeatable |  | answer a choice: name=answer (repeatable) |
| `--json` | boolean | off | print JSON |

<!-- smith:end recipe check -->

- `--all` checks every combination of answers (up to 64), not only the one given or
  defaulted. In the trial it printed the same report as the default combination and the
  exit code was 0.
- `--json` prints the same for tools.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The recipe is valid. |
| 1 | The recipe is not valid, or cannot be read. |

## Files and environment

Reads the recipe file, the recipes it extends, and `smith.recipes.lock` for URLs. Writes
nothing.

## Examples

A recipe with a `language` choice:

```sh
smith recipe check team.recipe.toml
```

```text
Recipe team 0.1.0
  chain: team 0.1.0 (team.recipe.toml)

Choices
  language     python         (default; options python, node) Language?

Plugins
  language     python                   team.recipe.toml:16 (recipe team)
  structure    vertical-slices          team.recipe.toml:13 (recipe team)

The recipe is valid.
```

Answering the choice:

```sh
smith recipe check team.recipe.toml --choice language=node
```

```text
Recipe team 0.1.0
  chain: team 0.1.0 (team.recipe.toml)

Choices
  language     node           (answered; options python, node) Language?

Plugins
  language     node                     team.recipe.toml:19 (recipe team)
  lint         eslint                   team.recipe.toml:20 (recipe team)
  structure    vertical-slices          team.recipe.toml:13 (recipe team)

The recipe is valid.
```

## Common errors

| Message | Cause and remedy |
|---|---|
| `error: --choice language=rust: "rust" is not one of the options of choice "language" (python, node), declared at team.recipe.toml:7 (recipe team)` | Answer with one of the options. |
| `error: reading the recipe nothing.toml: open nothing.toml: ...` | The path does not exist. |

## Related

[`smith recipe new`](recipe-new.md), [`smith explain`](explain.md), [Recipes](../reference/recipes.md).
