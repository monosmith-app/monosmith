# smith info

`smith info` shows the installed smith, the newest released one, and this workspace's
settings: its root, the number of projects by kind, the languages, the stacks and the default
for each language, the CI providers, the release branch, and the size of the local cache
against its limit. Outside a workspace it prints the
version and `not in a smith workspace`.

## When to use it

- To report the version in a bug report, or to see whether a newer smith exists.
- To confirm which workspace smith found (`SMITH_WORKSPACE_ROOT` and parent folders both
  decide this).
- Editors read `smith info --json`.

## Synopsis

```sh
smith info [--json]
```

## Arguments and flags

<!-- smith:begin info -->
Show the installed smith, the newest released one, and this workspace's settings

Usage: `smith info [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--json` | boolean | off | print the report as one JSON document |

<!-- smith:end info -->

`--json` prints one document: `cli` (`version`, `latest`, `updateAvailable`) and, inside a
workspace, `workspace` with `root`, `config` (the effective configuration) and `cache`
(`dir`, `bytes`, `limitBytes`; 0 means no limit). The cache line is a look at file sizes: it
hashes nothing and creates nothing. [`smith cache status`](cache-status.md) has the detail.

## Exit codes

0, also outside a workspace.

## Files and environment

Reads `smith.toml` and the size of the files in the cache directory (`SMITH_CACHE_DIR` moves it). Checks the npm registry for the newest release unless
`SMITH_NO_UPDATE_CHECK` is set; `latest` is then `null` and `updateAvailable` is false.

## Examples

```sh
smith info
```

```text
smith 1.62.0

workspace acme (~/work/acme)
  projects  4 (2 apps, 1 lib, 1 package)
  languages 1 go, 2 node, 1 python
  stacks    2 node-jest; default node-jest for node
  CI        github
  releases  from main
  cache     312 MB of 2048 MB (15%)

smith config list shows every setting; smith projects lists the projects.
```

Outside a workspace:

```text
smith 1.62.0
not in a smith workspace
```

## Related

[`smith config`](config.md), [`smith projects`](projects.md), [`smith cache`](cache.md), [Installing](../getting-started/installing.md).
