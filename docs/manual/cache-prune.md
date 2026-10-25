<!-- lore-master
parent: cache.md
-->
# smith cache prune

`smith cache prune` trims the local task cache now: it removes the least recently used
entries until what remains fits the limit, then removes the blobs that no entry names. The
limit is `[cache] max_size_mb`, or the size you pass with `--max-size-mb`. With `--dry-run` it
prints what would go and the size afterwards, and removes nothing.

smith already does this by itself after runs that stored results (see
[`smith cache`](cache.md#how-smith-keeps-the-cache-under-its-limit)), so this command is for
trimming sooner, trimming to a smaller size, or seeing what a smaller limit would cost.

## When to use it

- The cache is over its limit and you do not want to wait for the next check.
- You want room on the disk for something else: `smith cache prune --max-size-mb 256`.
- You are choosing a limit: `smith cache prune --max-size-mb 512 --dry-run` says which tasks
  would lose their cached results.

To delete everything use [`smith cache clear`](cache-clear.md). The remote cache is never pruned.

## Synopsis

```sh
smith cache prune [--max-size-mb N] [--dry-run] [--json]
```

## Arguments and flags

<!-- smith:begin cache prune -->
Trim the local cache to its limit now, removing the least recently used entries

Usage: `smith cache prune [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--dry-run` | boolean | off | show what would be removed and the size afterwards; remove nothing |
| `--json` | boolean | off | print the result as JSON |
| `--max-size-mb <max-size-mb>` | integer (megabytes) | `[cache] max_size_mb` | prune to this size in megabytes instead of [cache] max_size_mb |

<!-- smith:end cache prune -->

- `--max-size-mb` is a whole number of megabytes. `0` removes only the blobs no entry names and
  keeps every entry; a negative number is refused. (A `max_size_mb` of `0` in `smith.toml`
  means no limit, so a plain `smith cache prune` also removes only unused blobs.)
- The size counts the entries, the blobs and the last-run pointers. An entry's blobs count once
  however many entries share them.
- `--dry-run` runs the same selection and changes nothing, so its numbers are what the real
  command would do a moment later.
- The report lists the tasks whose entries go, most entries first. After a real prune, the next
  run of such a task is a cache miss and runs the task again.
- A file another process holds open (on Windows, a build reading a blob) is skipped, not
  forced: the report says how many, and the next prune tries again.
- A blob written in the last minute that no entry names yet is kept, because a task may still
  be saving it.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The cache was pruned, was already within the limit, or does not exist yet; or the dry run printed. |
| 1 | Not in a workspace, a negative `--max-size-mb`, or the store could not be read or written. |

## Files and environment

Deletes files in the cache directory (`ac/`, `cas/`, stale files in `tmp/`) and rewrites
`prune-state.json` beside them, which a real prune updates so automatic checks do not repeat
the work. Variable: `SMITH_CACHE_DIR`.

## Examples

What a smaller limit would cost:

```sh
smith cache prune --max-size-mb 2 --dry-run
```

```text
Would remove 2 entries and 2 blobs (5.0 MB of 6.0 MB); 1.0 MB would be left in 3 entries.
  core:bundle2             1 entry
  core:bundle3             1 entry
```

For real, to the configured limit of 10 MB, on a cache holding 21 MB:

```sh
smith cache prune
```

```text
Pruned the cache: removed 3 entries and 3 blobs (14.0 MB of 21.0 MB); 7.0 MB left in 5 entries.
  core:bundle3             1 entry
  core:bundle5             1 entry
  core:bundle6             1 entry
```

When there is nothing to do:

```sh
smith cache prune
```

```text
Nothing to prune: the cache is 7.0 MB in 5 entries, within its limit of 10.0 MB.
```

As JSON (the dry run above):

```sh
smith cache prune --max-size-mb 2 --dry-run --json
```

```json
{
  "dir": "~/work/acme/.smith/cache",
  "limitBytes": 2097152,
  "dryRun": true,
  "exists": true,
  "entries": 2,
  "blobs": 2,
  "freedBytes": 5245496,
  "sizeBeforeBytes": 6298112,
  "sizeAfterBytes": 1052616,
  "entriesKept": 3,
  "skipped": 0,
  "tasks": {
    "core:bundle2": 1,
    "core:bundle3": 1
  }
}
```

## Common errors

| Message | Cause and remedy |
|---|---|
| `✖ --max-size-mb -3 is not a size` | Pass a number of megabytes, or 0 to remove only unused blobs. `smith cache clear` deletes the whole cache. |
| `✖ could not prune the cache at <dir>: <reason>` with `→ run smith cache verify` | The store could not be read or written. Check it with `smith cache verify`, or close other smith processes and retry. |
| `! 2 files could not be removed (another process has them open)` | A build was reading them. Run the command again when no build is running. |
| `The cache at <dir> is empty: nothing to prune.` | Nothing has been cached yet, or the cache was cleared. |

## Related

[`smith cache`](cache.md), [`smith cache status`](cache-status.md),
[`smith cache clear`](cache-clear.md), [`smith config set`](config-set.md) (`cache.max_size_mb`).
