# recipe-widget-team

A smith **recipe**: a team's choice of plugins and settings, written as data. It chooses the widget
plugins of the other examples: every project gets the `widget` language, the `widget-lint` plugin in
the `lint` slot and the `widget-layout` structure rules; a `service` also gets the `widget-service`
scaffold with port 9000 and the `widget-bundle` deploy plugin.

Guide: [Creating a recipe and a stack](../../../plugins/creating/recipe-and-stack.md).

## Files

| File | What it is |
|---|---|
| `widget-team.recipe.toml` | the recipe |
| `answers.toml` | the answers `smith recipe new --from-answers` reads to write a recipe like it |

## Check, use

In a workspace that declares the five widget plugins (see the [examples index](../README.md#using-them-together)):

```sh
smith recipe check widget-team.recipe.toml --all
smith recipe check widget-team.recipe.toml --choice kind=service
smith recipe new widget-team-generated.recipe.toml --from-answers answers.toml
```

```toml
# smith.toml
[workspace]
recipe = "recipes/widget-team.recipe.toml"
```

```sh
smith add app billing --choice kind=service
smith add lib money
```

Put the recipe in a folder of its own: the built-in `recipe` plugin treats a `*.recipe.toml` that is
alone in its folder as a package project.
