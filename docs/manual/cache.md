# smith cache

`smith cache` looks after the local task cache: the folder where smith keeps the results of
tasks that ran, so that the next run can replay them. It shows what the cache holds, trims it
to its size limit, deletes it, and checks it for damage. It acts on the cache on this
machine only; the shared [remote cache](../reference/tasks-and-cache.md#remote-cache) is never
read, pruned or deleted by it.

You rarely need to run it. smith keeps the cache under `[cache] max_size_mb` by itself (see
below); these commands are for looking, for trimming sooner or smaller, and for repairing a
cache after a crash or a full disk.

## When to use it

- To see how big the cache is and which tasks take the room: [`smith cache status`](cache-status.md).
- To make room now, or to a smaller size than the setting: [`smith cache prune`](cache-prune.md).
- To start over, for example before measuring a cold build: [`smith cache clear`](cache-clear.md).
- After a crash, a full disk or a `smith doctor` warning about unusable entries:
  [`smith cache verify`](cache-verify.md).

## Subcommands

<!-- smith:begin cache -->
Look after the local task cache: status, prune, clear, verify

Usage: `smith cache <command>`

| Subcommand | Meaning |
|---|---|
| `smith cache clear` | Delete the local cache (the remote cache is not touched) |
| `smith cache prune` | Trim the local cache to its limit now, removing the least recently used entries |
| `smith cache status` | Show the local cache: where it is, how big, how full, and the largest entries |
| `smith cache verify` | Check the local cache for damage; --fix removes what is broken |

No flags.

<!-- smith:end cache -->

## How smith keeps the cache under its limit

`[cache] max_size_mb` (default 2048) is the size the local cache is kept under. `0` means no
limit, and then smith never prunes by itself. Set it with
[`smith config set cache.max_size_mb 512`](config-set.md).

After a run that stored new results, smith checks the cache. The check is cheap on purpose:

- A small file beside the store (`prune-state.json`) remembers when the store was last
  measured, how big it was, and how much runs have stored since. Most runs only add to that
  total.
- The store is walked, and pruned if it is over the limit, only when it was never measured, when an
  hour has passed since it was, or when a single run stored 4 MB or more and the total says
  the cache is now over the limit. On a store of 4 000 entries the walk takes about 0.2 s and a
  run that does not walk adds about 4 ms.
- Pruning removes the least recently used entries first (a cache hit counts as a use) until the
  rest fit, then the blobs no entry names.
- It never touches what another smith process may be using: an entry used or a blob written in the
  last five minutes stays (so the cache can be over the limit for a few minutes, and smith looks
  again after six minutes), a file it cannot delete is skipped, and only one smith prunes at a time.
- A problem while pruning is a warning on standard error and never a failed run:

  ```text
  smith: warning: could not prune the cache to its limit: <reason>
    → run `smith cache prune` to try again, or `smith cache verify` to check the store
  ```

It does this in CI too: the local cache is on that machine like any other. A run in a pull
request that may not write to the remote cache still prunes its own local one.

## Where the cache is

The directory is `[cache] dir` (default `.smith/cache`, relative to the workspace root), or
`SMITH_CACHE_DIR` when that is set. Inside it, smith's files are in a `v1` folder; the commands
here touch nothing else in the directory. See
[The cache key, hits and misses](../architecture/cache.md#the-local-store) for the layout.

## Exit codes

0 for the help; see the subcommands.

## Files and environment

`smith.toml` (`[cache] dir`, `max_size_mb`, `remote`, `remote_read_only`), the cache directory.
Variables: `SMITH_CACHE_DIR`, `SMITH_CACHE_REMOTE`, `SMITH_CACHE_KEY`, `SMITH_CACHE_READ_ONLY`
(read to describe the remote). The commands find the workspace like every other command, so
`SMITH_WORKSPACE_ROOT` applies; they start no plugins and do not need git.

## Examples

A scratch workspace with a Go lib `core` and six tasks that each write a few megabytes
(`bundle1` writes 1 MB, `bundle2` 2 MB, and so on), with the limit set to 10 MB:

```sh
smith config set cache.max_size_mb 10
smith build
smith run bundle1 bundle2 bundle3 bundle4 bundle5 bundle6
smith run bundle1 bundle2 bundle3 bundle4 bundle5 bundle6
```

```text
✔ 6/6 tasks succeeded (0 from cache) in 378ms
✔ 6/6 tasks succeeded (6 from cache, 2.035s saved) in 11ms
```

The first run stored 21 MB. Everything it stored is less than five minutes old, so smith left
it; the cache is over its limit until a later run, or:

```sh
smith cache status
smith cache prune --dry-run
smith cache prune
smith cache status
```

```text
local cache  ~/work/acme/.smith/cache
  entries     8
  blobs       7
  size        21.0 MB of 10.0 MB (210.1% full)
  last prune  2026-10-25 16:49
  remote      none
  ✖ the cache is over its limit
    → smith cache prune brings it under 10.0 MB
largest entries
  core:bundle6                 6.0 MB  1 entry
  core:bundle5                 5.0 MB  1 entry
  core:bundle4                 4.0 MB  1 entry
  core:bundle3                 3.0 MB  1 entry
  core:bundle2                 2.0 MB  1 entry
Would remove 3 entries and 3 blobs (14.0 MB of 21.0 MB); 7.0 MB would be left in 5 entries.
  core:bundle3             1 entry
  core:bundle5             1 entry
  core:bundle6             1 entry
Pruned the cache: removed 3 entries and 3 blobs (14.0 MB of 21.0 MB); 7.0 MB left in 5 entries.
  core:bundle3             1 entry
  core:bundle5             1 entry
  core:bundle6             1 entry
local cache  ~/work/acme/.smith/cache
  entries     5
  blobs       4
  size        7.0 MB of 10.0 MB (70.1% full)
  last prune  2026-10-25 16:50
  remote      none
largest entries
  core:bundle4                 4.0 MB  1 entry
  core:bundle2                 2.0 MB  1 entry
  core:bundle1                 1.0 MB  1 entry
  core:build                      0 B  1 entry
  util:build                      0 B  1 entry
```

The entries that went were the ones used longest ago. Checking, damaging and repairing the
store, and deleting it, are on the pages of [`verify`](cache-verify.md) and
[`clear`](cache-clear.md).

[`smith info`](info.md) shows the size against the limit on one line, and
[`smith doctor`](doctor.md) warns when the cache is over its limit or holds entries that can
never be a hit:

```text
  cache     7 MB of 10 MB (70%)
```

## Common errors

| Message | Cause and remedy |
|---|---|
| `smith: warning: could not prune the cache to its limit` | A run stored results but the prune failed (a read-only folder, a full disk, a store being deleted). The run itself succeeded. Run `smith cache prune` to see the error, or `smith cache verify`. |
| The cache stays over its limit after a run | What it holds is less than five minutes old, or the limit is below what one run stores. `smith cache prune` trims at once. |
| `✖ there is no smith workspace here` | The commands read `smith.toml` for the cache directory and limit. Run them inside a workspace. |

## Related

[`smith why-miss`](why-miss.md), [`smith doctor`](doctor.md), [`smith info`](info.md),
[`smith config`](config.md), [The cache key, hits and misses](../architecture/cache.md),
[Tasks and the cache](../reference/tasks-and-cache.md).
