<!-- lore-master
parent: plugin.md
-->
# smith plugin list

`smith plugin list` lists the built-in plugins and the ones installed on the plugin search
path. For each: the levels it plays, the task slots it fills, and what it requires.

## Synopsis

```sh
smith plugin list [--json]
```

## Arguments and flags

<!-- smith:begin plugin list -->
List the built-in plugins and the ones installed on the plugin search path

Usage: `smith plugin list [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--json` | boolean | off | print JSON |

<!-- smith:end plugin list -->

`--json` prints an array: `name`, `source` (`built-in`, or where it was installed), `roles`
and `capabilities` (`discover`, `generate`, `version`, `publish`, `ci`, `doctor`).

## Exit codes

0.

## Files and environment

Reads `smith.toml` and the search path (see [`smith plugin`](plugin.md#where-installed-plugins-are-found)).

## Examples

```sh
smith plugin list
```

```text
PLUGIN           LEVELS                       SLOTS                REQUIRES
bicep            language,toolchain,scaffold  -                    -
...
node             language,toolchain,scaffold  -                    -
...
angular          scaffold,toolchain           build,typecheck      language plugin node
eslint           toolchain                    lint                 language plugin node
jest             toolchain                    test                 language plugin node
vertical-slices  structure                    -                    -
...

Installed (search path: ~/work\acme\.smith\plugins, ~\.smith\plugins):
  none
```

(Twelve language plugins, then twelve toolchain, scaffold, structure and deploy plugins;
rows between `...` are cut.)

## Related

[`smith plugin`](plugin.md), [`smith plugin check`](plugin-check.md), [`smith stacks`](stacks.md).
