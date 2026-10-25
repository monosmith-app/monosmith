# smith recipe

A recipe is a team's choice of plugins and settings, written as a TOML file: which language,
scaffold, lint, test, structure and deploy plugins make up a project, with questions the
person adding a project answers. `smith recipe` checks a recipe and writes one. Using a
recipe is `[workspace] recipe = "<path or https URL>"` in `smith.toml`, or
`smith new --recipe`; then [`smith add`](add.md) asks the recipe's choices.

The recipe format, validation rules and the answers file are in
[Recipes](../reference/recipes.md).

## When to use it

- [`smith recipe new`](recipe-new.md) to write a recipe by answering questions.
- [`smith recipe check`](recipe-check.md) to validate one and see which plugin each level
  would use and where each value came from.

## Subcommands

<!-- smith:begin recipe -->
Check recipes: a team's choice of plugins and settings

Usage: `smith recipe <command>`

| Subcommand | Meaning |
|---|---|
| `smith recipe check` | Resolve and validate a recipe, showing what it would use and where each value came from |
| `smith recipe new` | Write a recipe by answering questions |

No flags.

<!-- smith:end recipe -->

## Exit codes

0 for the help; see the subcommands.

## Files and environment

A recipe is a local file or an https URL pinned by content hash in `smith.recipes.lock`.

## Related

[Recipes](../reference/recipes.md), [`smith add`](add.md), [`smith explain`](explain.md)
(`--recipe`), [`smith plugin`](plugin.md).
