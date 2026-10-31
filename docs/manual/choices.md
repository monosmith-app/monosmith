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

```text
framework
  html         a page in TypeScript with no framework, built with Vite  [vite]
  react        a React 19 app in TypeScript, built with Vite  [vite]
  svelte       a Svelte 5 app in TypeScript, built with Vite and checked by svelte-check  [vite]
  vue          a Vue 3 app in TypeScript, built with Vite and type-checked by vue-tsc  [vite]

library
  p5           draw with p5: a bouncing-ball slice (pure motion, a p5 sketch, a lazy component)  [vite]
               requires react

lint
  eslint       lint and format with @monosmith/eslint-config  [eslint]  (default)

style
  bootstrap    Bootstrap 5 styles, imported by the entry file  [vite]
  chakra       Chakra UI 3 components; the app is wrapped in its provider (React only)  [vite]
               requires react
  tailwind     Tailwind CSS 4 through its Vite plugin, imported by the entry file  [vite]

test
  jest         test with Jest (ts-jest, or @swc/jest when the siblings use it)  [jest]
  vitest       test with Vitest  [vitest]  (default)
```

Each line is a choice, its description and the plugin that owns it in brackets. `(default)` marks
the choice that fills an axis a selection needs and leaves out. `requires` and `conflicts` follow a
choice that has them, and a choice that cannot be used shows `✖ cannot be used:` and the reason.
`--json` prints the same as data.

## Common errors

`there is no smith workspace here`: run it inside a workspace.

## Related

[`smith add`](add.md), [`smith stacks`](stacks.md), [Stacks](../reference/stacks.md).
