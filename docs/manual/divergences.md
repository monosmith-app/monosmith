# smith divergences

`smith divergences` lists the places where this workspace made a different choice from smith's
defaults, each with the reason it is listed and how to go back. A divergence is a choice, not an
error: the command changes nothing and never fails. A short list means a tidy workspace.

| Kind | What it is |
|---|---|
| `override` | a task setting in `smith.toml` that replaces what a plugin infers (`[tasks.*]`, `[projects.*.tasks.*]`), or a project's stack |
| `ejected` | a derived file you own |
| `drift` | a derived file edited by hand, or one smith cannot edit |
| `differs` | a file that differs from the shape of its sibling projects |

Derived files are listed only when `[derive] enabled = true`.

## Synopsis

```sh
smith divergences [--json]
```

## Arguments and flags

<!-- smith:begin divergences -->
List where the workspace differs from smith's defaults, each with why and how to undo it

Usage: `smith divergences [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--json` | boolean | off | print the list as JSON |

<!-- smith:end divergences -->

## Example

```text
1 place where this workspace made its own choice. It is not an error.

override
  web test cache
    why:  [projects.web.tasks.test] sets cache instead of what the plugin or the workspace gives
    undo: remove cache from [projects.web.tasks.test] in smith.toml
```

`--json` prints a list of `{kind, subject, why, undo}` for editors. See
[ADR 0064](../architecture/decisions/0064-reasons-and-exceptions-are-data-before-they-are-screens.md).
