<!-- lore-master
parent: cache.md
-->
# smith cache clear

`smith cache clear` deletes the local task cache: every entry, blob, last-run pointer and
record under the cache directory's `v1` folder. The next run of each task is a cache miss and
runs the task again. Plain `smith cache clear` never touches the remote cache.

`--remote` deletes this workspace's objects on the remote cache instead: every entry and blob
under the remote's prefix. `--all` deletes the local cache and the remote objects. The team and
CI share the remote cache, so their next runs build again what it held.

In a terminal it asks first, and the question names what goes. Without a terminal it refuses
unless you pass `--yes`. `--dry-run` prints the same question and deletes nothing.

## When to use it

- To measure a cold build, or to see whether a problem is a stale cache.
- To reclaim all the disk space the cache uses.
- When [`smith cache verify --fix`](cache-verify.md) cannot repair the store.
- `--remote`: the remote holds results that should not be replayed (a bad toolchain was cached),
  or you leave a remote you no longer want to pay for.

To keep recent results and trim the rest use [`smith cache prune`](cache-prune.md). To run
one task without the cache once, `smith run <task> --no-cache` is enough.

## Synopsis

```sh
smith cache clear [--remote | --all] [--dry-run] [--yes] [--json]
```

## Arguments and flags

<!-- smith:begin cache clear -->
Delete the local cache; --remote deletes this workspace's remote cache, --all both

Usage: `smith cache clear [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--all` | boolean | off | delete the local cache and this workspace's objects on the remote cache |
| `--dry-run` | boolean | off | print what would be deleted, with the question; delete nothing |
| `--json` | boolean | off | print what was deleted as JSON |
| `--remote` | boolean | off | delete this workspace's objects on the remote cache, not the local cache |
| `-y, --yes` | boolean | off | do not ask for confirmation (needed without a terminal) |

<!-- smith:end cache clear -->

- The local question is `Delete the local cache at <dir> (<n> entries, <size>)? The remote
  cache is not touched. [y/N]`.
- The remote question names the store (host and prefix), the workspace and the count, and says
  that the team and CI share the cache. Only `y` goes ahead; anything else ends with
  `clear cancelled`.
- A plain `https://` store cannot list its objects. smith then deletes only the entries and
  blobs this machine's cache holds, says `up to <n> objects`, and says that the rest stays.
- The token needs list and delete rights. For Azure Blob that is a SAS token with `d` and `l`;
  `smith cache setup` makes tokens with read, write, delete, create and list (`rwdcl`). A token
  made before this change has no delete right: issue a new one. For S3 the key needs
  `s3:ListBucket` and `s3:DeleteObject`.
- A read-only remote (`remote_read_only`, `SMITH_CACHE_READ_ONLY`, a pull-request build) is
  refused.
- Only the `v1` folder under the cache directory is deleted locally. If `[cache] dir` points at a
  folder that holds other files, they stay.

In the interactive shell the `Clear cache` button asks first which cache goes (the local cache,
the remote cache or both) when the workspace has a remote, then shows this same question with
the count. Without a remote it clears the local cache as before.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The cache was deleted, there was none, or the dry run printed. |
| 1 | Not in a workspace, `--remote` with `--all`, the question was answered no or could not be asked, the remote refused or did not answer, or files or objects could not be deleted. |

## Files and environment

Deletes `<cache dir>/v1`, and with `--remote` or `--all` the objects under the remote's prefix.
Variables: `SMITH_CACHE_DIR`, `SMITH_CACHE_REMOTE`, `SMITH_CACHE_KEY`, `SMITH_CACHE_READ_ONLY` and
the store's credentials. Plain `smith cache clear` makes no network request.

## Examples

Without a terminal:

```text
× refusing to delete the cache without a terminal to ask in
  → run `smith cache clear --yes` to delete 3.0 MB at ~/work/acme/.smith/cache
```

```sh
smith cache clear --yes
```

```text
Deleted the local cache at ~/work/acme/.smith/cache: 4 entries and 3 blobs (3.0 MB).
```

The remote cache. These samples ran against a local imitation of S3 (an `httptest` server
that lists, deletes and checks the signature header), not a real bucket:

```sh
smith cache clear --remote --dry-run
```

```text
Reading what s3://smith-cache/acme/main holds for acme…
Delete 9 objects of the remote cache s3://smith-cache/acme/main for the workspace acme?
The team and CI share the remote cache: their next runs build again what it held.
Dry run: nothing was deleted.
```

Without a terminal, then with `--yes`:

```text
× refusing to delete the cache without a terminal to ask in
  → run `smith cache clear --remote --yes` to delete 9 objects of the remote cache s3://smith-cache/acme/main for the workspace acme, which the team and CI share
```

```text
✓ Deleted 9 of 9 objects of the remote cache s3://smith-cache/acme/main for the workspace acme.
```

As JSON (`--remote --yes --json`): `{"remote": {"store": …, "listed": true, "found": 9,
"deleted": 9, "failed": 0}, "dryRun": false}`; `--all` adds `local` with the fields of the local
result. Plain `smith cache clear --json` prints the local result as before:

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
| `× refusing to delete the cache without a terminal to ask in` | Standard input is not a terminal. Pass `--yes`. |
| `clear cancelled` | You answered something other than `y`. Nothing was deleted. |
| `× could not delete the cache at <dir>` | A build or an antivirus scanner holds a file. Close other smith processes and run it again. |
| `× the remote cache <store> refused the credentials (it answered 403 Forbidden)` | The credentials lack list or delete rights. Issue a new SAS token (`smith cache setup`), or give the S3 key `s3:ListBucket` and `s3:DeleteObject`. |
| `× --remote and --all cannot go together` | Pick one. |

## Related

[`smith cache`](cache.md), [`smith cache push`](cache-push.md), [`smith cache prune`](cache-prune.md),
[`smith cache status`](cache-status.md), [`smith run`](run.md) (`--no-cache`).
