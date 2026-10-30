<!-- lore-master
parent: plugin.md
-->
# smith plugin list

`smith plugin list` lists the built-in plugins, the external plugins the workspace uses, and
the ones installed on the plugin search path. For each built-in plugin: the levels it plays, the
task slots it fills, and what it requires. For each external plugin the workspace declares in
`smith.toml`: where it comes from, whether it is trusted and whether `smith.toml` pins it, and the
environment variables it may see beyond the baseline (names, never values).

## Synopsis

```sh
smith plugin list [--json]
```

## Arguments and flags

<!-- smith:begin plugin list -->
List the built-in plugins, the external ones the workspace uses (trusted or not) and the ones installed on the plugin search path

Usage: `smith plugin list [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--json` | boolean | off | print JSON |

<!-- smith:end plugin list -->

`--json` prints an array: `name`, `source` (`built-in`, or where it was installed), `roles`
and `capabilities` (`discover`, `generate`, `version`, `publish`, `ci`, `doctor`). An external
plugin the workspace uses also has `used`, `command`, `env`, `trusted` and `pinned`.

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
applescript      language,toolchain,scaffold  -                    -
...
bicep            language,toolchain,scaffold  -                    -
...
lua              language,toolchain,scaffold  -                    -
swift            language,toolchain,scaffold  -                    -
...
vba              language,toolchain,scaffold  -                    -
...
node             language,toolchain,scaffold  -                    -
...
angular          scaffold,toolchain           build,typecheck      language plugin node
eslint           toolchain                    lint                 language plugin node
jest             toolchain                    test                 language plugin node
vertical-slices  structure                    -                    -
...

External plugins this workspace uses (smith starts them only when trusted):
  none

Installed (search path: ~/work\acme\.smith\plugins, ~\.smith\plugins):
  none
```

With a plugin declared in `smith.toml` (`widget`, installed in the workspace and not yet trusted):

```text
External plugins this workspace uses (smith starts them only when trusted):
  PLUGIN         FROM               TRUSTED  PINNED  MAY SEE BEYOND THE BASELINE
  widget         .smith/plugins     no       no      GITHUB_TOKEN, NPM_TOKEN
```

`FROM` is `smith.toml` (its `command`), `.smith/plugins`, `SMITH_PLUGIN_PATH` or
`$SMITH_HOME/plugins`.

(Sixteen language plugins, then twelve toolchain, scaffold, structure and deploy plugins;
rows between `...` are cut.)

## Related

[`smith plugin`](plugin.md), [`smith plugin check`](plugin-check.md), [`smith stacks`](stacks.md).
