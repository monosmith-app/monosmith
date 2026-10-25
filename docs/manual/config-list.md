<!-- lore-master
parent: config.md
-->
# smith config list

`smith config list` prints every setting that `smith config` can edit, with the value
written in `smith.toml` (or `(default)`) and a one-line description that states the
default.

## When to use it

To find a key's name, see whether it is set, or read its default. For one value as it
is written, use [`smith config get`](config-get.md).

## Synopsis

```sh
smith config list [--json]
```

## Arguments and flags

<!-- smith:begin config list -->
List the settings, their values and what they do

Usage: `smith config list [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--json` | boolean | off | print the settings as JSON |

<!-- smith:end config list -->

`--json` prints an array of objects, one per setting: `path`, `type` (`string`, `boolean`,
`integer`, `array`), `description`, `set` (true when the file has it), `value` (the TOML
text of the value, only when set) and the schema's limits (`minimum`, `minLength`, and
so on) where there are some.

## Exit codes

0 on success; 1 outside a workspace.

## Files and environment

Reads `smith.toml`. Writes nothing.

## Examples

```sh
smith config list
```

```text
workspace.exclude                    (default)               Workspace-relative globs (examples/**) of directories whose manifests are not projects, in any language.
workspace.module_prefix              "github.com/acme/acme"  Go module path prefix, e.g. github.com/acme/repo.
workspace.name                       "acme"                  The workspace's name.
...
ci.branches                          (default)               Branches the generated GitHub and Azure pipelines run on, for pushes and pull requests. ...
ci.devcontainer                      true                    Write .devcontainer/devcontainer.json with the toolchains CI installs. Default: false.
ci.providers                         ["github"]              Pipelines to generate. Default: ["github"].
release.branch                       (default)               The only branch a real release runs from. Default: main.
...
cache.dir                            (default)               The task cache directory, workspace-relative or absolute. Default: .smith/cache.
cache.max_size_mb                    (default)               Prune the least recently used entries past this size. Default: 2048.
...
plugins.enabled                      (default)               Built-in language plugins to load. Default: all of them.
default_stacks.node                  "node-jest"             the stack of node projects that record none, and of smith add without --stack (for example node-jest)
```

(The listing is cut with `...`; the real output has 29 rows in this version.)

## Common errors

`there is no smith workspace here (no smith.toml in this folder or above it)`: run it
inside a workspace, or `smith new`.

## Related

[`smith config get`](config-get.md), [`smith config set`](config-set.md),
[Configuration](../reference/configuration.md).
