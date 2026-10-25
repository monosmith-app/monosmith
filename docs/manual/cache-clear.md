<!-- lore-master
parent: cache.md
-->
# smith cache clear

`smith cache clear` deletes the local task cache: every entry, blob, last-run pointer and
record under the cache directory's `v1` folder. The next run of each task is a cache miss and
runs the task again. The remote cache is never touched. In a terminal it asks first; without
one it refuses unless you pass `--yes`.

## When to use it

- To measure a cold build, or to see whether a problem is a stale cache.
- To reclaim all the disk space the cache uses.
- When [`smith cache verify --fix`](cache-verify.md) cannot repair the store.

To keep recent results and trim the rest use [`smith cache prune`](cache-prune.md). To run
one task without the cache once, `smith run <task> --no-cache` is enough.

## Synopsis

```sh
smith cache clear [--yes] [--json]
```

## Arguments and flags

<!-- smith:begin cache clear -->
Delete the local cache (the remote cache is not touched)

Usage: `smith cache clear [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--json` | boolean | off | print what was deleted as JSON |
| `-y, --yes` | boolean | off | do not ask for confirmation |

<!-- smith:end cache clear -->

- In a terminal smith asks `Delete the local cache at <dir> (<n> entries, <size>)? The remote
  cache is not touched. [y/N]`; only `y` goes ahead, and anything else ends with `clear cancelled`.
- Without a terminal (a script, CI) it refuses with a remedy; pass `--yes` there.
- Only the `v1` folder under the cache directory is deleted. If `[cache] dir` points at a
  folder that holds other files, they stay.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The cache was deleted, or there was none. |
| 1 | Not in a workspace, the question was answered no or could not be asked, or the files could not be deleted. |

## Files and environment

Deletes `<cache dir>/v1`. Variable: `SMITH_CACHE_DIR`. Nothing outside the local cache
directory is read or written, and no network request is made.

## Examples

Without a terminal:

```sh
smith cache clear
```

```text
✖ refusing to delete the cache without a terminal to ask in
  → run `smith cache clear --yes` to delete 3.0 MB at ~/work/acme/.smith/cache
```

```sh
smith cache clear --yes
```

```text
Deleted the local cache at ~/work/acme/.smith/cache: 4 entries and 3 blobs (3.0 MB).
```

When nothing is there:

```text
The cache at ~/work/acme/.smith/cache is already empty.
```

As JSON:

```sh
smith cache clear --yes --json
```

```json
{
  "dir": "~/work/acme/.smith/cache",
  "entries": 5,
  "blobs": 4,
  "bytes": 6298112
}
```

## Common errors

| Message | Cause and remedy |
|---|---|
| `✖ refusing to delete the cache without a terminal to ask in` | Standard input is not a terminal. Pass `--yes`. |
| `clear cancelled` | You answered something other than `y`. Nothing was deleted. |
| `✖ could not delete the cache at <dir>` | A build or an antivirus scanner holds a file. Close other smith processes and run it again. |

## Related

[`smith cache`](cache.md), [`smith cache prune`](cache-prune.md),
[`smith cache status`](cache-status.md), [`smith run`](run.md) (`--no-cache`).
