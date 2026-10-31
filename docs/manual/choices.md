# smith choices

`smith choices` lists the choices a project can be built from, grouped by axis (framework, test,
style, lint), with the plugin that owns each. A choice is one option on an axis. You pick several
with [`smith add --with`](add.md), and smith checks that they fit.

```sh
smith add app web --with react,vitest,bootstrap
```

## When to use it

- To find the names to give to `--with`.
- To see why a choice cannot be used: its language plugin is not enabled, or a choice it
  requires is offered by no plugin.

It needs a workspace, because the choices come from the plugins the workspace enables.

## Synopsis

```sh
smith choices [--json]
```

## Arguments and flags

<!-- smith:begin choices -->
List the choices a project can be built from, by axis (framework, test, style)

Usage: `smith choices [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--json` | boolean | off | print JSON |

<!-- smith:end choices -->

## Exit codes

0; 1 outside a workspace.

## Files and environment

Reads `smith.toml` (to know which plugins are enabled) and asks each plugin for its choices.

## Examples

```sh
smith choices
```

## Common errors

`there is no smith workspace here`: run it inside a workspace.

## Related

[`smith add`](add.md), [`smith stacks`](stacks.md), [Stacks](../reference/stacks.md).
